package control

import (
	"log"
	"time"

	"../orders"
	. "../utilities"
)

const debugEventManager = false

func EventManager(
	newOrder chan bool,
	floor int,
	elevatorStatusChannel chan Elevator,
	broadcastOrderChannel chan<- OrderMessage,
	broadcastBackupChannel chan<- BackupMessage,
	orderCompleteChannel chan OrderMessage,
	floorReached <-chan int,
	lightChannel chan<- ElevatorLight,
	motorChannel chan int, localIP string) {

	doorTimeout := make(chan bool)
	doorTimerReset := make(chan bool)

	var state int = Idle
	var direction int

	// initial sync
	syncFloor(floor, localIP, broadcastBackupChannel, elevatorStatusChannel)
	go doorTimer(doorTimeout, doorTimerReset)

	for {
		select {
		case <-newOrder:
			//log.Println("newOrder state: " + StateEventManager[state])
			switch state {
			case Idle:
				direction = syncDirection(orders.ChooseDirection(floor, direction, localIP), localIP, broadcastBackupChannel, elevatorStatusChannel)
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					state = syncState(DoorOpen, localIP, broadcastBackupChannel, elevatorStatusChannel)

				} else {
					motorChannel <- direction
					state = syncState(Moving, localIP, broadcastBackupChannel, elevatorStatusChannel)
				}
			case Moving: // Ignore
			case DoorOpen:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
				}
			default: // Insert error handling
				log.Println("[eventManager]\t Invalid state in newOrder")
			}
		case floor = <-floorReached:
			//log.Println("floorReached state: " + StateEventManager[state])
			syncFloor(floor, localIP, broadcastBackupChannel, elevatorStatusChannel)
			//log.Println("Floor reached: " + strconv.Itoa(floor+1))
			switch state {
			case Idle:
			case Moving:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					motorChannel <- Stop
					state = syncState(DoorOpen, localIP, broadcastBackupChannel, elevatorStatusChannel)
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
				orders.RemoveFloorOrders(floor, direction, localIP, broadcastOrderChannel, orderCompleteChannel)
				//printEventManager("eventDoorTimeout, Idle: direction: " + MotorStatus[direction+1])
				direction = syncDirection(orders.ChooseDirection(floor, direction, localIP), localIP, broadcastBackupChannel, elevatorStatusChannel)
				//printEventManager("Door closing, new direction is " + MotorStatus[direction+1] + ".  Elevator " + localIP)
				if direction == Stop {
					state = syncState(Idle, localIP, broadcastBackupChannel, elevatorStatusChannel)
				} else {
					motorChannel <- direction // Is this necessary?
					state = syncState(Moving, localIP, broadcastBackupChannel, elevatorStatusChannel)
				}
			default: // Insert error handling here - elevator might possibly need to be restarted ()
				log.Println("[eventManager]\t Invalid state in doorTimeout")
			}
		}
	}
}

func doorTimer(timeout chan<- bool, reset <-chan bool) {
	timer := time.NewTimer(0)
	timer.Stop()
	for {
		select {
		case <-reset:
			timer.Reset(DoorOpenTime * time.Second)
		case <-timer.C:
			timer.Stop()
			timeout <- true
		}
	}
}

func syncFloor(floor int, localIP string, broadcastBackupChannel chan<- BackupMessage, elevatorStatusChannel chan Elevator) {
	//ElevatorStatus[localIP].Floor = floor //TODO: send on channel to main
	elevatorStatusChannel <- Elevator{Floor: floor, LocalIP: localIP}
	broadcastBackupChannel <- BackupMessage{State: *ElevatorStatus[localIP], Event: EventElevatorBackup, AskerIP: localIP}
	//log.Println("Sendt ElevatorStatus sync message from syncFloor")

}

func syncDirection(direction int, localIP string, broadcastBackupChannel chan<- BackupMessage, elevatorStatusChannel chan Elevator) int {
	//ElevatorStatus[localIP].Direction = direction //TODO: send on channel to main
	elevatorStatusChannel <- Elevator{Direction: direction, LocalIP: localIP}
	broadcastBackupChannel <- BackupMessage{State: *ElevatorStatus[localIP], Event: EventElevatorBackup, AskerIP: localIP}
	//log.Println("Sendt ElevatorStatus sync message from syncDirection")
	return direction

}

func syncState(state int, localIP string, broadcastBackupChannel chan<- BackupMessage, elevatorStatusChannel chan Elevator) int {

	//ElevatorStatus[localIP].State = state //TODO: send on channel to main
	elevatorStatusChannel <- Elevator{State: state, LocalIP: localIP}
	broadcastBackupChannel <- BackupMessage{State: *ElevatorStatus[localIP], Event: EventElevatorBackup, AskerIP: localIP}
	//log.Println("Sendt ElevatorStatus sync message from syncState")
	return state
}

func printEventManager(s string) {
	if debugEventManager {
		log.Println("[eventManager]\t", s)
	}
}
