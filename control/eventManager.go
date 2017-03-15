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

	// initial sync of flor status
	setFloor(floor, localIP)
	go doorTimer(doorTimeout, doorTimerReset)

	for {
		select {
		case <-newOrder:
			switch state {
			case Idle:
				direction = setDirection(orders.ChooseDirection(floor, direction, localIP), localIP)
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					state = setState(DoorOpen, localIP)

				} else {
					motorChannel <- direction
					state = setState(Moving, localIP)
				}
			case Moving: // not used
			case DoorOpen:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
				}
			default:
				log.Println("[eventManager]\t Invalid state in newOrder")
			}
		case floor = <-floorReached:
			setFloor(floor, localIP)
			switch state {
			case Idle:
			case Moving:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					motorChannel <- Stop
					state = setState(DoorOpen, localIP)
				}
			case DoorOpen: // not used
			default:
				log.Println("[eventManager]\t Invalid state in floorReached")
			}
		case <-doorTimeout:
			//log.Println("doorTimeout state: " + StateEventManager[state])
			switch state {
			case Idle: // not used
			case Moving: // not used
			case DoorOpen:
				lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: false}
				orders.RemoveFloorOrders(floor, direction, localIP, broadcastOrderChannel, orderCompleteChannel)
				printEventManager("eventDoorTimeout, Idle: direction: " + MotorStatus[direction+1])
				direction = setDirection(orders.ChooseDirection(floor, direction, localIP), localIP)
				printEventManager("Door closing, new direction is " + MotorStatus[direction+1] + ".  Elevator " + localIP)
				if direction == Stop {
					state = setState(Idle, localIP)
				} else {
					motorChannel <- direction
					state = setState(Moving, localIP)
				}
			default:
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

func setFloor(floor int, localIP string) {
	ElevatorStatusMutex.Lock()
	ElevatorStatus[localIP].Floor = floor
	ElevatorStatusMutex.Unlock()
}

func setDirection(direction int, localIP string) int {
	ElevatorStatusMutex.Lock()
	ElevatorStatus[localIP].Direction = direction
	ElevatorStatusMutex.Unlock()
	return direction
}

func setState(state int, localIP string) int {
	ElevatorStatusMutex.Lock()
	ElevatorStatus[localIP].State = state
	ElevatorStatusMutex.Unlock()
	return state
}

func printEventManager(s string) {
	if debugEventManager {
		log.Println("[eventManager]\t", s)
	}
}
