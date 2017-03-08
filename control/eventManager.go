package control

import (
	"log"
	"time"

	. "../utilities"

	"../orders"
)

const (
	idle = iota
	moving
	doorOpen
)

var state int = idle
var floor int = FloorInvalid // to initialize or not to initialize?
var direction int

// Need three functions in an orders.go file to work:
// ShouldStop(floor, direction, localIP)
// ChooseDirection(floor, direction, localIP)
// RemoveFloorOrders(floor, localIP)

func eventManager(newOrder chan bool, floorReached chan int,
	lightChannel chan ElevatorLight, motorChannel chan int, localIP string) {
	// if restore order from file do ..., else:
	const pollDelay = 5 * time.Millisecond
	//floor = GoToFloorBelow(motorChannel, pollDelay)

	doorTimeout := make(chan bool)
	doorTimerReset := make(chan bool)
	go doorTimer(doorTimeout, doorTimerReset)
	go stateIndicator() // use for debug purpose only - remove later?

	for {
		select {
		case <-newOrder:
			eventNewOrder(lightChannel, motorChannel, doorTimerReset, localIP)
		case floor = <-floorReached:
			eventFloorReached(lightChannel, motorChannel, doorTimerReset, localIP)
		case <-doorTimeout:
			eventDoorTimeout(lightChannel, motorChannel, localIP)
		}
	}
}

func eventNewOrder(lightChannel chan ElevatorLight, motorChannel chan int, doorTimerReset chan bool, localIP string) {
	switch state {
	case idle:
		direction = orders.ChooseDirection(floor, direction, localIP)
		if orders.ShouldStop(floor, direction, localIP) {
			doorTimerReset <- true
			orders.RemoveFloorOrders(floor, direction, localIP)
			//queue.RemoveFloorOrders(floor, direction, localIP sendMessageChannel) // change the above function with this later
			lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
			state = doorOpen
		} else {
			motorChannel <- direction
			state = moving
		}
	case moving: // Ignore
	case doorOpen:
		if orders.ShouldStop(floor, direction, localIP) {
			doorTimerReset <- true
			orders.RemoveFloorOrders(floor, direction, localIP)
		}
	default:
		// Insert error handling
	}
}

func eventFloorReached(lightChannel chan ElevatorLight, motorChannel chan int, doorTimerReset chan bool, localIP string) {
	//SetFloorIndicator(floor)
	switch state {
	case moving:
		if orders.ShouldStop(floor, direction, localIP) {
			doorTimerReset <- true
			orders.RemoveFloorOrders(floor, direction, localIP)
			lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
			direction = MotorStop
			motorChannel <- MotorStop
			state = doorOpen
		}
	default:
		// Insert error handling
	}
}

func eventDoorTimeout(lightChannel chan ElevatorLight, motorChannel chan int, localIP string) {
	switch state {
	case doorOpen:
		lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: false}
		direction = orders.ChooseDirection(floor, direction, localIP)
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
