package control

import (
	"log"
	"strconv"
	"time"

	"../cost"
	"../orders/"
	. "../utilities/"
)

const debugElevatorControl = true

func MessageLoop(
	newOrder chan bool,
	buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	broadcastOrderChannel chan OrderMessage,
	receiveOrderChannel chan OrderMessage,
	broadcastBackupChannel chan BackupMessage,
	receiveBackupChannel chan BackupMessage,
	OnlineElevators map[string]bool,
	ElevatorStatus map[string]*Elevator,
	HallOrderMatrix [NumFloors][2]HallOrder,
	localIP string) {

	floorReached := make(chan int)
	go eventManager(newOrder, broadcastOrderChannel, floorReached, lightChannel, motorChannel, localIP)
	//go setPanelLights(lightChannel, localIP)

	for {
		select {
		case button := <-buttonChannel: // Hardware
			printElevatorControl("New button push from " + localIP + " of type '" + ButtonType[button.Kind] + "' at floor " + strconv.Itoa(button.Floor+1))
			switch button.Kind {
			case ButtonCallUp, ButtonCallDown:
				orderAssignedTo, err := cost.AssignOrderToElevator(button.Floor, button.Kind, OnlineElevators, ElevatorStatus, HallOrderMatrix)
				//printElevatorControl("Local assign order to " + orderAssignedTo)
				CheckError("[elevatorControl] Failed to assign Order to Elevator ", err)
				order := OrderMessage{
					Time:       time.Now(),
					Floor:      button.Floor,
					ButtonType: button.Kind,
					AssignedTo: orderAssignedTo,
					OriginIP:   localIP,
					SenderIP:   localIP,
					Event:      EventNewOrder,
				}
				broadcastOrderChannel <- order

			case ButtonCommand:
				orders.AddCabOrder(button, localIP)
				newOrder <- true

				/*
					broadcastBackupChannel <- BackupMessage{
						AskerIP: localIP,
						Event:   EventElevatorBackup,
						State: Elevator{
							LocalIP: localIP,
							//LastFloor: <-floorChannel ,
							//	Direction: ,
							//	IsMoving: ,
							//	DoorStatus: ,
							// CabOrders[button.Floor]: true, // why does this not work
						},
						Cab: CabOrder{
							LocalIP:  localIP,
							OriginIP: localIP,
							Floor:    button.Floor,
							//ConfirmedBy: ,
							Timer: time.Now(),
						},
					}
				*/
			case ButtonStop:
				//motorChannel <- MotorStop
				//lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
				log.Println("Stop button pressed. Elevator will come to a halt.")
				//time.Sleep(1 * time.Second)
				//lightChannel <- ElevatorLight{Kind: ButtonStop, Active: false}
				//os.Exit(1)
			}
		case floor := <-floorChannel: // Hardware
			//floorHandler(floor)
			floorReached <- floor
			printElevatorControl("Elevator " + localIP + " reaced floor " + strconv.Itoa(floor+1))
			// Add cases for tickers
		}
	}
}

func setPanelLights(lightChannel chan ElevatorLight, localIP string) {
	for {
		for f := 0; f < NumFloors; f++ {
			if ElevatorStatus[localIP].CabOrders[f] == true {
				lightChannel <- ElevatorLight{Floor: f, Kind: ButtonCommand, Active: true}
				printElevatorControl("Set panelLight for cabOrder " + strconv.Itoa(f+1))
			} else {
				lightChannel <- ElevatorLight{Floor: f, Kind: ButtonCommand, Active: false}
			}
			for k := ButtonCallUp; k <= ButtonCallDown; k++ {
				if HallOrderMatrix[f][k].Status == Awaiting || HallOrderMatrix[f][k].Status == UnderExecution {
					lightChannel <- ElevatorLight{Floor: f, Kind: k, Active: true}
				} else {
					lightChannel <- ElevatorLight{Floor: f, Kind: k, Active: false}
				}
			}
		}
	}
}

func printElevatorControl(s string) {
	if debugElevatorControl {
		log.Println("[elevatorControl] \t", s)
	}
}
