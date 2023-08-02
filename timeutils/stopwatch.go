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

var zeroTime time.Time

// IsRunning indicates whether the stopwatch is currently running.
func (s *Stopwatch) IsRunning() bool {
	return s.isRunning
}

// Restart reset everything and starts the stopwatch regardless of whether it is running or not.
func (s *Stopwatch) Restart() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.isRunning = true
	s.startTime = time.Now()
	s.elapsedTime = 0
}

// Start starts or resumes the stopwatch.
// Do nothing when stopwatch is currently running.
func (s *Stopwatch) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.isRunning {
		if s.startTime == zeroTime {
			// startTime 为零，说明 stopwatch 未曾启动或经由 Stop() 停止了，
			// 而不是由 Pause() 暂停的，所以无需继续计时，而是重新开始。
			s.elapsedTime = 0
		}

		s.isRunning = true
		s.startTime = time.Now()
	}
}

// Stop stops the stopwatch and record the ElapsedTime.
// Do nothing when stopwatch is not currently running.
func (s *Stopwatch) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.isRunning {
		s.isRunning = false
		// 继续累加时间。因为有可能是经过 Pause() 后又 Start() 了。
		s.elapsedTime += time.Since(s.startTime)
		// zeroTime 主要供判断是否经过 Pause() 后暂停。
		s.startTime = zeroTime
	}
}

// Pause pauses the stopwatch and update the ElapsedTime.
// Do nothing when stopwatch is not currently running.
func (s *Stopwatch) Pause() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.isRunning {
		s.isRunning = false
		s.elapsedTime += time.Since(s.startTime)
	}
}

// ElapsedTime returns the elapsed time of the Stopwatch.
//
// Returns a time.Duration representing the elapsed time.
func (s *Stopwatch) ElapsedTime() time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.isRunning {
		return s.elapsedTime + time.Since(s.startTime)
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
	return s.ElapsedTime(), err
}
