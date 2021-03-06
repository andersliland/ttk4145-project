package orders

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"time"

	. "../utilities"
)

const debugCost = false

type orderCost struct {
	Cost int
	IP   string
}

type orderCosts []orderCost

// Calculate cost for each elevator, add to slice, sort for and return IP of elevator with lowest cost
func AssignOrderToElevator(Floor int, Kind int,
	OnlineElevators map[string]bool,
	ElevatorStatus map[string]*Elevator) (ip string, err error) {

	numOnlineElevators := len(OnlineElevators)
	if numOnlineElevators == 0 {
		return "", errors.New("[cost] Cannot Assign new order with zero active elevators")
	}
	cost := orderCosts{} // initialize slice with empty interface

	for ip, _ := range OnlineElevators { // key, value
		floorCount, stopCount := calculateOrderCost(ip, Floor, Kind, ElevatorStatus[ip])
		cost_num := floorCount*TimeBetweenFloors + stopCount*DoorOpenTime
		cost = append(cost, orderCost{cost_num, ip})
		printCost("Cost of order is " + strconv.Itoa(cost_num) + " for IP: " + ip)
		printCost("floorCount: " + strconv.Itoa(floorCount) + " stopCount: " + strconv.Itoa(stopCount))
	}
	sort.Sort(cost) // smallest value at index 0
	ip = cost[0].IP // ip with lowest cost
	printCost("Assigned order to " + ip + " with cost " + strconv.Itoa(cost[0].Cost))
	return ip, err
}

// for each floor: loop each button, increment floorNum end of each loop
// for each button: figure out if there exsist order below, increment button end of each loop
func calculateOrderCost(ip string,
	orderFloor int,
	orderButtonKind int,
	elevator *Elevator) (floorCount, stopCount int) {

	direction := elevator.Direction
	prevFloor := elevator.Floor
	//state := elevator.State // Yet to be set

	floorCount = 0
	stopCount = 0

	printCost("Elevator direction: " + MotorStatus[direction+1])

	// Elevator is idle at the ordered floor
	if direction == Stop && prevFloor == orderFloor {
		return floorCount, stopCount
	}

	searchDirection := direction
	if orderFloor > prevFloor {
		if !(searchDirection == Down && anyRequestsBelow(prevFloor, ip)) {
			searchDirection = Up
		}
	} else if orderFloor < prevFloor {
		if !(searchDirection == Up && anyRequestsAbove(prevFloor, ip)) {
			searchDirection = Down
		}
	}

	printCost("Search direction: " + MotorStatus[searchDirection+1])
	printCost("Elevator state: " + strconv.Itoa(ElevatorStatus[ip].State))

	// increment floor based on direction of order
	for f := prevFloor + searchDirection; f < NumFloors && f >= Floor1; f += searchDirection {
		time.Sleep(200 * time.Millisecond) // TODO: remove
		floorCount++
		printCost("Current floor in cost loop, f = " + strconv.Itoa(f+1))
		if f == orderFloor {
			if f == Floor1 || f == NumFloors-1 {
				return floorCount, stopCount
			} else if (direction == Down && orderButtonKind == ButtonCallDown) ||
				(direction == Up && orderButtonKind == ButtonCallUp) {
				printCost("Order continuing same direction as elevator direction")
				return floorCount, stopCount
			} else {
				if searchDirection == Up && !anyRequestsAbove(orderFloor, ip) {
					return floorCount, stopCount
				} else if searchDirection == Down && !anyRequestsBelow(orderFloor, ip) {
					return floorCount, stopCount
				}
			}
		}

		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			// HallOrderMatrix[f][k].Status is never set to UnderExecution - should be done in systemControl.go
			if HallOrderMatrix[f][k].AssignedTo == ip && HallOrderMatrix[f][k].Status == UnderExecution {
				stopCount++
				break
			}
			if ElevatorStatus[ip].CabOrders[f] {
				stopCount++
				break
			}
		}

		if f == NumFloors-1 {
			direction = Down
		} else if f == Floor1 {
			direction = Up
		}
	}
	return floorCount, stopCount
}

// Implement sort.Interface - Len, Less and Swap of type orderCost
// so we can use the sort packages generic Sort function
// Number of elements in collection
func (s orderCosts) Len() int {
	return len(s)
}

// Less reports whether the element with
// index i should sort before the element with index j
func (s orderCosts) Less(i, j int) bool {
	if s[i].Cost != s[j].Cost {
		return s[i].Cost < s[j].Cost
	}
	return s[i].IP < s[j].IP // if equal cost, choose Elevator with lowest IP
}

// Swaps the elements with indexes i and j
func (s orderCosts) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func printCost(s string) {
	if debugCost {
		log.Println("[cost]\t\t", s)
	}
}
