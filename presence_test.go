package locket_test

import (
	"time"

	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/locket"
	"github.com/hashicorp/consul/api"

	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Presence", func() {
	var (
		presenceKey   string
		presenceValue []byte

		consulClient consuladapter.Client

		presenceRunner  ifrit.Runner
		presenceProcess ifrit.Process
		retryInterval   time.Duration
		logger          lager.Logger
	)

	getPresenceValue := func() ([]byte, error) {
		kvPair, _, err := consulClient.KV().Get(presenceKey, nil)
		if err != nil {
			return nil, err
		}

		if kvPair == nil || kvPair.Session == "" {
			return nil, consuladapter.NewKeyNotFoundError(presenceKey)
		}

		return kvPair.Value, nil
	}

	BeforeEach(func() {
		consulClient = consulRunner.NewConsulClient()

		presenceKey = "some-key"
		presenceValue = []byte("some-value")

		retryInterval = 500 * time.Millisecond
		logger = lagertest.NewTestLogger("locket")
	})

	JustBeforeEach(func() {
		clock := clock.NewClock()
		presenceRunner = locket.NewPresence(logger, consulClient, presenceKey, presenceValue, clock, retryInterval, 5*time.Second)
	})

	AfterEach(func() {
		ginkgomon.Kill(presenceProcess)
	})

	Context("When consul is running", func() {
		Context("an error occurs while acquiring the presence", func() {
			BeforeEach(func() {
				presenceKey = ""
			})

			It("continues to retry", func() {
				presenceProcess = ifrit.Background(presenceRunner)

				Eventually(presenceProcess.Ready()).Should(BeClosed())
				Consistently(presenceProcess.Wait()).ShouldNot(Receive())

				Eventually(logger).Should(Say("failed-setting-presence"))
				Eventually(logger).Should(Say("recreating-session"))
			})
		})

		Context("and the presence is available", func() {
			It("acquires the presence", func() {
				presenceProcess = ifrit.Background(presenceRunner)
				Eventually(presenceProcess.Ready()).Should(BeClosed())
				Eventually(getPresenceValue).Should(Equal(presenceValue))
			})

			Context("and we have acquired the presence", func() {
				JustBeforeEach(func() {
					presenceProcess = ifrit.Background(presenceRunner)
					Eventually(presenceProcess.Ready()).Should(BeClosed())
				})

				Context("when consul shuts down", func() {
					JustBeforeEach(func() {
						consulRunner.Stop()
					})

					AfterEach(func() {
						consulRunner.Start()
						consulRunner.WaitUntilReady()
					})

					It("loses the presence and retries", func() {
						Eventually(presenceProcess.Wait()).ShouldNot(Receive())
						Eventually(logger).Should(Say("recreating-session"))
					})
				})

				Context("and the process is shutting down", func() {
					It("releases the presence and exits", func() {
						ginkgomon.Interrupt(presenceProcess)
						Eventually(presenceProcess.Wait()).Should(Receive(BeNil()))
						_, err := getPresenceValue()
						Expect(err).To(Equal(consuladapter.NewKeyNotFoundError(presenceKey)))
					})
				})
			})
		})

		Context("and the presence is unavailable", func() {
			var (
				otherSession   *consuladapter.Session
				otherValue     []byte
				otherSessionID string
			)

			BeforeEach(func() {
				otherValue = []byte("doppel-value")
				otherSession = consulRunner.NewSession("other-session")

				_, err := otherSession.SetPresence(presenceKey, otherValue)
				Expect(err).NotTo(HaveOccurred())
				Expect(getPresenceValue()).To(Equal(otherValue))
				otherSessionID = otherSession.ID()
			})

			AfterEach(func() {
				otherSession.Destroy()
			})

			It("waits for the presence to become available", func() {
				presenceProcess = ifrit.Background(presenceRunner)
				Eventually(presenceProcess.Ready()).Should(BeClosed())
				Expect(getPresenceValue()).To(Equal(otherValue))
			})

			Context("when consul shuts down", func() {
				JustBeforeEach(func() {
					presenceProcess = ifrit.Background(presenceRunner)
					Eventually(presenceProcess.Ready()).Should(BeClosed())

					consulRunner.Stop()
				})

				AfterEach(func() {
					consulRunner.Start()
					consulRunner.WaitUntilReady()
				})

				It("continues to wait for the presence", func() {
					Consistently(presenceProcess.Ready()).Should(BeClosed())
					Consistently(presenceProcess.Wait()).ShouldNot(Receive())

					Eventually(logger).Should(Say("failed-setting-presence"))
					Eventually(logger).Should(Say("recreating-session"))
				})
			})

			Context("and the session is destroyed", func() {
				XIt("should recreate the session and continue to retry", func() {
					var err error
					presenceProcess = ifrit.Background(presenceRunner)
					Eventually(presenceProcess.Ready()).Should(BeClosed())

					var sessions []*api.SessionEntry
					Eventually(func() int {
						sessions, _, err = consulClient.Session().List(nil)
						Expect(err).NotTo(HaveOccurred())
						return len(sessions)
					}).Should(Equal(2))

					var originalSessionID string
					for _, session := range sessions {
						if session.ID != otherSessionID {
							originalSessionID = session.ID
							break
						}
					}

					_, err = consulClient.Session().Destroy(originalSessionID, nil)
					Expect(err).NotTo(HaveOccurred())

					Eventually(logger, 6*time.Second).Should(Say("consul-error"))
					Eventually(logger).Should(Say("recreating-session"))

					Eventually(func() int {
						sessions, _, err = consulClient.Session().List(nil)
						Expect(err).NotTo(HaveOccurred())
						return len(sessions)
					}).Should(Equal(2))

					var newSessionID string
					for _, session := range sessions {
						if session.ID != otherSessionID {
							newSessionID = session.ID
							break
						}
					}

					Expect(newSessionID).NotTo(Equal(originalSessionID))
				})
			})

			Context("and the process is shutting down", func() {
				It("exits", func() {
					presenceProcess = ifrit.Background(presenceRunner)
					Eventually(presenceProcess.Ready()).Should(BeClosed())

					ginkgomon.Interrupt(presenceProcess)
					Eventually(presenceProcess.Wait()).Should(Receive(BeNil()))
				})
			})

			Context("and the presence is released", func() {
				It("acquires the presence", func() {
					presenceProcess = ifrit.Background(presenceRunner)
					Eventually(presenceProcess.Ready()).Should(BeClosed())
					Expect(getPresenceValue()).To(Equal(otherValue))

					otherSession.Destroy()

					Eventually(getPresenceValue).Should(Equal(presenceValue))
				})
			})
		})
	})

	Context("When consul is down", func() {
		BeforeEach(func() {
			consulRunner.Stop()
		})

		AfterEach(func() {
			consulRunner.Start()
			consulRunner.WaitUntilReady()
		})

		It("continues to retry creating the session", func() {
			presenceProcess = ifrit.Background(presenceRunner)

			Eventually(presenceProcess.Ready()).Should(BeClosed())
			Consistently(presenceProcess.Wait()).ShouldNot(Receive())

			Eventually(logger).Should(Say("failed-setting-presence"))
			Eventually(logger).Should(Say("recreating-session"))
			Eventually(logger).Should(Say("recreating-session"))
		})

		Context("when consul starts up", func() {
			It("acquires the presence", func() {
				presenceProcess = ifrit.Background(presenceRunner)
				Eventually(presenceProcess.Ready()).Should(BeClosed())

				Eventually(logger).Should(Say("failed-setting-presence"))
				Eventually(logger).Should(Say("recreating-session"))
				Consistently(presenceProcess.Wait()).ShouldNot(Receive())

				consulRunner.Start()
				consulRunner.WaitUntilReady()

				Eventually(getPresenceValue).Should(Equal(presenceValue))
			})
		})
	})
})
