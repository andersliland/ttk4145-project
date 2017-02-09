package fsm

import (
//. "../config"
//. "../driver"
//"../network"
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

/*
func buttonHandler(buttonChannel chan ElevatorButton, lightChannel <-chan ElevatorLight) {

	for {
		select {
		case b := <-buttonChannel:
			//lightChannel <- driver.ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}

		}


	}

}

*/
