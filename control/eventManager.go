package control

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"../driver"
	"../orders"
	. "../utilities"
)

const debugEventManager = false

// Need three functions in an orders.go file to work:
// ShouldStop(floor, direction, localIP)
// ChooseDirection(floor, direction, localIP)
// RemoveFloorOrders(floor, localIP)

func eventManager(
	newOrder chan bool,
	broadcastOrderChannel chan OrderMessage,
	broadcastBackupChannel chan BackupMessage,
	floorReached chan int,
	lightChannel chan ElevatorLight,
	motorChannel chan int, localIP string) {

	var state int = Idle
	var floor int // to initialize or not to initialize?
	var direction int

	// if restore order from file do ..., else:
	const pollDelay = 5 * time.Millisecond

	if err := ElevatorStatus[localIP].LoadFromFile("backupElevator"); err == nil {
		log.Println("[eventManager]\t Executing restored orders")
		for f := 0; f < NumFloors; f++ {
			if ElevatorStatus[localIP].CabOrders[f] {
				newOrder <- true
				break
			}
		}
	}
	floor = driver.GoToFloorBelow(localIP, motorChannel, pollDelay)

	fmt.Print(ColorWhite)
	log.Println("[eventManager]\t New elevator "+localIP+" starting at floor "+strconv.Itoa(floor+1), ColorNeutral)
	time.Sleep(1 * time.Second)
	syncFloor(floor, localIP, broadcastBackupChannel)

	doorTimeout := make(chan bool)
	doorTimerReset := make(chan bool)

	go doorTimer(doorTimeout, doorTimerReset)

	for {
		select {
		case <-newOrder:
			//log.Println("newOrder state: " + StateEventManager[state])
			switch state {
			case Idle:
				direction = syncDirection(orders.ChooseDirection(floor, direction, localIP), localIP, broadcastBackupChannel)
				if orders.ShouldStop(floor, direction, localIP) {
					printEventManager("Stopped at floor " + strconv.Itoa(floor+1))
					doorTimerReset <- true
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					state = syncState(DoorOpen, localIP, broadcastBackupChannel)

				} else {
					motorChannel <- direction
					state = syncState(Moving, localIP, broadcastBackupChannel)
					//newState <- Moving
				}
			case Moving: // Ignore
			case DoorOpen:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
				}
			default: // Insert error handling
			}

		case floor = <-floorReached:
			//log.Println("floorReached state: " + StateEventManager[state])
			syncFloor(floor, localIP, broadcastBackupChannel)
			//log.Println("Floor reached: " + strconv.Itoa(floor+1))
			switch state {
			case Idle:
				printEventManager("Elevator reached floor " + strconv.Itoa(floor+1) + " in state IDLE")

			case Moving:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					motorChannel <- Stop
					state = syncState(DoorOpen, localIP, broadcastBackupChannel)
				}
			case DoorOpen: // not applicable
			default: // Insert error handling
			}
		case <-doorTimeout:
			//log.Println("doorTimeout state: " + StateEventManager[state])
			switch state {
			case Idle: // not applicable
			case Moving: // not applicable
			case DoorOpen:
				lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: false}
				orders.RemoveFloorOrders(floor, direction, localIP, broadcastOrderChannel)

				printEventManager("eventDoorTimeout, Idle: direction: " + MotorStatus[direction+1])
				direction = syncDirection(orders.ChooseDirection(floor, direction, localIP), localIP, broadcastBackupChannel)
				printEventManager("Door closing, new direction is " + MotorStatus[direction+1] + ".  Elevator " + localIP)
				if direction == Stop {
					state = syncState(Idle, localIP, broadcastBackupChannel)
				} else {
					motorChannel <- direction // Is this necessary?
					state = syncState(Moving, localIP, broadcastBackupChannel)
				}
			default: // Insert error handling here - elevator might possibly need to be restarted ()
			}

		}
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

func syncFloor(floor int, localIP string, broadcastBackupChannel chan<- BackupMessage) {
	ElevatorStatus[localIP].Floor = floor
	broadcastBackupChannel <- BackupMessage{State: *ElevatorStatus[localIP], Event: EventElevatorBackup, AskerIP: localIP}
	//log.Println("Sendt ElevatorStatus sync message from syncFloor")

}

func syncDirection(direction int, localIP string, broadcastBackupChannel chan<- BackupMessage) int {
	ElevatorStatus[localIP].Direction = direction
	broadcastBackupChannel <- BackupMessage{State: *ElevatorStatus[localIP], Event: EventElevatorBackup, AskerIP: localIP}
	//log.Println("Sendt ElevatorStatus sync message from syncDirection")

	return direction

}

func syncState(state int, localIP string, broadcastBackupChannel chan<- BackupMessage) int {
	ElevatorStatus[localIP].State = state
	broadcastBackupChannel <- BackupMessage{State: *ElevatorStatus[localIP], Event: EventElevatorBackup, AskerIP: localIP}
	//log.Println("Sendt ElevatorStatus sync message from syncState")
	return state
}

func printEventManager(s string) {
	if debugEventManager {
		log.Println("[eventManager]\t", s)
	}
}
