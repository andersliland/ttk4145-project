package cost

import . "../config"

func ElevatorCostCalulation(newElevatorOrder ElevatorOrderMessage) (assignedOrder ElevatorOrderMessage, err error) {

	//TODO: calculate cost
	newElevatorOrder.AssignedTo = newElevatorOrder.OriginIP

	return newElevatorOrder, nil
}
