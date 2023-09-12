package fileutils

import (
	"errors"
	"hash"
	"os"
)

// FileChecksumCalculationProvider defines the interface for calculating the checksum.
//
// See [ChecksumCalculateFunc], [HeaderChecksumReadyFunc] and [FullChecksumReadyFunc]
// for details of ChecksumCalculator, HeaderReadyHandler and FullReadyHandler.
type FileChecksumCalculationProvider[T any] interface {
	FileInfo() os.FileInfo                         // The file info of the file being processed. Valid when calculation is done.
	HeaderChecksum() T                             // The checksum of the file header.
	FullChecksum() T                               // The checksum of the whole file.
	IsHeaderChecksumReady() bool                   // Whether the header checksum is ready.
	IsFullChecksumReady() bool                     // Whether the full checksum is ready.
	ChecksumCalculator(buffer []byte) (int, error) // The function to calculate the checksum of the file segment.
	HeaderReadyHandler(os.FileInfo, bool) error    // The function to handle the checksum calculation when header is calculated.
	FullReadyHandler(os.FileInfo) error            // The function to handle the checksum calculation when whole file is calculated.
	Reset()                                        // Reset all information for next calculation.
}

/*
GetFileChecksum calculates the checksum for a file. This function is responsible for file operations,
and only delegates the checksum calculation methods to the caller to simplify operations.

Parameters:
  - filename: Name of the file to process.
  - headerSize: Length of the file header. Can be greater than or equal to the file length.
  - buffer: Buffer for reading the file.
  - provider: The object that performs the checksum calculation, cannot be nil.
  - isNeedHeaderChecksum: If calculating header checksum is required. Cannot be false when isNeeeFullChecksum is false.
  - isNeeeFullChecksum: If calculating full checksum is required. Cannot be false when isNeedHeaderChecksum is false.

Returns:
  - an error if any of the arguments are invalid or an error occurs while calculating the checksum.

GetFileChecksum 计算文件的校验值。本函数负责文件操作，仅把校验各计算方法将由调用者实现，简化其操作。

参数:
  - filename: 待处理的文件名。
  - headerSize: 文件头长度。可能大于等于文件长度。
  - buffer: 读取文件的缓冲区。
  - provider: 执行校验和计算的对象，不能为 nil。
  - isNeedHeaderChecksum: 是否需要头部校验值。不能与 isNeeeFullChecksum 同时为 false。
  - isNeeeFullChecksum: 是否需要完整校验值。不能与 isNeedHeaderChecksum 同时为 false。

返回:
  - 错误信息。
*/
func GetFileChecksumWithProvider[T any](
	filename string,
	headerSize int,
	buffer []byte,
	provider FileChecksumCalculationProvider[T],
	isNeedHeaderChecksum bool,
	isNeeeFullChecksum bool,
) error {
	if provider == nil {
		return errors.New("provider must not be nil")
	} else if !isNeedHeaderChecksum && !isNeeeFullChecksum {
		return errors.New("isNeedHeaderChecksum and isNeeeFullChecksum must not be false at the same time")
	}

	provider.Reset()

	if isNeedHeaderChecksum && isNeeeFullChecksum {
		return GetFileChecksum(filename, headerSize, buffer, provider.ChecksumCalculator, provider.HeaderReadyHandler, provider.FullReadyHandler)
	} else if isNeedHeaderChecksum {
		return GetFileChecksum(filename, headerSize, buffer, provider.ChecksumCalculator, provider.HeaderReadyHandler, nil)
	} else { // isNeeeFullChecksum
		return GetFileChecksum(filename, headerSize, buffer, provider.ChecksumCalculator, nil, provider.FullReadyHandler)
	}
}

/*
CommonFileChecksumProvider implements the [FileChecksumCalculationProvider] interface for calculating the checksum.

CommonFileChecksumProvider 实现了 [FileChecksumCalculationProvider] 接口，用于计算校验值。
*/
type CommonFileChecksumProvider[T any] struct {
	fileInfo              os.FileInfo
	headerChecksum        T
	fullChecksum          T
	isHeaderChecksumReady bool
	isFullChecksumReady   bool
	hash                  hash.Hash
	sumFunc               func() T
}

