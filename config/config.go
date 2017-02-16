package config

import (
	"log"
	"os"
	"time"
)

const debug = false

//TODO:make channel structs
// Channels vs Channel?
type HardwareChannels struct {
}

type NetworkChannels struct {
}

type EventChannels struct {
}

const (
	EvNewOrder = iota
	EvExecuteOrder
	EvRestoreOrder
	AckExecuteOrder
	EvElevatorAliveMessage
)

const NumButtons = 3
const NumFloors = 4

const (
	ButtonCallUp = iota
	ButtonCallDown
	ButtonCommand
	ButtonStop
	DoorIndicator
	FloorSensor
	FloorIndicator
)

const (
	FloorInvalid = iota - 1
	Floor1
	Floor2
	Floor3
	Floor4
)

const (
	MotorStop = iota
	MotorUp
	MotorDown
)

type ElevatorOrderMessage struct {
	Time       time.Time
	Floor      int
	ButtonType int
	AssignedTo string
	OriginIP   string
	SenderIP   string
	Event      int
}

type ElevatorBackupMessage struct {
	Time     time.Time
	OriginIP string
	Event    int
}

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

// Console colors
const (
	ColorWhite   = "\x1b[37;1m"
	ColorNeutral = "\x1b[0m"
)
