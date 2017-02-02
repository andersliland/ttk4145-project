package driver

import (
	. "../config"
	"log"
	"time"
)

var lampMatrix = [NumFloors][NumButtons]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}

var buttonMatrix = [NumFloors][NumButtons]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

type ElevButton struct {
	Floor int
	Kind  int
}

type ElevLight struct {
	Floor  int
	Kind   int
	Active bool
}

/* 	<-chan : accepts a channel for SENDING values
chan<- : accepts a channel for RECEIVING values
chan : bidirectional
*/
func Init(buttonChannel chan<- ElevButton, lightChannel <-chan ElevLight, motorChannel chan int, floorChannel chan<- int, pollDelay time.Duration) error {
	if err := ioInit(); err != nil {
		log.Println("Failed: ioInit()")
		return err
	}
	resetAllLights()

	go lightController(lightChannel)

	go readInputs(buttonChannel, pollDelay)

	log.Println("Success: Driver initialization")
	return nil
}

func resetAllLights() {
	for floor := 0; floor < NumFloors; floor++ {
		for kind := ButtonCallUp; kind <= ButtonCommand; kind++ {
			ioClearBit(lampMatrix[floor][kind])
		}
	}
	ioClearBit(LIGHT_DOOR_OPEN)
	ioClearBit(LIGHT_STOP)
}

func lightController(lightChannel <-chan ElevLight) {
	var command ElevLight
	for {
		select {
		case command = <-lightChannel:
			switch command.Kind {
			case ButtonCallUp, ButtonCallDown, ButtonCommand:
				if command.Active {
					ioSetBit(lampMatrix[command.Floor][command.Kind])
				} else {
					ioClearBit(lampMatrix[command.Floor][command.Kind])
				}
			}
		}
	}
}

func readInputs(buttonChannel chan<- ElevButton, pollDelay time.Duration) {
	inputMatrix := [NumFloors][NumButtons]bool{}
	//stopButton := false
	for {
		for floor := 0; floor < NumFloors; floor++ {
			for kind := ButtonCallUp; kind <= ButtonCommand; kind++ {
				input := ioReadBit(buttonMatrix[floor][kind])
				if input && !inputMatrix[floor][kind] { // First occurence of this specific input
					inputMatrix[floor][kind] = true
					buttonChannel <- ElevButton{floor, kind}
				} else {
					inputMatrix[floor][kind] = false
				}
			}
		}
		time.Sleep(pollDelay)
	}
}

//--- HELP FUNCTIONS - TO BE DELETED ---//

func SetLight(floor int, kind int) {
	ioSetBit(lampMatrix[floor][kind])
}
