package fileutils

import (
	"io"
	"os"
	"path/filepath"
)

/*
WalkOption defines the options for walk through a path.
*/
type WalkOption struct {
	Recursive        bool              // whether scan the directory recursively. default is true.
	PathErrorHandler filepath.WalkFunc // error hander when filepath.Walk encounters an error.
}

/*
NewWalkOption creates a new WalkOption with scan directory recursively and bypass permission denied error.

NewWalkOption 创建默认的 WalkOption。包含递归扫描目录及跳过没有权限的文件及目录。
*/
func NewWalkOption() *WalkOption {
	return &WalkOption{ // 默认忽略无权限的文件及目录，并且遍历子目录。
		Recursive: true,
		PathErrorHandler: func(path string, info os.FileInfo, err error) error {
			if os.IsPermission(err) {
				return nil // 跳过没有权限的文件及目录。
			}
			return err
		},
	}
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

Returns:
  - an error if any occurred during the copy process.

CopyDir 复制目录。包含其下的文件和子目录。

参数:
  - source: 要复制的源路径。
  - target: 要复制的目标路径。

返回:
  - 错误信息。
*/
func CopyDir(source, target string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 按相同的目录结构在 target 下创建目录
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		abspath := filepath.Join(target, relPath)
		if info.IsDir() {
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
  - includeSubDir: whether to include sub directories. default is true.

Returns:
  - the number of directories.
  - the number of files.
  - the size of the directory.
  - an error if any occurred during the process.

GetDirStatistics 返回目录统计信息。

参数:
  - dir: 目录路径。
  - includeSubDir: 是否包含子目录。默认为 true。

返回:
  - 目录数量。
  - 文件数量。
  - 目录整体字节大小。
  - 错误信息。
*/
func GetDirStatistics(dir string, option *WalkOption) (stat *DirStatistics, err error) {
	if option == nil { // 保证 option 不为 nil。
		option = NewWalkOption()
	}

	stat = &DirStatistics{}
	isSubDir := false

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if option.PathErrorHandler != nil {
				return option.PathErrorHandler(path, info, err)
			}
			return err
		}

		if info.IsDir() {
			if !option.Recursive {
				// 第一次到达这里，必然是整个函数的参数 dir 目录，所以 isSubDir 为 false。
				if isSubDir {
					return filepath.SkipAll
				}
				// 以后是子目录了，所以设置 isSubDir 为 true。
				isSubDir = true
			}

			stat.DirCount++
		} else {
			stat.FileCount++
			stat.TotalSize += info.Size()
		}

		return nil
	})

	return stat, err
}
