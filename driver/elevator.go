package driver

import (
	"log"
	"time"

	. "../utilities"
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

type ElevatorButton struct {
	Floor int
	Kind  int
}

type ElevatorLight struct {
	Floor  int
	Kind   int
	Active bool
}

const motorSpeed = 2800

/* 	<-chan : accepts a channel for SENDING values
chan<- : accepts a channel for RECEIVING values
chan : bidirectional
*/
func Init(buttonChannel chan<- ElevatorButton,
	lightChannel <-chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan<- int,
	pollDelay time.Duration) {

	ioInit()
	resetAllLights()
	go lightController(lightChannel)
	go motorController(motorChannel)
	go floorSensorPoller(floorChannel, pollDelay)
	go buttonPoller(buttonChannel, pollDelay)
	log.Println("SUCCESS [driver] Initialization")
}

func resetAllLights() {
	for f := 0; f < NumFloors; f++ {
		for k := ButtonCallUp; k <= ButtonCommand; k++ {
			ioClearBit(lampMatrix[f][k])
		}
	}
	ioClearBit(LIGHT_DOOR_OPEN)
	ioClearBit(LIGHT_STOP)
}

func lightController(lightChannel <-chan ElevatorLight) {
	var command ElevatorLight
	for {
		select {
		case command = <-lightChannel:
			switch command.Kind {
			case ButtonStop:
				if command.Active {
					ioSetBit(LIGHT_STOP)
				} else {
					ioClearBit(LIGHT_STOP)
				}
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
			floorChannel <- f
		}
		prevFloor = f
		time.Sleep(pollDelay)
	}
}

func buttonPoller(buttonChannel chan<- ElevatorButton, pollDelay time.Duration) {
	inputMatrix := [NumFloors][NumButtons]bool{}
	buttonStopActivated := false
	for {
		for f := 0; f < NumFloors; f++ {
			for k := ButtonCallUp; k <= ButtonCommand; k++ {
				b := ioReadBit(buttonMatrix[f][k])
				if b && inputMatrix[f][k] != b {
					buttonChannel <- ElevatorButton{f, k}
				}
				inputMatrix[f][k] = b
			}
			if s := ioReadBit(STOP); s {
				if !buttonStopActivated {
					buttonChannel <- ElevatorButton{Kind: ButtonStop}
				}
				buttonStopActivated = s
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

func SetFloorIndicator(floor int) {
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

func goToFloorBelow(motorChannel chan int, pollDelay time.Duration) int {
	if readFloorSensor() == FloorInvalid {
		motorChannel <- MotorDown
		for {
			if floor := readFloorSensor(); floor != FloorInvalid {
				motorChannel <- MotorStop
				break
			} else {
				time.Sleep(pollDelay)
			}
		}
	}
	return floor
}
