package fileutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDirStatistics(t *testing.T) {
	dirCount, fileCount, size, err := GetDirStatistics("../test-data/fileutils/extension")

	assert.Nil(t, err)
	assert.Equal(t, 3, dirCount)
	assert.Equal(t, 8, fileCount)
	assert.Equal(t, int64(368), size)
}
