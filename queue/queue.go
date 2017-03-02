package queue

import (
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

func AddLocalOrder(button ElevatorButton) {
	cabOrders[button.Floor] = true
}

func ShouldStop(floor, direction int) bool {
	return cabOrders[floor]
}

func ChooseDirection(floor, direction int) int {
	var nextFloor int
	for index, active := range cabOrders {
		if active == true {
			nextFloor = index
			break
		}
	}
	// THIS IS WRONG! Implement algorithm from example project at github.
	switch {
	case floor == nextFloor:
		return MotorStop
	case floor < nextFloor:
		return MotorUp
	case floor > nextFloor:
		return MotorDown
	default:
		// Error handling
		return MotorStop
	}
}

func RemoveOrder(floor, direction int) {
	cabOrders[floor] = false
}

// --- //

func anyRequestsAbove() {
	for f := 0; f < NumFloors; f++ {
		if CabOrderMatrix[f].status == Awaiting {
			return true
		}
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[f][k] == Awaiting {
				return true
			}
		}
	}
	return false
}

func anyRequestsBelow() {

}
