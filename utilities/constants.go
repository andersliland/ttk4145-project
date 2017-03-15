package utilities

import "time"

const NumElevators = 3
const NumButtons = 3
const NumFloors = 4

const TimeBetweenFloors = 5
const DoorOpenTime = 3
const OrderTimeout = DoorOpenTime + TimeBetweenFloors

const PollDelay = 5 * time.Millisecond
const ElevatorPollDelay = 50 * time.Millisecond

var EventType = []string{
	// BackupMessage Events
	"EventElevatorOnline",
	// OrderMessage Events
	"EventNewOrder",
	"EventAckNewOrder",
	"EventOrderConfirmed",
	"EventAckOrderConfirmed",
	"EventAckOrderCompleted",
	"EventOrderCompleted",
	"EventReassignOrder",
}

const (
	// BackupMessage Events
	EventElevatorOnline = iota //  = 0
	// OrderMessage Events
	EventNewOrder
	EventAckNewOrder
	EventOrderConfirmed
	EventAckOrderConfirmed
	EventAckOrderCompleted
	EventOrderCompleted
	EventReassignOrder
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
	ButtonCallUp = iota //0
	ButtonCallDown
	ButtonCommand
	ButtonStop
	DoorIndicator
	FloorSensor
	FloorIndicator
)

var MotorStatus = []string{
	"Down",
	"Stop",
	"Up",
}

const (
	Down = iota - 1
	Stop
	Up
)

var StateEventManager = []string{
	"Idle",
	"Moving",
	"DoorOpen",
}

const (
	Idle = iota
	Moving
	DoorOpen
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
	TimeoutAckNewOrder = iota
	TimeoutAckOrderConfirmed
	TimeoutOrderExecution
)

const (
	FloorInvalid = iota - 1
	Floor1
	Floor2
	Floor3
	Floor4
)

// Console colors
const (
	ColorDarkGrey = "\x1b[30;1m"
	ColorMagenta  = "\x1b[35;1m"
	ColorCyan     = "\x1b[36;1m"
	ColorRed      = "\x1b[31;1m"
	ColorGreen    = "\x1b[32;1m"
	ColorYellow   = "\x1b[33;1m"
	ColorBlue     = "\x1b[34;1m"
	ColorWhite    = "\x1b[37;1m"
	ColorNeutral  = "\x1b[0m"
)
