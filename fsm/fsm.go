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
			if b.Floor == 1 && b.Kind == ButtonCallUp {
			//	motorChannel <- 1
			log.Println("test")
				lightChannel <- ElevatorLight{ Floor : 2, Kind : 0, Active : true}
			}
			/*
			if b.Floor == 0  && b.Kind == 0 {
					log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
					lightChannel <- ElevatorLight{ Floor : b.Floor, Kind : b.Kind, Active : true}
				}
			if b.Floor == 2 && b.Kind == BUTTON_COMMAND2 {
					motorChannel <- 2
			}
			if b.Floor == 3 && b.Kind == BUTTON_COMMAND3 {
					motorChannel <- 0
			}
			*/
	}
 }
}
