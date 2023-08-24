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
	records     []time.Duration
	lock        sync.RWMutex
}

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

	stop(s)
	reset(s)
	start(s)
}

/*
Reset resets all information in the Stopwatch. If the Stopwatch is already running, there is no effect.

Start 重置计时器。如果 Stopwatch 当前正在运行，则无操作。
*/
func (s *Stopwatch) Reset() {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.isRunning {
		reset(s)
	}
}

func reset(s *Stopwatch) {
	s.elapsedTime = 0
	s.records = s.records[0:0]
}

/*
Start or resume timing. If the Stopwatch is already running, there is no effect.

Start 开始或继续计时。如果 Stopwatch 当前正在运行，则无操作。
*/
func (s *Stopwatch) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.isRunning {
		start(s)
	}
}

func start(s *Stopwatch) {
	s.isRunning = true
	s.startTime = time.Now()
}

/*
Stop stops the stopwatch. If the Stopwatch is not running, there is no effect.

Stop 停止计时。如果 Stopwatch 当前未运行，则无操作。
*/
func (s *Stopwatch) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.isRunning {
		stop(s)
	}
}

func stop(s *Stopwatch) {
	s.isRunning = false
	s.elapsedTime += time.Since(s.startTime)
}

/*
Record records the lap time or split time when stopwatch is running.

Returns:
  - Elapsed time array, arranged in the order of calling Record().
    The elapsed time for all is calculated from the first Start().

Record 在 Stopwatch 正在运行时记录当前的一段时间。

返回:
  - 耗时数组，按调用 Record() 的顺序排列。所有耗时时间都是从第一次 Start() 开始计算的。
*/
func (s *Stopwatch) Record() []time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.isRunning {
		elapsed := s.elapsedTime + time.Since(s.startTime)
		s.records = append(s.records, elapsed)
	}

	return append([]time.Duration{}, s.records...)
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
