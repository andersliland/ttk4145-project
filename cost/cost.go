package cost

import . "../utilities"

// calculate cost for new orders
func CalculateCost(knownElevators map[string]*Elevator,
	activeElevators map[string]bool,
	HallOrderMatrix [NumFloors][NumButtons - 1]ElevatorOrder,
	Floor int,
	Type int) (string, err error) {

	// based on input, figure out which elevator is best suited to preform the order,
	// return this elevator

	return "", err
}
