package fileutils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileExtension describes file extension information.
//
// FileExtension 描述文件扩展名信息。
type FileExtension struct {
	/*
		Name of the file extension, including the dot.

		For example:
			".txt"
			".html"
			"" means no extension.
	*/
	Name  string
	Count int    // occurrence count
	Size  int64  // total file size in byte
	key   string // key is an internal key used for sorting
}

/*
FileExtensionConsumer is a function type that is called when traversing the given path.
It is mainly used to notify the caller about each file and directory processed when traversing the path.
The traversal can be terminated by returning SkipDir and SkipAll.

Parameters:
  - path: The path being processed. Can be a directory or file.
  - info: Information of the file being processed.
  - extension: Extension information of the file being processed. nil indicates it is a directory.

Returns:
  - Error message.

FileExtensionConsumer 是在遍历给定路径时被调用的函数类型。主要用于遍历路径下的文件时将处理的每个文件和目录通知调用者。
可以通过返回 SkipDir 和 SkipAll 终止遍历。

参数：
  - path: 当前处理路径。可能是目录或文件。
  - info: 当前处理的文件信息。
  - extension: 当前处理的文件扩展名信息。为 nil 表示当前处理的是目录。

返回：
  - 错误信息。
*/
type FileExtensionConsumer func(path string, info os.FileInfo, extension *FileExtension) error

/*
NewFileExtension creates a new [FileExtension] object with the given file extension.

Parameters:
  - extension: the name of the file extension, including the dot. empty string means no extenion.

Returns:
  - a pointer to the newly created [FileExtension] object.

NewFileExtension 创建 [FileExtension] 对象。

参数:
  - extension: 文件扩展名，包括点(.)。空字符串表示没有扩展名。

返回：
  - 指向新创建的 [FileExtension] 对象的指针。
*/
func NewFileExtension(extension string) *FileExtension {
	return &FileExtension{Name: extension, Count: 0, Size: 0, key: strings.ToLower(extension)}
}

/*
GetFileExtensions scans and collects extension information of all files under the given path.

Parameters:
  - path: Path to be scanned.
  - caseSensitive: Whether to distinguish case for extensions.
  - consumer: This function will be invoked whenever a new file or directory is processed to notify the caller. Can be nil.

Returns:
  - An unsorted array of [FileExtension].
  - nil if processed successfully, otherwise the error message.

GetFileExtensions 扫描并统计给定路径下所有文件的扩展名信息。

参数:
  - path: 待扫描的路径。
  - caseSensitive: 扩展名是否区分大小写。
  - consumer: 每处理一个新的文件或目录都将尝试调用该函数，从而通知调用者。可为 nil。

返回:
  - 未经排序的文件扩展名信息数组。
  - 处理正常时为 nil，否则为错误信息。
*/
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
			if os.IsPermission(err) {
				return nil // 没有权限则跳过。
			}
			return err
		} else if info.IsDir() {
			if consumer != nil {
				return consumer(path, info, nil) // 将开始处理新目录通知外部调用者。
			}
			return nil
		}

		ext := filepath.Ext(path)
		if !caseSensitive {
			ext = strings.ToLower(ext)
		}

		if _, ok := extMap[ext]; !ok {
			extMap[ext] = NewFileExtension(ext) // 该扩展名第一次出现，创建对象。
		}

		extMap[ext].Count++
		extMap[ext].Size += info.Size()

		if consumer != nil {
			return consumer(path, info, extMap[ext]) // 将处理新文件通知外部调用者。
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

/*
SortFileExtensionsByName sorts the given list of [FileExtension] objects by name, asec. The function modifies the given slice in-place.

Parameters:
  - extensions: a slice of [FileExtension] objects.

SortFileExtensionsByName 按名称升序排列。将直接修改给定的切片。

参数：
  - extensions: 待排序的 [FileExtension] 数组。
*/
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

/*
SortFileExtensionsByCount sorts the given list of [FileExtension] objects by count, desc. The function modifies the given slice in-place.

Parameters:
  - extensions: a slice of [FileExtension] objects.

SortFileExtensionsByCount 按数量降序排列。将直接修改给定的切片。

参数：
  - extensions: 待排序的 [FileExtension] 数组。
*/
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

/*
SortFileExtensionsBySize sorts the given list of [FileExtension] objects by total file size, desc. The function modifies the given slice in-place.

Parameters:
  - extensions: a slice of [FileExtension] objects.

SortFileExtensionsBySize 按文件大小降序排列。将直接修改给定的切片。

参数：
  - extensions: 待排序的 [FileExtension] 数组。
*/
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
