package main

/*
#include <stdlib.h>
*/
import "C"

import "unsafe"

// GoWSK_FreeCString frees C strings returned by GoWSK functions to prevent memory leaks.
//
//export GoWSK_FreeCString
func GoWSK_FreeCString(s *C.char) {
	if s != nil {
		C.free(unsafe.Pointer(s))
	}
}
