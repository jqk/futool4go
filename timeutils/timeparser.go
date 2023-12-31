package timeutils

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

/*
RequireDateTimeFieldValid defines whether to require the date time fields to be within valid ranges.
Default is true. This is a global setting and will affect all subsequent operations.

Example:

	RequireDateTimeFieldValid = true
	tm = ParseDateTime("2010-13-23")	// nil because the month is 13. it is out of range.

	RequireDateTimeFieldValid = false
	tm = ParseDateTime("2010-13-23")	// 2011-01-23 00:00:00. It is what GO is going to parse.

RequireDateTimeFieldValid 定义是否要求日期时间各字段的值都在范围内。默认为 true。
该值为全局设置，会影响后续所有操作。
*/
var RequireDateTimeFieldValid = true

// regexUnixTime 是用于 UnixTime 的正则表达式：
// 1. 可以有字符前缀及后缀。
// 2. 需要至少 10 位数字代表秒数。
// 3. 可以再有 3 位数字代表毫秒数。不足不算。
var regexUnixTime = regexp.MustCompile(`^.*?(\d{1,10})(\d{3})?.*`)

/*
ParseUnixTime separates consecutive 1 to 13 digit numbers from the input string and
converts them to time variables using Unix time format.
Seconds with 1 to 10 digits. Followed by milliseconds with 3 digits, insufficient digits are ignored.

Parameters:
  - s: The string to parse

Returns:
  - The parsed time. nil is returned on failure.

Example:

	tm := ParseUnixTime("snapshot_1553867509757.png") // 2019-03-29 21:51:49.757
	tm = ParseUnixTime("155386750975abcd")            // 2019-03-29 21:51:49.000

ParseUnixTime 从字符串中分离出连续的 1 到 13 位的数字，并将其按 Unix 时间格式转转为时间变量。
秒数 1 到 10 位。后续紧跟毫秒数 3 位，不足不算。

参数:
  - s: 待解析的字符串。

返回:
  - 解析后的时间。失败均返回 nil。
*/
func ParseUnixTime(s string) *time.Time {
	subs := regexUnixTime.FindStringSubmatch(s)
	count := len(subs)

	if count <= 1 {
		// 没有配置的 unix 时间截字符串。
		return nil
	}

	// 第 1 个匹配是 10 位，代表秒数。到此处必定存在。
	var nanosecond int64 = 0
	second, _ := strconv.ParseInt(subs[1], 10, 64)

	// 第 2 个匹配是 3 位，代表毫秒数，要转换为纳秒。可能不存在。
	if count > 2 {
		nanosecond, _ = strconv.ParseInt(subs[2], 10, 64)
		nanosecond *= 1000_000
	}

	result := time.Unix(second, nanosecond).In(time.Local)
	return &result
}

// regexDateTimeNoSep 是用于无分隔符的日期时间正则表达式：
//  1. 可以有字符前缀及后缀。
//  2. 日期数字之间无分隔符，需要至少 8 位数字表示 YYYYMMDD。
//  3. 日期与时间之间可以有为分隔符，也可以无分隔符。分隔符可以是“_”、“-”、“.”、“ ”和“T”。
//  4. 时间数字之间无分隔符，可以是 4 位数字，6 位数字，或者 9 位数字。
//     分别表示 HHMM，HHMMSS 及 HHMMSSSS。也就是说，可以精确到分钟、秒或毫秒。
//  5. 毫秒数为 3 位，与秒数之间可以有“.”作为分隔符，也可以无分隔符。
//
// note: 最后的 (\.?(\d{3}))? 不要外圈这括号也行，但加上后解析结果数组与 regexDateTimeHasSep 一致。
var regexDateTimeNoSep = regexp.MustCompile(
	`^.*?(\d{4})(\d{2})(\d{2})[-|_|\.| |T]?` +
		`(\d{2})(\d{2})((\d{2})(\.?(\d{3}))?)?.*`)

// regexDateTimeHasSep 是用于有分隔符的日期时间正则表达式：
//  1. 可以有字符前缀及后缀。
//  2. 日期数字之间有分隔符，年 4 位，月 1 或 2 位，日 1 或 2 位。
//  3. 日期与时间之间必须有为分隔符，可以是“_”、“-”、“.”、“ ”和“T”。
//  4. 时间数字之间有分隔符，可以是“_”、“-”、“.”、“:”。秒与毫秒之间只能是“.”。
//     小时 1 或 2 位，分钟和秒都是 2 位，毫秒是 3 位。可以精确到分钟、秒或毫秒。
var regexDateTimeHasSep = regexp.MustCompile(
	`^.*?(\d{4})[-|_|\.|](\d{1,2})[-|_|\.|](\d{1,2})[-|_|\.| |T]` +
		`(\d{1,2})[-|_|\.|\:|](\d{2})([-|_|\.|\:|](\d{2})(\.(\d{3}))?)?.*`)

