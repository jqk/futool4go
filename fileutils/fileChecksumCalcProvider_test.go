package fileutils

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"hash/crc64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZeroLengthFile64(t *testing.T) {
	buffer := make([]byte, 10240)
	p := newCrc64Provider()

	// 文件头和整个文件都要计算。
	err := GetFileChecksumWithProvider(
		"../test-data/fileutils/extension/zero-length.properties",
		2000, buffer, true, true, p,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.Equal(t, uint64(0), p.HeaderChecksumValue())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0), p.FullChecksumValue())

	// 不计算文件头。
	err = GetFileChecksumWithProvider(
		"../test-data/fileutils/extension/zero-length.properties",
		2000, buffer, false, true, p,
	)

	assert.Nil(t, err)
	assert.False(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0), p.FullChecksumValue())
}

func TestGetLargeFileChecksum64(t *testing.T) {
	buffer := make([]byte, 10240)
	p := newCrc64Provider()

	// 文件头和整个文件都要计算。
	err := GetFileChecksumWithProvider(
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, true, true, p,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0x15ca02b42efc56d9), p.HeaderChecksumValue())
	assert.Equal(t, uint64(0xb8b5323611968f17), p.FullChecksumValue())

	// 不计算文件头。
	err = GetFileChecksumWithProvider(
		"../test-data/fileutils/filter/001.MD",
		-1, buffer, false, true, p,
	)

	assert.Nil(t, err)
	assert.False(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0xb8b5323611968f17), p.FullChecksumValue())

	// 不计算整个文件。
	err = GetFileChecksumWithProvider(
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, true, false, p,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.False(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0x15ca02b42efc56d9), p.HeaderChecksumValue())
}

// 下面这些代码模拟自定义结构实现 CommonFileChecksumProvider 相同的功能，附加新属性。
type crc64Provider struct {
	CommonFileChecksumProvider
}

func newCrc64Provider() *crc64Provider {
	return &crc64Provider{
		CommonFileChecksumProvider: CommonFileChecksumProvider{
			method:                "crc64",
			hash:                  crc64.New(crc64.MakeTable(crc64.ISO)),
			headerChecksum:        nil,
			fullChecksum:          nil,
			isHeaderChecksumReady: false,
			isFullChecksumReady:   false,
		},
	}
}

func (c *crc64Provider) HeaderChecksumValue() uint64 {
	if c.isHeaderChecksumReady {
		//return binary.LittleEndian.Uint64(c.headerChecksum)
		return binary.BigEndian.Uint64(c.headerChecksum)
	}
	return uint64(0)
}

func (c *crc64Provider) FullChecksumValue() uint64 {
	if c.isFullChecksumReady {
		// return binary.LittleEndian.Uint64(c.fullChecksum)
		return binary.BigEndian.Uint64(c.fullChecksum)
	}
	return uint64(0)
}

func TestGetLargeFileChecksumDrivedProvider64(t *testing.T) {
	buffer := make([]byte, 10240)

	p := newCrc64Provider()

	// 文件头和整个文件都要计算。
	err := GetFileChecksumWithProvider(
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, true, true, p,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0x15ca02b42efc56d9), p.HeaderChecksumValue())
	assert.Equal(t, uint64(0xb8b5323611968f17), p.FullChecksumValue())

	// 不计算文件头。
	err = GetFileChecksumWithProvider(
		"../test-data/fileutils/filter/001.MD",
		-1, buffer, false, true, p,
	)

	assert.Nil(t, err)
	assert.False(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0xb8b5323611968f17), p.FullChecksumValue())

	// 不计算整个文件。
	err = GetFileChecksumWithProvider(
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, true, false, p,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.False(t, p.IsFullChecksumReady())
	assert.Equal(t, uint64(0x15ca02b42efc56d9), p.HeaderChecksumValue())
}

func TestGetLargeFileChecksumDrivedProviderMD5(t *testing.T) {
	buffer := make([]byte, 10240)
	p := NewCommonFileChecksumProvider("MD5", md5.New())

	// 文件头和整个文件都要计算。
	err := GetFileChecksumWithProvider(
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, true, true, p,
	)

	header := []byte{199, 85, 44, 115, 143, 23, 243, 52, 237, 88, 199, 105, 89, 15, 101, 103}
	full := []byte{47, 122, 214, 188, 119, 125, 116, 142, 29, 186, 194, 159, 89, 176, 209, 159}

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.True(t, bytes.Equal(header, p.HeaderChecksum()))
	assert.True(t, bytes.Equal(full, p.FullChecksum()))

	// 不计算文件头。
	err = GetFileChecksumWithProvider(
		"../test-data/fileutils/filter/001.MD",
		-1, buffer, false, true, p,
	)

	assert.Nil(t, err)
	assert.False(t, p.IsHeaderChecksumReady())
	assert.True(t, p.IsFullChecksumReady())
	assert.True(t, bytes.Equal(full, p.FullChecksum()))

	// 不计算整个文件。
	err = GetFileChecksumWithProvider(
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, true, false, p,
	)

	assert.Nil(t, err)
	assert.True(t, p.IsHeaderChecksumReady())
	assert.False(t, p.IsFullChecksumReady())
	assert.True(t, bytes.Equal(header, p.HeaderChecksum()))
}
