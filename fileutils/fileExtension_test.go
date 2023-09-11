package fileutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExtensionsWithoutConsumer(t *testing.T) {
	option := NewWalkExtensionOption()
	option.CaseSensitive = true

	extensions, err := GetFileExtensions("../test-data/fileutils/extension", option, nil)
	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 8, len(extensions))

	option.CaseSensitive = false
	extensions, err = GetFileExtensions("../test-data/fileutils/extension", option, nil)
	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 4, len(extensions))
}

func TestGetExtensionsWithConsumer(t *testing.T) {
	option := NewWalkExtensionOption()
	option.CaseSensitive = true

	extensions, err := GetFileExtensions("../test-data/fileutils/extension", option,
		func(path string, info os.FileInfo, extension *FileExtension) error {
			// 直接停止，所以结果为空数组。
			return filepath.SkipAll
		})

	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 0, len(extensions))

	extensions, err = GetFileExtensions("../test-data/fileutils/extension", option,
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
	assert.Equal(t, 7, len(extensions))

	extensions, err = GetFileExtensions("../test-data/fileutils/extension", option,
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
	assert.Equal(t, 6, len(extensions))
}

func TestSortExtensions(t *testing.T) {
	fs := []FileExtension{
		{
			Name:  ".txt",
			Count: 1,
			Size:  1000,
			key:   ".txt",
		},
		{
			Name:  ".Txt",
			Count: 4,
			Size:  50,
			key:   ".txt",
		},
		{
			Name:  ".log",
			Count: 5,
			Size:  100,
			key:   ".log",
		},
		{
			Name:  ".md",
			Count: 1,
			Size:  100,
			key:   ".md",
		},
	}

	SortFileExtensionsByName(fs)

	assert.Equal(t, ".log", fs[0].Name)
	assert.Equal(t, ".md", fs[1].Name)
	assert.Equal(t, ".txt", fs[2].Name)
	assert.Equal(t, ".Txt", fs[3].Name)

	SortFileExtensionsByCount(fs)

	assert.Equal(t, ".log", fs[0].Name)
	assert.Equal(t, ".Txt", fs[1].Name)
	assert.Equal(t, ".txt", fs[2].Name)
	assert.Equal(t, ".md", fs[3].Name)

	SortFileExtensionsBySize(fs)

	assert.Equal(t, ".txt", fs[0].Name)
	assert.Equal(t, ".log", fs[1].Name)
	assert.Equal(t, ".md", fs[2].Name)
	assert.Equal(t, ".Txt", fs[3].Name)
}
