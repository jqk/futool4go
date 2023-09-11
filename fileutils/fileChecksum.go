package fileutils

import (
	"bufio"
	"errors"
	"io"
	"os"
	"reflect"
)

/*
ChecksumCalculateFunc defines function type that calculates the checksum. It only needs to calculate the checksum for the given byte array.

Parameters:
  - the bytes to be caculated.

Returns:
  - the count of byte is being calculated.
  - an error if anything wrong during calculating the checksum.

ChecksumCalculateFunc 定义了执行检验和计算的函数类型。它只需计算给定的字节数组的校验和。

参数:
  - 待计算的字节数组。

返回:
  - 计算的字节数。
  - 错误信息。
*/
type ChecksumCalculateFunc func([]byte) (int, error)

/*
HeaderChecksumReadyFunc defines the function type that is called after the header checksum calculation is completed.
It is usually used to perform operations like saving the header checksum.

Parameters:
  - the os.FileInfo of the file being processed.
  - Whether the file has ended. When the preset header length is greater than or equal to the processed file length,
    this is true after header processing is completed; otherwise it is false, indicating the file is not fully processed.

Returns:
  - an error if anything wrong during the calculation.

HeaderChecksumReadyFunc 定义了在文件头部校验值计算完成后被调用的函数类型。一般用于执行保存文件头部校验和之类的操作。

参数:
  - 当前正在处理的文件信息。
  - 是否是文件已结束。当预设的文件头长度大于等于被处理的文件长度时，完成文件头处理后，该值为 true；否则为 false，说明文件还未处理完。

返回：
  - 错误信息。
*/
type HeaderChecksumReadyFunc func(os.FileInfo, bool) error

/*
FullChecksumReadyFunc defines function type that is called after the full file checksum is calculated.

Parameters:
  - the os.FileInfo of the file.

Returns:
  - an error if anything wrong during calculation.

FullChecksumReadyFunc 定义了在整个文件的完整校验值计算后被调用的函数类型。

参数:
  - 当前正在处理的文件信息。

返回：
  - 错误信息。
*/
type FullChecksumReadyFunc func(os.FileInfo) error

