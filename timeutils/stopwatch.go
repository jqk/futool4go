package timeutils

import (
	"sync"
	"time"
)

/*
Stopwatch is a stopwatch.

Stopwatch 定义了一个计时器。
*/
type Stopwatch struct {
	isRunning   bool
	startTime   time.Time
	elapsedTime time.Duration
	lock        sync.RWMutex
}

var zeroTime time.Time

/*
IsRunning indicates whether the stopwatch is currently running.

IsRunning 返回 Stopwatch 是否正在运行。
*/
func (s *Stopwatch) IsRunning() bool {
	return s.isRunning
}

/*
Restart resets all information in the Stopwatch and starts timing again.

Restart 重置 Stopwatch 的所有信息，并且重新开始计时。
*/
func (s *Stopwatch) Restart() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.isRunning = true
	s.startTime = time.Now()
	s.elapsedTime = 0
}

/*
Start If previously paused by [Pause], resumes the timing. Otherwise, restarts the timing.
If the Stopwatch is already running, there is no effect.

Start 如果前一次由 [Pause] 暂停，则继续计时。否则，重新开始计时。如果 Stopwatch 当前正在运行，则无操作。
*/
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

/*
Stop stops the stopwatch. If the Stopwatch is not running, there is no effect.

Stop 停止计时。如果 Stopwatch 当前未运行，则无操作。
*/
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

/*
Pause pauses the stopwatch. If the Stopwatch is not running, there is no effect.

Pause 暂停计时。如果 Stopwatch 当前未运行，则无操作。
*/
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
// ElapsedTime 返回 Stopwatch 的运行时间。
func (s *Stopwatch) ElapsedTime() time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.isRunning {
		return s.elapsedTime + time.Since(s.startTime)
	} else {
		return s.elapsedTime
	}
}

/*
Elapsing runs the given function and returns the elapsed time.

Parameters:
  - task: The function to execute. Can't be nil.

Returns:
  - The elapsed time.
  - Error message.

Elapsing 运行给定的函数并返回运行时间。

参数:
  - task: 要执行的函数。不能为 nil。

返回:
  - 运行时长。
  - 错误信息。
*/
func (s *Stopwatch) Elapsing(task func() error) (time.Duration, error) {
	s.Restart()
	defer s.Stop()

	err := task()
	return s.ElapsedTime(), err
}
