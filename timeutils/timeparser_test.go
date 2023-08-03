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
	RequireDateTimeInRange = true

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

	// 无效的字符串，因为没有时间部分。
	tm = ParseDateTime("20100223")
	assert.Nil(t, tm)

	// 无效的字符串，因为月份字段超过范围。
	tm = ParseDateTime("201022231234")
	assert.Nil(t, tm)

	// 有效的字符串，因为不检查字段范围，所以月份字段超过范围仍进行解析，这是 go 内置的功能。
	RequireDateTimeInRange = false
	tm = ParseDateTime("201022231234")
	assert.NotNil(t, tm)
	// 加上了超范围的月份数。
	assert.Equal(t, "2011-10-23 12:34:00", tm.Format("2006-01-02 15:04:05"))

	// 恢复默认设置。
	RequireDateTimeInRange = true
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

	// 无效的字符串，因为没有时间部分。
	tm = ParseDateTime("2010-02-23")
	assert.Nil(t, tm)
}

func TestParseDate(t *testing.T) {
	// 有前、后缀，虽然有分钟，但仅处理日期部分。
	tm := ParseDate("abc2010-2-23-15:34ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 00:00:00", tm.Format("2006-01-02 15:04:05"))

	// 无前缀、有后缀。
	tm = ParseDate("20100223153456.789ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "2010-02-23 00:00:00", tm.Format("2006-01-02 15:04:05"))
}

func TestParseTime(t *testing.T) {
	// 有前、后缀，精确到分钟。
	tm := ParseTime("abc15:34ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "15:34:00", tm.Format("15:04:05"))

	// 有前、后缀，精确到秒。
	tm = ParseTime("abc15:34.56ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "15:34:56", tm.Format("15:04:05"))

	// 有前、后缀，精确到毫秒。
	tm = ParseTime("abc15:34.56.789ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "15:34:56.789", tm.Format("15:04:05.000"))

	// 有前、后缀，毫秒位不足，会被忽略。
	tm = ParseTime("abc15:34-56.78ddd.jpg")
	assert.NotNil(t, tm)
	assert.Equal(t, "15:34:56.000", tm.Format("15:04:05.000"))

	// 无前、后缀，无分隔符。
	tm = ParseTime("153456789")
	assert.NotNil(t, tm)
	assert.Equal(t, "15:34:56.789", tm.Format("15:04:05.000"))

	// 注意 PasrseTime() 使用的时区是 time.Local，而 time.Parse() 使用的时区是 time.UTC。
	tp, err := time.Parse("15:04:05.000", "15:34:56.789")
	assert.Nil(t, err)
	assert.Equal(t, "15:34:56.789", tp.Format("15:04:05.000"))
	// 由于时区关系，此处必然不相等。
	assert.NotEqual(t, tp, *tm)

	tp, err = time.ParseInLocation("15:04:05.000", "15:34:56.789", time.Local)
	assert.Nil(t, err)
	assert.Equal(t, "15:34:56.789", tp.Format("15:04:05.000"))
	// 指定时区，此处必然相等。
	assert.Equal(t, tp, *tm)
}