/*
ParseDateTime parses date time strings into time variables.

Parameters:
  - s: The string to parse. Milliseconds must be 3 digits, otherwise that value is not parsed.

Returns:
  - The parsed time. nil is returned on failure.

Example:

	// no seperator between fields.
	tm = ParseDateTime("abc20100223-1534ddd.jpg")      // 2010-02-23 15:34:00
	tm = ParseDateTime("20100223.153456.789ddd.jpg")   // 2010-02-23 15:34:56.789
	tm = ParseDateTime("20100223153456789")            // 2010-02-23 15:34:56.789
	tm = ParseDateTime("20100223 1534567")             // 2010-02-23 15:34:56.000

	tm = ParseDateTime("20100223")                     // nil because no time fields.
	tm = ParseDateTime("201022231234")                 // nil because month is 22. it's out of range.

	// has seperator between fields.
	tm = ParseDateTime("abc2010-02-23-15:34ddd.jpg")   // 2010-02-23 15:34:00
	tm = ParseDateTime("2010_2-23.5.34.56.789ddd.jpg") // 2010-02-23 05:34:56.789
	tm = ParseDateTime("2010.02.23T15-34_56.789")      // 2010-02-23 15:34:56.789
	tm = ParseDateTime("2010-02-23 15:34:56.7")        // 2010-02-23 15:34:56.000

ParseDateTime 将日期时间字符串转换为时间变量。

参数:
  - s: 待解析的字符串。毫秒必需是 3 位，否则不解析该值。

返回:
  - 解析后的时间。失败均返回 nil。
*/
func ParseDateTime(s string) *time.Time {
	parse := func(s string, regex *regexp.Regexp) *time.Time {
		subs := regex.FindStringSubmatch(s)
		if len(subs) == 0 {
			// 没有配置的日期时间字符串，所以数组长度为 0，返回 nil 说明转换不成功。
			return nil
		}

		year, _ := strconv.Atoi(subs[1])
		m, _ := strconv.Atoi(subs[2])
		month := time.Month(m)
		day, _ := strconv.Atoi(subs[3])
		hour, _ := strconv.Atoi(subs[4])
		minute, _ := strconv.Atoi(subs[5])
		// subs[6] 包含了秒和毫秒。
		second, _ := strconv.Atoi(subs[7])

		if IsDateTimeFieldValid(year, m, day, hour, minute, second) != nil {
			return nil
		}

		// subs[8] 包含了"."和毫秒。
		millisecond, _ := strconv.Atoi(subs[9])
		nanosecond := millisecond * 1000_000

		result := time.Date(year, month, day, hour, minute, second, nanosecond, time.Local)
		return &result
	}

	result := parse(s, regexDateTimeHasSep)
	if result != nil {
		return result
	}

	return parse(s, regexDateTimeNoSep)
}

// regexDateHasSep 是用于有分隔符的日期正则表达式：
//  1. 可以有字符前缀及后缀。
//  2. 日期数字之间有分隔符，年 4 位，月 1 或 2 位，日 1 或 2 位。
var regexDateHasSep = regexp.MustCompile(`^.*?(\d{4})[-|_|\.|](\d{1,2})[-|_|\.|](\d{1,2}).*`)

// regexDateNoSep 是用于无分隔符的日期正则表达式：
//  1. 可以有字符前缀及后缀。
//  2. 日期数字之间无分隔符，需要至少 8 位数字表示 YYYYMMDD。
var regexDateNoSep = regexp.MustCompile(`^.*?(\d{4})(\d{2})(\d{2}).*`)

/*
ParseDate parses date strings into time variables.

Parameters:
  - s: The string to parse. see examples of [ParseDateTime].

Returns:
  - The parsed date. nil is returned on failure.

ParseDate 将日期字符串转换为时间变量。

参数:
  - s: 待解析的字符串。参考 [ParseDateTime] 的示例。

返回:
  - 解析后的日期。失败均返回 nil。
*/
func ParseDate(s string) *time.Time {
	parse := func(s string, regex *regexp.Regexp) *time.Time {
		subs := regex.FindStringSubmatch(s)
		if len(subs) == 0 {
			// 没有配置的日期字符串，所以数组长度为 0，返回 nil 说明转换不成功。
			return nil
		}

		year, _ := strconv.Atoi(subs[1])
		m, _ := strconv.Atoi(subs[2])
		month := time.Month(m)
		day, _ := strconv.Atoi(subs[3])

		if IsDateTimeFieldValid(year, m, day, 0, 0, 0) != nil {
			return nil
		}

		result := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		return &result
	}

	result := parse(s, regexDateHasSep)
	if result != nil {
		return result
	}

	return parse(s, regexDateNoSep)
}

