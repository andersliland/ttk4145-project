package utilities

import "time"

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

// Console colors
const (
	ColorWhite   = "\x1b[37;1m"
	ColorNeutral = "\x1b[0m"
)
