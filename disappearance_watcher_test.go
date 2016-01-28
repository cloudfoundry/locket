package locket_test

import (
	"time"

	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/locket"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Disappearance Watcher", func() {
	const retryInterval = 500 * time.Millisecond

	var (
		consulClient   consuladapter.Client
		watcherRunner  ifrit.Runner
		watcherProcess ifrit.Process

		disappearChan <-chan []string

		logger lager.Logger
	)

	BeforeEach(func() {
		consulClient = consulRunner.NewConsulClient()
		logger = lagertest.NewTestLogger("test")
		clock := clock.NewClock()
		watcherRunner, disappearChan = locket.NewDisappearanceWatcher(logger, consulClient, "under", clock)
		watcherProcess = ifrit.Invoke(watcherRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(watcherProcess)
	})

	var addThenRemovePresence = func(presenceKey string) {
		presenceRunner := locket.NewPresence(logger, consulClient, presenceKey, []byte("value"), clock.NewClock(), retryInterval, 10*time.Second)

		presenceProcess := ifrit.Invoke(presenceRunner)
		Eventually(func() int {
			sessions, _, err := consulClient.Session().List(nil)
			Expect(err).NotTo(HaveOccurred())
			return len(sessions)
		}).Should(Equal(1))

		ginkgomon.Kill(presenceProcess)
	}

	Context("when the watch starts first", func() {
		Context("when there are keys", func() {
			It("detects removals of keys", func() {
				addThenRemovePresence("under/here")

				Eventually(disappearChan, 10*time.Second).Should(Receive(Equal([]string{"under/here"})))
			})

			Context("with other prefixes", func() {
				It("does not detect removal of keys under other prefixes", func() {
					addThenRemovePresence("other")

					Consistently(disappearChan).ShouldNot(Receive())
				})
			})

			Context("when destroying the session", func() {
				It("closes the disappearance channel", func() {
					// session.Destroy()

					Eventually(disappearChan, 15).Should(BeClosed())
				})
			})

			Context("when an error occurs", func() {
				It("retries", func() {
					consulRunner.Stop()

					Consistently(disappearChan).ShouldNot(Receive())

					consulRunner.Start()
					consulRunner.WaitUntilReady()

					time.Sleep(1 * time.Second) // allow the watch to retry

					addThenRemovePresence("under/here")

					Eventually(disappearChan).Should(Receive(Equal([]string{"under/here"})))
				})
			})
		})
	})

	Context("when the watch starts later", func() {
		It("detects removals of keys", func() {
			presenceRunner := locket.NewPresence(logger, consulClient, "under/here", []byte("value"), clock.NewClock(), retryInterval, 10*time.Second)
			presenceProcess := ifrit.Invoke(presenceRunner)

			time.Sleep(1 * time.Second) // allow the watch to retry
			ginkgomon.Kill(presenceProcess)

			Eventually(disappearChan).Should(Receive(Equal([]string{"under/here"})))
		})
	})
})
