package systemstop

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Signaler - interface to get stop signal from system or main goroutine
type Signaler interface {
	Signal() <-chan struct{}
	Done()
}

// signalimpl - sync.WaitGroup with system stop notifications
type signalimpl struct {
	wg   *sync.WaitGroup
	done chan struct{}
	term chan os.Signal
}

var instance *signalimpl

// initialize WG
func init() {
	instance = &signalimpl{}
	instance.wg = &sync.WaitGroup{}
	instance.done = make(chan struct{})
	instance.term = make(chan os.Signal, 1)
	signal.Notify(instance.term, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-instance.term
		close(instance.done)
	}()
}

// Subscribe - return interface to get system stop signal
func Subscribe() Signaler {
	instance.wg.Add(1)
	return instance
}

// StopAll - send stop event
func StopAll() {
	close(instance.term)
}

// Wait - wait stop all goroutines
func Wait() {
	instance.wg.Wait()
}

// Signal - get channel with stop event
func (s *signalimpl) Signal() <-chan struct{} {
	return s.done
}

// Done - event of goroutine ends
func (s *signalimpl) Done() {
	s.wg.Done()
}
