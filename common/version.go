package common

import (
	"regexp"
	"strconv"
	"strings"
)

/*
subVersionInfo 定义子版本号结构。子版本号由数字（可选），字符串后缀组成。
*/
type subVersionInfo struct {
	number int    // 数字部分的值，默认为 0。
	suffix string // 后缀部分的值，默认为空字符串。
}

/*
CompareVersions compares two version numbers.

Parameters:
  - version1: The first version number.
  - version2: The second version number.

Returns:
  - -1: version1 < version2.
  - 0: version1 = version2.
  - 1: version1 > version2.

Example:

	// same length, only compare numbers.
	assert.Equal(t, -1, CompareVersions("1.1.0.20", "1.1.1.5"))
	assert.Equal(t, 1, CompareVersions("1.1.1.20", "1.1.1.5"))
	// case insensitive, numbers are different.
	assert.Equal(t, -1, CompareVersions("1.1a.0", "1.1A.1"))
	// case insensitive, numbers are same.
	assert.Equal(t, 0, CompareVersions("1.1a.0", "1.1A.0"))
	// alphatbit order if numbers are same.
	assert.Equal(t, -1, CompareVersions("1.1-a.1", "1.1-b.1"))
	// 'a' equals to '0a'.
	assert.Equal(t, 1, CompareVersions("1.1.a", "1.1.0"))
	assert.Equal(t, -1, CompareVersions("1.1.a", "1.1.1"))
	// no suffix in subverison is newer.
	assert.Equal(t, -1, CompareVersions("1.1.0", "1.1b.1"))
	// different version length.
	assert.Equal(t, -1, CompareVersions("1.1", "1.1.1"))
	assert.Equal(t, 0, CompareVersions("1.1", "1.1.0"))
	// different version length with newer subversion.
	assert.Equal(t, 1, CompareVersions("1.2", "1.1.1"))
	// prefix and suffix dots are trimed.
	// spaces in subversion are ignored.
	assert.Equal(t, 0, CompareVersions(" 1. 2 .", ".1. 2"))
	// prefix is treated as a string.
	// 'v1' is a string, and it's subversion number is 0,
	// compared with '1' in second parameter, which subversion number is 1.
	assert.Equal(t, -1, CompareVersions("v1.1", "1.1"))
	// ' -234' is trimed and treated as string '-234'.
	assert.Equal(t, 0, CompareVersions("1.1 -234", "1.1-234"))

CompareVersion 比较两个版本号。版本号必须以 "." 分隔。

参数:
  - version1: 第一个版本号。
  - version2: 第二个版本号。

返回:
  - -1: version1 < version2。
  - 0: version1 = version2。
  - 1: version1 > version2。
*/
func CompareVersions(version1, version2 string) int {
	// 去年前后的 "."，并以 "." 作为分隔符分离成字符串数组，即子版本号数组。
	subVerionStrings1 := strings.Split(strings.Trim(version1, "."), ".")
	subVerionStrings2 := strings.Split(strings.Trim(version2, "."), ".")

	// 使用子版本号数量较大的值。
	count := len(subVerionStrings1)
	temp := len(subVerionStrings2)
	if count < temp {
		count = temp
	}

	subVersions1 := getSubVersions(subVerionStrings1, count)
	subVersions2 := getSubVersions(subVerionStrings2, count)

	for i := 0; i < count; i++ {
		// 从左侧开始逐一比较子版号。先比较数字部分，再比较后缀部分。
		if subVersions1[i].number < subVersions2[i].number {
			return -1
		} else if subVersions1[i].number > subVersions2[i].number {
			return 1
		} else if subVersions1[i].suffix != subVersions2[i].suffix {
			return strings.Compare(subVersions1[i].suffix, subVersions2[i].suffix)
		}
	}

	return 0
}

/*
getSubVersions 解析子版本号数组。

参数:
  - subVersions: 子版本号数组。
  - count: 数组长度。必须大于等于 subVersions 的长度。

返回:
  - 子版本号信息数组。
*/
func getSubVersions(subVersions []string, count int) []*subVersionInfo {
	result := make([]*subVersionInfo, count)

	for i, s := range subVersions {
		result[i] = getSubVersionInfo(strings.TrimSpace(s))
	}

	for i := len(subVersions); i < count; i++ {
		// 使用默认值补足空位。此处填写了结构内的字段，与默认值相同，仅为明确标识值。
		result[i] = &subVersionInfo{
			number: 0,
			suffix: "",
		}
	}

	return result
}

// regexSubVersion 是用于解析子版本号的正则表达式：
// 1. 可以没有数字。
// 2. 若有数字，可以有字符串后缀。
var regexSubVersion = regexp.MustCompile(`^(\d*)(.*)`)

// getSubVersionInfo 解析子版本号。
func getSubVersionInfo(s string) *subVersionInfo {
	vers := regexSubVersion.FindStringSubmatch(s)
	result := &subVersionInfo{}

	// 如果是空字符串，取得的值是 0。
	result.number, _ = strconv.Atoi(strings.TrimSpace(vers[1]))
	// 去除前后空格，并转小写。
	result.suffix = strings.ToLower(strings.TrimSpace(vers[2]))
	return result
}
