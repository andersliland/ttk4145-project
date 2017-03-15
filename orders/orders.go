package orders

import (
	"fmt"
	"log"
	"os"
	"strconv"

	. "../utilities"
)

const debugOrders = false

func ShouldStop(floor, direction int, localIP string) bool {
	// cabOrders are checked first, do not depend on direction
	if floor == FloorInvalid {
		fmt.Printf(ColorRed)
		log.Println("Invalid floor "+strconv.Itoa(floor+1)+". The system will terminate.", ColorNeutral)
		os.Exit(1)
	}

	// stop at floor if there is a cabOrder there
	if ElevatorStatus[localIP].CabOrders[floor] == true {
		printOrders("There is a cabOrder at floor " + strconv.Itoa(floor+1) + " for " + localIP)
		return true
	}

	// Check hallOrders up and down for special cases (top floor, bottom floor, no more orders in direction)
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
		log.Println("[orders]\t\t Invalid direction in ShouldStop. Ignoring... ")
		return false
	}
	return false
}

func ChooseDirection(floor, direction int, localIP string) int {
	switch direction {
	case Stop:
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
		log.Println("[orders]\t\t Invalid direction in ChooseDirection. Ignoring... ")
		return Stop
	}
}

func RemoveFloorOrders(floor, direction int, localIP string, broadcastOrderChannel chan<- OrderMessage, orderCompleteChannel chan OrderMessage) {
	if ElevatorStatus[localIP].CabOrders[floor] == true {
		printOrders("Removed CabOrder at floor " + strconv.Itoa(floor+1) + " for " + localIP)
	}
	ElevatorStatusMutex.Lock()
	ElevatorStatus[localIP].CabOrders[floor] = false
	ElevatorStatusMutex.Unlock()

	if err := SaveBackup("backupElevator", ElevatorStatus[localIP].CabOrders); err != nil {
		fmt.Printf(ColorRed)
		log.Println("Write CabOrder backup to file failed: ", err, ColorNeutral)
	}

	switch direction {
	case Stop:
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[floor][k].AssignedTo == localIP {
				orderCompleteChannel <- resolveRemoveOrderMessage(floor, k, localIP) // send to systemControl routine
				broadcastOrderChannel <- resolveRemoveOrderMessage(floor, k, localIP)
				printOrders("Removed HallOrder at floor " + strconv.Itoa(floor+1) + " for direction " + MotorStatus[direction+1] + ". Ip " + localIP)

			}
		}
	case Up:
		if HallOrderMatrix[floor][ButtonCallUp].AssignedTo == localIP {
			orderCompleteChannel <- resolveRemoveOrderMessage(floor, ButtonCallUp, localIP) // send to systemControl routine
			broadcastOrderChannel <- resolveRemoveOrderMessage(floor, ButtonCallUp, localIP)

		}
		if !anyRequestsAbove(floor, localIP) && HallOrderMatrix[floor][ButtonCallDown].AssignedTo == localIP {
			orderCompleteChannel <- resolveRemoveOrderMessage(floor, ButtonCallDown, localIP) // send to systemControl routine
			broadcastOrderChannel <- resolveRemoveOrderMessage(floor, ButtonCallDown, localIP)
			printOrders("Direction up at floor " + strconv.Itoa(floor+1) + ". No new orders above this floor. Removed down order. Elevator: " + localIP)
		}
		printOrders("Removed HallOrder at floor " + strconv.Itoa(floor+1) + " for direction " + MotorStatus[direction+1] + ". Ip " + localIP)

	case Down:
		if HallOrderMatrix[floor][ButtonCallDown].AssignedTo == localIP {
			orderCompleteChannel <- resolveRemoveOrderMessage(floor, ButtonCallDown, localIP) // send to systemControl routine
			broadcastOrderChannel <- resolveRemoveOrderMessage(floor, ButtonCallDown, localIP)
		}
		if !anyRequestsBelow(floor, localIP) && HallOrderMatrix[floor][ButtonCallUp].AssignedTo == localIP {
			orderCompleteChannel <- resolveRemoveOrderMessage(floor, ButtonCallUp, localIP) // send to systemControl routine
			broadcastOrderChannel <- resolveRemoveOrderMessage(floor, ButtonCallUp, localIP)
			printOrders("Direction down at floor " + strconv.Itoa(floor+1) + ". No new orders above this floor. Removed up order. Elevator: " + localIP)
		}
		printOrders("Removed HallOrder at floor" + strconv.Itoa(floor+1) + " for direction " + MotorStatus[direction+1] + ". Ip" + localIP)
	default:
		fmt.Printf(ColorRed)
		log.Println("[order]\t\t Undefined direction for RemoveFloorOrders. Ignoring...", ColorNeutral)
	}

}

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

func resolveRemoveOrderMessage(floor int, button int, localIP string) OrderMessage {
	return OrderMessage{
		Floor:      floor,
		ButtonType: button,
		AssignedTo: localIP,
		OriginIP:   localIP,
		Event:      EventOrderCompleted,
	}
}

func printOrders(s string) {
	if debugOrders {
		log.Println("[orders]\t\t", s)
	}
}
