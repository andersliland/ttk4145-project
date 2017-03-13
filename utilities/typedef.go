package utilities

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const debug = false

const NumElevators = 3
const NumButtons = 3
const NumFloors = 4
const OrderTimeout = 14 //seconds

var HallOrderMatrix [NumFloors][2]HallOrder
var ElevatorStatus = make(map[string]*Elevator) // containing last known state
var OnlineElevators = make(map[string]bool)

var EventType = []string{
	// BackupMessage Events
	"EventElevatorOnline",
	"EventElevatorBackup",
	"EventRequestBackup",
	"EventBackupReturned",

	// OrderMessage Events
	"EventNewOrder",
	"EventAckNewOrder",
	"EventOrderConfirmed",
	"EventAckOrderConfirmed",
	"EventAckOrderCompleted",
	"EventOrderCompleted",
	"EventReassignOrder",
}

// TODO: UPDATE network module if any changes in events
//TODO: remember to update IsValid when change Events
const (
	// BackupMessage Events
	EventElevatorOnline = iota //  = 0
	EventElevatorBackup
	EventRequestBackup
	EventBackupReturned
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

const (
	Idle = iota
	Moving
	DoorOpen
)

var StateEventManager = []string{
	"Idle",
	"Moving",
	"DoorOpen",
}

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

type ExtendedHallOrder struct {
	Floor        int
	ButtonType   int
	TimeoutState int
	OriginIP     string
	Order        HallOrder
}

type HallOrder struct {
	Status      int
	AssignedTo  string
	ConfirmedBy map[string]bool
	Timer       *time.Timer
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
	ColorRed     = "\x1b[31;1m"
	ColorGreen   = "\x1b[32;1m"
	ColorYellow  = "\x1b[33;1m"
	ColorBlue    = "\x1b[34;1m"
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
	//if m.AskerIP == m.ResponderIP {
	//	return false
	//}
	//if m.Event > 4 || m.Event < 0 {
	//	return false
	//}
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
	//if m.Event > 10 || m.Event < 5 {
	//	return false
	//}
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

func (order *HallOrder) StopTimer() bool {
	if order.Timer != nil {
		return order.Timer.Stop()
	}
	return false

}

func (state *Elevator) SaveToFile(filename string) error {
	data, err := json.Marshal(&state)
	if err != nil {
		log.Println("json.Marshal() error: Failed to marshal backup")
		return err
	}
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		log.Println("ioutil.WriteFile() error: Failed to save backup")
		return err
	}
	return nil
}

func (state *Elevator) LoadFromFile(filename string) error {
	if _, fileNotFound := os.Stat(filename); fileNotFound == nil {
		log.Println("Backup file found")
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println("loadFromDisk() error: Failed to read file")
		}
		if err := json.Unmarshal(data, state); err != nil {
			log.Println("loadFromDisk() error: Failed to unmarshal")
		}
		return nil
	} else {
		log.Println("Backup file not found")
		return fileNotFound
	}

}
