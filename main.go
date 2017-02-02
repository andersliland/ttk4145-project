package main

import (
	"./cost"
	"./driver"
	"./network"
	"fmt"
)

func main() {
	fmt.Println("Hello World from main")

	driver.Elevator()
	driver.Io()
	network.Network()
	cost.Cost()
	network.Udp()

	network.InitUDP()

}
