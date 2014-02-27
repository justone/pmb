package main

import (
	"os"
	"reflect"
	"strings"
	"unsafe"
)

// from http://stackoverflow.com/questions/14926020/setting-process-name-as-seen-by-ps-in-go
func MaskProcessArg(index int) error {
	existing := os.Args[index]
	argv0str := (*reflect.StringHeader)(unsafe.Pointer(&os.Args[index]))
	argv0 := (*[1 << 30]byte)(unsafe.Pointer(argv0str.Data))[:argv0str.Len]

	n := copy(argv0, strings.Repeat("x", len(existing)))
	if n < len(argv0) {
		argv0[n] = 0
	}

	return nil
}
