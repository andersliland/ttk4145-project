package driver

import (
	"log"
	"time"

	. "../config"
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

const motorSpeed = 2800

/* 	<-chan : accepts a channel for SENDING values
chan<- : accepts a channel for RECEIVING values
chan : bidirectional
*/
func Init(buttonChannel chan<- ElevButton, lightChannel <-chan ElevLight, motorChannel chan int, floorChannel chan<- int, pollDelay time.Duration) {
	ioInit()
	resetAllLights()
	go lightController(lightChannel)
	go motorController(motorChannel)

	goToFloorBelow(motorChannel, pollDelay) // move to fsm, before for-select (include shouldStop() )
	go floorSensorPoller(floorChannel, pollDelay)
	go readInputs(buttonChannel, pollDelay)

	log.Println("SUCCESS: Driver initialization")
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
			case DoorIndicator:
				if command.Active {
					ioSetBit(LIGHT_DOOR_OPEN)
				} else {
					ioClearBit(LIGHT_DOOR_OPEN)
				}
			default:
				log.Println("ERROR [driver]: Invalid light command")
			}
		}
	}
}

func motorController(motorChannel chan int) {
	var command int
	for {
		select {
		case command = <-motorChannel:
			switch command {
			case MotorStop:
				ioWriteAnalog(MOTOR, 0)
			case MotorUp:
				ioClearBit(MOTORDIR)
				ioWriteAnalog(MOTOR, motorSpeed)
			case MotorDown:
				ioSetBit(MOTORDIR)
				ioWriteAnalog(MOTOR, motorSpeed)
			default:
				log.Println("ERROR [driver]: Invalid motor command")
			}
		}
	}
}

func floorSensorPoller(floorChannel chan<- int, pollDelay time.Duration) {
	prevFloor := FloorInvalid
	for {
		f := readFloorSensor()
		if f != prevFloor && f != -1 {
			setFloorIndicator(f) // Move to fsm
			floorChannel <- f
		}
		prevFloor = f
		time.Sleep(pollDelay)
	}
}

func readInputs(buttonChannel chan<- ElevButton, pollDelay time.Duration) { // rename: buttonPoller
	inputMatrix := [NumFloors][NumButtons]bool{}
	for {
		for floor := 0; floor < NumFloors; floor++ {
			for kind := ButtonCallUp; kind <= ButtonCommand; kind++ {
				input := ioReadBit(buttonMatrix[floor][kind])
				if input && inputMatrix[floor][kind] != input { // First occurrence of this specific input
					buttonChannel <- ElevButton{floor, kind}
				}
				inputMatrix[floor][kind] = input
			}
		}
		time.Sleep(pollDelay)
	}
}

func readFloorSensor() int {
	if ioReadBit(SENSOR_FLOOR1) {
		return Floor1
	} else if ioReadBit(SENSOR_FLOOR2) {
		return Floor2
	} else if ioReadBit(SENSOR_FLOOR3) {
		return Floor3
	} else if ioReadBit(SENSOR_FLOOR4) {
		return Floor4
	} else {
		return FloorInvalid
	}
}

func setFloorIndicator(floor int) {
	if floor < 0 || floor >= NumFloors {
		log.Printf("ERROR [driver]: Floor %d out of range!\n", floor)
		return
	}

	if floor&0x02 > 0 {
		ioSetBit(LIGHT_FLOOR_IND1)
	} else {
		ioClearBit(LIGHT_FLOOR_IND1)
	}
	if floor&0x01 > 0 {
		ioSetBit(LIGHT_FLOOR_IND2)
	} else {
		ioClearBit(LIGHT_FLOOR_IND2)
	}
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

func goToFloorBelow(motorChannel chan int, pollDelay time.Duration) {
	if readFloorSensor() == FloorInvalid {
		motorChannel <- MotorUp
		for {
			if readFloorSensor() != FloorInvalid {
				motorChannel <- MotorStop
				break
			} else {
				time.Sleep(pollDelay)
			}
		}
	}
}

//--- HELP FUNCTIONS - TO BE DELETED ---//

func SetLight(floor int, kind int) {
	ioSetBit(lampMatrix[floor][kind])
}
