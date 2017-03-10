package wrapper

/*
#cgo CFLAGS: -std=gnu11
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
*/
import "C"
import "log"
import "os"

func IoInit() {
	if err := int(C.io_init()); err == 0 {
		log.Println("FAILED [io]: IO initialization")
		os.Exit(-1)
	}
}

func IoSetBit(channel int) {
	C.io_set_bit(C.int(channel))
}

func IoClearBit(channel int) {
	C.io_clear_bit(C.int(channel))
}

func IoReadBit(channel int) bool {
	return int(C.io_read_bit(C.int(channel))) != 0
}

func IoReadAnalog(channel int) int {
	return int(C.io_read_analog(C.int(channel)))
}

func IoWriteAnalog(channel int, value int) {
	C.io_write_analog(C.int(channel), C.int(value))
}
