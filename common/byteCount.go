package common

import "fmt"

/*
ByteCount defines type for counting bytes.

ByteCount 定义了可用于表示字节数的类型。
*/
type ByteCount interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64
}

var kb float64 = 1024
var mb = kb * kb
var gb = mb * kb
var tb = gb * kb
var pb = tb * kb

/*
ToSizeString converts a byte count to a string with proper units (KB, MB, GB, or TB) and formatted with precision.

Parameters:
    - size: Byte count.
    - precision: Precision. Precision must be between 0 and 9. Default is 3.

Returns:
    - Formatted string.

ToSizeString 将字节数转换为正确单位(KB, MB, GB, 或 TB)的字符串，并按精度格式化。

参数:
	- size: 字节数。
	- precision: 精度。范围 0 到 9，默认为 3。

返回:
	- 格式化后的字符串。
*/
func ToSizeString[T ByteCount](size T, precision ...int) string {
	// 未指定 precision 参数时，默认为 3。指定多个参数时也只有第一个有效。
	p := 3
	if len(precision) > 0 {
		p = precision[0]
		if p < 0 || p > 9 {
			panic("invalid precision, must be between 0 and 9")
		}
	}

	format := func(unit string) string {
		return fmt.Sprintf("%%.%df %s", p, unit)
	}
	value := float64(size)

	if value < kb {
		return fmt.Sprintf("%.0f bytes", value)
	} else if value < mb {
		return fmt.Sprintf(format("KB"), value/kb)
	} else if value < gb {
		return fmt.Sprintf(format("MB"), value/mb)
	} else if value < tb {
		return fmt.Sprintf(format("GB"), value/gb)
	} else if value < pb {
		return fmt.Sprintf(format("TB"), value/tb)
	} else {
		return fmt.Sprintf(format("PB"), value/pb)
	}
}
