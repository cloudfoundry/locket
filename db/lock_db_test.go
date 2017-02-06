package db_test

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func validateLockInDB(db *sql.DB, res *models.Resource, expectedIndex int64) error {
	var key, owner, value string
	var index int64

	lockQuery := helpers.RebindForFlavor(
		"SELECT * FROM locks WHERE path = ?",
		dbFlavor,
	)

	row := db.QueryRow(lockQuery, res.Key)
	Expect(row.Scan(&key, &owner, &value, &index)).To(Succeed())
	errMsg := ""
	if res.Key != key {
		errMsg += fmt.Sprintf("mismatch key (%s, %s),", res.Key, key)
	}
	if res.Owner != owner {
		errMsg += fmt.Sprintf("mismatch owner (%s, %s),", res.Owner, owner)
	}
	if res.Value != value {
		errMsg += fmt.Sprintf("mismatch value (%s, %s),", res.Value, value)
	}
	if expectedIndex != index {
		errMsg += fmt.Sprintf("mismatch index (%d, %d),", expectedIndex, index)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

var _ = Describe("Lock", func() {
	var resource, emptyResource *models.Resource

	BeforeEach(func() {
		resource = &models.Resource{
			Key:   "quack",
			Owner: "iamthelizardking",
			Value: "i can do anything",
		}

		emptyResource = &models.Resource{Key: "quack"}
	})

	Context("Lock", func() {
		Context("when the lock does not exist", func() {
			Context("because the row does not exist", func() {
				It("inserts the lock for the owner", func() {
					err := sqlDB.Lock(logger, resource, 10)
					Expect(err).NotTo(HaveOccurred())
					Expect(validateLockInDB(db, resource, 1)).To(Succeed())
				})

				It("expires the lock after the ttl", func() {
					err := sqlDB.Lock(logger, resource, 10)
					Expect(err).NotTo(HaveOccurred())
					Expect(validateLockInDB(db, resource, 1)).To(Succeed())

					fakeClock.Increment(9 * time.Second)
					Consistently(func() error {
						return validateLockInDB(db, resource, 1)
					}).ShouldNot(HaveOccurred())

					fakeClock.Increment(2 * time.Second)
					Eventually(func() error {
						return validateLockInDB(db, emptyResource, 2)
					}).ShouldNot(HaveOccurred())
				})
			})

			Context("because the contents of the lock are empty", func() {
				BeforeEach(func() {
					query := helpers.RebindForFlavor(
						`INSERT INTO locks (path, owner, value, modified_index) VALUES (?, ?, ?, ?);`,
						dbFlavor,
					)
					result, err := db.Exec(query, resource.Key, "", "", 300)
					Expect(err).NotTo(HaveOccurred())
					Expect(result.RowsAffected()).To(BeEquivalentTo(1))
				})

				It("inserts the lock for the owner", func() {
					err := sqlDB.Lock(logger, resource, 10)
					Expect(err).NotTo(HaveOccurred())
					Expect(validateLockInDB(db, resource, 301)).To(Succeed())
				})
			})
		})

		Context("when the lock does exist", func() {
			BeforeEach(func() {
				err := sqlDB.Lock(logger, resource, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(validateLockInDB(db, resource, 1)).To(Succeed())
			})

			Context("and the desired owner is different", func() {
				It("returns an error without grabbing the lock", func() {
					newResource := &models.Resource{
						Key:   "quack",
						Owner: "jim",
						Value: "i have never seen the princess bride and never will",
					}

					err := sqlDB.Lock(logger, newResource, 10)
					Expect(err).To(Equal(models.ErrLockCollision))
					Expect(validateLockInDB(db, resource, 1)).To(Succeed())
				})
			})

			Context("and the desired owner is the same", func() {
				It("increases the modified_index", func() {
					err := sqlDB.Lock(logger, resource, 10)
					Expect(err).NotTo(HaveOccurred())
					Expect(validateLockInDB(db, resource, 2)).To(Succeed())
				})
			})
		})
	})

	Context("Release", func() {
		Context("when the lock exists", func() {
			var currentIndex int64

			BeforeEach(func() {
				currentIndex = 500
				query := helpers.RebindForFlavor(
					`INSERT INTO locks (path, owner, value, modified_index) VALUES (?, ?, ?, ?);`,
					dbFlavor,
				)
				result, err := db.Exec(query, resource.Key, resource.Owner, resource.Value, currentIndex)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RowsAffected()).To(BeEquivalentTo(1))
			})

			It("empties out the lock from the lock table", func() {
				err := sqlDB.Release(logger, resource)
				Expect(err).NotTo(HaveOccurred())
				Expect(validateLockInDB(db, emptyResource, 501)).To(Succeed())
			})

			Context("when the lock is owned by another owner", func() {
				It("returns an error", func() {
					err := sqlDB.Release(logger, &models.Resource{
						Key:   "test",
						Owner: "not jim",
						Value: "beep boop",
					})
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when the lock does not exist", func() {
			It("returns an error", func() {
				err := sqlDB.Release(logger, resource)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("Fetch", func() {
		var lock *models.Resource

		BeforeEach(func() {
			lock = &models.Resource{
				Key:   "test",
				Owner: "jim",
				Value: "locks stuff for days",
			}
		})

		Context("when the lock exists", func() {
			BeforeEach(func() {
				query := helpers.RebindForFlavor(
					`INSERT INTO locks (path, owner, value) VALUES (?, ?, ?);`,
					dbFlavor,
				)
				result, err := db.Exec(query, lock.Key, lock.Owner, lock.Value)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RowsAffected()).To(BeEquivalentTo(1))
			})

			It("returns the lock from the database", func() {
				resource, err := sqlDB.Fetch(logger, "test")
				Expect(err).NotTo(HaveOccurred())
				Expect(resource).To(Equal(lock))
			})
		})

		Context("when the lock does not exist", func() {
			Context("because the row does not exist", func() {
				It("returns an error", func() {
					_, err := sqlDB.Fetch(logger, "test")
					Expect(err).To(HaveOccurred())
				})
			})

			Context("because the contents of the lock are empty", func() {
				BeforeEach(func() {
					query := helpers.RebindForFlavor(
						`INSERT INTO locks (path, owner, value) VALUES (?, ?, ?);`,
						dbFlavor,
					)
					result, err := db.Exec(query, "test", "", "")
					Expect(err).NotTo(HaveOccurred())
					Expect(result.RowsAffected()).To(BeEquivalentTo(1))
				})

				It("returns an error", func() {
					_, err := sqlDB.Fetch(logger, "test")
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
