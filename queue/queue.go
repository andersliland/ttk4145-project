package queue

import (
	"time"

	. "../driver"
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

func ChooseDirection(floor, dir int) int {
	var nextFloor int
	for _, nextFloor = range cabOrders {
		if nextFloor == true {
			break
		}
	}
	switch floor {
	case floor < nextFloor:
		return MotorUp
	case floor > nextFloor:
		return MotorDown
	case floor == nextFloor:
		return MotorStop
		//default:
		// Error handling
	}
}

func RemoveOrder(floor, dir int) {
	cabOrders[floor] = false
}

//RemoveRemoteOrderAt
