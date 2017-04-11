package db_test

import (
	"database/sql"
	"errors"
	"fmt"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func validateLockInDB(rawDB *sql.DB, res *models.Resource, expectedIndex, expectedTTL int64) error {
	var key, owner, value, lockType string
	var index, ttl int64

	lockQuery := helpers.RebindForFlavor(
		"SELECT path, owner, value, type, modified_index, ttl FROM locks WHERE path = ?",
		dbFlavor,
	)

	row := rawDB.QueryRow(lockQuery, res.Key)
	Expect(row.Scan(&key, &owner, &value, &lockType, &index, &ttl)).To(Succeed())
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
	if res.Type != lockType {
		errMsg += fmt.Sprintf("mismatch value (%s, %s),", res.Type, lockType)
	}
	if expectedIndex != index {
		errMsg += fmt.Sprintf("mismatch index (%d, %d),", expectedIndex, index)
	}
	if expectedTTL != ttl {
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
			Type:  "fooooobar",
		}

		emptyResource = &models.Resource{Key: "quack"}
	})

	Context("Lock", func() {
		Context("when the lock does not exist", func() {
			Context("because the row does not exist", func() {
				It("inserts the lock for the owner", func() {
					lock, err := sqlDB.Lock(logger, resource, 10)
					Expect(err).NotTo(HaveOccurred())
					Expect(lock).To(Equal(&db.Lock{
						Resource:      resource,
						ModifiedIndex: 1,
						TtlInSeconds:  10,
					}))
					Expect(validateLockInDB(rawDB, resource, 1, 10)).To(Succeed())
				})
			})

			Context("because the contents of the lock are empty", func() {
				BeforeEach(func() {
					query := helpers.RebindForFlavor(
						`INSERT INTO locks (path, owner, value, modified_index) VALUES (?, ?, ?, ?);`,
						dbFlavor,
					)
					result, err := rawDB.Exec(query, resource.Key, "", "", 300)
					Expect(err).NotTo(HaveOccurred())
					Expect(result.RowsAffected()).To(BeEquivalentTo(1))
				})

				It("inserts the lock for the owner", func() {
					lock, err := sqlDB.Lock(logger, resource, 10)
					Expect(err).NotTo(HaveOccurred())
					Expect(lock).To(Equal(&db.Lock{
						Resource:      resource,
						ModifiedIndex: 301,
						TtlInSeconds:  10,
					}))
					Expect(validateLockInDB(rawDB, resource, 301, 10)).To(Succeed())
				})
			})
		})

		Context("when the lock does exist", func() {
			BeforeEach(func() {
				_, err := sqlDB.Lock(logger, resource, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(validateLockInDB(rawDB, resource, 1, 10)).To(Succeed())
			})

			Context("and the desired owner is different", func() {
				It("returns an error without grabbing the lock", func() {
					newResource := &models.Resource{
						Key:   "quack",
						Owner: "jim",
						Value: "i have never seen the princess bride and never will",
					}

					_, err := sqlDB.Lock(logger, newResource, 10)
					Expect(err).To(Equal(models.ErrLockCollision))
					Expect(validateLockInDB(rawDB, resource, 1, 10)).To(Succeed())
				})
			})

			Context("and the desired owner is the same", func() {
				It("increases the modified_index", func() {
					lock, err := sqlDB.Lock(logger, resource, 10)
					Expect(err).NotTo(HaveOccurred())
					Expect(lock).To(Equal(&db.Lock{
						Resource:      resource,
						ModifiedIndex: 2,
						TtlInSeconds:  10,
					}))
					Expect(validateLockInDB(rawDB, resource, 2, 10)).To(Succeed())
				})
			})
		})
	})

	Context("Release", func() {
		Context("when the lock exists", func() {
			var currentIndex, currentTTL int64

			BeforeEach(func() {
				currentIndex = 500
				currentTTL = 501
				query := helpers.RebindForFlavor(
					`INSERT INTO locks (path, owner, value, modified_index, ttl) VALUES (?, ?, ?, ?, ?);`,
					dbFlavor,
				)
				result, err := rawDB.Exec(query, resource.Key, resource.Owner, resource.Value, currentIndex, currentTTL)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RowsAffected()).To(BeEquivalentTo(1))
			})

			It("empties out the lock from the lock table", func() {
				err := sqlDB.Release(logger, resource)
				Expect(err).NotTo(HaveOccurred())
				Expect(validateLockInDB(rawDB, emptyResource, 501, currentTTL)).To(Succeed())
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
				Type:  "FOR DAYS",
			}
		})

		Context("when the lock exists", func() {
			BeforeEach(func() {
				query := helpers.RebindForFlavor(
					`INSERT INTO locks (path, owner, value, type, modified_index, ttl) VALUES (?, ?, ?, ?, ?, ?);`,
					dbFlavor,
				)
				result, err := rawDB.Exec(query, lock.Key, lock.Owner, lock.Value, lock.Type, 434, 5)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RowsAffected()).To(BeEquivalentTo(1))
			})

			It("returns the lock from the database", func() {
				resource, err := sqlDB.Fetch(logger, "test")
				Expect(err).NotTo(HaveOccurred())
				Expect(resource).To(Equal(&db.Lock{
					Resource:      lock,
					ModifiedIndex: 434,
					TtlInSeconds:  5,
				}))
			})
		})

		Context("when the lock does not exist", func() {
			Context("because the row does not exist", func() {
				It("returns an resource not found error", func() {
					_, err := sqlDB.Fetch(logger, "test")
					Expect(err).To(Equal(models.ErrResourceNotFound))
				})
			})

			Context("because the contents of the lock are empty", func() {
				BeforeEach(func() {
					query := helpers.RebindForFlavor(
						`INSERT INTO locks (path, owner, value) VALUES (?, ?, ?);`,
						dbFlavor,
					)
					result, err := rawDB.Exec(query, "test", "", "")
					Expect(err).NotTo(HaveOccurred())
					Expect(result.RowsAffected()).To(BeEquivalentTo(1))
				})

				It("returns an error", func() {
					_, err := sqlDB.Fetch(logger, "test")
					Expect(err).To(Equal(models.ErrResourceNotFound))
				})
			})
		})
	})

	Context("FetchAll", func() {
		var dogLock, humanLock *db.Lock

		BeforeEach(func() {
			query := helpers.RebindForFlavor(
				`INSERT INTO locks (path, owner, value, type, modified_index, ttl) VALUES (?, ?, ?, ?, ?, ?);`,
				dbFlavor,
			)
			result, err := rawDB.Exec(query, "test1", "jake", "thedog", "dog", 10, 20)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RowsAffected()).To(BeEquivalentTo(1))

			result, err = rawDB.Exec(query, "test2", "", "", "", 10, 20)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RowsAffected()).To(BeEquivalentTo(1))

			result, err = rawDB.Exec(query, "test3", "finn", "thehuman", "human", 10, 20)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RowsAffected()).To(BeEquivalentTo(1))

			dogLock = &db.Lock{
				Resource: &models.Resource{
					Key:   "test1",
					Owner: "jake",
					Value: "thedog",
					Type:  "dog",
				},
				ModifiedIndex: 10,
				TtlInSeconds:  20,
			}
			humanLock = &db.Lock{
				Resource: &models.Resource{
					Key:   "test3",
					Owner: "finn",
					Value: "thehuman",
					Type:  "human",
				},
				ModifiedIndex: 10,
				TtlInSeconds:  20,
			}
		})

		It("retrieves a list of all locks with owners", func() {
			locks, err := sqlDB.FetchAll(logger, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(locks).To(ConsistOf(dogLock, humanLock))
		})

		Context("when a type is specified", func() {
			It("filters the locks returned by that type", func() {
				locks, err := sqlDB.FetchAll(logger, "human")
				Expect(err).NotTo(HaveOccurred())
				Expect(locks).To(ConsistOf(humanLock))
			})
		})

		Context("when fetching the locks fail", func() {
			BeforeEach(func() {
				_, err := rawDB.Exec("DROP TABLE locks")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := sqlDB.CreateLockTable(logger)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := sqlDB.FetchAll(logger, "")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("Count", func() {
		BeforeEach(func() {
			query := helpers.RebindForFlavor(
				`INSERT INTO locks (path, owner, value, type, modified_index, ttl) VALUES (?, ?, ?, ?, ?, ?);`,
				dbFlavor,
			)
			result, err := rawDB.Exec(query, "test1", "jake", "thedog", "dog", 10, 20)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RowsAffected()).To(BeEquivalentTo(1))

			result, err = rawDB.Exec(query, "test2", "", "", "", 10, 20)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RowsAffected()).To(BeEquivalentTo(1))

			result, err = rawDB.Exec(query, "test3", "finn", "thehuman", "human", 10, 20)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RowsAffected()).To(BeEquivalentTo(1))
		})

		It("retrieves a count of the locks", func() {
			count, err := sqlDB.Count(logger, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(2))
		})

		It("filters based on lock type", func() {
			count, err := sqlDB.Count(logger, "dog")
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(1))
		})

		Context("when the database errors", func() {
			BeforeEach(func() {
				_, err := rawDB.Exec("DROP TABLE locks")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := sqlDB.CreateLockTable(logger)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := sqlDB.Count(logger, "")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
