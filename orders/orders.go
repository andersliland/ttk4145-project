package orders

import (
	"log"
	"strconv"

	. "../utilities"
)

const debugOrders = true

func AddCabOrder(button ElevatorButton, localIP string) {
	ElevatorStatus[localIP].CabOrders[button.Floor] = true
	printOrders("Added CabOrder for " + localIP + " at floor " + strconv.Itoa(button.Floor+1))
}

func ShouldStop(floor, direction int, localIP string) bool {
	// cabOrders are checked first, do not depend on direction
	if ElevatorStatus[localIP].CabOrders[floor] == true {
		printOrders("There is a cabOrder at floor " + strconv.Itoa(floor+1) + " for " + localIP)
		return true
	}

	// Check both hallOrders, and for special cases (top floor, bottom floor, no more orders in direction)
	switch direction {
	case MotorStop:

		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[floor][k].AssignedTo == localIP && HallOrderMatrix[floor][k].Status == Awaiting {
				return true
			}
		}
	case MotorUp:
		if HallOrderMatrix[floor][ButtonCallUp].AssignedTo == localIP && HallOrderMatrix[floor][ButtonCallUp].Status == Awaiting {
			return true
		}
		return floor == NumFloors-1 || !anyRequestsAbove(floor, localIP)
	case MotorDown:
		if HallOrderMatrix[floor][ButtonCallDown].AssignedTo == localIP && HallOrderMatrix[floor][ButtonCallDown].Status == Awaiting {
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
		if anyRequestsAbove(floor, localIP) && floor < NumFloors-1 {
			return MotorUp
		} else if anyRequestsBelow(floor, localIP) {
			return MotorDown
		} else {
			return MotorStop
		}
	case MotorUp:
		if anyRequestsAbove(floor, localIP) {
			return MotorUp
		} else if anyRequestsBelow(floor, localIP) {
			return MotorDown
		} else {
			return MotorStop
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

// Does this function also need to send a message on the broadcastOrderChannel to notify that it has removed an order?
func RemoveFloorOrders(floor, direction int, localIP string) {

	if ElevatorStatus[localIP].CabOrders[floor] == true {
		printOrders("Removed CabOrder at floor " + strconv.Itoa(floor+1) + " for " + localIP)
	}
	ElevatorStatus[localIP].CabOrders[floor] = false

	switch direction {
	case MotorUp:
		HallOrderMatrix[floor][ButtonCallUp].Status = NotActive
		printOrders("Removed HallOrder at floor " + strconv.Itoa(floor) + " for direction " + MotorStatus[direction] + ". Ip " + localIP)
		// send order done
		/*
			broadcastOrderChannel <- ElevatorOrderMessage{
				Floor: floor,
				ButtonType: ButtonCallUp,
				//OriginIP: ,
				Event: EventOrderCompleted,
			}
		*/

	case MotorDown:
		HallOrderMatrix[floor][ButtonCallDown].Status = NotActive
		printOrders("Removed HallOrder at floor" + strconv.Itoa(floor) + " for direction " + MotorStatus[direction] + ". Ip" + localIP)

	default:
		log.Println("ERROR [order]: Undefined direction for RemoveFloorOrders")
	}
}

// --- //

func anyRequestsAbove(floor int, localIP string) bool {
	for f := floor + 1; f < NumFloors; f++ { // floor+1 : check floor above
		if ElevatorStatus[localIP].CabOrders[f] {
			printOrders("There is a cabOrder above floor " + strconv.Itoa(floor+1) + " at floor " + strconv.Itoa(f+1) + " for elevator " + localIP)
			return true
		}
		for k := 0; k < NumButtons-1; k++ { // -1 to remove Cab buttons
			if HallOrderMatrix[f][k].AssignedTo == localIP && HallOrderMatrix[f][k].Status == Awaiting {
				printOrders("There is a hallOrder above floor " + strconv.Itoa(floor+1) + " at floor " + strconv.Itoa(f+1) + " for elevator " + localIP)
				return true
			}
		}
	}
	return false
}

func anyRequestsBelow(floor int, localIP string) bool {
	for f := 0; f < floor; f++ {
		if ElevatorStatus[localIP].CabOrders[f] {
			printOrders("There is a cabOrder below floor " + strconv.Itoa(floor+1) + " at floor " + strconv.Itoa(f+1) + " for elevator " + localIP)
			return true
		}
		for k := 0; k < NumButtons-1; k++ {
			if HallOrderMatrix[f][k].AssignedTo == localIP && HallOrderMatrix[f][k].Status == Awaiting {
				printOrders("There is a hallOrder below floor " + strconv.Itoa(floor+1) + " at floor " + strconv.Itoa(f+1) + " for elevator " + localIP)
				return true
			}
		}
	}
	return false
}

func printOrders(s string) {
	if debugOrders {
		log.Println("[orders]\t\t", s)
	}
}
