package expiration_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager"
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

			It("does not release the lock", func() {
				lockPick.RegisterTTL(logger, lock)

				fakeClock.WaitForWatcherAndIncrement(ttl)

				Eventually(fakeLockDB.FetchCallCount).Should(Equal(1))
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
			})
		})

		Context("when there is already a check process running", func() {
			BeforeEach(func() {
				lockPick.RegisterTTL(logger, lock)
				Eventually(fakeClock.WatcherCount).Should(Equal(1))
			})

			It("does not add a new check process", func() {
				lockPick.RegisterTTL(logger, lock)

				Eventually(fakeClock.WatcherCount).Should(Equal(2))
				Consistently(fakeClock.WatcherCount).Should(Equal(2))
				fakeClock.WaitForWatcherAndIncrement(ttl)

				Eventually(fakeLockDB.FetchCallCount).Should(Equal(1))
				_, key := fakeLockDB.FetchArgsForCall(0)
				Expect(key).To(Equal(lock.Key))

				Eventually(logger).Should(gbytes.Say("cancelling-old-check"))
				Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(1))
				Consistently(fakeLockDB.ReleaseCallCount).Should(Equal(1))
			})

			Context("and competes with a newer lock on checking expiry", func() {
				var newLock db.Lock
				BeforeEach(func() {
					newLock = *lock
					newLock.ModifiedIndex += 1

					trigger := true
					fakeLockDB.FetchStub = func(logger lager.Logger, key string) (*db.Lock, error) {
						if trigger {
							// second expiry goroutine
							lockPick.RegisterTTL(logger, &newLock)
						}
						trigger = false
						return &newLock, nil
					}
				})

				It("checks the expiration of the lock", func() {
					// first expiry goroutine proceeds into timer case statement
					fakeClock.WaitForWatcherAndIncrement(ttl)
					Eventually(fakeLockDB.FetchCallCount).Should(Equal(1))

					// third expiry goroutine, cancels the second expiry goroutine
					lockPick.RegisterTTL(logger, &newLock)

					Eventually(fakeClock.WatcherCount).Should(Equal(2))
					fakeClock.WaitForWatcherAndIncrement(ttl)

					Eventually(fakeLockDB.FetchCallCount).Should(Equal(2))
					Consistently(fakeLockDB.FetchCallCount).Should(Equal(2))

					Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(1))
					_, resource := fakeLockDB.ReleaseArgsForCall(0)
					Expect(resource).To(Equal(newLock.Resource))
				})
			})

			Context("when an older lock is registered", func() {
				var oldLock db.Lock

				BeforeEach(func() {
					oldLock = *lock
					oldLock.ModifiedIndex -= 1
				})

				It("does nothing", func() {
					l := oldLock
					lockPick.RegisterTTL(logger, &l)
					Eventually(logger).Should(gbytes.Say("found-newer-expiration-goroutine"))
				})

				Context("and the previous lock has already expired", func() {
					BeforeEach(func() {
						fakeClock.WaitForWatcherAndIncrement(ttl)
						Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(1))
					})

					It("checks the expiration of the lock", func() {
						l := oldLock
						lockPick.RegisterTTL(logger, &l)
						Eventually(fakeClock.WatcherCount).Should(Equal(1))
						fakeClock.WaitForWatcherAndIncrement(ttl)

						Eventually(fakeLockDB.FetchCallCount).Should(Equal(2))
						_, key := fakeLockDB.FetchArgsForCall(0)
						Expect(key).To(Equal(l.Key))
					})
				})
			})

			Context("when another lock is registered", func() {
				var newLock *db.Lock
				BeforeEach(func() {
					newLock = &db.Lock{
						Resource: &models.Resource{
							Key:   "another",
							Owner: "myself",
							Value: "hi",
						},
						TtlInSeconds:  lock.TtlInSeconds,
						ModifiedIndex: 9,
					}

					fakeLockDB.FetchStub = func(logger lager.Logger, key string) (*db.Lock, error) {
						switch {
						case key == lock.Key:
							return lock, nil
						case key == newLock.Key:
							return newLock, nil
						default:
							return nil, errors.New("unknown lock")
						}
					}
				})

				It("does not effect the other check goroutines", func() {
					lockPick.RegisterTTL(logger, newLock)

					Eventually(fakeClock.WatcherCount).Should(Equal(2))
					Consistently(fakeClock.WatcherCount).Should(Equal(2))

					lockPick.RegisterTTL(logger, lock)

					Eventually(fakeClock.WatcherCount).Should(Equal(3))
					fakeClock.WaitForWatcherAndIncrement(ttl)

					Eventually(fakeLockDB.FetchCallCount).Should(Equal(2))
					_, key1 := fakeLockDB.FetchArgsForCall(0)
					_, key2 := fakeLockDB.FetchArgsForCall(1)
					Expect([]string{key1, key2}).To(ContainElement(lock.Key))
					Expect([]string{key1, key2}).To(ContainElement(newLock.Key))

					Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(2))
					Consistently(fakeLockDB.ReleaseCallCount).Should(Equal(2))
				})
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
