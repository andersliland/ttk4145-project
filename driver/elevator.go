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
	Type  int
	Floor int
}

type ElevLight struct {
	Type   int
	Floor  int
	Active bool
}

/* 	<-chan : accepts a channel for SENDING values
chan<- : accepts a channel for RECEIVING values
chan : bidirectional
*/
func Init(buttonChannel chan<- ElevButton, lightChannel <-chan ElevLight, motorChannel chan int, floorChannel chan<- int, pollDelay time.Duration) error {
	if err := IOInit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
