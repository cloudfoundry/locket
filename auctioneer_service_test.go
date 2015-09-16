package locket_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"

	"github.com/cloudfoundry-incubator/locket"
	"github.com/cloudfoundry-incubator/locket/presence"
	"github.com/pivotal-golang/clock/fakeclock"
)

var _ = Describe("Receptor Service Registry", func() {
	var clock *fakeclock.FakeClock
	var locketClient locket.Client
	var logger *lagertest.TestLogger

	BeforeEach(func() {
		clock = fakeclock.NewFakeClock(time.Now())
		logger = lagertest.NewTestLogger("test")
		locketClient = locket.NewClient(consulSession, clock, logger)
	})

	Describe("AuctioneerAddress", func() {
		Context("when able to get an auctioneer presence", func() {
			var heartbeater ifrit.Process
			var auctioneerPresence presence.AuctioneerPresence

			JustBeforeEach(func() {
				locketClient := locket.NewClient(consulSession, clock, logger)
				auctioneerLock, err := locketClient.NewAuctioneerLock(auctioneerPresence, 100*time.Millisecond)
				Expect(err).NotTo(HaveOccurred())
				heartbeater = ifrit.Invoke(auctioneerLock)
			})

			AfterEach(func() {
				heartbeater.Signal(os.Interrupt)
				Eventually(heartbeater.Wait()).Should(Receive(BeNil()))
			})

			Context("when the auctionner address is present", func() {
				BeforeEach(func() {
					auctioneerPresence = presence.NewAuctioneerPresence("auctioneer-id", "auctioneer.example.com")
				})

				It("returns the address", func() {
					address, err := locketClient.AuctioneerAddress()
					Expect(err).NotTo(HaveOccurred())
					Expect(address).To(Equal(auctioneerPresence.AuctioneerAddress))
				})
			})
		})

		Context("when unable to get any auctioneer presences", func() {
			It("returns ErrServiceUnavailable", func() {
				_, err := locketClient.AuctioneerAddress()
				Expect(err).To(Equal(locket.ErrServiceUnavailable))
			})
		})
	})
})
