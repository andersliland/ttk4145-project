package utilities

import "time"

const debug = false

const NumButtons = 3
const NumFloors = 4

var EventType = []string{
	"EvNewOrder",
	"EvExecuteOrder",
	"EvRestoreOrder",
	"AckExecuteOrder",
	"EvElevatorAliveMessage",
	"EvRequestState",
}

const (
	EvNewOrder = iota
	EvExecuteOrder
	EvRestoreOrder
	AckExecuteOrder
	EvElevatorAliveMessage
	EvRequestState
)

var ButtonType = []string{
	"ButtonCallUp",
	"ButtonCallDown",
	"ButtonCommand",
	"ButtonStop",
	"DoorIndicator",
	"FloorSensor",
	"FloorIndicator",
}

const (
	ButtonCallUp = iota
	ButtonCallDown
	ButtonCommand
	ButtonStop
	DoorIndicator
	FloorSensor
	FloorIndicator
)

var OrderStatus = []string{
	"NotActive",
	"Awaiting",
	"UnderExecution",
}

const (
	NotActive = iota
	Awaiting
	UnderExecution
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

type ElevatorOrder struct {
	Status      int
	AssignedTo  string
	ConfirmedBy map[string]bool
	Timer       time.Timer
}

type ElevatorState struct {
	LocalIP    string
	LastFloor  int
	Direction  int
	IsMoving   int
	DoorStatus int
}

type ElevatorBackupMessage struct {
	AskerIP string
	State   ElevatorState
	Event   int
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
