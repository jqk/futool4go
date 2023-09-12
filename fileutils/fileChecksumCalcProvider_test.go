package fileutils

import (
	"hash"
	"hash/crc64"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZeroLengthFile64(t *testing.T) {
	buffer := make([]byte, 10240)
	p := newChecksumProvider()

	// 文件头和整个文件都要计算。
	err := GetFileChecksumWithProvider[uint64](
		"../test-data/fileutils/extension/zero-length.properties",
		2000, buffer, p, true, true,
	)

	assert.Nil(t, err)
	assert.Equal(t, uint64(0), p.HeaderChecksum())
	assert.Equal(t, uint64(0), p.FullChecksum())

	// 不计算文件头。
	err = GetFileChecksumWithProvider[uint64](
		"../test-data/fileutils/extension/zero-length.properties",
		2000, buffer, p, false, true,
	)

	assert.Nil(t, err)
	assert.Equal(t, uint64(0), p.HeaderChecksum())
	assert.Equal(t, uint64(0), p.FullChecksum())
}

func TestGetLargeFileChecksum64(t *testing.T) {
	buffer := make([]byte, 10240)
	p := newChecksumProvider()

	// 文件头和整个文件都要计算。
	err := GetFileChecksumWithProvider[uint64](
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, p, true, true,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0x15ca02b42efc56d9), p.HeaderChecksum())
	assert.Equal(t, uint64(0xb8b5323611968f17), p.FullChecksum())

	reset32()

	// 不计算文件头。
	err = GetFileChecksumWithProvider[uint64](
		"../test-data/fileutils/filter/001.MD",
		-1, buffer, p, false, true,
	)

	assert.Nil(t, err)
	assert.False(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0xb8b5323611968f17), p.FullChecksum())

	reset32()

	// 不计算整个文件。
	err = GetFileChecksumWithProvider[uint64](
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, p, true, false,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.False(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0x15ca02b42efc56d9), p.HeaderChecksum())
}

type checksumProvider struct {
	hashCrc64             hash.Hash64
	fileInfo              os.FileInfo
	headerChecksum        uint64
	fullChecksum          uint64
	isHeaderChecksumReady bool
	isFullChecksumReady   bool
}

func newChecksumProvider() *checksumProvider {
	return &checksumProvider{
		hashCrc64:             crc64.New(crc64.MakeTable(crc64.ISO)),
		headerChecksum:        0,
		fullChecksum:          0,
		isHeaderChecksumReady: false,
		isFullChecksumReady:   false,
	}
}

func (c *checksumProvider) FileInfo() os.FileInfo {
	return c.fileInfo
}

func (c *checksumProvider) HeaderChecksum() uint64 {
	return c.headerChecksum
}

func (c *checksumProvider) FullChecksum() uint64 {
	return c.fullChecksum
}

func (c *checksumProvider) IsHeaderChecksumReady() bool {
	return c.isHeaderChecksumReady
}

func (c *checksumProvider) IsFullChecksumReady() bool {
	return c.isFullChecksumReady
}

func (c *checksumProvider) ChecksumCalculator(buffer []byte) (int, error) {
	return c.hashCrc64.Write(buffer)
}

func (c *checksumProvider) HeaderReadyHandler(info os.FileInfo, fullIsReady bool) error {
	c.headerChecksum = c.hashCrc64.Sum64()
	c.fileInfo = info
	c.isHeaderChecksumReady = true

	if fullIsReady {
		c.isFullChecksumReady = true
		c.fullChecksum = c.headerChecksum
	}
	return nil
}

func (c *checksumProvider) FullReadyHandler(info os.FileInfo) error {
	c.fullChecksum = c.hashCrc64.Sum64()
	c.fileInfo = info
	c.isFullChecksumReady = true
	return nil
}

func (c *checksumProvider) Reset() {
	c.hashCrc64.Reset()
	c.isHeaderChecksumReady, c.isFullChecksumReady = false, false
}

func TestGetLargeFileChecksumDrivedProvider(t *testing.T) {
	buffer := make([]byte, 10240)

	p := NewCommonFileChecksumProvider[uint64](func() (hash.Hash, func() uint64) {
		hash := crc64.New(crc64.MakeTable(crc64.ISO))
		f := func() uint64 {
			return hash.Sum64()
		}
		return hash, f
	}())

	// 文件头和整个文件都要计算。
	err := GetFileChecksumWithProvider[uint64](
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, p, true, true,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0x15ca02b42efc56d9), p.HeaderChecksum())
	assert.Equal(t, uint64(0xb8b5323611968f17), p.FullChecksum())

	reset32()

	// 不计算文件头。
	err = GetFileChecksumWithProvider[uint64](
		"../test-data/fileutils/filter/001.MD",
		-1, buffer, p, false, true,
	)

	assert.Nil(t, err)
	assert.False(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0xb8b5323611968f17), p.FullChecksum())

	reset32()

	// 不计算整个文件。
	err = GetFileChecksumWithProvider[uint64](
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, p, true, false,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.False(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0x15ca02b42efc56d9), p.HeaderChecksum())
}
