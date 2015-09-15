package status_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/locket/shared"
	"github.com/cloudfoundry-incubator/locket/status"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/clock/fakeclock"
)

var _ = Describe("Cell Service Registry", func() {
	const retryInterval = time.Second
	var (
		clock *fakeclock.FakeClock

		presenceStatus     *status.PresenceStatus
		presence1          ifrit.Process
		presence2          ifrit.Process
		firstCellPresence  models.CellPresence
		secondCellPresence models.CellPresence
	)

	BeforeEach(func() {
		clock = fakeclock.NewFakeClock(time.Now())
		presenceStatus = status.NewPresenceStatus(consulSession, clock, lagertest.NewTestLogger("test"))

		firstCellPresence = models.NewCellPresence("first-rep", "1.2.3.4", "the-zone", models.NewCellCapacity(128, 1024, 3), []string{}, []string{})
		secondCellPresence = models.NewCellPresence("second-rep", "4.5.6.7", "the-zone", models.NewCellCapacity(128, 1024, 3), []string{}, []string{})

		presence1 = nil
		presence2 = nil
	})

	AfterEach(func() {
		if presence1 != nil {
			presence1.Signal(os.Interrupt)
			Eventually(presence1.Wait()).Should(Receive(BeNil()))
		}

		if presence2 != nil {
			presence2.Signal(os.Interrupt)
			Eventually(presence2.Wait()).Should(Receive(BeNil()))
		}
	})

	Describe("MaintainCellPresence", func() {
		BeforeEach(func() {
			presence1 = ifrit.Invoke(presenceStatus.NewCellPresence(firstCellPresence, retryInterval))
		})

		It("should put /cell/CELL_ID in the store", func() {
			expectedJSON, err := models.ToJSON(firstCellPresence)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() []byte {
				value, _ := consulSession.GetAcquiredValue(shared.CellSchemaPath(firstCellPresence.CellID))
				return value
			}, time.Second).Should(MatchJSON(expectedJSON))
		})
	})

	Describe("CellById", func() {
		Context("when the cell exists", func() {
			BeforeEach(func() {
				presence1 = ifrit.Invoke(presenceStatus.NewCellPresence(firstCellPresence, retryInterval))
			})

			It("returns the correct CellPresence", func() {
				Eventually(func() (models.CellPresence, error) {
					return presenceStatus.CellById(firstCellPresence.CellID)
				}).Should(Equal(firstCellPresence))
			})
		})

		Context("when the cell does not exist", func() {
			It("returns ErrStoreResourceNotFound", func() {
				_, err := presenceStatus.CellById(firstCellPresence.CellID)
				Expect(err).To(Equal(shared.ErrStoreResourceNotFound))
			})
		})
	})

	Describe("Cells", func() {
		Context("when there are available Cells", func() {
			BeforeEach(func() {
				presence1 = ifrit.Invoke(presenceStatus.NewCellPresence(firstCellPresence, retryInterval))
				presence2 = ifrit.Invoke(presenceStatus.NewCellPresence(secondCellPresence, retryInterval))
			})

			It("should get from /v1/cell/", func() {
				Eventually(func() ([]models.CellPresence, error) {
					return presenceStatus.Cells()
				}).Should(ConsistOf(firstCellPresence, secondCellPresence))
			})

			Context("when there is unparsable JSON in there...", func() {
				BeforeEach(func() {
					err := consulSession.AcquireLock(shared.CellSchemaPath("blah"), []byte("ß"))
					Expect(err).NotTo(HaveOccurred())

					Eventually(func() map[string][]byte {
						cells, err := consulSession.ListAcquiredValues(shared.CellSchemaRoot)
						Expect(err).NotTo(HaveOccurred())
						return cells
					}, 1, 50*time.Millisecond).Should(HaveLen(3))
				})

				It("should ignore the unparsable JSON and move on", func() {
					cellPresences, err := presenceStatus.Cells()
					Expect(err).NotTo(HaveOccurred())
					Expect(cellPresences).To(HaveLen(2))
					Expect(cellPresences).To(ContainElement(firstCellPresence))
					Expect(cellPresences).To(ContainElement(secondCellPresence))
				})
			})
		})

		Context("when there are none", func() {
			It("should return empty", func() {
				reps, err := presenceStatus.Cells()
				Expect(err).NotTo(HaveOccurred())
				Expect(reps).To(BeEmpty())
			})
		})
	})

	Describe("CellEvents", func() {
		var receivedEvents <-chan status.CellEvent
		var otherSession *consuladapter.Session

		setPresences := func() {
			presence1 = ifrit.Invoke(presenceStatus.NewCellPresence(firstCellPresence, retryInterval))

			Eventually(func() ([]models.CellPresence, error) {
				return presenceStatus.Cells()
			}).Should(HaveLen(1))
		}

		BeforeEach(func() {
			otherSession = consulRunner.NewSession("other-session")
			otherPresenceStatus := status.NewPresenceStatus(otherSession, clock, lagertest.NewTestLogger("test"))
			receivedEvents = otherPresenceStatus.CellEvents()
		})

		Context("when the store is up", func() {
			Context("when cells are present and then one disappears", func() {
				BeforeEach(func() {
					otherSession = consulRunner.NewSession("other-session")
					otherPresenceStatus := status.NewPresenceStatus(otherSession, clock, lagertest.NewTestLogger("test"))
					receivedEvents = otherPresenceStatus.CellEvents()

					setPresences()
					ginkgomon.Interrupt(presence1)

					Eventually(func() ([]models.CellPresence, error) {
						return presenceStatus.Cells()
					}).Should(HaveLen(0))
				})

				AfterEach(func() {
					otherSession.Destroy()
				})

				It("receives a CellDisappeared event", func() {
					Eventually(receivedEvents).Should(Receive(Equal(
						status.CellDisappearedEvent{IDs: []string{firstCellPresence.CellID}},
					)))
				})
			})
		})

		Context("when the store is down", func() {
			BeforeEach(func() {
				otherSession = consulRunner.NewSession("other-session")
				otherPresenceStatus := status.NewPresenceStatus(otherSession, clock, lagertest.NewTestLogger("test"))
				receivedEvents = otherPresenceStatus.CellEvents()

				consulRunner.Stop()
			})

			It("attaches when the store is back", func() {
				consulRunner.Start()
				consulRunner.WaitUntilReady()

				By("setting presences")
				newSession, err := consulSession.Recreate()
				Expect(err).NotTo(HaveOccurred())
				presenceStatus = status.NewPresenceStatus(newSession, clock, lagertest.NewTestLogger("test"))

				setPresences()

				Eventually(func() ([]models.CellPresence, error) {
					return presenceStatus.Cells()
				}).Should(HaveLen(1))

				time.Sleep(2 * time.Second) //wait for a watch fail cycle

				By("stopping the presence")
				ginkgomon.Interrupt(presence1)

				Eventually(func() ([]models.CellPresence, error) {
					return presenceStatus.Cells()
				}).Should(HaveLen(0))

				Eventually(receivedEvents).Should(Receive(Equal(
					status.CellDisappearedEvent{IDs: []string{firstCellPresence.CellID}},
				)))
			})
		})
	})
})
