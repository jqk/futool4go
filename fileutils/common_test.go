package fileutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDirStatisticsIncludeSubDir(t *testing.T) {
	dirCount, fileCount, size, err := GetDirStatistics("../test-data/fileutils/extension")

	assert.Nil(t, err)
	assert.Equal(t, 3, dirCount)
	assert.Equal(t, 8, fileCount)
	assert.Equal(t, int64(368), size)
}

func TestGetDirStatisticsExcludeSubDir(t *testing.T) {
	dirCount, fileCount, size, err := GetDirStatistics("../test-data/fileutils/extension", false)

	assert.Nil(t, err)
	assert.Equal(t, 1, dirCount)
	assert.Equal(t, 4, fileCount)
	assert.Equal(t, int64(176), size)
}
