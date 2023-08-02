package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseUnixTime(t *testing.T) {
	// 数字是 13 位，有秒有毫秒。尾部再长，超过 13 位，将被忽略。
	tm := ParseUnixTime("snapshot_1553867509757.png")
	assert.NotNil(t, tm)
	assert.Equal(t, int64(1553867509757), tm.UnixMilli())

	expect, err := time.ParseInLocation("2006-01-02T15:04:05.000", `2019-03-29T21:51:49.757`, time.Local)
	assert.Nil(t, err)
	assert.Equal(t, expect, *tm)

	// 由于仅 12 位，所以毫秒数不够 3 位，不取毫秒，因此只处理 10 位。
	tm = ParseUnixTime("155386750975abcd")
	assert.NotNil(t, tm)
	assert.Equal(t, int64(1553867509000), tm.UnixMilli())

	// 后缀 Z 表示 UTC 时间。
	expect, err = time.Parse("2006-01-02T15:04:05Z", `2019-03-29T13:51:49Z`)
	assert.Nil(t, err)

	// 必须转换一下时区，虽然即使不转换，绝对时间点也相同。但不转的话，比较时会报错。
	expect = expect.In(time.Local)
	assert.Equal(t, expect, *tm)
}

func TestParseDateTimeNoSeperator(t *testing.T) {
	// 有前、后缀，“-”作为分隔符，精确到分钟的字符串。
	tm := ParseDateTime("abc20100223-1534ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 15:34:00", tm.Format("2006-01-02 15:04:05"))

	// 无前缀、有后缀，“.”作为分隔符，精确到毫秒的字符串。
	tm = ParseDateTime("20100223.153456.789ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 15:34:56.789", tm.Format("2006-01-02 15:04:05.000"))

	// 无前缀、无后缀，无分隔符，精确到毫秒的字符串。
	tm = ParseDateTime("20100223153456789")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 15:34:56.789", tm.Format("2006-01-02 15:04:05.000"))

	// 有前缀、无后缀，空格作为无分隔符，精确到秒的字符串。
	// 注意，最后一位”7“由于在秒之后，又组不成 3 位毫秒，所以被忽略。
	tm = ParseDateTime("20100223 1534567")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 15:34:56", tm.Format("2006-01-02 15:04:05"))

	// 无效的字符串。
	tm = ParseDateTime("20100223")
	assert.Nil(t, tm)
}

func TestParseDateTimeHasSeperator(t *testing.T) {
	// 有前、后缀，精确到分钟的字符串。
	tm := ParseDateTime("abc2010-02-23-15:34ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 15:34:00", tm.Format("2006-01-02 15:04:05"))

	// 无前缀、有后缀，精确到毫秒的字符串。注意月份和小时都只有 1 位。
	tm = ParseDateTime("2010_2-23.5.34.56.789ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 05:34:56.789", tm.Format("2006-01-02 15:04:05.000"))

	// 无前缀、无后缀，精确到毫秒的字符串。
	tm = ParseDateTime("2010.02.23T15-34_56.789")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 15:34:56.789", tm.Format("2006-01-02 15:04:05.000"))

	// 有前缀、无后缀，精确到秒的字符串。
	// 注意，最后一位”7“由于在秒之后，又组不成 3 位毫秒，所以被忽略。
	tm = ParseDateTime("2010-02-23 15:34:56.7")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 15:34:56", tm.Format("2006-01-02 15:04:05"))

	// 无效的字符串。
	tm = ParseDateTime("2010-02-23")
	assert.Nil(t, tm)
}