/*
NewCommonFileChecksumProvider creates a new CommonFileChecksumProvider[T].

Parameters:
  - hashInstance: The hash instance to use.
  - sumFuncOfhashInstance: The function that get the sum of the calculation. It must be a function of the hashInstance object instance.

Example:

	hash := crc32.NewIEEE()
	f := func() uint32 {
		return hash.Sum32()
	}

	p := NewCommonFileChecksumProvider[uint32](hash, f)

	err := GetFileChecksumWithProvider[uint32](
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, p, false, true,
	)

	// or init like this
	p := NewCommonFileChecksumProvider[uint64](func() (hash.Hash, func() uint64) {
		hash := crc64.New(crc64.MakeTable(crc64.ISO))
		f := func() uint64 {
			return hash.Sum64()
		}
		return hash, f
	}())

	err := GetFileChecksumWithProvider[uint64](
		"../test-data/fileutils/filter/001.MD",
		2000, buffer, p, true, true,
	)

NewCommonFileChecksumProvider 创建一个新的 CommonFileChecksumProvider[T]。

参数:
  - hashInstance: 使用的哈希实例。
  - sumFuncOfhashInstance: 计算哈希值的函数。必须是 hashInstance 对象实例的函数。
*/
func NewCommonFileChecksumProvider[T any](hashInstance hash.Hash, sumFuncOfhashInstance func() T) *CommonFileChecksumProvider[T] {
	result := &CommonFileChecksumProvider[T]{
		fileInfo:              nil,
		isHeaderChecksumReady: false,
		isFullChecksumReady:   false,
		hash:                  hashInstance,
		sumFunc:               sumFuncOfhashInstance,
	}

	return result
}

// FileInfo returns the os.FileInfo of the CommonFileChecksumProvider[T]. Only valid when the calculation is done.
// At this time, either IsHeaderChecksumReady() or IsFullChecksumReady() is true.
//
// FileInfo 返回所计算的文件信息。仅在校验值计算完成后有效。此时，IsHeaderChecksumReady() 或 IsFullChecksumReady() 为 true。
func (c *CommonFileChecksumProvider[T]) FileInfo() os.FileInfo {
	return c.fileInfo
}

// HeaderChecksum returns the checksum of the file header. Only valid when the IsHeaderChecksumReady() is true.
//
// HeaderChecksum 返回文件头的校验值。仅当 IsHeaderChecksumReady() 返回 true 时有效。
func (c *CommonFileChecksumProvider[T]) HeaderChecksum() T {
	return c.headerChecksum
}

// FullChecksum returns the checksum of the whole file. Only valid when the IsFullChecksumReady() is true.
//
// FullChecksum 返回整个文件的校验值。仅当 IsFullChecksumReady() 返回 true 时有效。
func (c *CommonFileChecksumProvider[T]) FullChecksum() T {
	return c.fullChecksum
}

// IsHeaderChecksumReady returns true when the file header checksum is calculated.
func (c *CommonFileChecksumProvider[T]) IsHeaderChecksumReady() bool {
	return c.isHeaderChecksumReady
}

// IsFullChecksumReady returns true when the whole file checksum is calculated.
func (c *CommonFileChecksumProvider[T]) IsFullChecksumReady() bool {
	return c.isFullChecksumReady
}

// ChecksumCalculator calculates the checksum of the file segment.
func (c *CommonFileChecksumProvider[T]) ChecksumCalculator(buffer []byte) (int, error) {
	return c.hash.Write(buffer)
}

// HeaderReadyHandler handles the checksum calculation when header is calculated.
func (c *CommonFileChecksumProvider[T]) HeaderReadyHandler(info os.FileInfo, fullIsReady bool) error {
	c.headerChecksum = c.sumFunc()
	c.fileInfo = info
	c.isHeaderChecksumReady = true

	if fullIsReady {
		c.isFullChecksumReady = true
		c.fullChecksum = c.headerChecksum
	}
	return nil
}

// FullReadyHandler handles the checksum calculation when whole file is calculated.
func (c *CommonFileChecksumProvider[T]) FullReadyHandler(info os.FileInfo) error {
	c.fullChecksum = c.sumFunc()
	c.fileInfo = info
	c.isFullChecksumReady = true
	return nil
}

// Reset resets all information for next calculation.
func (c *CommonFileChecksumProvider[T]) Reset() {
	c.hash.Reset()
	c.isHeaderChecksumReady, c.isFullChecksumReady = false, false
	c.fileInfo = nil
}
