package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartStop(t *testing.T) {
	step := time.Millisecond * 50
	sw := Stopwatch{}

	sw.Start()
	time.Sleep(step)
	assert.True(t, sw.IsRunning())

	sw.Stop()
	assert.False(t, sw.IsRunning())
	assert.True(t, sw.ElapsedTime() >= step)
}

func TestPauseResume(t *testing.T) {
	step := time.Millisecond * 50
	sw := Stopwatch{}

	sw.Pause()
	assert.False(t, sw.IsRunning())

	sw.Start()
	assert.True(t, sw.IsRunning())
	time.Sleep(step)

	sw.Pause()
	assert.False(t, sw.IsRunning())
	assert.True(t, sw.ElapsedTime() >= step)
	time.Sleep(step)

	sw.Start()
	assert.True(t, sw.IsRunning())
	time.Sleep(step)

	sw.Stop()
	assert.False(t, sw.IsRunning())
	assert.True(t, sw.ElapsedTime() >= step*2)
}

func TestElapsing(t *testing.T) {
	step := time.Millisecond * 50
	sw := Stopwatch{}

	d, err := sw.Elapsing(func() error {
		time.Sleep(step)
		return nil
	})

	assert.Nil(t, err)
	assert.True(t, d >= step)
}
