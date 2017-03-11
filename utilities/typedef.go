package utilities

import "time"

const debug = false

const NumElevators = 3
const NumButtons = 3
const NumFloors = 4

var HallOrderMatrix [NumFloors][2]HallOrder

//var CabHallOrderMatrix [NumFloors]CabOrder

// key = IPaddr
var ElevatorStatus = make(map[string]*Elevator) // containing last known state

//var ElevatorLocalMap = make(map[string]*ElevatorLocal)

var OnlineElevators = make(map[string]bool)

var EventType = []string{
	// BackupMessage Events
	"EventElevatorOnline",
	"EventElevatorBackup",
	"EventRequestBackup",
	"EventBackupReturned",
	"EventCabOrder",
	"EventAckCabOrder",

	// OrderMessage Events

	"EventNewOrder",
	"EventAckNewOrder",
	"EventOrderCost",
	"EventAckOrderCost",
	"EventOrderCompleted",
	"EventAckOrderCompleted",
}

// TODO: UPDATE network module if any changes in events
const (
	// BackupMessage Events
	EventElevatorOnline = iota //  = 0
	EventElevatorBackup
	EventRequestBackup
	EventBackupReturned
	EventCabOrder
	EventAckCabOrder
	// OrderMessage Events
	EventNewOrder
	EventAckNewOrder
	EventOrderConfirmed
	EventAckOrderConfirmed
	EventOrderCompleted
	EventAckOrderCompleted
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
	Idle = iota
	Moving
	DoorOpen
)

const (
	Down = iota - 1
	Stop
	Up
)

var MotorStatus = []string{
	"Down",
	"Stop",
	"Up",
}

type CabOrder struct {
	LocalIP     string
	OriginIP    string
	Floor       int
	ConfirmedBy map[string]bool
	Timer       time.Time
}

type HallOrder struct {
	Status      int
	AssignedTo  string
	ConfirmedBy map[string]bool
	Timer       *time.Timer // *time.Timer 'json:"-"'
}

type Elevator struct { // syncronised for all elevators
	LocalIP         string
	Time            time.Time
	State           int //Idle, Moving, DoorOpen
	Floor           int // current floor for elevator
	Direction       int // current direction: Stop, Up, Down
	CabOrders       [NumFloors]bool
	HallOrderMatrix [NumFloors][2]HallOrder
}

/*
type ElevatorLocal struct {
	State           int //Idle, Moving, DoorOpen
	LastFloor       int // current floor for elevator
	Direction       int // current direction: Stop, Up, Down
	CabOrders       [NumFloors]bool
	HallOrderMatrix [NumFloors][2]HallOrder
}
*/

type OrderMessage struct {
	Time       time.Time
	Floor      int
	ButtonType int
	AssignedTo string
	OriginIP   string
	SenderIP   string
	Event      int
}

type BackupMessage struct {
	AskerIP         string
	ResponderIP     string
	Event           int
	State           Elevator
	Cab             CabOrder
	HallOrderMatrix [NumFloors][2]HallOrder
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

func ResolveElevator(e Elevator) *Elevator {
	return &Elevator{
		LocalIP:   e.LocalIP,
		Time:      time.Now(),
		CabOrders: e.CabOrders,
	}
}

func ResolveWatchdogKickMessage(e *Elevator) BackupMessage {
	return BackupMessage{
		//AskerIP:     "",
		ResponderIP: e.LocalIP,
		Event:       EventElevatorOnline,
		State:       *e,
	}

}

func ResolveBackupState(e *Elevator) BackupMessage {
	return BackupMessage{
		ResponderIP: e.LocalIP,
		Event:       EventElevatorBackup,
		State:       *e,
	}
}

// ----Type: BackupMessage ----
func (m BackupMessage) IsValid() bool {
	if m.AskerIP == m.ResponderIP {
		return false
	}
	if m.Event > 4 || m.Event < 0 {
		return false
	}
	return true
}

// ----Type: OrderMessage ----

func (m OrderMessage) IsValid() bool {
	if m.Floor > NumFloors || m.Floor < -1 {
		return false
	}
	if m.ButtonType > 2 || m.ButtonType < 0 {
		return false
	}
	if m.Event > 10 || m.Event < 5 {
		return false
	}
	return true
}

// ----Type: Elevator ----
func (e *Elevator) AddCabOrder(Floor int) {
	e.CabOrders[Floor] = true
}

func (e *Elevator) RemoveCabOrder(Floor int) {
	e.CabOrders[Floor] = false

}

func (e *Elevator) UpdateElevatorStatus(backup BackupMessage) {
	e.State = backup.State.State
	e.Floor = backup.State.Floor
	e.Direction = backup.State.Direction
	//e.CabOrders[backup.State.Floor] = //how to sync CabOrders
	// NEED TO SYNC HallOrderMatrix?
}

// ----Type: HallOrder ----
func (order *HallOrder) ClearConfirmedBy() {
	for key := range order.ConfirmedBy {
		delete(order.ConfirmedBy, key)
	}
	order.ConfirmedBy = make(map[string]bool)

}
