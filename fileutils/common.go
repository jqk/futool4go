package fileutils

import (
	"io"
	"os"
	"path/filepath"
)

/*
WalkOption defines the options for walk through a path.
See [NewWalkOption] for default settings.
*/
type WalkOption struct {
	/*
		whether scan the directory recursively. It is called indirectly like this:

		filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			....

			if info.IsDir() {
				if option.ShouldQuitForNonRecursive() {
					return filepath.SkipAll
				}

				....
			｝
			....
		})
	*/
	Recursive bool
	/*
		error hander when filepath.Walk encounters an error. It is only called like this:

		filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if option.PathErrorHandler != nil {
					return option.PathErrorHandler(path, info, err)
				}
				return err
			}

			....
		}
	*/
	PathErrorHandler filepath.WalkFunc

	isSubDir bool // 默认为 false。初始必须为 false。
}

/*
ShouldQuitForNonRecursive returns true if the current path should be skipped.

It alwasy returns false when WalkOption.Recursive is true.

When WalkOption.Recursive is false:
  - The first call to the function returns false.
  - Subsequent calls will all return true.

ShouldQuitForNonRecursive 返回是否需要跳过当前路径。

WalkOption.Recursive 为 true 时始终返回 false。

WalkOption.Recursive 为 false:
  - 第一次调用返回 false。
  - 后续调用都返回 true。
*/
func (option *WalkOption) ShouldQuitForNonRecursive() bool {
	if option.Recursive {
		return false
	}

	// 第一次到达这里，必然是整个 filepath.Walk() 函数的起始目录，所以 isSubDir 为 false。
	// 以后就是子目录了，再到这里，就是 true 了。
	if option.isSubDir {
		return true
	}

	// 以后到达这里都是子目录了，所以设置 isSubDir 为 true。
	option.isSubDir = true
	return false
}

/*
NewWalkOption creates a new WalkOption with scan directory recursively and bypass permission denied error.

NewWalkOption 创建默认的 WalkOption。包含递归扫描目录及跳过没有权限的文件及目录。
*/
func NewWalkOption() *WalkOption {
	return &WalkOption{
		Recursive:        true,
		PathErrorHandler: SkipPermissionError,
	}
}

/*
SkipPermissionError is an example for WalkOption.PathErrorHandler.
It is used for skipping permission denied error.
*/
func SkipPermissionError(path string, info os.FileInfo, err error) error {
	// 仅在 err 不为 nil 时被调用，所以不必检查该值。
	if os.IsPermission(err) {
		return nil // 跳过没有权限的文件及目录。
	}
	return err
}

/*
FileExists checks if a file or directory exists.

Parameters:
  - path: string representing the path to check. can be file or directory.

Returns:
  - a bool indicating if the file/directory exists.
  - a bool indicating it's a directory, true for directory and false for file.
  - an error if any occurred during the process.

FileExists 查看给定的路径是否存在，可以是文件名或目录名。

参数:
  - path: 要检查的路径。

返回:
  - 文件或目录是否存在。
  - path 是否为目录。
  - 错误信息。
*/
func FileExists(path string) (bool, bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return true, info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, false, nil
	}
	return false, false, err
}

/*
CopyDir copies the directory and its contents from the source path to the target path.

Parameters:
  - source: the source path of the directory to be copied.
  - target: the target path where the directory and its contents will be copied to.
  - option: the scan options. if nil, the default options will be used.

Returns:
  - an error if any occurred during the copy process.

CopyDir 复制目录。包含其下的文件和子目录。

参数:
  - source: 要复制的源路径。
  - target: 要复制的目标路径。
  - option: 扫描选项。如果为 nil 则使用默认选项。

返回:
  - 错误信息。
*/
func CopyDir(source, target string, option *WalkOption) error {
	if option == nil { // 保证 option 不为 nil。
		option = NewWalkOption()
	}

	walkErr := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if option.PathErrorHandler != nil {
				return option.PathErrorHandler(path, info, err)
			}
			return err
		}
		// 按相同的目录结构在 target 下创建目录
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		abspath := filepath.Join(target, relPath)

		if info.IsDir() {
			if option.ShouldQuitForNonRecursive() {
				return filepath.SkipAll
			}

			os.MkdirAll(abspath, os.ModePerm)
		} else {
			// 复制文件
			from, err := os.Open(path)
			if err != nil {
				return err
			}

			to, err := os.Create(abspath)
			if err != nil {
				from.Close()
				return err
			}

			_, err = io.Copy(to, from)
			from.Close()
			to.Close()
			if err != nil {
				return err
			}
		}

		return nil
	})

	if walkErr == filepath.SkipAll || walkErr == filepath.SkipDir {
		walkErr = nil
	}

	return walkErr
}

/*
DirStatistics defines the statistics of a directory.
*/
type DirStatistics struct {
	DirCount  int
	FileCount int
	TotalSize int64
}

/*
GetDirStatistics returns the statistics of a directory.

Parameters:
  - dir: the directory path.
  - option: the scan options. if nil, the default options will be used.

Returns:
  - the statistics of the directory.
  - an error if any occurred during the process.

GetDirStatistics 返回目录统计信息。

参数:
  - dir: 目录路径。
  - option: 扫描选项。如果为 nil 则使用默认选项。

返回:
  - 目录统计信息。
  - 错误信息。
*/
func GetDirStatistics(dir string, option *WalkOption) (stat *DirStatistics, err error) {
	if option == nil { // 保证 option 不为 nil。
		option = NewWalkOption()
	}

	stat = &DirStatistics{}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if option.PathErrorHandler != nil {
				return option.PathErrorHandler(path, info, err)
			}
			return err
		}

		if info.IsDir() {
			if option.ShouldQuitForNonRecursive() {
				return filepath.SkipAll
			}

			stat.DirCount++
		} else {
			stat.FileCount++
			stat.TotalSize += info.Size()
		}

		return nil
	})

	return stat, FilterFilePathSkipErrors(err)
}

/*
FilterFilePathSkipErrors returns nil if the error is SkipAll, SkipDir or nil. Otherwise, the given error is returned.

FilterFilePathSkipErrors 如果错误为 SkipAll、SkipDir 或 nil 则返回 nil；否则，直接返回给定的错误参数。
*/
func FilterFilePathSkipErrors(err error) error {
	if err == filepath.SkipAll || err == filepath.SkipDir {
		return nil
	}
	return err
}
