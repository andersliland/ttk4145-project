package queue

import (
	"log"
	"time"

	. "../utilities"
)

var cabOrders [NumFloors]bool

type orderStatus struct {
	active bool
	ip     string
	timer  time.Timer
}

//func (q queue) setOrder(floor, button int, status orderStatus){
//
//}

func AddLocalOrder(button ElevatorButton, localIP string) {
	RegisteredElevators[localIP].State.InternalOrders[button.Floor] = true
}

func ShouldStop(floor, direction int, localIP string) bool {
	// For cabOrders: stop if it exists
	if RegisteredElevators[localIP].State.InternalOrders[floor] == true {
		return true
	}

	// For hallOrders: Is THIS floor assigned to me?
	// Notice that it does not differentiate on the motor direction here,
	// that logic must be implemented in the cost function (if implemented)
	for k := ButtonCallUp; k <= ButtonCallDown; k++ {
		if OrderMatrix[floor][k].AssignedTo == localIP {
			return true
		}
	}

	return false

	// NOT THE WAY TO GO:
	// run cost function -- determine whether the elevator should stop based on this
	// cost function returns next floor
}

func ChooseDirection(floor, direction int) int {
	// For cabOrders: choose a direction based on

	// For hallOrders: is THERE A floor assigned to me?

	// NOT THE WAY TO GO:
	// return floor from cost function -- determine direction based on this

	var nextFloor int
	for index, active := range cabOrders {
		if active == true {
			nextFloor = index
			break
		}
	}
	log.Println(nextFloor)

	// nextFloor is now found, choose a direction based on that and the current floor the elevator is at

	// THIS IS WRONG! Implement algorithm from example project at github.
	switch direction {
	case MotorDown:
		return MotorStop
	case MotorUp:
	case MotorStop:
	default:
		// Error handling
		return MotorStop
	}
}

func RemoveFloorOrders(floor, direction int, localIP string) {
	RegisteredElevators[localIP].State.InternalOrders[floor] = false
	switch direction {
	case MotorUp:
		OrderMatrix[floor][ButtonCallUp].Status = NotActive
	case MotorDown:
		OrderMatrix[floor][ButtonCallDown].Status = NotActive
	default:
		log.Println("ERROR [queue]: Undefined direction for RemoveFloorOrders")
	}
}

// --- //

func anyRequestsAbove(floor int, localIP string) bool {
	for f := floor + 1; f < NumFloors; f++ {
		if RegisteredElevators[localIP].State.InternalOrders[f] {
			return true
		}
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if OrderMatrix[f][k].Status > 0 {
				return true
			}
		}
	}
	return false
}

func anyRequestsBelow(floor int, localIP string) bool {
	for f := 0; f < floor; f++ {
		if RegisteredElevators[localIP].State.InternalOrders[f] {
			return true
		}
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if OrderMatrix[f][k].Status > 0 {
				return true
			}
		}
	}
	return false
}
