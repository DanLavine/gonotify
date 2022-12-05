package gonotify_test

import (
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
}
