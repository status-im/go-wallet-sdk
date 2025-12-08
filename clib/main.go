//go:build cgo

package main

/*
#include <stdint.h>
*/
import "C"

// This main() is never executed—it's just to satisfy -buildmode=c-shared.
// All real exports live in the imported cshared package.
func main() {}
