package utilities

import "time"

const debug = false

const NumElevators = 3
const NumButtons = 3
const NumFloors = 4

var OrderMatrix [NumFloors][2]ElevatorOrder
var CabOrderMatrix [NumFloors]CabOrder

type CabOrderList map[string][NumFloors]CabOrder

// key = IPaddr
var RegisteredElevators = make(map[string]*Elevator) // containing last known state
var WorkingElevators = make(map[string]bool)

var EventType = []string{
	// BackupMessage Events
	"EventElevatorAlive",
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
	"EventOrderDone",
	"EventAckOrderDone",
}

// TODO: UPDATE network module if any changes in events
const (
	// BackupMessage Events
	EventElevatorAlive = iota //  = 0
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
	EventOrderDone
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

type Elevator struct { // TODO: remove
	State ElevatorState
	Time  time.Time
}
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
	Timer       time.Timer // *time.Timer 'json:"-"'
}

type ElevatorState struct {
	LocalIP    string
	LastFloor  int
	Direction  int
	IsMoving   bool
	DoorStatus bool
	Time       time.Time
	CabOrders  [NumFloors]bool
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

// functions

//set internal order for ElevatorOrder

func ResolveElevator(state ElevatorState) *Elevator {
	return &Elevator{State: state, Time: time.Now()}
}

func ResolveElevatorState(state ElevatorState) *ElevatorState {
	return &ElevatorState{
		LocalIP: state.LocalIP,
	}
}

func ResolveWatchdogKickMessage(elevator *Elevator) ElevatorBackupMessage {
	return ElevatorBackupMessage{
		AskerIP:     "",
		ResponderIP: elevator.State.LocalIP,
		Event:       EventElevatorAlive,
		State:       elevator.State}

}

func ResolveBackupState(elevator *Elevator) ElevatorBackupMessage {
	return ElevatorBackupMessage{
		ResponderIP: elevator.State.LocalIP,
		State:       elevator.State,
		Event:       EventElevatorBackup,
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

func (m ElevatorOrderMessage) IsValid() bool {
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

func (e *Elevator) AddCabOrder(Floor int) {
	e.State.CabOrders[Floor] = true
}

func (e *Elevator) RemoveCabOrder(Floor int) {
	e.State.CabOrders[Floor] = false

}
