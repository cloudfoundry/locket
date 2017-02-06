package handlers_test

import (
	"context"
	"errors"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/locket/db/dbfakes"
	"code.cloudfoundry.org/locket/handlers"
	"code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lock", func() {
	var (
		fakeLockDB    *dbfakes.FakeLockDB
		logger        *lagertest.TestLogger
		locketHandler models.LocketServer
		resource      *models.Resource
	)

	BeforeEach(func() {
		fakeLockDB = &dbfakes.FakeLockDB{}
		logger = lagertest.NewTestLogger("locket-handler")

		resource = &models.Resource{
			Key:   "test",
			Value: "test-value",
			Owner: "myself",
		}

		locketHandler = handlers.NewLocketHandler(logger, fakeLockDB)
	})

	Context("Lock", func() {
		var request *models.LockRequest

		BeforeEach(func() {
			request = &models.LockRequest{
				Resource:     resource,
				TtlInSeconds: 10,
			}
		})

		It("reserves the lock in the database", func() {
			_, err := locketHandler.Lock(context.Background(), request)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLockDB.LockCallCount()).Should(Equal(1))
			_, actualResource, ttl := fakeLockDB.LockArgsForCall(0)
			Expect(actualResource).To(Equal(resource))
			Expect(ttl).To(BeEquivalentTo(10))
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

		Context("when locking errors", func() {
			BeforeEach(func() {
				fakeLockDB.LockReturns(errors.New("Boom."))
			})

			It("returns the error", func() {
				_, err := locketHandler.Lock(context.Background(), request)
				Expect(err).To(HaveOccurred())
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
	})

	Context("Fetch", func() {
		BeforeEach(func() {
			fakeLockDB.FetchReturns(resource, nil)
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
	})
})
