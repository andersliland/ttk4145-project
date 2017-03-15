package utilities

import (
	"sync"
	"time"
)

var HallOrderMatrix [NumFloors][2]HallOrder
var ElevatorStatus = make(map[string]*Elevator) // containing last known state

var HallOrderMatrixMutex = &sync.Mutex{}
var ElevatorStatusMutex = &sync.Mutex{}

type ElevatorButton struct {
	Floor int
	Kind  int
}

type ElevatorLight struct {
	Floor  int
	Kind   int
	Active bool
}

type HallOrder struct {
	Status      int
	AssignedTo  string
	ConfirmedBy map[string]bool
	Timer       *time.Timer
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

type Elevator struct {
	LocalIP         string
	Time            time.Time
	State           int // Idle, Moving, DoorOpen
	Floor           int
	Direction       int // Stop, Up, Down
	CabOrders       [NumFloors]bool
	HallOrderMatrix [NumFloors][2]HallOrder
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

type BackupMessage struct {
	AskerIP         string
	ResponderIP     string
	Event           int
	State           Elevator
	Cab             CabOrder
	HallOrderMatrix [NumFloors][2]HallOrder
}

func ResolveElevator(e Elevator) *Elevator {
	return &Elevator{
		LocalIP:   e.LocalIP,
		Time:      time.Now(),
		CabOrders: e.CabOrders,
	}
}

func ResolveWatchdogKickMessage(e *Elevator) BackupMessage {
	return BackupMessage{
		ResponderIP: e.LocalIP,
		Event:       EventElevatorOnline,
		State:       *e,
	}

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
