package fsm

import (
	. "../config"
	. "../driver"
	//"../network"
	"log"
	"time"
	//"../watchdog"
)

const (
	idle = iota
	moving
	doorOpen
)

const watchdogTimeoutInterval = time.Second * 1
const watchdogKickInterval = watchdogTimeoutInterval / 3

func InitFSM() {

}

// ORDERS
func FSM(buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	sendMessageChannel chan ElevatorOrderMessage,
	receiveMessageChannel chan ElevatorOrderMessage,
	localIP string) {

	wdog := time.NewTicker(watchdogTimeoutInterval)
	defer wdog.Stop()

	wdogKick := time.NewTicker(watchdogKickInterval)
	defer wdogKick.Stop()

	for {
		select {
		case <-wdog.C:
			//log.Println("watchdog timeout")
			// implement timeout handling
			//os.Exit(1)

		case <-wdogKick.C:
			//log.Println("watchdog kick")
			// TODO: implement kick handling

		case b := <-buttonChannel:

			sendMessageChannel <- ElevatorOrderMessage{
				Floor:      b.Floor,
				ButtonType: b.Kind,
				AssignedTo: "AssignedTo",
				OriginIP:   localIP,
				SenderIP:   localIP,
				Event:      23,
			}

			log.Println(b)

			if b.Floor == 0 && b.Kind == 2 {
				motorChannel <- 0
				log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
				lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}
			}
			if b.Floor == 1 && b.Kind == 2 {
				motorChannel <- 1
				log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
				lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}
			}
			if b.Floor == 2 && b.Kind == 2 {
				motorChannel <- 2
				log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
				lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}
			}

		case f := <-floorChannel:
			if f != -1 {
				motorChannel <- 0
			}

		case msg := <-receiveMessageChannel:
			switch msg.Event {
			case idle:
			case moving:
				motorChannel <- MotorDown
			case doorOpen:

			}

		}
	}
}
