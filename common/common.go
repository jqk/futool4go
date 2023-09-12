package common

import "reflect"

func GetBuffer[T int | []byte](buf T) []byte {
	switch v := reflect.ValueOf(buf); v.Kind() {
	case reflect.Int:
		return make([]byte, int(v.Int()))
	case reflect.Slice:
		return v.Slice(0, int(v.Len())).Bytes()
	default:
		// 不会执行到这里。
		panic("T must be int or []byte")
	}
}
