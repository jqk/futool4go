package timeutils

import (
	"sync"
	"time"
)

type Stopwatch struct {
	isRunning   bool
	startTime   time.Time
	elapsedTime time.Duration
	lock        sync.RWMutex
}

func (s *Stopwatch) IsRunning() bool {
	return s.isRunning
}

func (s *Stopwatch) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.isRunning {
		s.isRunning = true
		s.startTime = time.Now()
	}
}

func (s *Stopwatch) Restart() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.isRunning = true
	s.startTime = time.Now()
}

func (s *Stopwatch) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.isRunning {
		s.isRunning = false
		s.elapsedTime = time.Since(s.startTime)
	}
}

// ElapsedTime returns the elapsed time of the Stopwatch.
//
// No parameters.
//
// Returns a time.Duration representing the elapsed time.
func (s *Stopwatch) ElapsedTime() time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.isRunning {
		return time.Since(s.startTime)
	} else {
		return s.elapsedTime
	}
}

// Elapsing executes the given function and returns the elapsed time and error.
//
// The function fn is executed after restarting the stopwatch. The stopwatch is then stopped
// after executing the function. The elapsed time is returned along with any error that occurred.
//
// The return type is time.Duration and error.
func (s *Stopwatch) Elapsing(fn func() error) (time.Duration, error) {
	s.Restart()
	defer s.Stop()

	err := fn()
	s.Stop()

	return s.ElapsedTime(), err
}