/*
GetFileChecksum calculates the checksum for a file. This function is responsible for file operations,
and only delegates the checksum calculation methods to the caller to simplify operations.

Parameters:
  - filename: Name of the file to process.
  - headerSize: Length of the file header. Can be greater than or equal to the file length.
  - buf: Buffer for reading the file. Can be []byte or int. The former directly provides the buffer for reuse;
    the latter sets the buffer length for the function to create the buffer itself.
  - calculator: The function that performs the checksum calculation, cannot be nil.
  - headerReadyHandler: Callback function after the header checksum is calculated.
    Can be nil, indicating no need to calculate header checksum separately.
    Cannot be nil if fullReadyHandler is nil.
  - fullReadyHandler: Callback function after the full checksum is calculated.
    Can be nil, indicating no need for full checksum.
    Cannot be nil if headerReadyHandler is nil.

Returns:
  - an error if any of the arguments are invalid or an error occurs while calculating the checksum.

GetFileChecksum 计算文件的校验值。本函数负责文件操作，仅把校验各计算方法将由调用者实现，简化其操作。

参数:
  - filename: 待处理的文件名。
  - headerSize: 文件头长度。可能大于等于文件长度。
  - buf: 读取文件的缓冲区。可以是 []byte 或 int。前者直接提供缓冲区，达到复用目的；后者设置缓冲区长度，由函数自己创建缓冲区。
  - calculator: 执行校验和计算的函数，不能为 nil。
  - headerReadyHandler: 头部校验值计算完成后的回调函数。可为 nil，表示不需要单独计算头部校验值。不能与 fullReadyHandler 同时为 nil。
  - fullReadyHandler: 全部校验值计算完成后的回调函数。可为 nil，表示不需要完整校验值。不能与 headerReadyHandler 同时为 nil。

返回:
  - 错误信息。
*/
func GetFileChecksum[T int | []byte](
	filename string,
	headerSize int,
	buf T,
	calculator ChecksumCalculateFunc,
	headerReadyHandler HeaderChecksumReadyFunc,
	fullReadyHandler FullChecksumReadyFunc,
) error {

	buffer := getBuffer(buf)
	if err := validateArguments(headerSize, len(buffer), calculator, headerReadyHandler, fullReadyHandler); err != nil {
		return err
	}

	// 打开文件的操作。
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 文件已打开，此处不会再有错误。
	info, _ := file.Stat()
	reader := bufio.NewReader(file)
	readCount := 0

	// 计算文件头的校验和。
	if headerReadyHandler != nil {
		// 根据规则，buffer 长度大于等于 headerSize。
		// 所以使用 buffer 可能读出超过头部长度的数据。
		readCount, err = io.ReadFull(reader, buffer)

		// err != nil 说明有问题，但有可能是如下两个不是问题的情况：
		// 1. 文件长度为 0，返回的是 io.EOF。
		// 2. 文件长度小于预定义的头部长度，返回的是 io.ErrUnexpectedEOF。
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}

		fullIsReady := false

		if readCount <= headerSize {
			// 因为前面使用了 ReadFull()，所以这里如果 readCount 小于等于 headerSize，
			// 则说明文件长度小于等于预定义的头部长度。
			// 此时得到的校验和，即是文件头的校验和，又是整个文件的校验和。
			fullIsReady = true
			if _, err = calculator(buffer[:readCount]); err != nil {
				return err
			}
		} else if _, err = calculator(buffer[:headerSize]); err != nil {
			return err
		}

		// 在 headerReady() 中做保存校验和等业务操作。
		if err = headerReadyHandler(info, fullIsReady); err != nil {
			return err
		}

		if fullIsReady || fullReadyHandler == nil {
			// 文件头校验和已是整体校验和，或者无需整体校验和，结束处理。
			return nil
		}

		// 运行到此处，readCount <= headerSize 的情况已导致 fullIsReady 为 true，并结束程序了。
		// 现在肯定是 readCount > headerSize，即读到的数据大于 headerSize。
		// 此时需计算本次读取的超出头部长度数据的校验和，以便继续计算整体校验和。
		if _, err = calculator(buffer[headerSize:readCount]); err != nil {
			return err
		}
	}

	// 到达此处时，fullReadyHandler 必然不为 nil。继续读取文件剩余部分，计算整体校验和。
	for {
		readCount, err = reader.Read(buffer)
		if err != nil {
			if err != io.EOF { // 说明确实有错误，不是读到结尾了，中断处理。
				return err
			}

			return fullReadyHandler(info) // 读到结尾了，处理整体校验和。
		}

		if _, err = calculator(buffer[:readCount]); err != nil {
			return err
		}
	}
}

func getBuffer[T int | []byte](buf T) []byte {
	switch v := reflect.ValueOf(buf); v.Kind() {
	case reflect.Int:
		return make([]byte, int(v.Int()))
	case reflect.Slice:
		return v.Slice(0, int(v.Len())).Bytes()
	default:
		// 不会执行到这里。
		panic("T must be int or []byte")
	}
}

func validateArguments(
	headerSize int,
	bufferSize int,
	calculator ChecksumCalculateFunc,
	headerReadyHandler HeaderChecksumReadyFunc,
	fullReadyHandler FullChecksumReadyFunc,
) error {
	if headerReadyHandler == nil && fullReadyHandler == nil {
		return errors.New("headerReadyHandler and fullReadyHandler must not be nil at the same time")
	} else if calculator == nil {
		return errors.New("calculator must not be nil")
	} else if headerSize <= 0 && headerReadyHandler != nil {
		// headerReadyHander 非空说明需处理 Header 校验和。此时 HeaderSize 必需大于 0。
		return errors.New("header size must be greater than 0")
	} else if bufferSize < headerSize && headerReadyHandler != nil {
		// headerReadyHander 非空说明需处理 Header 校验和。
		// 此时 HeaderSize 必然大于 0，BufferSize 必需大于等于 HeaderSize。
		return errors.New("buffer size can not be less than headerSize")
	}

	return nil
}
