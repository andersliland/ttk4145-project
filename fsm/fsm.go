package fsm

import (
	//. "../config"
	. "../driver"
	//"../network"
	"log"
)

const (
	idle = iota
	moving
	doorOpen
)

func InitFSM() {
	go shouldStop()
}

func FSM() {
	log.Println("")
	/*
		switch state {
		case idle:

		case moving:

		motorChannel <- MotorDown

		case doorOpen:
	*/
}

func shouldStop() {

}

func ButtonHandler(buttonChannel chan ElevatorButton, lightChannel chan ElevatorLight, motorChannel chan int) {
	for {
		select {
		case  b := <-buttonChannel:
			if b.Floor == 0 && b.Kind == 2 {
			motorChannel <- 0
			log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
			lightChannel <- ElevatorLight{ Floor : b.Floor, Kind : b.Kind, Active : true}
			}
			if b.Floor == 1 && b.Kind == 2 {
			motorChannel <- 1
			log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
			lightChannel <- ElevatorLight{ Floor : b.Floor, Kind : b.Kind, Active : true}
			}
			if b.Floor == 2 && b.Kind == 2 {
			motorChannel <- 2
			log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
			lightChannel <- ElevatorLight{ Floor : b.Floor, Kind : b.Kind, Active : true}
			}

	}
 }
}
