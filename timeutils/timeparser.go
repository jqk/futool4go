package timeutils

import (
	"regexp"
	"strconv"
	"time"
)

func ParseUnixTime(s string) *time.Time {
	// 前后都可以有字符，但需要至少 10 位数字代表秒数。
	// 可以再有 3 位数字代表毫秒数。
	re := regexp.MustCompile(`^.*?(\d{1,10})(\d{3})?.*`)
	subs := re.FindStringSubmatch(s)
	count := len(subs)

	if count <= 1 {
		// 没有配置的 unix 时间截字符串。
		return nil
	}

	// 第 1 个匹配是 10 位，代表秒数。到此处必定存在。
	var nanoseconds int64 = 0
	seconds, _ := strconv.ParseInt(subs[1], 10, 64)

	// 第 2 个匹配是 3 位，代表毫秒数，要转换为纳秒。可能不存在。
	if count > 2 {
		nanoseconds, _ = strconv.ParseInt(subs[2], 10, 64)
		nanoseconds *= 1000_000
	}

	result := time.Unix(seconds, nanoseconds)
	return &result
}
