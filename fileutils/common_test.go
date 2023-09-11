package fileutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDirStatisticsIncludeSubDir(t *testing.T) {
	option := NewWalkOption()
	stat, err := GetDirStatistics("../test-data/fileutils/extension", option)

	assert.Nil(t, err)
	assert.Equal(t, 3, stat.DirCount)
	assert.Equal(t, 8, stat.FileCount)
	assert.Equal(t, int64(368), stat.TotalSize)
}

func TestGetDirStatisticsExcludeSubDir(t *testing.T) {
	option := &WalkOption{
		Recursive: false,
	}
	stat, err := GetDirStatistics("../test-data/fileutils/extension", option)

	assert.Nil(t, err)
	assert.Equal(t, 1, stat.DirCount)
	assert.Equal(t, 4, stat.FileCount)
	assert.Equal(t, int64(176), stat.TotalSize)
}
