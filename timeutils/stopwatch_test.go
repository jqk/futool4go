package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartStop(t *testing.T) {
	step := time.Millisecond * 50
	sw := Stopwatch{}

	sw.Stop()
	assert.False(t, sw.IsRunning())

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

	sw.Start()
	assert.True(t, sw.IsRunning())
	time.Sleep(step)

	sw.Stop()
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

func TestReset(t *testing.T) {
	step := time.Millisecond * 50
	sw := Stopwatch{}

	sw.Start()
	assert.True(t, sw.IsRunning())
	time.Sleep(step)

	sw.Stop()
	assert.False(t, sw.IsRunning())
	assert.True(t, sw.ElapsedTime() >= step)
	time.Sleep(step)

	// only diff between TestPauseResume and TestReset.
	sw.Reset()

	sw.Start()
	assert.True(t, sw.IsRunning())
	time.Sleep(step)

	sw.Stop()
	assert.False(t, sw.IsRunning())
	assert.True(t, sw.ElapsedTime() >= step)
	assert.False(t, sw.ElapsedTime() >= step*2)
}

func TestRecord(t *testing.T) {
	step := time.Millisecond * 50
	sw := Stopwatch{}

	r := sw.Record()
	assert.Equal(t, 0, len(r))

	sw.Start()

	time.Sleep(step)
	r = sw.Record()
	assert.Equal(t, 1, len(r))
	assert.True(t, r[0] >= step)

	time.Sleep(step)
	r = sw.Record()
	assert.Equal(t, 2, len(r))
	assert.True(t, r[0] >= step)
	assert.True(t, r[1] >= step*2)

	time.Sleep(step)
	r = sw.Record()
	assert.Equal(t, 3, len(r))
	assert.True(t, r[0] >= step)
	assert.True(t, r[1] >= step*2)
	assert.True(t, r[2] >= step*3)

	sw.Stop()

	r = sw.Record()
	assert.Equal(t, 3, len(r))
	assert.True(t, r[0] >= step)
	assert.True(t, r[1] >= step*2)
	assert.True(t, r[2] >= step*3)
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
