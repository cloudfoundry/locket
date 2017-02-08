package expiration_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/db/dbfakes"
	"code.cloudfoundry.org/locket/expiration"
	"code.cloudfoundry.org/locket/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("LockPick", func() {
	var (
		lockPick expiration.LockPick

		logger     *lagertest.TestLogger
		fakeLockDB *dbfakes.FakeLockDB
		fakeClock  *fakeclock.FakeClock

		ttl time.Duration

		lock *db.Lock
	)

	BeforeEach(func() {
		lock = &db.Lock{
			Resource: &models.Resource{
				Key:   "funky",
				Owner: "town",
				Value: "won't you take me to",
			},
			TtlInSeconds:  25,
			ModifiedIndex: 6,
		}

		ttl = time.Duration(lock.TtlInSeconds) * time.Second

		fakeClock = fakeclock.NewFakeClock(time.Now())
		logger = lagertest.NewTestLogger("lock-pick")
		fakeLockDB = &dbfakes.FakeLockDB{}

		lockPick = expiration.NewLockPick(fakeLockDB, fakeClock)
	})

	Context("RegisterTTL", func() {
		BeforeEach(func() {
			fakeLockDB.FetchReturns(lock, nil)
		})

		It("checks that the lock expires after the ttl", func() {
			lockPick.RegisterTTL(logger, lock)

			fakeClock.WaitForWatcherAndIncrement(ttl)

			Eventually(fakeLockDB.FetchCallCount).Should(Equal(1))
			_, key := fakeLockDB.FetchArgsForCall(0)
			Expect(key).To(Equal(lock.Key))

			Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(1))
			_, resource := fakeLockDB.ReleaseArgsForCall(0)
			Expect(resource).To(Equal(lock.Resource))
		})

		Context("when the modified index has been incremented", func() {
			var returnedLock *db.Lock
			BeforeEach(func() {
				returnedLock = &db.Lock{
					Resource: &models.Resource{
						Key:   "funky",
						Owner: "town",
						Value: "won't you take me to",
					},
					TtlInSeconds:  25,
					ModifiedIndex: 7,
				}

				fakeLockDB.FetchReturns(returnedLock, nil)
			})

			It("does not release the lock", func() {
				lockPick.RegisterTTL(logger, lock)

				fakeClock.WaitForWatcherAndIncrement(ttl)

				Eventually(fakeLockDB.FetchCallCount).Should(Equal(1))
				_, key := fakeLockDB.FetchArgsForCall(0)
				Expect(key).To(Equal(lock.Key))

				Consistently(fakeLockDB.ReleaseCallCount).Should(Equal(0))
			})
		})

		Context("when fetching the lock fails", func() {
			BeforeEach(func() {
				fakeLockDB.FetchReturns(nil, errors.New("failed-to-fetch-lock"))
			})

			It("logs the error and does not release the lock", func() {
				lockPick.RegisterTTL(logger, lock)

				fakeClock.WaitForWatcherAndIncrement(ttl)

				Eventually(fakeLockDB.FetchCallCount).Should(Equal(1))

				Eventually(logger).Should(gbytes.Say("failed-to-fetch-lock"))
				Consistently(fakeLockDB.ReleaseCallCount).Should(Equal(0))
			})
		})

		Context("when releasing the lock fails", func() {
			BeforeEach(func() {
				fakeLockDB.ReleaseReturns(errors.New("failed-to-release-lock"))
			})

			It("logs the error", func() {
				lockPick.RegisterTTL(logger, lock)

				fakeClock.WaitForWatcherAndIncrement(ttl)

				Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(1))
				Eventually(logger).Should(gbytes.Say("failed-to-release-lock"))
			})
		})

		Context("when there is already a check process running", func() {
			BeforeEach(func() {
				lockPick.RegisterTTL(logger, lock)
			})

			It("does not add a new check process", func() {
				lockPick.RegisterTTL(logger, lock)

				Eventually(fakeClock.WatcherCount).Should(Equal(1))
				Consistently(fakeClock.WatcherCount).Should(Equal(1))
				fakeClock.WaitForWatcherAndIncrement(ttl)

				Eventually(fakeLockDB.FetchCallCount).Should(Equal(1))
				_, key := fakeLockDB.FetchArgsForCall(0)
				Expect(key).To(Equal(lock.Key))

				Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(1))
				Consistently(fakeLockDB.ReleaseCallCount).Should(Equal(1))
			})

			Context("and the check process finishes", func() {
				BeforeEach(func() {
					fakeClock.WaitForWatcherAndIncrement(ttl)

					Eventually(fakeLockDB.FetchCallCount).Should(Equal(1))
					_, key := fakeLockDB.FetchArgsForCall(0)
					Expect(key).To(Equal(lock.Key))

					Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(1))
					Consistently(fakeLockDB.ReleaseCallCount).Should(Equal(1))
				})

				It("performs the expiration check", func() {
					lockPick.RegisterTTL(logger, lock)

					Eventually(fakeClock.WatcherCount).Should(Equal(1))
					fakeClock.WaitForWatcherAndIncrement(ttl)

					Eventually(fakeLockDB.FetchCallCount).Should(Equal(2))
					_, key := fakeLockDB.FetchArgsForCall(0)
					Expect(key).To(Equal(lock.Key))

					Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(2))
					Consistently(fakeLockDB.ReleaseCallCount).Should(Equal(2))
				})
			})
		})
	})
})
