package char

import (
	"reflect"
	"unsafe"
)

func ByteToStr(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))

	sh := reflect.StringHeader{
		Data: sliceHeader.Data,
		Len:  sliceHeader.Len,
	}

	return *(*string)(unsafe.Pointer(&sh))
}

func StrToByte(str string) []byte {
	return *(*[]byte)(unsafe.Pointer(&str))
}

func ByteLen(b []byte) int {
	return (*reflect.SliceHeader)(unsafe.Pointer(&b)).Len
}
