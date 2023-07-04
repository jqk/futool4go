package fileutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExtensions(t *testing.T) {
	extensions, err := GetFileExtensions("../test-data/fileutils/extension", true, nil)
	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 7, len(extensions))

	extensions, err = GetFileExtensions("../test-data/fileutils/extension", false, nil)
	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 3, len(extensions))

	extensions, err = GetFileExtensions("../test-data/fileutils/extension", true,
		func(path string, info os.FileInfo, extension *FileExtension) error {
			// 直接停止，所以结果为空数组。
			return filepath.SkipAll
		})

	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 0, len(extensions))

	extensions, err = GetFileExtensions("../test-data/fileutils/extension", true,
		func(path string, info os.FileInfo, extension *FileExtension) error {
			if extension != nil {
				if strings.Index(path, "sub1") > 0 {
					// 已扫描完 extension 目录，再扫描了 sub1 中的第一个文件。
					return filepath.SkipDir
				}
			}
			return nil
		})

	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 6, len(extensions))

	extensions, err = GetFileExtensions("../test-data/fileutils/extension", true,
		func(path string, info os.FileInfo, extension *FileExtension) error {
			if extension == nil {
				if strings.Index(path, "sub1") > 0 {
					// 已扫描完 extension 目录，但不再扫描子目录 sub1。
					return filepath.SkipDir
				}
			}
			return nil
		})

	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 5, len(extensions))
}
