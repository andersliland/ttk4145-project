package orders

import (
	"log"

	. "../utilities"
)

func AddCabOrder(button ElevatorButton, localIP string) {
	ElevatorStatus[localIP].CabOrders[button.Floor] = true
}

func ShouldStop(floor, direction int, localIP string) bool {
	// cabOrders are checked first, do not depend on direction
	if ElevatorStatus[localIP].CabOrders[floor] == true {
		return true
	}

	// Check both hallOrders, and for special cases (top floor, bottom floor, no more orders in direction)
	switch direction {
	case MotorStop:
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[floor][k].AssignedTo == localIP {
				return true
			}
		}
	case MotorUp:
		if HallOrderMatrix[floor][ButtonCallUp].AssignedTo == localIP {
			return true
		}
		return floor == NumFloors-1 || !anyRequestsAbove(floor, localIP)
	case MotorDown:
		if HallOrderMatrix[floor][ButtonCallDown].AssignedTo == localIP {
			return true
		}
		return floor == 0 || !anyRequestsBelow(floor, localIP)
	default:
		// Insert error handling here
	}
	return false
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
	// Might be stuck in a loop (between say 2 floors), depends on implementation in eventManager.go
}

// Does this function also need to send a message on the sendBroadcastChannel to notify that it has removed an order?
func RemoveFloorOrders(floor, direction int, localIP string) {
	ElevatorStatus[localIP].CabOrders[floor] = false
	switch direction {
	case MotorUp:
		HallOrderMatrix[floor][ButtonCallUp].Status = NotActive
	case MotorDown:
		HallOrderMatrix[floor][ButtonCallDown].Status = NotActive
	default:
		log.Println("ERROR [order]: Undefined direction for RemoveFloorOrders")
	}
}

// --- //

func anyRequestsAbove(floor int, localIP string) bool {
	for f := floor + 1; f < NumFloors; f++ {
		if ElevatorStatus[localIP].CabOrders[f] {
			return true
		}
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[floor][k].AssignedTo == localIP {
				return true
			}
		}
	}
	return false
}

func anyRequestsBelow(floor int, localIP string) bool {
	for f := 0; f < floor; f++ {
		if ElevatorStatus[localIP].CabOrders[f] {
			return true
		}
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[floor][k].AssignedTo == localIP {
				return true
			}
		}
	}
	return false
}
