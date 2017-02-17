package utilities

import (
	"log"
	"os"
)

func CheckError(errMsg string, err error) {
	if err != nil {
		log.Println(errMsg, " :", err.Error())
		os.Exit(1)
	}
}

func printDebug(s string) {
	if debug {
		log.Println("CONFIG: \t", s)
	}
}
