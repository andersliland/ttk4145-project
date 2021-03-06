// Package driver provides an interface to controlling core functionality of the elevator
package driver

import (
	"log"
	"strconv"
	"time"

	// . "../simulator/simulatorCore" // uncomment to use simulator
	. "../utilities"
	. "../wrapper" // comment out when using simulator
)

var debugElevator = false

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

const motorSpeed = 2800

func Init(buttonChannel chan<- ElevatorButton,
	lightChannel <-chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan<- int,
	pollDelay time.Duration) {

	// SIMULATOR Uncomment below to run simulator
	//SimulatorInit() // Uncomment to start elevator

	// Wrapper Uncomment below to use wrapper
	IoInit()

	resetAllLights()
	go lightController(lightChannel)
	go motorController(motorChannel)
	go floorSensorPoller(floorChannel, pollDelay)
	go buttonPoller(buttonChannel, pollDelay)
}

func resetAllLights() {
	for f := 0; f < NumFloors; f++ {
		for k := ButtonCallUp; k <= ButtonCommand; k++ {
			IoClearBit(lampMatrix[f][k])
		}
	}
	IoClearBit(LIGHT_DOOR_OPEN)
	IoClearBit(LIGHT_STOP)
}

func lightController(lightChannel <-chan ElevatorLight) {
	var command ElevatorLight
	for {
		select {
		case command = <-lightChannel:
			switch command.Kind {
			case ButtonStop:
				if command.Active {
					IoSetBit(LIGHT_STOP)
				} else {
					IoClearBit(LIGHT_STOP)
				}
			case ButtonCallUp, ButtonCallDown, ButtonCommand:
				if command.Active {
					IoSetBit(lampMatrix[command.Floor][command.Kind])
				} else {
					IoClearBit(lampMatrix[command.Floor][command.Kind])
				}
			case DoorIndicator:
				if command.Active {
					IoSetBit(LIGHT_DOOR_OPEN)
				} else {
					IoClearBit(LIGHT_DOOR_OPEN)
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
			case Stop:
				IoWriteAnalog(MOTOR, 0)
			case Up:
				IoClearBit(MOTORDIR)
				IoWriteAnalog(MOTOR, motorSpeed)
			case Down:
				IoSetBit(MOTORDIR)
				IoWriteAnalog(MOTOR, motorSpeed)
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
		if f != prevFloor && f != FloorInvalid {
			prevFloor = f
			SetFloorIndicator(f) // Move to fsm
			floorChannel <- f
		}
		time.Sleep(pollDelay)
	}
}

func buttonPoller(buttonChannel chan<- ElevatorButton, pollDelay time.Duration) {
	inputMatrix := [NumFloors][NumButtons]bool{}
	buttonStopActivated := false
	for {
		for f := 0; f < NumFloors; f++ {
			for k := ButtonCallUp; k <= ButtonCommand; k++ {
				b := IoReadBit(buttonMatrix[f][k])
				if b && inputMatrix[f][k] != b {
					buttonChannel <- ElevatorButton{f, k}
				}
				inputMatrix[f][k] = b
			}
			if s := IoReadBit(STOP); s {
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
	if IoReadBit(SENSOR_FLOOR1) {
		return Floor1
	} else if IoReadBit(SENSOR_FLOOR2) {
		return Floor2
	} else if IoReadBit(SENSOR_FLOOR3) {
		return Floor3
	} else if IoReadBit(SENSOR_FLOOR4) {
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
		IoSetBit(LIGHT_FLOOR_IND1)
	} else {
		IoClearBit(LIGHT_FLOOR_IND1)
	}
	if floor&0x01 > 0 {
		IoSetBit(LIGHT_FLOOR_IND2)
	} else {
		IoClearBit(LIGHT_FLOOR_IND2)
	}
}

func GoToFloorBelow(localIP string, motorChannel chan int, pollDelay time.Duration) int {
	if readFloorSensor() == FloorInvalid {
		printElevator("ReadFloorSensor " + strconv.Itoa(readFloorSensor()))
		motorChannel <- Down
		for {
			if floor := readFloorSensor(); floor != FloorInvalid {
				motorChannel <- Stop
				return floor
			} else {
				time.Sleep(pollDelay)
			}
		}
	}
	return readFloorSensor()
}

func printElevator(s string) {
	if debugElevator {
		log.Println("[elevator]\t\t ", s)
	}
}
