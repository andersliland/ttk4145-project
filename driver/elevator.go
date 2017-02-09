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
func Init(buttonChannel chan<- ElevButton, lightChannel <-chan ElevLight, motorChannel chan int, floorChannel chan<- int, pollDelay time.Duration) error {
	if err := ioInit(); err != nil {
		log.Println("FAILED: ioInit()")
		return err
	}
	go lightController(lightChannel)
	go motorController(motorChannel)
	go floorController(floorChannel)
	go readInputs(buttonChannel, pollDelay)

	resetAllLights()
	goToFloorBelow()

	log.Println("SUCCESS: Driver initialization")
	return nil
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
	for {
		select {
		case command <- motorChannel:
			switch command.Kind {
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

func floorController(floorChannel chan<- int, pollDelay time.Duration) {
	prevFloor := FloorInvalid
	var tempFloor int // TODO: declare on one line
	for {
		tempFloor = readFloorSensor()
		if tempFloor != prevFloor {
			prevFloor = tempFloor
			setFloorIndicator(tempFloor)
			floorChannel <- tempFloor
		}
		time.Sleep(pollDelay)
	}
}

func readInputs(buttonChannel chan<- ElevButton, pollDelay time.Duration) {
	inputMatrix := [NumFloors][NumButtons]bool{}
	for {
		for floor := 0; floor < NumFloors; floor++ {
			for kind := ButtonCallUp; kind <= ButtonCommand; kind++ {
				input := ioReadBit(buttonMatrix[floor][kind])
				if input && !inputMatrix[floor][kind] { // First occurrence of this specific input
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

func readFloorSensor() {
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

func goToFloorBelow(pollDelay time.Duration) {
	if readFloorSensor() == FloorInvalid {
		motorChannel <- MotorDown
		for {
			if readFloorSensor() != FloorInvalid {
				floorChannel <- MotorStop
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
