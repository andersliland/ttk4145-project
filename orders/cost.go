package orders

import (
	"errors"
	"log"
	"sort"

	. "../utilities"
)

const debugCost = false

const timeBetweenFloor = 2 //seconds //TODO: time and test1
const timeAtFloor = 3      //seconds //TODO: update at the end

type orderCost struct {
	Cost int
	IP   string
}

type orderCosts []orderCost

// Calculate cost for each elevator, add to slice, sort for and return IP of elevator with lowest cost
func AssignOrderToElevator(Floor int, Kind int,
	OnlineElevators map[string]bool,
	ElevatorStatus map[string]*Elevator,
	HallHallOrderMatrix [NumFloors][2]HallOrder) (ip string, err error) {

	numOnlineElevators := len(OnlineElevators)
	if numOnlineElevators == 0 {
		return "", errors.New("[cost] Cannot Assign new order with zero active elevators")
	}
	cost := orderCosts{} // initialize slice with empty interface

	for ip, _ := range OnlineElevators { // key, value
		floorCount, stopCount := calculateOrderCost(ip, Floor, Kind, ElevatorStatus[ip], HallHallOrderMatrix)
		cost_num := floorCount*timeBetweenFloor + stopCount*timeAtFloor
		cost = append(cost, orderCost{cost_num, ip})
		printDebug(" Cost of order is " + string(cost_num) + " for IP: " + ip)
	}
	sort.Sort(cost) // smallest value at index 0
	ip = cost[0].IP // ip with lowest cost
	printDebug("Assigned order to " + ip + " with cost " + string(cost[0].Cost))
	return ip, err
}

// for each floor: loop each button, increment floorNum end of each loop
// for each button: figure out if there exsist order below, increment button end of each loop
func calculateOrderCost(localIP string,
	orderFloor int,
	orderButtonKind int,
	elevator *Elevator,
	HallHallOrderMatrix [NumFloors][2]HallOrder) (floorCount, stopCount int) {

	direction := elevator.Direction
	prevFloor := elevator.Floor
	state := elevator.State // Yet to be set

	// Elevator is Idle at the ordered floor
	if direction == Stop && state != Moving && prevFloor == orderFloor {
		return 0, 0
	}

	searchDirection := direction
	if orderFloor > prevFloor {
		if !(searchDirection == Down && anyRequestsBelow(prevFloor, localIP)) {
			searchDirection = Up
		}
	} else if orderFloor < prevFloor {
		if !(searchDirection == Up && anyRequestsAbove(prevFloor, localIP)) {
			searchDirection = Down
		}
	}

	// increment floor based on direction of order
	for f := prevFloor + searchDirection; f < NumFloors && f >= Floor1; f += searchDirection {
		floorCount++
		if f == orderFloor {
			if f == 0 || f == NumFloors-1 {
				return floorCount, stopCount
			}
			//else if direction == Down {

		}

	}

	return 0, 0
	//return floorCount, stopCount
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

func printDebug(s string) {
	if debugCost {
		log.Println("[cost]", s)
	}
}
