package driver

/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
*/

import (
	"C"
	"errors"
)

func ioInit() error {
	if err := int(C.io_init()); err == 0 {
		return errors.New("Elevator initialization unsuccessful!")
	}
	return nil
}

func ioSetBit(channel int) {
	C.io_set_bit(C.int(channel))
}

func ioClearBit(channel int) {
	C.io_clear_bit(C.int(channel))
}

func ioReadBit(channel int) bool {
	return int(C.io_read_bit(C.int(channel))) != 0
}

func ioReadAnalog(channel int) int {
	return int(C.io_read_analog(C.int(channel)))
}

func ioWriteAnalog(channel int, value int) {
	C.io_write_analog(C.int(channel), C.int(value))
}
