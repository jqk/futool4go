package timeutils

import (
	"regexp"
	"strconv"
	"time"
)

// 用于 UnixTime 的正则表达式：
// 1. 可以有字符前缀及后缀。
// 2. 需要至少 10 位数字代表秒数。
// 3. 可以再有 3 位数字代表毫秒数。不足不算。
var regexUnixTime = regexp.MustCompile(`^.*?(\d{1,10})(\d{3})?.*`)

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

	result := time.Unix(second, nanosecond)
	return &result
}

// 用于无分隔符的日期时间正则表达式：
//  1. 可以有字符前缀及后缀。
//  2. 日期数字之间无分隔符，需要至少 8 位数字表示 YYYYMMDD。
//  3. 日期与时间之间可以有为分隔符，也可以无分隔符。分隔符可以是“_”、“-”、“.”、“ ”和“T”。
//  4. 时间数字之间无分隔符，可以是 4 位数字，6 位数字，或者 9 位数字。
//     分别表示 HHMM，HHMMSS 及 HHMMSSSS。也就是说，可以精确到分钟、秒或毫秒。
//  5. 毫秒数为 3 位，与秒数之间可以有“.”作为分隔符，也可以无分隔符。
var regexDateTimeNoSep = regexp.MustCompile(
	`^.*?(\d{4})(\d{2})(\d{2})[-|_|\.| |T]?` +
		`(\d{2})(\d{2})((\d{2})\.?(\d{3})?)?.*`)

// 用于有分隔符的日期时间正则表达式：
//  1. 可以有字符前缀及后缀。
//  2. 日期数字之间有分隔符，年 4 位，月 1 或 2 位，日 1 或 2 位。
//  3. 日期与时间之间必须有为分隔符，可以是“_”、“-”、“.”、“ ”和“T”。
//  4. 时间数字之间有分隔符，可以是“_”、“-”、“.”、“:”。秒与毫秒之间只能是“.”。
//     小时 1 或 2 位，分钟和秒都是 2 位，毫秒是 3 位。可以精确到分钟、秒或毫秒。
var regexDateTimeHasSep = regexp.MustCompile(
	`^.*?(\d{4})[-|_|\.|](\d{1,2})[-|_|\.|](\d{1,2})[-|_|\.| |T]` +
		`(\d{1,2})[-|_|\.|\:|](\d{2})([-|_|\.|\:|](\d{2})(\.(\d{3}))?)?.*`)

func ParseDateTime(s string) *time.Time {
	result := parseDateTimeHasSeperator(s)
	if result != nil {
		return result
	}

	return parseDateTimeNoSeperator(s)
}

func parseDateTimeNoSeperator(s string) *time.Time {
	subs := regexDateTimeNoSep.FindStringSubmatch(s)
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
	millisecond, _ := strconv.Atoi(subs[8])
	nanosecond := millisecond * 1000_000

	result := time.Date(year, month, day, hour, minute, second, nanosecond, time.Local)
	return &result
}

func parseDateTimeHasSeperator(s string) *time.Time {
	subs := regexDateTimeHasSep.FindStringSubmatch(s)
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
	// subs[8] 包含了"."和毫秒。
	millisecond, _ := strconv.Atoi(subs[9])
	nanosecond := millisecond * 1000_000

	result := time.Date(year, month, day, hour, minute, second, nanosecond, time.Local)
	return &result
}
