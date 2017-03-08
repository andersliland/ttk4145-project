package simulatorCore

//const NumFloors = 4
//const NumButtons = 3

//Motor commands
//const MotorUp = 1
//const MotorStop = 0
//const MotorDown = 2

//---------------SIMULATOR CONFIGURATION PARAMETERS--------------
const DistancePassingFloors = 1820000
const DistanceBetweenFloors = 4200000
const TravelTimeBetweenFloors_ms = 1500 * 2
const TravelTimePassingFloor_ms = 1000
const BtnDepressedTime_ms = 200

//const PortToInterface int = 44044
const PortFromInterface int = 44033

const (
	S_stoppedBetweenFloors = iota
	S_stoppedAtFloor
	S_stoppedInsideSensor
	S_movingUp
	S_movingDown
	S_movingUpInsideSensor
	S_movingDownInsideSensor
)

var MotorStates = [7]string{
	"S_stoppedBetweenFloors",
	"S_stoppedAtFloor",
	"S_stoppedInsideSensor",
	"S_movingUp",
	"S_movingDown",
	"S_movingUpInsideSensor",
	"S_movingDownInsideSensor",
}

type SimulatorElevator struct {
	FloorSensor       [4]bool
	ButtonMatrix      [4][3]bool
	ButtonLightMatrix [4][3]bool
	ObstructionButton bool
	StopButton        bool
	StopButtonLight   bool
	Direction         int
	MotorSpeed        int
	DoorOpen          bool
	LastFloor         int //0-3
}
