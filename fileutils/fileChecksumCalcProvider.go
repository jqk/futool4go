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
type FileChecksumCalculationProvider interface {
	Method() string                                // Can be any non-empty string. The digest algorithm name is suggested.
	FileInfo() os.FileInfo                         // The file info of the file being processed. Valid when calculation is done.
	HeaderChecksum() []byte                        // The checksum of the file header.
	FullChecksum() []byte                          // The checksum of the whole file.
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
  - isNeedHeaderChecksum: If calculating header checksum is required. Cannot be false when isNeeeFullChecksum is false.
  - isNeeeFullChecksum: If calculating full checksum is required. Cannot be false when isNeedHeaderChecksum is false.
  - provider: The object that performs the checksum calculation, cannot be nil.

Returns:
  - an error if any of the arguments are invalid or an error occurs while calculating the checksum.

GetFileChecksum 计算文件的校验值。本函数负责文件操作，仅把校验各计算方法将由调用者实现，简化其操作。

参数:
  - filename: 待处理的文件名。
  - headerSize: 文件头长度。可能大于等于文件长度。
  - buffer: 读取文件的缓冲区。
  - isNeedHeaderChecksum: 是否需要头部校验值。不能与 isNeeeFullChecksum 同时为 false。
  - isNeeeFullChecksum: 是否需要完整校验值。不能与 isNeedHeaderChecksum 同时为 false。
  - provider: 执行校验和计算的对象，不能为 nil。

返回:
  - 错误信息。
*/
func GetFileChecksumWithProvider(
	filename string,
	headerSize int,
	buffer []byte,
	isNeedHeaderChecksum bool,
	isNeeeFullChecksum bool,
	provider FileChecksumCalculationProvider,
) error {
	if provider == nil {
		return errors.New("provider must not be nil")
	} else if !isNeedHeaderChecksum && !isNeeeFullChecksum {
		return errors.New("isNeedHeaderChecksum and isNeeeFullChecksum must not be false at the same time")
	}

	provider.Reset()

	var headerHandler HeaderChecksumReadyFunc = nil
	var fullHandler FullChecksumReadyFunc = nil

	if isNeedHeaderChecksum {
		headerHandler = provider.HeaderReadyHandler
	}
	if isNeeeFullChecksum {
		fullHandler = provider.FullReadyHandler
	}

	return GetFileChecksum(filename, headerSize, buffer, provider.ChecksumCalculator, headerHandler, fullHandler)
}

/*
CommonFileChecksumProvider implements the [FileChecksumCalculationProvider] interface for calculating the checksum.

CommonFileChecksumProvider 实现了 [FileChecksumCalculationProvider] 接口，用于计算校验值。
*/
type CommonFileChecksumProvider struct {
	method                string
	fileInfo              os.FileInfo
	headerChecksum        []byte
	fullChecksum          []byte
	isHeaderChecksumReady bool
	isFullChecksumReady   bool
	hash                  hash.Hash
}

/*
NewCommonFileChecksumProvider creates a new CommonFileChecksumProvider object.

Parameters:
  - method: The digest algorithm name.
  - hashInstance: The hash instance to use.

Example:

		// using crc32.
		p := NewCommonFileChecksumProvider("crc32", crc32.NewIEEE())

		err := GetFileChecksumWithProvider(
			"../test-data/fileutils/filter/001.MD",
			2000, buffer, false, true, p,
		)

		if err == nil && p.IsHeaderChecksumReady() {
			// Ok, here.
		}

		// or using MD5
		hash := md5.New()
		p := NewCommonFileChecksumProvider("MD5", hash)

		err := GetFileChecksumWithProvider[[]byte](
			"../test-data/fileutils/filter/001.MD",
			2000, buffer, p, true, true,
		)

		if err == nil && p.IsHeaderChecksumReady() && bytes.Equal(p.HeaderChecksum(), []byte{....}) {
			// Ok, here
		}

		// or define your own type with additional prperties.
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
	    		return binary.BigEndian.Uint64(c.headerChecksum)
	    	}
	    	return uint64(0)
	    }

	    func (c *crc64Provider) FullChecksumValue() uint64 {
	    	if c.isFullChecksumReady {
	    		return binary.BigEndian.Uint64(c.fullChecksum)
	    	}
	    	return uint64(0)
	    }

NewCommonFileChecksumProvider 创建一个 CommonFileChecksumProvider 对象。

参数:
  - method: 哈希算法名称。
  - hashInstance: 使用的哈希实例。
*/
func NewCommonFileChecksumProvider(method string, hashInstance hash.Hash) *CommonFileChecksumProvider {
	result := &CommonFileChecksumProvider{
		method:                method,
		fileInfo:              nil,
		isHeaderChecksumReady: false,
		isFullChecksumReady:   false,
		hash:                  hashInstance,
	}

	return result
}

// Method returns the digest algorithm name.
//
// Method 返回哈希算法名称。
func (c *CommonFileChecksumProvider) Method() string {
	return c.method
}

// FileInfo returns the os.FileInfo of the CommonFileChecksumProvider[T]. Only valid when the calculation is done.
// At this time, either IsHeaderChecksumReady() or IsFullChecksumReady() is true.
//
// FileInfo 返回所计算的文件信息。仅在校验值计算完成后有效。此时，IsHeaderChecksumReady() 或 IsFullChecksumReady() 为 true。
func (c *CommonFileChecksumProvider) FileInfo() os.FileInfo {
	return c.fileInfo
}

// HeaderChecksum returns the checksum of the file header. Only valid when the IsHeaderChecksumReady() is true.
//
// HeaderChecksum 返回文件头的校验值。仅当 IsHeaderChecksumReady() 返回 true 时有效。
func (c *CommonFileChecksumProvider) HeaderChecksum() []byte {
	checksum := make([]byte, len(c.headerChecksum))
	copy(checksum, c.headerChecksum)
	return checksum
}

// FullChecksum returns the checksum of the whole file. Only valid when the IsFullChecksumReady() is true.
//
// FullChecksum 返回整个文件的校验值。仅当 IsFullChecksumReady() 返回 true 时有效。
func (c *CommonFileChecksumProvider) FullChecksum() []byte {
	checksum := make([]byte, len(c.fullChecksum))
	copy(checksum, c.fullChecksum)
	return checksum
}

// IsHeaderChecksumReady returns true when the file header checksum is calculated.
func (c *CommonFileChecksumProvider) IsHeaderChecksumReady() bool {
	return c.isHeaderChecksumReady
}

// IsFullChecksumReady returns true when the whole file checksum is calculated.
func (c *CommonFileChecksumProvider) IsFullChecksumReady() bool {
	return c.isFullChecksumReady
}

// ChecksumCalculator calculates the checksum of the file segment.
func (c *CommonFileChecksumProvider) ChecksumCalculator(buffer []byte) (int, error) {
	return c.hash.Write(buffer)
}

// HeaderReadyHandler handles the checksum calculation when header is calculated.
func (c *CommonFileChecksumProvider) HeaderReadyHandler(info os.FileInfo, fullIsReady bool) error {
	c.headerChecksum = c.hash.Sum(nil)
	c.fileInfo = info
	c.isHeaderChecksumReady = true

	if fullIsReady {
		c.isFullChecksumReady = true
		c.fullChecksum = c.headerChecksum
	}
	return nil
}

// FullReadyHandler handles the checksum calculation when whole file is calculated.
func (c *CommonFileChecksumProvider) FullReadyHandler(info os.FileInfo) error {
	c.fullChecksum = c.hash.Sum(nil)
	c.fileInfo = info
	c.isFullChecksumReady = true
	return nil
}

// Reset resets all information for next calculation.
func (c *CommonFileChecksumProvider) Reset() {
	c.hash.Reset()
	c.isHeaderChecksumReady, c.isFullChecksumReady = false, false
	c.headerChecksum, c.fullChecksum = nil, nil
	c.fileInfo = nil
}
