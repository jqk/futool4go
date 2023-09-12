package fileutils

import (
	"hash/crc32"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var hashCrc32 = crc32.NewIEEE()

var headerChecksum32, fullChecksum32 uint32
var headerReadyHanderIsRun32, fullReadyHandlerIsRun32 bool

func TestZeroLengthFile(t *testing.T) {
	buffer := make([]byte, 10240)
	reset32()

	// 文件头和整个文件都要计算。
	err := GetFileChecksum(
		"../test-data/fileutils/extension/zero-length.properties",
		2000,
		buffer,
		calculateChecksum32,
		headerReadyHander32,
		fullReadyHandler32,
	)

	assert.Nil(t, err)
	assert.Equal(t, uint32(0), headerChecksum32)
	assert.Equal(t, uint32(0), fullChecksum32)

	// 不计算文件头。
	err = GetFileChecksum(
		"../test-data/fileutils/extension/zero-length.properties",
		2000,
		buffer,
		calculateChecksum32,
		nil,
		fullReadyHandler32,
	)

	assert.Nil(t, err)
	assert.Equal(t, uint32(0), headerChecksum32)
	assert.Equal(t, uint32(0), fullChecksum32)
}

func TestGetLargeFileChecksum(t *testing.T) {
	buffer := make([]byte, 10240)
	reset32()

	// 文件头和整个文件都要计算。
	err := GetFileChecksum(
		"../test-data/fileutils/filter/001.MD",
		2000,
		buffer,
		calculateChecksum32,
		headerReadyHander32,
		fullReadyHandler32,
	)

	assert.Nil(t, err)
	assert.Equal(t, uint32(3222652411), headerChecksum32)
	assert.Equal(t, uint32(3230993970), fullChecksum32)
	assert.True(t, headerReadyHanderIsRun32)
	assert.True(t, fullReadyHandlerIsRun32)

	reset32()

	// 不计算文件头。
	err = GetFileChecksum(
		"../test-data/fileutils/filter/001.MD",
		-1,
		buffer,
		calculateChecksum32,
		nil,
		fullReadyHandler32,
	)

	assert.Nil(t, err)
	assert.Equal(t, uint32(0), headerChecksum32)
	assert.Equal(t, uint32(3230993970), fullChecksum32)
	assert.Equal(t, false, headerReadyHanderIsRun32)
	assert.Equal(t, true, fullReadyHandlerIsRun32)

	reset32()

	// 不计算整个文件。
	err = GetFileChecksum(
		"../test-data/fileutils/filter/001.MD",
		2000,
		buffer,
		calculateChecksum32,
		headerReadyHander32,
		nil,
	)

	assert.Nil(t, err)
	assert.Equal(t, uint32(3222652411), headerChecksum32)
	assert.Equal(t, uint32(0), fullChecksum32)
	assert.Equal(t, true, headerReadyHanderIsRun32)
	assert.Equal(t, false, fullReadyHandlerIsRun32)
}

func TestGetSmallFileChecksum(t *testing.T) {
	buffer := make([]byte, 10240)
	reset32()

	// 文件小于文件头的长度。
	err := GetFileChecksum(
		"../test-data/fileutils/filter/002.txt",
		2000,
		buffer,
		calculateChecksum32,
		headerReadyHander32,
		fullReadyHandler32,
	)

	assert.Nil(t, err)
	assert.Equal(t, uint32(4245835769), headerChecksum32)
	assert.Equal(t, uint32(4245835769), fullChecksum32)
	assert.Equal(t, true, headerReadyHanderIsRun32)
	assert.Equal(t, false, fullReadyHandlerIsRun32)
}

func reset32() {
	hashCrc32.Reset()
	headerChecksum32, fullChecksum32 = 0, 0
	headerReadyHanderIsRun32, fullReadyHandlerIsRun32 = false, false
}

func calculateChecksum32(data []byte) (int, error) {
	return hashCrc32.Write(data)
}

func headerReadyHander32(info os.FileInfo, fullIsReady bool) error {
	headerReadyHanderIsRun32 = true
	headerChecksum32 = hashCrc32.Sum32()

	if fullIsReady {
		fullChecksum32 = headerChecksum32
	}
	return nil
}

func fullReadyHandler32(info os.FileInfo) error {
	fullReadyHandlerIsRun32 = true
	fullChecksum32 = hashCrc32.Sum32()
	return nil
}
