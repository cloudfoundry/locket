package handlers_test

import (
	"context"
	"errors"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/db/dbfakes"
	"code.cloudfoundry.org/locket/expiration/expirationfakes"
	"code.cloudfoundry.org/locket/handlers"
	"code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Lock", func() {
	var (
		fakeLockDB    *dbfakes.FakeLockDB
		fakeLockPick  *expirationfakes.FakeLockPick
		logger        *lagertest.TestLogger
		locketHandler models.LocketServer
		resource      *models.Resource
		exitCh        chan struct{}
	)

	BeforeEach(func() {
		fakeLockDB = &dbfakes.FakeLockDB{}
		fakeLockPick = &expirationfakes.FakeLockPick{}
		logger = lagertest.NewTestLogger("locket-handler")
		exitCh = make(chan struct{}, 1)

		resource = &models.Resource{
			Key:   "test",
			Value: "test-value",
			Owner: "myself",
		}

		locketHandler = handlers.NewLocketHandler(logger, fakeLockDB, fakeLockPick, exitCh)
	})

	Context("Lock", func() {
		var (
			request      *models.LockRequest
			expectedLock *db.Lock
		)

		BeforeEach(func() {
			request = &models.LockRequest{
				Resource:     resource,
				TtlInSeconds: 10,
			}

			expectedLock = &db.Lock{
				Resource:      resource,
				TtlInSeconds:  10,
				ModifiedIndex: 2,
			}

			fakeLockDB.LockReturns(expectedLock, nil)
		})

		It("reserves the lock in the database", func() {
			_, err := locketHandler.Lock(context.Background(), request)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLockDB.LockCallCount()).Should(Equal(1))
			_, actualResource, ttl := fakeLockDB.LockArgsForCall(0)
			Expect(actualResource).To(Equal(resource))
			Expect(ttl).To(BeEquivalentTo(10))
		})

		It("registers the lock and ttl with the lock pick", func() {
			_, err := locketHandler.Lock(context.Background(), request)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLockPick.RegisterTTLCallCount()).To(Equal(1))
			_, lock := fakeLockPick.RegisterTTLArgsForCall(0)
			Expect(lock).To(Equal(expectedLock))
		})

		Context("when request does not have TTL", func() {
			BeforeEach(func() {
				request = &models.LockRequest{
					Resource: resource,
				}
			})

			It("returns a validation error", func() {
				_, err := locketHandler.Lock(context.Background(), request)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrInvalidTTL))
			})
		})

		Context("when the request does not have an owner", func() {
			BeforeEach(func() {
				resource.Owner = ""
			})

			It("returns a validation error", func() {
				_, err := locketHandler.Lock(context.Background(), request)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrInvalidOwner))
			})
		})

		Context("when locking errors", func() {
			var (
				err error
			)

			BeforeEach(func() {
				fakeLockDB.LockReturns(nil, errors.New("Boom."))
			})

			JustBeforeEach(func() {
				_, err = locketHandler.Lock(context.Background(), request)
			})

			It("returns the error", func() {
				Expect(err).To(HaveOccurred())
			})

			It("logs the error", func() {
				Expect(logger).To(gbytes.Say("Boom."))
			})

			Context("when lock collision error occurs", func() {
				BeforeEach(func() {
					fakeLockDB.LockReturns(nil, models.ErrLockCollision)
				})

				It("does not log the error", func() {
					Expect(logger).NotTo(gbytes.Say("lock-collision"))
				})
			})
		})

		Context("when an unrecoverable error is returned", func() {
			BeforeEach(func() {
				fakeLockDB.LockReturns(nil, helpers.ErrUnrecoverableError)
			})

			It("logs and writes to the exit channel", func() {
				locketHandler.Lock(context.Background(), request)
				Expect(logger).To(gbytes.Say("unrecoverable-error"))
				Expect(exitCh).To(Receive())
			})
		})
	})

	Context("Release", func() {
		It("releases the lock in the database", func() {
			_, err := locketHandler.Release(context.Background(), &models.ReleaseRequest{Resource: resource})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLockDB.ReleaseCallCount()).Should(Equal(1))
			_, actualResource := fakeLockDB.ReleaseArgsForCall(0)
			Expect(actualResource).To(Equal(resource))
		})

		Context("when releasing errors", func() {
			BeforeEach(func() {
				fakeLockDB.ReleaseReturns(errors.New("Boom."))
			})

			It("returns the error", func() {
				_, err := locketHandler.Release(context.Background(), &models.ReleaseRequest{Resource: resource})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when an unrecoverable error is returned", func() {
			BeforeEach(func() {
				fakeLockDB.ReleaseReturns(helpers.ErrUnrecoverableError)
			})

			It("logs and writes to the exit channel", func() {
				locketHandler.Release(context.Background(), &models.ReleaseRequest{Resource: resource})
				Expect(logger).To(gbytes.Say("unrecoverable-error"))
				Expect(exitCh).To(Receive())
			})
		})
	})

	Context("Fetch", func() {
		BeforeEach(func() {
			fakeLockDB.FetchReturns(&db.Lock{Resource: resource}, nil)
		})

		It("fetches the lock in the database", func() {
			fetchResp, err := locketHandler.Fetch(context.Background(), &models.FetchRequest{Key: "test-fetch"})
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchResp.Resource).To(Equal(resource))

			Expect(fakeLockDB.FetchCallCount()).Should(Equal(1))
			_, key := fakeLockDB.FetchArgsForCall(0)
			Expect(key).To(Equal("test-fetch"))
		})

		Context("when fetching errors", func() {
			BeforeEach(func() {
				fakeLockDB.FetchReturns(nil, errors.New("boom"))
			})

			It("returns the error", func() {
				_, err := locketHandler.Fetch(context.Background(), &models.FetchRequest{Key: "test-fetch"})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when an unrecoverable error is returned", func() {
			BeforeEach(func() {
				fakeLockDB.FetchReturns(nil, helpers.ErrUnrecoverableError)
			})

			It("logs and writes to the exit channel", func() {
				locketHandler.Fetch(context.Background(), &models.FetchRequest{Key: "test-fetch"})
				Expect(logger).To(gbytes.Say("unrecoverable-error"))
				Expect(exitCh).To(Receive())
			})
		})
	})

	Context("FetchAll", func() {
		var expectedResources []*models.Resource
		BeforeEach(func() {
			expectedResources = []*models.Resource{
				resource,
				&models.Resource{Key: "cell", Owner: "cell-1", Value: "{}"},
			}

			var locks []*db.Lock
			for _, r := range expectedResources {
				locks = append(locks, &db.Lock{Resource: r})
			}
			fakeLockDB.FetchAllReturns(locks, nil)
		})

		It("fetches all the locks in the database", func() {
			fetchResp, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{Type: "dawg"})
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchResp.Resources).To(Equal(expectedResources))
			Expect(fakeLockDB.FetchAllCallCount()).Should(Equal(1))
			_, lockType := fakeLockDB.FetchAllArgsForCall(0)
			Expect(lockType).To(Equal("dawg"))
		})

		Context("when fetching errors", func() {
			BeforeEach(func() {
				fakeLockDB.FetchAllReturns(nil, errors.New("boom"))
			})

			It("returns the error", func() {
				_, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when an unrecoverable error is returned", func() {
			BeforeEach(func() {
				fakeLockDB.FetchAllReturns(nil, helpers.ErrUnrecoverableError)
			})

			It("logs and writes to the exit channel", func() {
				locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{})
				Expect(logger).To(gbytes.Say("unrecoverable-error"))
				Expect(exitCh).To(Receive())
			})
		})
	})
})
