package config

import (
	"log"
	"os"
	"time"
)

const debug = false

const NumButtons = 3
const NumFloors = 4

const (
	EvNewOrder = iota
	EvExecuteOrder
	EvRestoreOrder
	AckExecuteOrder
	EvElevatorAliveMessage
)

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

type ElevatorState struct {
	LocalIP    string
	LastFloor  int
	Direction  int
	IsMoving   int
	DoorStatus int
}

type ElevatorBackupMessage struct {
	Time     time.Time
	OriginIP string
	Event    int
}

type Elevator struct {
	State ElevatorState
	Time  time.Time
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
