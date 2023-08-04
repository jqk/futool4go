package fileutils

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// 一组预定义的文件未满足过滤条件的原因的错误类型。
var (
	ErrReasonIsDir        = errors.New("file is a directory")
	ErrReasonMinSize      = errors.New("file size is less than min size")
	ErrReasonMaxSize      = errors.New("file size is larger than max size")
	ErrReasonInExclude    = errors.New("file name matches exclude")
	ErrReasonNotInInclude = errors.New("file name does not match include")
)

/*
Filter defines conditions to filter files.

Filter 定义了针对文件的过滤条件。
*/
type Filter struct {
	CaseSensitive bool     `mapstructure:"caseSensitive"` // Case sensitive flag. If true, include and exclude patterns are case sensitive.
	Include       []string `mapstructure:"include"`       // Only files matching at least one pattern will be included. Supports glob patterns.
	Exclude       []string `mapstructure:"exclude"`       // Files matching at least one pattern will be excluded. Supports glob patterns.
	MinFileSize   int64    `mapstructure:"minFileSize"`   // Minimum file size in bytes. Files smaller than this will be excluded. 0 means no limit.
	MaxFileSize   int64    `mapstructure:"maxFileSize"`   // Maximum file size in bytes. Files larger than this will be excluded. 0 means no limit.
}

/*
FilteredFileHandler is a function type that receives and processes filtered files.

Parameters:
  - path: Path of the file that meets the filter condition.
  - info: Information of the file that meets the filter condition.

Returns:
  - an error if any, or filepath.SkipDir and filepath.SkipAll to terminate scan.

FilteredFileHandler 是一个函数类型，它接收并处理过滤后文件。

参数:
  - path: 符合过滤条件的文件路径。
  - info: 符合过滤条件的文件信息。

返回:
  - 错误信息，或者 filepath.SkipDir 及 filepath.SkipAll 中断处理。
*/
type FilteredFileHandler func(path string, info os.FileInfo) error

/*
IsRefusedReason checks if the given error is one of the predefined refused reasons.

Parameters:
  - err: the error to check.

Returns:
  - true if the error is one of the predefined refused reasons.

IsRefusedReason 检查给定的错误是否为预定义的拒绝原因。

参数:
  - err: 待检查的错误。

返回:
  - 如果是预定义的拒绝原因，返回 true。
*/
func IsRefusedReason(err error) bool {
	return err == ErrReasonInExclude || err == ErrReasonNotInInclude ||
		err == ErrReasonIsDir || err == ErrReasonMinSize || err == ErrReasonMaxSize
}

/*
GetEachFile scans the specified directory and calls [FilteredFileHandler] to process each file that meets the filter condition.

Parameters:
  - root: The directory to scan.
  - recursive: Whether to scan subdirectories recursively.
  - handler: Callback function to handle files that meet the filter condition. Cannot be nil.

Returns:
  - Error message.

GetEachFile 扫描指定的目录，并调用 [FilteredFileHandler] 处理每个满足过滤条件的文件。

参数:
  - root: 要扫描的目录。
  - recursive: 是否递归扫描子目录。
  - handler: 处理满足过滤条件的文件回调函数。不能为 nil。

返回:
  - 错误信息。
*/
func (f *Filter) GetEachFile(root string, recursive bool, handler FilteredFileHandler) error {
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

		if f.IsQulifiedFile(info) == nil {
			err = handler(path, info)
		}

		return err
	})

	if walkErr != nil && walkErr != filepath.SkipAll && walkErr != filepath.SkipDir {
		return nil
	}

	return walkErr
}

/*
GetFiles returns all file names under the given directory that meet the filter condition.

Parameters:
  - root: The directory to search.
  - recursive: Whether to search subdirectories recursively.

Returns:
  - Array of file names.
  - Error message.

GetFiles 返回所有给定目录下符合过滤条件的文件名。

参数:
  - root: 要搜索的目录。
  - recursive: 是否递归搜索子目录。

返回:
  - 文件名数组。
  - 错误信息。
*/
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

/*
IsQulifiedFile checks whether the given file should meet the filter condition.

Parameters:
  - fileInfo: The file info object.

Returns:
  - Error message. Returns nil if the file meets the filter condition.

IsQulifiedFile 检查给定的文件是否应符合过滤条件。

参数:
  - fileInfo: 文件信息对象。

返回:
  - 错误信息。符合过滤条件返回 nil。
*/
func (f *Filter) IsQulifiedFile(fileInfo os.FileInfo) error {
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
