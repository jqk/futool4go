package fileutils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileExtension describes file extension information.
type FileExtension struct {
	Name string // Name of the file extension, including the dot.
	// For example:
	//   ".txt"
	//   ".html"
	//   "" means no extenion.
	Count int    // occurrence count
	Size  int64  // total file size
	key   string // key is an internal key used for sorting
}

// FileExtensionConsumer is a function type that provides the FileExtension and consumes them.
//
// Takes in file's full path and its info object, and FileExtension object.
// FileExtension is nil means path is a directory.
//
// Returns an error if any, or filepath.SkipDir and filepath.SkipAll to terminate scan.
type FileExtensionConsumer func(path string, info os.FileInfo, extension *FileExtension) error

// NewFileExtension creates a new FileExtension object with the given file extension.
//
// Parameters:
// - extension: the name of the file extension, including the dot. "" means no extenion.
//
// Returns:
// - *FileExtension: a pointer to the newly created FileExtension object.
func NewFileExtension(extension string) *FileExtension {
	return &FileExtension{Name: extension, Count: 0, Size: 0, key: strings.ToLower(extension)}
}

// GetFileExtensions returns a slice of FileExtension structs by walking through the
// directory tree rooted at a given path and counting the files of each unique
// extension type.
//
// The function takes three arguments: the path string and a bool
// value indicating whether the extension names should be case sensitive or not.
// The last one is the consumer function, could be nil.
//
// It returns a unsorted slice of FileExtension structs and an error value.
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

	return extensions, nil
}

// SortFileExtensionsByName sorts the given list of FileExtension structs by name, asec.
//
// The function takes a slice of FileExtension structs as the parameter.
//
// The function modifies the given slice in-place.
func SortFileExtensionsByName(extensions []FileExtension) {
	sort.Slice(extensions, func(i, j int) bool {
		key_i := extensions[i].key
		key_j := extensions[j].key

		if key_i == key_j {
			// 这样做可以在区分大小写的情况下将 key 相同但大小写不同的扩展名排在一起。
			return extensions[i].Name > extensions[j].Name
		}

		// 升序。
		return key_i < key_j
	})
}

// SortFileExtensionsByCount sorts the given list of file extensions by count in descending order.
//
// The function takes a slice of FileExtension structs as input. The FileExtension struct should have
// two fields: Count (which represents the count of the file extension) and Size (which represents the
// size of the file extension). The function sorts the list of file extensions based on the count in
// descending order. If the count of two file extensions is the same, it compares the size in descending
// order. The function modifies the given slice in-place.
func SortFileExtensionsByCount(extensions []FileExtension) {
	sort.Slice(extensions, func(i, j int) bool {
		count_i := extensions[i].Count
		count_j := extensions[j].Count

		if count_i == count_j {
			return extensions[i].Size > extensions[j].Size
		}

		// 降序。
		return count_i > count_j
	})
}

// SortFileExtensionsBySize sorts the given list of file extensions by size in descending order.
//
// The function takes in a slice of FileExtension structs as the parameter. Each FileExtension struct
// represents a file extension and contains the size and count of files with that extension. The function
// sorts the extensions based on their size, with larger sizes appearing first. If two extensions have the
// same size, the function sorts them based on their count in descending order.
//
// The function modifies the given slice in-place.
func SortFileExtensionsBySize(extensions []FileExtension) {
	sort.Slice(extensions, func(i, j int) bool {
		size_i := extensions[i].Size
		size_j := extensions[j].Size

		if size_i == size_j {
			return extensions[i].Count > extensions[j].Count
		}

		// 降序。
		return size_i > size_j
	})
}
