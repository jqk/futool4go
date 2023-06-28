package fileutils

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

var ErrReasonIsDir = errors.New("file is a directory")
var ErrReasonMinSize = errors.New("file size is less than min size")
var ErrReasonMaxSize = errors.New("file size is larger than max size")
var ErrReasonInExclude = errors.New("file name matches exclude")
var ErrReasonNotInInclude = errors.New("file name does not match include")

type Filter struct {
	CaseSensitive bool     `mapstructure:"caseSensitive"`
	Include       []string `mapstructure:"include"`
	Exclude       []string `mapstructure:"exclude"`
	MinFileSize   int64    `mapstructure:"minFileSize"`
	MaxFileSize   int64    `mapstructure:"maxFileSize"`
}

// FilterConsumer is a function type that filters file and consumes them.
//
// Takes in file's full path and its info object.
//
// Returns an error if any, or filepath.SkipDir and filepath.SkipAll to terminate scan.
type FilterConsumer func(path string, info os.FileInfo) error

// IsRefusedReason checks if the given error is one of the refused reasons.
//
// err - the error to be checked. It returned by AcceptFile()，
// which is usally not treated as an error.
//
// Returns a boolean indicating whether the error is a refused reason or not.
func IsRefusedReason(err error) bool {
	return err == ErrReasonInExclude || err == ErrReasonNotInInclude ||
		err == ErrReasonIsDir || err == ErrReasonMinSize || err == ErrReasonMaxSize
}

// GetEachFile scans the specified directory and applies the FilterComsumer function
// to every file that matches the filter. The recursive parameter specifies if
// subdirectories should also be scanned. Returns an error if the Filter is invalid
// or if there is an error scanning the directory.
//
// root: the root directory to scan.
//
// recursive: whether to scan subdirectories.
//
// comsumer: the function to apply to each file that matches the filter.
//
// Returns an error if there is one.
func (f *Filter) GetEachFile(root string, recursive bool, comsumer FilterConsumer) error {
	// 先保证 Filter 中的配置项有效。
	if err := f.Validate(); err != nil {
		return err
	}

	skipAll := false

	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
			// 在 Walk() 中，对每个目录都重生以下操作：先给目录，再给目录下的文件，最后给目录下的子目录。
			if !recursive {
				// 第一次走到这里，就是 Root 目录。
				// 第二次走到这里，就是 Root 目录下的第一个子目录，依此类推。
				// 在两次到达此处之间，上一次目录下的文件都会在后面的逻辑中处理。
				if skipAll {
					return filepath.SkipAll
				}

				// 第一次时为 false，此处设为 ture，第二次是就直接返回 SkipAll 了。
				skipAll = true
			}
			return nil
		}

		if f.AcceptFile(info) == nil {
			err = comsumer(path, info)
		}

		return err
	})

	return walkErr
}

// GetFiles returns a slice of files from a given root directory and boolean
// flag indicating whether or not to include subdirectories. The function
// preallocates space to avoid multiple expansions of the slice. It returns a
// pointer to the resulting slice and an error, if any.
//
// root: the root directory to search for files
//
// recursive: a boolean flag indicating whether or not to include
// subdirectories in the search
//
// *[]string, error: a pointer to a slice of filename and an error, if any
func (f *Filter) GetFiles(root string, recursive bool) (*[]string, error) {
	// 遍历目录可能获得多个文件，为避免过多的对数组进行扩展，预分配空间。
	// 也不必过大，因为毕竟有时返回数量也较小。。
	result := make([]string, 0, 1000)

	err := f.GetEachFile(root, recursive, func(path string, info os.FileInfo) error {
		result = append(result, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (f *Filter) AcceptFile(fileInfo os.FileInfo) error {
	if fileInfo.IsDir() {
		return ErrReasonIsDir
	} else if fileInfo.Size() < f.MinFileSize && f.MinFileSize > 0 {
		return ErrReasonMinSize
	} else if fileInfo.Size() > f.MaxFileSize && f.MaxFileSize > 0 {
		return ErrReasonMaxSize
	}

	filename := fileInfo.Name()
	if !f.CaseSensitive {
		filename = strings.ToLower(filename)
	}

	ext := filepath.Ext(filename)

	for _, pattern := range f.Exclude {
		if matchPattern(pattern, filename, ext) {
			// 在 Exclude 中，不合格。
			return ErrReasonInExclude
		}
	}

	for _, pattern := range f.Include {
		if matchPattern(pattern, filename, ext) {
			// 在 Include 中，合格。
			return nil
		}
	}

	return ErrReasonNotInInclude
}

func (f *Filter) Diff(other *Filter) string {
	if f == other {
		return ""
	}
	if f.CaseSensitive != other.CaseSensitive {
		return "Filter.CaseSensitive"
	}
	if f.MaxFileSize != other.MaxFileSize {
		return "Filter.MaxFileSize"
	}
	if f.MinFileSize != other.MinFileSize {
		return "Filter.MinFileSize"
	}
	if !reflect.DeepEqual(f.Include, other.Include) {
		return "Filter.Include"
	}
	if !reflect.DeepEqual(f.Exclude, other.Exclude) {
		return "Filter.Exclude"
	}

	return ""
}

func (f *Filter) Validate() error {
	if f.MaxFileSize < 0 {
		f.MaxFileSize = 0
	}
	if f.MinFileSize < 0 {
		f.MinFileSize = 0
	}

	if f.MinFileSize > f.MaxFileSize && f.MaxFileSize != 0 {
		return errors.New("Filter.MaxFileSize must be greater than or equal to Filter.MinFileSize")
	}

	if exts, err := validateExtensions(&f.Exclude, f.CaseSensitive); err != nil {
		return err
	} else {
		f.Exclude = *exts
	}

	if exts, err := validateExtensions(&f.Include, f.CaseSensitive); err != nil {
		return err
	} else {
		f.Include = *exts
	}

	if len(f.Include) == 0 {
		return errors.New("Filter.Include must not be empty")
	}

	return nil
}

func validateExtensions(exts *[]string, caseSensitive bool) (*[]string, error) {
	// 使用 map 是为了过滤掉相同的扩展名。
	extMap := make(map[string]bool, len(*exts))

	for _, ext := range *exts {
		ext = strings.TrimSpace(ext)
		if !caseSensitive {
			ext = strings.ToLower(ext)
		}

		// 预先调用 Match()，可以提前发现 ext 格式是否正确。
		if matched, err := filepath.Match(ext, ""); err != nil {
			return nil, err
		} else {
			extMap[ext] = matched
		}
	}

	result := make([]string, 0, len(extMap))
	for ext := range extMap {
		result = append(result, ext)
	}

	sort.Strings(result)
	return &result, nil
}

func matchPattern(pattern string, filename string, ext string) bool {
	// 在调用本函数之前，应保证 Include 和 Exclude 已使用 Validate() 校验过了。
	// 这样 pattern 都是有效的。所以 Match() 不会返回 error，即无需处理。
	if pattern == "" && ext == "" {
		// 文件没有扩展名时 ext 为空字符串，而 pattern 日空字符串，两者匹配。
		return true
	} else if matched, _ := filepath.Match(pattern, filename); matched {
		return true
	}

	return false
}
