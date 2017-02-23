package control

import (
	"time"

	"../driver"
)

const (
	idle = iota
	moving
	doorOpen
)

var state int = idle
var floor int = FloorInvalid // to initialize or not to initialize?
var direction int

// Need three functions in a queue.go file to work:
// ShouldStop(floor, direction)
// ChooseDirection(floor, direction)
// RemoveOrdersAt(floor)

func eventManager(newOrder chan bool, floorReached chan int,
	lightChannel <-chan ElevatorLight, motorChannel chan int) {
	// if restore order from file do ..., else:
	floor = driver.goToFloorBelow()

	doorTimeout := make(chan bool)
	doorTimerReset := make(chan bool)
	go doorTimer(doorTimeout, doorTimerReset)
	// go stateIndicator(state) - implement for debug purpose only

	for {
		select {
		case <-newOrder:
			eventNewOrder(lightChannel, doorTimerReset)
		case floor = <-floorReached:
			eventFloorReached(lightChannel, motorChannel, doorTimerReset)
		case <-doorTimeout:
			eventDoorTimeout(lightChannel, motorChannel)
		}
	}
}

func eventNewOrder(lightChannel <-chan ElevatorLight, doorTimerReset <-chan bool) {
	switch state {
	case idle:
		direction = queue.ChooseDirection(floor, direction)
		if queue.ShouldStop(floor, direction) {
			doorTimerReset <- true
			queue.RemoveOrdersAt(floor, sendMessageChannel)
			lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
			state = doorOpen
		}
	case moving: // Ignore
	case doorOpen:
		if queue.ShouldStop(floor, direction) {
			doorTimerReset <- true
			queue.RemoveOrdersAt(floor, sendMessageChannel)
		}
	default:
		// Insert error handling
	}
}

func eventFloorReached(lightChannel <-chan ElevatorLight, motorChannel chan int, doorTimerReset <-chan bool) {
	driver.SetFloorIndicator(floor)
	switch state {
	case moving:
		if queue.ShouldStop(floor, direction) { // not implemented yet
			doorTimerReset <- true
			queue.RemoveOrdersAt()
			lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
			direction = MotorStop
			motorChannel <- MotorStop
			state = doorOpen
		}
	default:
		// Insert error handling
	}
}

func eventDoorTimeout(lightChannel <-chan ElevatorLight, motorChannel chan int) {
	switch state {
	case doorOpen:
		lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: false}
		direction = queue.ChooseDirection(floor, direction) // not implemented yet
		if direction == MotorStop {
			state = idle
		} else {
			state = moving
		}
	default:
		// Insert error handling here - elevator might possibly need to be restarted ()
	}
}

func doorTimer(timeout chan<- bool, reset <-chan bool) {
	const doorOpenTime = 3 * time.Second
	timer := time.NewTimer(0)
	timer.Stop()

	for {
		select {
		case <-reset:
			timer.Reset(doorOpenTime)
		case <-timer.C:
			timer.Stop()
			timeout <- true
		}
	}
}
