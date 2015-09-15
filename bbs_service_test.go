package locket_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"

	"github.com/cloudfoundry-incubator/locket"
	"github.com/cloudfoundry-incubator/locket/shared"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/clock/fakeclock"
)

var _ = Describe("BBS Presence", func() {
	var clock *fakeclock.FakeClock
	var locketClient *locket.Locket
	var logger *lagertest.TestLogger

	BeforeEach(func() {
		clock = fakeclock.NewFakeClock(time.Now())
		logger = lagertest.NewTestLogger("test")
		locketClient = locket.New(consulSession, clock, logger)
	})

	Describe("BBSMasterURL", func() {
		Context("when able to get a master bbs presence", func() {
			var heartbeater ifrit.Process
			var bbsPresence models.BBSPresence

			JustBeforeEach(func() {
				locketClient := locket.New(consulSession, clock, logger)
				bbsLock, err := locketClient.NewBBSMasterLock(bbsPresence, 100*time.Millisecond)
				Expect(err).NotTo(HaveOccurred())
				heartbeater = ifrit.Invoke(bbsLock)
			})

			AfterEach(func() {
				heartbeater.Signal(os.Interrupt)
				Eventually(heartbeater.Wait()).Should(Receive(BeNil()))
			})

			Context("when the master bbs URL is present", func() {
				BeforeEach(func() {
					bbsPresence = models.NewBBSPresence("a-bbs-id", "https://database-z1-0.database.consul.cf.internal:7085")
				})

				It("returns the URL", func() {
					url, err := locketClient.BBSMasterURL()
					Expect(err).NotTo(HaveOccurred())
					Expect(url).To(Equal(bbsPresence.URL))
				})
			})
		})

		Context("when unable to get any bbs presences", func() {
			It("returns ErrServiceUnavailable", func() {
				_, err := locketClient.BBSMasterURL()
				Expect(err).To(Equal(shared.ErrServiceUnavailable))
			})
		})
	})
})