func IsDateTimeFieldValid(year, month, day, hour, minute, second int) error {
	if !RequireDateTimeFieldValid {
		return nil
	}

	if month < 1 || month > 12 {
		return fmt.Errorf("invalid month: %d", month)
	}
	if hour < 0 || hour > 23 {
		return fmt.Errorf("invalid hour: %d", hour)
	}
	if minute < 0 || minute > 59 {
		return fmt.Errorf("invalid minute: %d", minute)
	}
	if second < 0 || second > 59 {
		return fmt.Errorf("invalid second: %d", second)
	}
	if day < 1 || day > 31 {
		return fmt.Errorf("invalid day: %d", day)
	}
	if (month == 4 || month == 6 || month == 9 || month == 11) && day > 30 {
		return fmt.Errorf("invalid day: %d", day)
	}
	if month == 2 {
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			// 闰年 2 月份的最大天数为 29
			if day > 29 {
				return fmt.Errorf("invalid day for leap year: %d", day)
			}
		} else if day > 28 {
			return fmt.Errorf("invalid day: %d", day)
		}
	}

	return nil
}

// regexTimeNoSep 是用于无分隔符的时间正则表达式：
//  1. 可以有字符前缀及后缀。
//  2. 时间数字之间无分隔符，可以是 4 位数字，6 位数字，或者 9 位数字。
//     分别表示 HHMM，HHMMSS 及 HHMMSSSS。也就是说，可以精确到分钟、秒或毫秒。
//  3. 毫秒数为 3 位，与秒数之间可以有“.”作为分隔符，也可以无分隔符。
//
// note: 最后的 (\.?(\d{3}))? 不要外圈这括号也行，但加上后解析结果数组与 regexTimeHasSep 一致。
var regexTimeNoSep = regexp.MustCompile(`^.*?(\d{2})(\d{2})((\d{2})(\.?(\d{3}))?)?.*`)

// regexTimeHasSep 是用于有分隔符的时间正则表达式：
//  1. 可以有字符前缀及后缀。
//  2. 时间数字之间有分隔符，可以是“_”、“-”、“.”、“:”。秒与毫秒之间只能是“.”。
//     小时 1 或 2 位，分钟和秒都是 2 位，毫秒是 3 位。可以精确到分钟、秒或毫秒。
var regexTimeHasSep = regexp.MustCompile(`^.*?(\d{1,2})[-|_|\.|\:|](\d{2})([-|_|\.|\:|](\d{2})(\.(\d{3}))?)?.*`)

/*
ParseTime parses time strings into time variables.

Parameters:
  - s: The string to parse.  Milliseconds must be 3 digits, otherwise that value is not parsed.
    see examples of [ParseDateTime].

Returns:
  - The parsed time. nil is returned on failure.

ParseTime 将日期字符串转换为时间变量。

参数:
  - s: 待解析的字符串。毫秒必需是 3 位，否则不解析该值。参考 [ParseDateTime] 的示例。

返回:
  - 解析后的时间。失败均返回 nil。
*/
func ParseTime(s string) *time.Time {
	parse := func(s string, regex *regexp.Regexp) *time.Time {
		subs := regex.FindStringSubmatch(s)
		if len(subs) == 0 {
			// 没有配置的日期字符串，所以数组长度为 0，返回 nil 说明转换不成功。
			return nil
		}

		hour, _ := strconv.Atoi(subs[1])
		minute, _ := strconv.Atoi(subs[2])
		second, _ := strconv.Atoi(subs[4])
		millisecond, _ := strconv.Atoi(subs[6])
		nanosecond := millisecond * 1000_000

		if IsDateTimeFieldValid(0, 1, 1, hour, minute, second) != nil {
			return nil
		}

		// time.Parse() 只解析时间时，使用的日期就是 0，1，1。
		result := time.Date(0, 1, 1, hour, minute, second, nanosecond, time.Local)
		return &result
	}

	result := parse(s, regexTimeHasSep)
	if result != nil {
		return result
	}

	return parse(s, regexTimeNoSep)
}
