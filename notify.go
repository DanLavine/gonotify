package gontify

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Notify struct {
	// if Noify has been canceled
	forceDoneOnce *sync.Once
	forceDone     chan struct{}

	// if Noify has been canceled
	doneOnce *sync.Once
	done     chan struct{}

	// counter to know how many items are waiting
	readyCount   *atomic.Int32
	readyTrigger chan struct{}

	// Can be used to signal that there is a new item to read
	ready chan *struct{}
}

func New() *Notify {
	notify := &Notify{
		forceDoneOnce: new(sync.Once),
		forceDone:     make(chan struct{}),

		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		readyCount: new(atomic.Int32),

		// a buffer of 1 should allow all Add() funcs to not block
		readyTrigger: make(chan struct{}, 1),

		// Notifies a caller that something should be ready to read
		ready: make(chan *struct{}),
	}

	go notify.run()

	return notify
}

func (n *Notify) haveReady() {
	currentCount := n.readyCount.Add(-1)

	// If there is a number of items in the queue, this will re-trigger itslef.
	// Otherwise the next time Add() is called, will trigger this loop from the 'run()' <-readTrigger again
	if currentCount > 0 {
		select {
		case n.readyTrigger <- struct{}{}:
			// enqueu another 'run()' loop
		default:
			// something is already in the queue, so we will hit this again on the next 'run()' trigger
		}
	}

	select {
	case n.ready <- &struct{}{}:
		// block on ready
	case <-n.forceDone:
		// on force done we are not gracefully waiting anymore. Time to give up
	}
}

func (n *Notify) run() {
	for {
		select {
		case <-n.done:
			select {
			case <-n.forceDone:
				// close everything and return
				close(n.readyTrigger)
				close(n.ready)
				return
			default:
				// Ensure all messages have been drained
				if n.readyCount.Load() == 0 {
					n.ForceStop() // trigger a close on the force chan just to be safe
					close(n.readyTrigger)
					close(n.ready)
					return
				}

				// message to drain
				n.haveReady()
			}
		case <-n.readyTrigger:
			n.haveReady()
		}
	}
}

// Add a counter to notify and trigger a ready call
func (n *Notify) Add() error {
	select {
	case <-n.done:
		return fmt.Errorf("Notify has been Canceled")
	case <-n.forceDone:
		return fmt.Errorf("Notify has been Canceled")
	default:
		n.readyCount.Add(1)

		select {
		case n.readyTrigger <- struct{}{}:
			// trigger that there is a item that is ready
		default:
			// in this case, we already know something is ready and fall through.
			// Don't wan't to spin up a lot of goroutines to know that there is
			// data available. That is handled by 'run()' directly
		}

		return nil
	}
}

// Used to know if there is a message ready. If this is "nil", then all messages have been
// drained and the Notifyer has been closed. No more messages should be sent on the shared data
// structure this is protecting
func (n *Notify) Ready() <-chan *struct{} {
	return n.ready
}

// Close is the graceful shutdown mechanism for our notification process. This will allow all currently
// enqued counters to be notfied by the Read() chan.
func (n *Notify) Close() {
	n.doneOnce.Do(func() {
		close(n.done)
	})
}

// ForceStop is the destructive shutdown that does not allow for Ready() to be fully drained.
func (n *Notify) ForceStop() {
	n.forceDoneOnce.Do(func() {
		close(n.forceDone)
		close(n.done)
	})
}
