package queue

import (
	"log"
	"time"

	. "../utilities"
)

// TO BE REMOVED? What is this? [comment by: Sondre]
type orderStatus struct {
	active bool
	ip     string
	timer  time.Timer
}

func AddCabOrder(button ElevatorButton, localIP string) {
	RegisteredElevators[localIP].State.CabOrders[button.Floor] = true
}

func ShouldStop(floor, direction int, localIP string) bool {
	// For cabOrders: stop if it exists
	if RegisteredElevators[localIP].State.CabOrders[floor] == true {
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

func ChooseDirection(floor, direction int, localIP string) int {
	switch direction {
	case MotorStop:
		// For MotorStop, a "closestOrderedFloor" could possibly be used for further optimization
		if anyRequestsAbove(floor, localIP) {
			return MotorUp
		} else if anyRequestsBelow(floor, localIP) {
			return MotorDown
		} else {
			return MotorUp
		}
	case MotorUp:
		if anyRequestsAbove(floor, localIP) {
			return MotorUp
		} else if anyRequestsBelow(floor, localIP) {
			return MotorDown
		} else {
			return MotorUp
		}
	case MotorDown:
		if anyRequestsBelow(floor, localIP) {
			return MotorDown
		} else if anyRequestsAbove(floor, localIP) {
			return MotorUp
		} else {
			return MotorStop
		}
	default:
		// Insert error handling here
		return MotorStop
	}
}

// Does this function also need to send a message on the sendMessageChannel to notify that it has removed an order?
func RemoveFloorOrders(floor, direction int, localIP string) {
	RegisteredElevators[localIP].State.CabOrders[floor] = false
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
		if RegisteredElevators[localIP].State.CabOrders[f] {
			return true
		}
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if OrderMatrix[floor][k].AssignedTo == localIP {
				return true
			}
		}
	}
	return false
}

func anyRequestsBelow(floor int, localIP string) bool {
	for f := 0; f < floor; f++ {
		if RegisteredElevators[localIP].State.CabOrders[f] {
			return true
		}
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if OrderMatrix[floor][k].AssignedTo == localIP {
				return true
			}
		}
	}
	return false
}
