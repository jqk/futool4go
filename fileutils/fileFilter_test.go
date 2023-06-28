package fileutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var filter *Filter = &Filter{
	CaseSensitive: false,
	Include: []string{
		"*.md",
		"*.txt",
		"",
	},
	Exclude: []string{
		"*.log",
	},
	MinFileSize: 1024,
	MaxFileSize: 3000,
}

var testPath = "../test-data/fileutils/filter"

func TestGetEachFileIncludingSubDir(t *testing.T) {
	result := make(map[string]bool)
	filter.CaseSensitive = false

	err := filter.GetEachFile(testPath, true, func(path string, info os.FileInfo) error {
		result[info.Name()] = true
		return nil
	})

	assert.Nil(t, err)
	// 文件名前 3 位是 001 至 005。
	assert.Equal(t, 5, len(result))

	result = make(map[string]bool)
	filter.CaseSensitive = true

	err = filter.GetEachFile(testPath, true, func(path string, info os.FileInfo) error {
		result[info.Name()] = true
		return nil
	})

	assert.Nil(t, err)
	// 大小写敏感，将过滤掉两个文件。
	// 文件名前 3 位是 002 至 004。
	assert.Equal(t, 3, len(result))
}

func TestGetEachFileExcludingSubDir(t *testing.T) {
	result := make(map[string]bool)
	filter.CaseSensitive = false

	err := filter.GetEachFile(testPath, false, func(path string, info os.FileInfo) error {
		result[info.Name()] = true
		return nil
	})

	assert.Nil(t, err)
	// 文件名前 3 位是 001 至 004。005 在 sub 中，未遍历。
	assert.Equal(t, 4, len(result))

	result = make(map[string]bool)
	filter.CaseSensitive = true

	err = filter.GetEachFile(testPath, false, func(path string, info os.FileInfo) error {
		result[info.Name()] = true
		return nil
	})

	assert.Nil(t, err)
	// 大小写敏感，将过滤掉一个文件。
	assert.Equal(t, 3, len(result))
}

func TestGetEachFileSkipDir(t *testing.T) {
	result := make(map[string]bool)
	filter.CaseSensitive = false
	count := 0

	err := filter.GetEachFile(testPath, true, func(path string, info os.FileInfo) error {
		result[info.Name()] = true
		count++

		if count == 2 {
			// 查找两个文件后再放弃遍历。
			return filepath.SkipDir
		}

		return nil
	})

	assert.Nil(t, err)
	// 文件 001、002。
	assert.Equal(t, count, len(result))
}

func TestGetFiles(t *testing.T) {
	filter.CaseSensitive = false

	result, err := filter.GetFiles(testPath, true)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(*result))
}
