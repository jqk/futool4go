package fileutils

import (
	"bufio"
	"errors"
	"io"
	"os"
	"reflect"
)

type ChecksumCalculator func([]byte) (int, error)
type HeaderChecksumReadyHandler func(os.FileInfo, bool) error
type FullChecksumReadyHandler func(os.FileInfo) error

// GetFileChecksum calculates the checksum of a file using the provided calculator.
//
// filename is the path to the file to be checksummed. headerSize is the number of bytes to be calculated for the file header.
// buf is the buffer to be used for reading the file, either a []byte or a int. calculator is the checksum calculator to be used.
// headerReadyHandler is the function to be called after the header checksum is calculated.
// Being nil indicates that the header checksum is not required.
// fullReadyHandler is the function to be called after the full file checksum is calculated.
// Being nil indicates that the full file checksum is not required.
//
// Returns an error if any of the arguments are invalid or an error occurs while calculating the checksum.
func GetFileChecksum[T int | []byte](
	filename string,
	headerSize int,
	buf T,
	calculator ChecksumCalculator,
	headerReadyHandler HeaderChecksumReadyHandler,
	fullReadyHandler FullChecksumReadyHandler,
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
		if err != nil && err != io.ErrUnexpectedEOF {
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

	// 到达此处时，fullReadyHandler 必然不为 nil。
	// 继续读取文件剩余部分，计算整体校验和。
	for {
		readCount, err = reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				// 说明确实有错误，不是读到结尾了，中断处理。
				return err
			}

			// 读到结尾了，处理整体校验和。
			return fullReadyHandler(info)
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
	calculator ChecksumCalculator,
	headerReadyHandler HeaderChecksumReadyHandler,
	fullReadyHandler FullChecksumReadyHandler,
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
