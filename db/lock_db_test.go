package db_test

import (
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lock", func() {
	Context("Lock", func() {
		Context("when the lock does not exist", func() {
			It("inserts the lock and grabs it", func() {
				newLock := models.Resource{
					Key:   "quack",
					Owner: "iamthelizardking",
					Value: "i can do anything",
				}

				err := sqlDB.Lock(logger, &newLock)
				Expect(err).NotTo(HaveOccurred())

				lockQuery := helpers.RebindForFlavor(
					"SELECT * FROM Locks WHERE key = ?",
					dbFlavor,
				)

				var key, owner, value string
				row := db.QueryRow(lockQuery, newLock.Key)
				Expect(row.Scan(&key, &owner, &value)).To(Succeed())
				Expect(key).To(Equal(newLock.Key))
				Expect(owner).To(Equal(newLock.Owner))
				Expect(value).To(Equal(newLock.Value))
			})
		})

		Context("when the lock does exist", func() {
			It("returns an error without grabbing the lock", func() {
				oldLock := models.Resource{
					Key:   "quack",
					Owner: "iamthelizardking",
					Value: "i can do anything",
				}

				err := sqlDB.Lock(logger, &oldLock)
				Expect(err).NotTo(HaveOccurred())

				newLock := models.Resource{
					Key:   "quack",
					Owner: "jim",
					Value: "i have never seen the princess bride and never will",
				}

				err = sqlDB.Lock(logger, &newLock)
				Expect(err).To(Equal(models.ErrLockCollision))
			})
		})
	})

	Context("Release", func() {
		var lock models.Resource

		BeforeEach(func() {
			lock = models.Resource{
				Key:   "test",
				Owner: "jim",
				Value: "locks stuff for days",
			}
		})

		Context("when the lock exists", func() {
			BeforeEach(func() {
				query := helpers.RebindForFlavor(
					`INSERT INTO locks (key, owner, value) VALUES (?, ?, ?);`,
					dbFlavor,
				)
				result, err := db.Exec(query, lock.Key, lock.Owner, lock.Value)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RowsAffected()).To(BeEquivalentTo(1))
			})

			It("removes the lock from the lock table", func() {
				err := sqlDB.Release(logger, &lock)
				Expect(err).NotTo(HaveOccurred())

				rows, err := db.Query(`SELECT key FROM locks;`)
				Expect(err).NotTo(HaveOccurred())
				Expect(rows.Next()).To(BeFalse())
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
				err := sqlDB.Release(logger, &lock)
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
					`INSERT INTO locks (key, owner, value) VALUES (?, ?, ?);`,
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
			It("returns the lock from the database", func() {
				_, err := sqlDB.Fetch(logger, "test")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
