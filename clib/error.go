package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

func handleError(errOut **C.char, err error) {
	if errOut != nil && err != nil {
		*errOut = C.CString(err.Error())
	}
}
