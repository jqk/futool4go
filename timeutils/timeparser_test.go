package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseUnixTime(t *testing.T) {
	// 数字是 13 位，有秒有毫秒。
	tm := ParseUnixTime("snapshot_1553867509757.png")
	assert.Equal(t, int64(1553867509757), tm.UnixMilli())

	expect, err := time.ParseInLocation("2006-01-02T15:04:05.000", `2019-03-29T21:51:49.757`, time.Local)
	assert.Nil(t, err)
	assert.Equal(t, expect, *tm)

	// 由于仅 12 位，所以毫秒数不够 3 位，不取毫秒。
	tm = ParseUnixTime("155386750975abcd")
	assert.Equal(t, int64(1553867509000), tm.UnixMilli())

	// 后缀 Z 表示 UTC 时间。
	expect, err = time.Parse("2006-01-02T15:04:05Z", `2019-03-29T13:51:49Z`)
	assert.Nil(t, err)

	// 必须转换一下时区，虽然即使不转换，绝对时间点也相同。但不转的话，比较时会报错。
	expect = expect.In(time.Local)
	assert.Equal(t, expect, *tm)
}
