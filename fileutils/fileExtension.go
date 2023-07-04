package fileutils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileExtension struct {
	Name  string
	Count int
	Size  int64
	key   string
}

// FileExtensionConsumer is a function type that provides the FileExtension and consumes them.
//
// Takes in file's full path and its info object, and FileExtension object.
// FileExtension is nil means path is a directory.
//
// Returns an error if any, or filepath.SkipDir and filepath.SkipAll to terminate scan.
type FileExtensionConsumer func(path string, info os.FileInfo, extension *FileExtension) error

func NewFileExtension(name string) *FileExtension {
	return &FileExtension{Name: name, Count: 0, Size: 0, key: strings.ToLower(name)}
}

// GetFileExtensions returns a slice of FileExtension structs by walking through the
// directory tree rooted at a given path and counting the files of each unique
// extension type.
//
// The function takes three arguments: the path string and a bool
// value indicating whether the extension names should be case sensitive or not.
// The last one is the consumer function, could be nil.
//
// It returns a sorted slice of FileExtension structs and an error value.
func GetFileExtensions(path string, caseSensitive bool, consumer FileExtensionConsumer) ([]FileExtension, error) {
	pathExists, isDir, outerErr := FileExists(path)
	if outerErr != nil {
		return nil, outerErr
	} else if !pathExists {
		return nil, fmt.Errorf("path does not exist: %s", path)
	} else if !isDir {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}

	// 使用 map 主要是为了合并同名扩展名，统计各个扩展名出现的次数。
	extMap := make(map[string]*FileExtension)

	outerErr = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
			if consumer != nil {
				return consumer(path, info, nil)
			}
			return nil
		}

		ext := filepath.Ext(path)
		if !caseSensitive {
			ext = strings.ToLower(ext)
		}

		if _, ok := extMap[ext]; !ok {
			// 该扩展名第一次出现，创建对象。
			extMap[ext] = NewFileExtension(ext)
		}

		extMap[ext].Count++
		extMap[ext].Size += info.Size()

		if consumer != nil {
			return consumer(path, info, extMap[ext])
		}

		return nil
	})

	if outerErr != nil && outerErr != filepath.SkipAll && outerErr != filepath.SkipDir {
		return nil, outerErr
	}

	// 将 map 中的内容保存到数组中。
	extensions := make([]FileExtension, 0, len(extMap))
	for _, ext := range extMap {
		extensions = append(extensions, *ext)
	}

	// 排序。
	sort.Slice(extensions, func(i, j int) bool {
		key_i := extensions[i].key
		key_j := extensions[j].key

		if key_i == key_j {
			// 这样做可以在区分大小写的情况下将 key 相同但大小写不同的扩展名排在一起。
			return extensions[i].Name > extensions[j].Name
		}

		return key_i < key_j
	})

	return extensions, nil
}
