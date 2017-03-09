package utilities

import "time"

const debug = false

const NumElevators = 3
const NumButtons = 3
const NumFloors = 4

var HallOrderMatrix [NumFloors][2]ElevatorOrder
var CabHallOrderMatrix [NumFloors]CabOrder

type CabOrderList map[string][NumFloors]CabOrder

// key = IPaddr
var ElevatorStatus = make(map[string]*Elevator) // containing last known state
var OnlineElevators = make(map[string]bool)

var EventType = []string{
	// BackupMessage Events
	"EventElevatorOnline",
	"EventElevatorBackup",
	"EventRequestBackup",
	"EventElevatorBackupReturned",
	"EventCabOrder",
	"EventAckCabOrder",

	// OrderMessage Events

	"EventNewOrder",
	"EventAckNewOrder",
	"EventOrderConfirmed",
	"EventAckOrderConfirmed",
	"EventOrderCompleted",
	"EventAckOrderCompleted",
}

// TODO: UPDATE network module if any changes in events
const (
	// BackupMessage Events
	EventElevatorOnline = iota //  = 0
	EventElevatorBackup
	EventRequestBackup
	EventElevatorBackupReturned
	EventCabOrder
	EventAckCabOrder
	// OrderMessage Events
	EventNewOrder
	EventAckNewOrder
	EventOrderConfirmed
	EventAckOrderConfirmed
	EventOrderCompleted
	EventAckOrderDone
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

type CabOrder struct {
	LocalIP     string
	OriginIP    string
	Floor       int
	ConfirmedBy map[string]bool
	Timer       time.Time
}

type ElevatorOrder struct {
	Status      int
	AssignedTo  string
	ConfirmedBy map[string]bool
	Timer       *time.Timer // *time.Timer 'json:"-"'
}

type Elevator struct {
	LocalIP    string
	LastFloor  int
	Direction  int
	IsMoving   bool
	DoorStatus bool
	Time       time.Time
	CabOrders  [NumFloors]bool
}

type OrderMessage struct {
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
	State       Elevator
	Cab         CabOrder
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
		LocalIP:    e.LocalIP,
		LastFloor:  e.LastFloor,
		Direction:  e.Direction,
		IsMoving:   e.IsMoving,
		DoorStatus: e.DoorStatus,
		Time:       time.Now(),
		CabOrders:  e.CabOrders,
	}
}

func ResolveWatchdogKickMessage(e *Elevator) ElevatorBackupMessage {
	return ElevatorBackupMessage{
		//AskerIP:     "",
		ResponderIP: e.LocalIP,
		Event:       EventElevatorOnline,
		State:       *e,
	}

}

func ResolveBackupState(e *Elevator) ElevatorBackupMessage {
	return ElevatorBackupMessage{
		ResponderIP: e.LocalIP,
		Event:       EventElevatorBackup,
		State:       *e,
	}
}



// ----Type: ElevatorBackupMessage ----
func (m ElevatorBackupMessage) IsValid() bool {
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

// ----Type: ElevatorOrder ----
func (order *ElevatorOrder) InitConfirmedBy() {
	for key := range order.ConfirmedBy {
		delete(order.ConfirmedBy, key)
	}
	order.ConfirmedBy = make(map[string]bool)

}
