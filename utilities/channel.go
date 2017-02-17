package utilities

import "os"

type Channels struct {
	Driver
	Network
	safeKill chan os.Signal
}

type Driver struct {
	button chan driver.ElevatorButton
	light  chan driver.ElevatorLight
	motor  chan int
	floor  chan int
}

type Network struct {
	sendMessage   chan ElevetorOrderMessage
	receiveOrder  chan ElevatorOrderMessage
	sendBackup    chan ElevatorBackupMessage
	receiveBackup chan ElevatorBackup
}