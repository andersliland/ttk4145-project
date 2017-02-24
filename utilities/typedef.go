package utilities

import "time"

const debug = false

const NumButtons = 3
const NumFloors = 4

var EventType = []string{
	// BackupMessage Events
	"EvIAmAlive",
	"EvBackupState",
	"EvRequestBackupState",
	"EvBackupStateReturned",
	// OrderMessage Events

	"EvNewOrder",
	"EvAckNewOrder",
	"EvOrderConfirmed",
	"EvAckOrderConfirmed",
	"EvOrderDone",
	"EvAckOrderDone",
	"EvReassignOrder",
}

const (
	// BackupMessage Events
	EvIAmAlive = iota //  = 0
	EvBackupState
	EvRequestBackupState
	EvBackupStateReturned
	// OrderMessage Events
	EvNewOrder
	EvAckNewOrder
	EvOrderConfirmed
	EvAckOrderConfirmed
	EvOrderDone
	EvAckOrderDone
	EvReassignOrder
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

type Elevator struct {
	State ElevatorState
	Time  time.Time
}

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
	AskerIP     string
	ResponderIP string
	Event       int
	State       ElevatorState
	//ExternalOrderMatrix
}

type ElevatorButton struct {
	Floor int
	Kind  int
}

type ElevatorLight struct {
	Floor  int
	Kind   int
	Active bool
}

// Console colors
const (
	ColorWhite   = "\x1b[37;1m"
	ColorNeutral = "\x1b[0m"
)

// functions

func ResolveElevator(state ElevatorState) *Elevator {
	return &Elevator{state, time.Now()}
}

func ResolveWatchdogKickMessage(elevator *Elevator) ElevatorBackupMessage {
	return ElevatorBackupMessage{
		ResponderIP: elevator.State.LocalIP,
		Event:       EvIAmAlive,
		State:       elevator.State}

}

func (m ElevatorBackupMessage) IsValid() bool {
	if m.AskerIP == m.ResponderIP {
		return false
	}
	if m.Event > 3 || m.Event < 0 {
		return false
	}
	return true
}

func (m ElevatorOrderMessage) IsValid() bool {
	if m.Floor > NumFloors || m.Floor < -1 {
		return false
	}
	if m.ButtonType > 2 || m.ButtonType < 0 {
		return false
	}
	if m.Event > 10 || m.Event < 4 {
		return false
	}
	return true
}
