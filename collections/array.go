package collections

import "reflect"

/*
MapToArray convert a map to an array.

MapToArray 将 map 转换为数组。
*/
func MapToArray[K comparable, V any](m map[K]V) []V {
	v := reflect.ValueOf(m)
	ret := reflect.MakeSlice(reflect.SliceOf(v.Type().Elem()), v.Len(), v.Len())

	for i, key := range v.MapKeys() {
		ret.Index(i).Set(v.MapIndex(key))
	}

	return ret.Interface().([]V)
}
