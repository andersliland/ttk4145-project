package control

import (
	"log"
	"time"

	. "../utilities"

	. "../driver"
	"../queue"
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
	lightChannel chan ElevatorLight, motorChannel chan int) {
	// if restore order from file do ..., else:
	const pollDelay = 5 * time.Millisecond
	floor = GoToFloorBelow(motorChannel, pollDelay)

	doorTimeout := make(chan bool)
	doorTimerReset := make(chan bool)
	go doorTimer(doorTimeout, doorTimerReset)
	go stateIndicator() // use for debug purpose only - remove later?

	for {
		select {
		case <-newOrder:
			eventNewOrder(lightChannel, motorChannel, doorTimerReset)
		case floor = <-floorReached:
			eventFloorReached(lightChannel, motorChannel, doorTimerReset)
		case <-doorTimeout:
			eventDoorTimeout(lightChannel, motorChannel)
		}
	}
}

func eventNewOrder(lightChannel chan ElevatorLight, motorChannel chan int, doorTimerReset chan bool) {
	switch state {
	case idle:
		direction = queue.ChooseDirection(floor, direction)
		if queue.ShouldStop(floor, direction) {
			doorTimerReset <- true
			queue.RemoveOrder(floor, direction)
			//queue.RemoveOrdersAt(floor, sendMessageChannel) // change the above function with this later
			lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
			state = doorOpen
		} else {
			motorChannel <- direction
			state = moving
		}
	case moving: // Ignore
	case doorOpen:
		if queue.ShouldStop(floor, direction) {
			doorTimerReset <- true
			queue.RemoveOrder(floor, direction)
			//queue.RemoveOrdersAt(floor, sendMessageChannel)
		}
	default:
		// Insert error handling
	}
}

func eventFloorReached(lightChannel chan ElevatorLight, motorChannel chan int, doorTimerReset chan bool) {
	SetFloorIndicator(floor)
	switch state {
	case moving:
		if queue.ShouldStop(floor, direction) { // not implemented yet
			doorTimerReset <- true
			queue.RemoveOrder(floor, direction)
			//queue.RemoveOrdersAt()
			lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
			direction = MotorStop
			motorChannel <- MotorStop
			state = doorOpen
		}
	default:
		// Insert error handling
	}
}

func eventDoorTimeout(lightChannel chan ElevatorLight, motorChannel chan int) {
	switch state {
	case doorOpen:
		lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: false}
		direction = queue.ChooseDirection(floor, direction) // not implemented yet
		if direction == MotorStop {
			state = idle
		} else {
			motorChannel <- direction
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

func stateIndicator() {
	prevState := idle
	for {
		if state != prevState {
			switch state {
			case idle:
				log.Println("STATE [eventManager]: idle")
			case moving:
				log.Println("STATE [eventManager]: moving")
			case doorOpen:
				log.Println("STATE [eventManager]: doorOpen")
			}
			prevState = state
		}
	}
}
