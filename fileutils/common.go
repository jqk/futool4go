package fileutils

import (
	"io"
	"os"
	"path/filepath"
)

// FileExists checks if a file or directory exists at the given path.
//
// path: string representing the path to check.
//
// Returns a bool indicating if the file/directory exists, a bool indicating
// if it's a directory, and an error if any.
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
