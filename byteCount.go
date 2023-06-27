package futool4go

import "fmt"

type ByteCount interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64
}

var kb float64 = 1024
var mb = kb * kb
var gb = mb * kb
var tb = gb * kb

// ToSizeString converts the byte count to the appropriate unit (KB, MB, GB, or TB)
// and formats it with the specified precision if provided.
// If no precision is provided, it defaults to 3 decimal places.
// Precision must be between 0 and 9。
func ToSizeString[T ByteCount](size T, precision ...int) string {
	p := 3
	n := len(precision)

	if n > 1 {
		panic("too many precision")
	} else if n == 1 {
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
	} else {
		return fmt.Sprintf(format("TB"), value/tb)
	}
}
