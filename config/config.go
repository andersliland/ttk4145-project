package config

import (
	"log"
	"os"
)

//TODO:make channel structs
// Channels vs Channel?
type HardwareChannels struct {
}

type NetworkChannels struct {
}

type EventChannels struct {
}

const NumButtons = 3
const NumFloors = 4

const (
	ButtonCallUp = iota
	ButtonCallDown
	ButtonCommand
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
	Floor      int
	ButtonType int
	AssignedTo string
	OriginIP   string
	SenderIP   string
	Event      int
}

func CheckError(errMsg string, err error) {
	if err != nil {
		log.Println(errMsg, " :", err.Error())
		os.Exit(1)
	}

}

// Console colors
const (
	ColorWhite   = "\x1b[37;1m"
	ColorNeutral = "\x1b[0m"
)
