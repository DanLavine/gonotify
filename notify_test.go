package gonotify_test

import (
	"fmt"
	"testing"

	"github.com/DanLavine/gonotify"
	. "github.com/onsi/gomega"
)

func TestAdd(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("when Stop has already been called", func(t *testing.T) {
		notify := gonotify.New()
		err := notify.Add()
		g.Expect(err).ToNot(HaveOccurred())

		notify.Stop()
		err = notify.Add()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Notify has been stopped already"))
	})

	t.Run("when ForceStop has already been called", func(t *testing.T) {
		notify := gonotify.New()
		err := notify.Add()
		g.Expect(err).ToNot(HaveOccurred())

		notify.ForceStop()
		err = notify.Add()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Notify has been stopped already"))
	})
}

func TestRemove(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it performs a no-op if the notifier is empty", func(t *testing.T) {
		notify := gonotify.New()

		ready := notify.Ready()
		g.Consistently(ready).ShouldNot(Receive())

		notify.Remove()

		g.Eventually(ready).ShouldNot(Receive())
		notify.Stop()
	})

	t.Run("it removes a ready counter", func(t *testing.T) {
		notify := gonotify.New()

		ready := notify.Ready()
		g.Consistently(ready).ShouldNot(Receive())

		g.Expect(notify.Add()).ToNot(HaveOccurred())
		notify.Remove()

		g.Eventually(ready).ShouldNot(Receive())
		notify.Stop()
	})
}

func TestReady(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("is notified of a message when Add is called", func(t *testing.T) {
		notify := gonotify.New()

		ready := notify.Ready()
		g.Consistently(ready).ShouldNot(Receive())

		// add a counter
		g.Expect(notify.Add()).ToNot(HaveOccurred())

		g.Eventually(ready).Should(Receive())
		notify.Stop()
	})

	t.Run("ready chan receives for each call of Add", func(t *testing.T) {
		notify := gonotify.New()
		ready := notify.Ready()

		// add a counter
		g.Expect(notify.Add()).ToNot(HaveOccurred())
		g.Expect(notify.Add()).ToNot(HaveOccurred())
		g.Expect(notify.Add()).ToNot(HaveOccurred())
		g.Expect(notify.Add()).ToNot(HaveOccurred())

		g.Eventually(ready).Should(Receive())
		g.Eventually(ready).Should(Receive())
		g.Eventually(ready).Should(Receive())
		g.Eventually(ready).Should(Receive())
		notify.Stop()
	})

	t.Run("ready chan drains for each call of Add on a Stop", func(t *testing.T) {
		notify := gonotify.New()
		ready := notify.Ready()

		// add a counter
		g.Expect(notify.Add()).ToNot(HaveOccurred())
		g.Expect(notify.Add()).ToNot(HaveOccurred())
		g.Expect(notify.Add()).ToNot(HaveOccurred())
		g.Expect(notify.Add()).ToNot(HaveOccurred())

		notify.Stop()

		g.Eventually(ready).Should(Receive(Equal(&struct{}{})))
		g.Eventually(ready).Should(Receive(Equal(&struct{}{})))
		g.Eventually(ready).Should(Receive(Equal(&struct{}{})))
		g.Eventually(ready).Should(Receive(Equal(&struct{}{})))
	})

	t.Run("ready chan returns nil on a ForceStop", func(t *testing.T) {
		notify := gonotify.New()
		ready := notify.Ready()

		// add a counter
		g.Expect(notify.Add()).ToNot(HaveOccurred())
		g.Expect(notify.Add()).ToNot(HaveOccurred())
		g.Expect(notify.Add()).ToNot(HaveOccurred())
		g.Expect(notify.Add()).ToNot(HaveOccurred())

		notify.ForceStop()

		g.Eventually(ready).ShouldNot(Receive())
	})

	t.Run("it runs all common commands asynchrnously", func(t *testing.T) {
		notify := gonotify.New()
		ready := notify.Ready()

		// set 100 notifications
		addErrChan := make(chan error)
		for i := 0; i < 100; i++ {
			go func() {
				addErrChan <- notify.Add()
			}()
		}

		// accept up to 100 notifications
		go func() {
			for i := 0; i < 75; i++ {
				go func() {
					<-ready
				}()
			}
		}()

		// drop up to 25 notifications
		go func() {
			for i := 0; i < 25; i++ {
				go func() {
					notify.Remove()
				}()
			}
		}()

		for i := 0; i < 100; i++ {
			g.Eventually(addErrChan).Should(Receive(BeNil()))
		}
	})

	t.Run("it can properly drains all operations asynchronously", func(t *testing.T) {
		notify := gonotify.New()
		ready := notify.Ready()

		// set 100 notifications
		addErrChan := make(chan error)
		for i := 0; i < 100; i++ {
			go func() {
				addErrChan <- notify.Add()
			}()
		}

		// accept up to 100 notifications
		go func() {
			for i := 0; i < 75; i++ {
				go func() {
					<-ready
				}()
			}
		}()

		// drop up to 25 notifications
		go func() {
			for i := 0; i < 25; i++ {
				go func() {
					notify.Remove()
				}()
			}
		}()

		notify.Stop()
		for i := 0; i < 100; i++ {
			g.Eventually(addErrChan).Should(Receive(Or(BeNil(), Equal(fmt.Errorf("Notify has been stopped already")))))
		}

		g.Eventually(ready).Should(BeClosed())
	})

	t.Run("it can immediately stops all operations asynchronously", func(t *testing.T) {
		notify := gonotify.New()
		ready := notify.Ready()

		// set 100 notifications
		addErrChan := make(chan error)
		for i := 0; i < 100; i++ {
			go func() {
				addErrChan <- notify.Add()
			}()
		}

		// accept up to 100 notifications
		go func() {
			for i := 0; i < 75; i++ {
				go func() {
					<-ready
				}()
			}
		}()

		// drop up to 25 notifications
		go func() {
			for i := 0; i < 25; i++ {
				go func() {
					notify.Remove()
				}()
			}
		}()

		notify.ForceStop()
		for i := 0; i < 100; i++ {
			g.Eventually(addErrChan).Should(Receive(Or(BeNil(), Equal(fmt.Errorf("Notify has been stopped already")))))
		}

		g.Eventually(ready).Should(BeClosed())
	})
}
