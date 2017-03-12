package orders

import (
	"log"
	"os"
	"strconv"

	. "../utilities"
)

const debugOrders = false

func AddCabOrder(button ElevatorButton, localIP string) {
	ElevatorStatus[localIP].CabOrders[button.Floor] = true
	printOrders("Added CabOrder for " + localIP + " at floor " + strconv.Itoa(button.Floor+1))
}

func ShouldStop(floor, direction int, localIP string) bool {
	// cabOrders are checked first, do not depend on direction
	if floor == FloorInvalid {
		printOrders("Invalid floor " + strconv.Itoa(floor+1) + ". The system will terminate.")
		os.Exit(1)
	}

	if ElevatorStatus[localIP].CabOrders[floor] == true {
		printOrders("There is a cabOrder at floor " + strconv.Itoa(floor+1) + " for " + localIP)
		return true
	}

	// Check both hallOrders, and for special cases (top floor, bottom floor, no more orders in direction)
	switch direction {
	case Stop:
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[floor][k].AssignedTo == localIP && HallOrderMatrix[floor][k].Status == Awaiting {
				printOrders("Found order at floor " + strconv.Itoa(floor+1) + " of type " + ButtonType[k] + " (direction: " + MotorStatus[direction+1] + ")")
				return true
			}
		}
	case Up:
		if HallOrderMatrix[floor][ButtonCallUp].AssignedTo == localIP && HallOrderMatrix[floor][ButtonCallUp].Status == Awaiting {
			printOrders("Found order at floor " + strconv.Itoa(floor+1) + " of type " + ButtonType[ButtonCallUp] + " (direction: " + MotorStatus[direction+1] + ")")
			return true
		}
		return floor == NumFloors-1 || !anyRequestsAbove(floor, localIP)
	case Down:
		if HallOrderMatrix[floor][ButtonCallDown].AssignedTo == localIP && HallOrderMatrix[floor][ButtonCallDown].Status == Awaiting {
			printOrders("Found order at floor " + strconv.Itoa(floor+1) + " of type " + ButtonType[ButtonCallDown] + " (direction: " + MotorStatus[direction+1] + ")")
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
	case Stop:
		// For Stop, a "closestOrderedFloor" could possibly be used for further optimization
		if anyRequestsAbove(floor, localIP) && floor < NumFloors-1 {
			return Up
		} else if anyRequestsBelow(floor, localIP) {
			return Down
		} else {
			return Stop
		}
	case Up:
		if anyRequestsAbove(floor, localIP) {
			return Up
		} else if anyRequestsBelow(floor, localIP) {
			return Down
		} else {
			return Stop
		}
	case Down:
		if anyRequestsBelow(floor, localIP) {
			return Down
		} else if anyRequestsAbove(floor, localIP) {
			return Up
		} else {
			return Stop
		}
	default:
		// Insert error handling here
		return Stop
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
	case Stop:
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[floor][k].AssignedTo == localIP {
				HallOrderMatrix[floor][k].Status = NotActive
				printOrders("Removed HallOrder at floor " + strconv.Itoa(floor+1) + " for direction " + MotorStatus[direction+1] + ". Ip " + localIP)

			}
		}
	case Up:
		if HallOrderMatrix[floor][ButtonCallUp].AssignedTo == localIP {
			HallOrderMatrix[floor][ButtonCallUp].Status = NotActive

		}
		if !anyRequestsAbove(floor, localIP) {
			HallOrderMatrix[floor][ButtonCallDown].Status = NotActive
			printOrders("Direction up at floor " + strconv.Itoa(floor+1) + ". No new orders above this floor. Removed down order. Elevator: " + localIP)
		}
		//if floor == NumFloors-1 { // Edge case: top floor reached
		//	HallOrderMatrix[floor][ButtonCallDown].Status = NotActive
		//}
		printOrders("Removed HallOrder at floor " + strconv.Itoa(floor+1) + " for direction " + MotorStatus[direction+1] + ". Ip " + localIP)
		// send order done
		/*
			broadcastOrderChannel <- ElevatorOrderMessage{
				Floor: floor,
				ButtonType: ButtonCallUp,
				//OriginIP: ,
				Event: EventOrderCompleted,
			}
		*/

	case Down:
		if HallOrderMatrix[floor][ButtonCallDown].AssignedTo == localIP {
			HallOrderMatrix[floor][ButtonCallDown].Status = NotActive
		}
		if !anyRequestsBelow(floor, localIP) {
			HallOrderMatrix[floor][ButtonCallUp].Status = NotActive
			printOrders("Direction down at floor " + strconv.Itoa(floor+1) + ". No new orders above this floor. Removed up order. Elevator: " + localIP)
		}
		//if floor == Floor1 { // Edge case: bottom floor reached
		//	HallOrderMatrix[floor][ButtonCallUp].Status = NotActive
		//}
		printOrders("Removed HallOrder at floor" + strconv.Itoa(floor+1) + " for direction " + MotorStatus[direction+1] + ". Ip" + localIP)

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
