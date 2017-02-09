package config

const NumButtons = 3
const NumFloors = 4

const (
	ButtonCallUp = iota
	ButtonCallDown
	ButtonCommand
	DoorIndicator
	FloorSensor
	FloorIndicator
)

const (
	FloorInvalid = iota - 1
	Floor0
	Floor1
	Floor2
	Floor3
)

const (
	MotorStop = iota
	MotorUp
	MotorDown
)
