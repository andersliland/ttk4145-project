package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"time"

	"./control"
	"./driver"
	"./network"
	"./orders"
	. "./utilities"
)

const debugSystemControl = false
const debugElevatorControl = false

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	broadcastOrderChannel := make(chan OrderMessage, 5)
	receiveOrderChannel := make(chan OrderMessage, 5)
	broadcastBackupChannel := make(chan BackupMessage, 5)
	receiveBackupChannel := make(chan BackupMessage, 5)
	orderCompleteChannel := make(chan OrderMessage, 5) // send OrderComplete from RemoveOrders to SystemControl

	buttonChannel := make(chan ElevatorButton, 10)
	lightChannel := make(chan ElevatorLight)
	motorChannel := make(chan int)
	floorChannel := make(chan int)

	newOrder := make(chan bool, 5)
	timeoutChannel := make(chan ExtendedHallOrder)

	safeKillChannel := make(chan os.Signal, 5)
	floorReached := make(chan int, 5)

	var onlineElevators = make(map[string]bool)

	// random timeout to prevent all elevator from timing out at the same time, first to timeout is new OriginIP
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var orderTimeout = OrderTimeout*time.Second + time.Duration(r.Intn(2000))*time.Millisecond

	var localIP string
	var err error
	localIP, err = network.Init(broadcastOrderChannel, receiveOrderChannel, broadcastBackupChannel, receiveBackupChannel)
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[main]\t\t Could not initiate network", err, ColorNeutral)
	}

	driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, ElevatorPollDelay)

	// init by going to  nearest floor below and broadcast
	floor := driver.GoToFloorBelow(localIP, motorChannel, PollDelay)
	floorReached <- floor
	fmt.Print(ColorWhite)
	log.Println("[main]\t\t New elevator "+localIP+" starting at floor "+strconv.Itoa(floor+1), ColorNeutral)

	ElevatorStatusMutex.Lock()
	ElevatorStatus[localIP] = ResolveElevator(Elevator{LocalIP: localIP})
	ElevatorStatusMutex.Unlock()

	onlineElevators = updateOnlineElevators(ElevatorStatus, onlineElevators, localIP, WatchdogLimit)

	go control.EventManager(newOrder, floor, broadcastOrderChannel, broadcastBackupChannel, orderCompleteChannel, floorReached, lightChannel, motorChannel, localIP)
	signal.Notify(safeKillChannel, os.Interrupt)
	go safeKill(safeKillChannel, motorChannel)
	go setPanelLights(lightChannel, localIP)

	if err := LoadBackup("backupElevator", &ElevatorStatus[localIP].CabOrders); err == nil {
		log.Println("[main]\t\t Loading and executing CabOrder restored from backup")
		for f := 0; f < NumFloors; f++ {
			if ElevatorStatus[localIP].CabOrders[f] {
				newOrder <- true
				break
			}
		}
	}

	watchdogTimer := time.NewTicker(WatchdogLimit)
	defer watchdogTimer.Stop()
	WatchdogKickTimer := time.NewTicker(WatchdogKickTime)
	defer WatchdogKickTimer.Stop()

	for {
		select {
		case <-WatchdogKickTimer.C:
			broadcastBackupChannel <- ResolveWatchdogKickMessage(ElevatorStatus[localIP])

		case <-watchdogTimer.C:
			onlineElevators = updateOnlineElevators(ElevatorStatus, onlineElevators, localIP, WatchdogLimit)

		case button := <-buttonChannel:
			printElevatorControl("New button push from " + localIP + " of type '" + ButtonType[button.Kind] + "' at floor " + strconv.Itoa(button.Floor+1))
			switch button.Kind {
			case ButtonCallUp, ButtonCallDown:
				if _, ok := onlineElevators[localIP]; !ok {
					log.Println("[main]\t\t Elevator offline, cannot accept new order")
				} else {
					orderAssignedTo, _ := orders.AssignOrderToElevator(button.Floor, button.Kind, onlineElevators, ElevatorStatus, HallOrderMatrix)
					broadcastOrderChannel <- OrderMessage{
						Floor:      button.Floor,
						ButtonType: button.Kind,
						AssignedTo: orderAssignedTo,
						OriginIP:   localIP,
						SenderIP:   localIP,
						Event:      EventNewOrder,
					}
				}
			case ButtonCommand:
				ElevatorStatusMutex.Lock()
				ElevatorStatus[localIP].CabOrders[button.Floor] = true
				ElevatorStatusMutex.Unlock()

				if err := SaveBackup("backupElevator", ElevatorStatus[localIP].CabOrders); err != nil {
					log.Println("[main]\t\t Save Backup failed: ", err)
				}
				resetTimerForAllAssignedOrders(orderTimeout, localIP) // reset timer on Cabbutton spamming
				fmt.Printf(ColorGreen)
				log.Println("[main]\t\t CabOrder "+ButtonType[button.Kind]+"\ton floor "+strconv.Itoa(button.Floor+1), ColorNeutral)
				newOrder <- true

			case ButtonStop:
				motorChannel <- Stop
				lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
				log.Println("[main]\t\t Stop button pressed. Elevator will come to a halt.")
				log.Println("[main]\t\t You need to restart system to init elevator again.")
				time.Sleep(200 * time.Millisecond)
				lightChannel <- ElevatorLight{Kind: ButtonStop, Active: false}
				os.Exit(1)

			}
		case floor := <-floorChannel:
			floorReached <- floor
			fmt.Print(ColorYellow)
			log.Println("[main]\t\t Elevator "+localIP+" reached floor "+strconv.Itoa(floor+1), ColorNeutral)
			resetTimerForAllAssignedOrders(orderTimeout, localIP)

		case backup := <-receiveBackupChannel:
			//log.Printf("[main]\t receivedBackupChannel with event %v from %v]", EventType[backup.Event], backup.AskerIP)
			switch backup.Event {
			case EventElevatorOnline:
				if _, ok := ElevatorStatus[backup.ResponderIP]; ok {
					ElevatorStatusMutex.Lock()
					ElevatorStatus[backup.ResponderIP].Time = time.Now()
					ElevatorStatusMutex.Unlock()

				} else {
					log.Println("[main]\t\t Received EventElevatorOnline from a new elevator with IP " + backup.ResponderIP)
					ElevatorStatusMutex.Lock()
					ElevatorStatus[backup.ResponderIP] = ResolveElevator(backup.State)
					ElevatorStatusMutex.Unlock()

				}
				onlineElevators = updateOnlineElevators(ElevatorStatus, onlineElevators, localIP, WatchdogLimit)

			default:
				log.Println("[main]\t\tReceived invalid BackupMessage from", backup.ResponderIP)
			}

		case order := <-receiveOrderChannel:
			//printSystemControl("Received an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)
			switch order.Event {
			case EventNewOrder:
				HallOrderMatrixMutex.Lock()
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // create new instance of ConfirmedBy map
				HallOrderMatrixMutex.Unlock()

				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: order.AssignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventAckNewOrder,
				}

				if order.OriginIP == localIP {
					//printSystemControl("Starting ack timer on new order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(AckTimeLimit, func() {
						log.Println("[main]\t\t ACK-TIMEOUT\t newOrder at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " not ACK'ed by all ")
						timeoutChannel <- ExtendedHallOrder{
							Floor:        order.Floor,
							ButtonType:   order.ButtonType,
							TimeoutState: TimeoutAckNewOrder,
							Order: HallOrder{
								AssignedTo: order.AssignedTo,
							},
						}
					})
				}
			case EventAckNewOrder:
				// OriginIP is responsible for registering ack from other elevators
				if order.OriginIP == localIP {
					HallOrderMatrixMutex.Lock()
					HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
					HallOrderMatrixMutex.Unlock()

					if allElevatorsHaveAcked(onlineElevators, HallOrderMatrix, order) {
						//printSystemControl("All elevators have ack'ed NewOrder at Floor " + strconv.Itoa(order.Floor+1) + " of  type " + ButtonType[order.ButtonType])
						HallOrderMatrixMutex.Lock()
						HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()
						HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy()
						HallOrderMatrixMutex.Unlock()

						broadcastOrderChannel <- OrderMessage{
							Floor:      order.Floor,
							ButtonType: order.ButtonType,
							AssignedTo: order.AssignedTo,
							OriginIP:   order.OriginIP,
							SenderIP:   localIP,
							Event:      EventOrderConfirmed,
						}
					}
				}

			case EventOrderConfirmed:
				fmt.Print(ColorGreen)
				log.Println("[main]\t\t Order " + ButtonType[order.ButtonType] + "\ton floor " + strconv.Itoa(order.Floor+1) + " is assigned to " + order.AssignedTo + ColorNeutral)
				HallOrderMatrixMutex.Lock()
				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = order.AssignedTo
				HallOrderMatrix[order.Floor][order.ButtonType].Status = Awaiting
				HallOrderMatrixMutex.Unlock()

				if order.AssignedTo == localIP {
					newOrder <- true // start executing order
				}

				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: order.AssignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventAckOrderConfirmed,
				}

				// All other elevators than OriginIP start timer on order execution.
				// The elevator which the order is AssignedTo must timeout before the others, to be removed by watchdog.
				if order.OriginIP != localIP {
					//printSystemControl("Start execution timeout on order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					timeout := orderTimeout
					if order.AssignedTo != localIP {
						timeout = 2 * orderTimeout
					}
					//printSystemControl("Starting execution timer [EventOrderConfirmed] on order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(timeout, func() {
						fmt.Print(ColorDarkGrey)
						log.Println("[main]\t\t Order "+ButtonType[order.ButtonType]+"\t on floor "+strconv.Itoa(order.Floor+1)+" could not be completed by "+order.AssignedTo+". OriginIP "+order.OriginIP+" (OriginIP != localIP) ", ColorNeutral)
						timeoutChannel <- ExtendedHallOrder{
							Floor:        order.Floor,
							ButtonType:   order.ButtonType,
							OriginIP:     order.OriginIP,
							TimeoutState: TimeoutOrderExecution,
							Order: HallOrder{
								AssignedTo: order.AssignedTo,
							},
						}
					})
				}

			case EventAckOrderConfirmed:
				//printSystemControl("case: EventAckOrderConfirmed")
				if order.OriginIP == localIP {
					HallOrderMatrixMutex.Lock()
					HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
					HallOrderMatrixMutex.Unlock()

					if allElevatorsHaveAcked(onlineElevators, HallOrderMatrix, order) {
						//printSystemControl("All elevators have ack'ed OrderConfirmed at Floor " + strconv.Itoa(order.Floor+1) + " of  type " + ButtonType[order.ButtonType])
						HallOrderMatrixMutex.Lock()
						HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()
						HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy()
						HallOrderMatrixMutex.Unlock()

						// Timehout handler on order when there is only one elevator on the network
						timeout := orderTimeout
						if order.AssignedTo != localIP {
							timeout = 2 * orderTimeout
						}
						//log.Println("[systemConrtol]\t OriginIP start execution timer [EventOrderConfirmed] on order "+ButtonType[order.ButtonType]+" on floor "+strconv.Itoa(order.Floor+1)+" Timer: ", HallOrderMatrix[order.Floor][order.ButtonType].Timer)
						HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(timeout, func() {
							fmt.Print(ColorDarkGrey)
							log.Println("[main]\t\t Order "+ButtonType[order.ButtonType]+"\t on floor "+strconv.Itoa(order.Floor+1)+" could not be completed by "+order.AssignedTo+". OriginIP "+order.OriginIP+" (OriginIP == localIP) ", ColorNeutral)
							timeoutChannel <- ExtendedHallOrder{
								Floor:        order.Floor,
								ButtonType:   order.ButtonType,
								OriginIP:     order.OriginIP,
								TimeoutState: TimeoutOrderExecution,
								Order: HallOrder{
									AssignedTo: order.AssignedTo,
								},
							}
						})

					}
				}

			case EventOrderCompleted:
				printSystemControl("case: EventOrderCompleted at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " for " + order.AssignedTo)

				// TODO: move to allElevatorsHaveAcked. Orders should not be removed untill all elevators have ack
				HallOrderMatrixMutex.Lock()
				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = ""
				HallOrderMatrix[order.Floor][order.ButtonType].Status = NotActive
				HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()
				//log.Println("[main]\t\t EventOrderComplete stop timer at order  "+ButtonType[order.ButtonType]+" on floor "+strconv.Itoa(order.Floor+1)+" Timer: ", HallOrderMatrix[order.Floor][order.ButtonType].Timer)
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy()
				HallOrderMatrixMutex.Unlock()

				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: order.AssignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventAckOrderCompleted,
				}

				if order.AssignedTo == localIP {
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(AckTimeLimit, func() {
						log.Println("[main]\t\t ACK-TIMEOUT\t orderCompleted at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " not ACK'ed by all ")
						broadcastOrderChannel <- OrderMessage{
							Floor:      order.Floor,
							ButtonType: order.ButtonType,
							AssignedTo: order.AssignedTo,
							OriginIP:   order.OriginIP,
							SenderIP:   localIP,
							Event:      EventAckOrderCompleted,
						}
					})
				}

			case EventAckOrderCompleted:
				//printSystemControl("case: EventAckOrderCompleted")

				HallOrderMatrixMutex.Lock()
				HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
				HallOrderMatrixMutex.Unlock()

				if allElevatorsHaveAcked(onlineElevators, HallOrderMatrix, order) {
					fmt.Printf(ColorBlue)
					log.Println("[main]\t\t Order "+ButtonType[order.ButtonType]+"\ton floor "+strconv.Itoa(order.Floor+1)+" is completed and ack'ed by all", ColorNeutral)
					HallOrderMatrixMutex.Lock()
					HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()
					HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy()
					HallOrderMatrixMutex.Unlock()

				}

			case EventReassignOrder:
				//printSystemControl("case: EventReassignOrder")
				HallOrderMatrixMutex.Lock()
				HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy()
				HallOrderMatrix[order.Floor][order.ButtonType].Status = NotActive
				HallOrderMatrixMutex.Unlock()

				assignedTo, _ := orders.AssignOrderToElevator(order.Floor, order.ButtonType, onlineElevators, ElevatorStatus, HallOrderMatrix)
				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: assignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventNewOrder,
				}
				fmt.Printf(ColorWhite)
				log.Println("[main]\t\t Order "+ButtonType[order.ButtonType]+"\ton floor "+strconv.Itoa(order.Floor+1)+" is reassigned from "+order.AssignedTo+" to "+assignedTo+" with OriginIP "+order.OriginIP, ColorNeutral)

			default:
				log.Println("[main]\t\t Received an invalid OrderMessage from " + order.SenderIP)

			}

		case order := <-orderCompleteChannel: // set HallOrders to NotActive when there is no network connection
			//printSystemControl("case: orderCompleteChannel at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " for " + order.AssignedTo)
			if !elevatorIsOnline(order.AssignedTo, onlineElevators) {
				HallOrderMatrixMutex.Lock()
				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = ""
				HallOrderMatrix[order.Floor][order.ButtonType].Status = NotActive
				HallOrderMatrix[order.Floor][order.ButtonType].StopTimer() // stops timer set in EventAckOrderConfirmed
				HallOrderMatrixMutex.Unlock()

				log.Println("[main]\t\t Order " + ButtonType[order.ButtonType] + "\t at floor " + strconv.Itoa(order.Floor+1) + " set to NotActive")
				fmt.Printf(ColorBlue)
				log.Println("[main]\t\t Order "+ButtonType[order.ButtonType]+"\ton floor "+strconv.Itoa(order.Floor+1)+" is completed while elevator is offline", ColorNeutral)
			}
		case t := <-timeoutChannel:
			switch t.TimeoutState {
			case TimeoutAckNewOrder:
				log.Println("[main]\t\t Not all elevators ACKed newOrder at floor " + strconv.Itoa(t.Floor+1) + " for " + ButtonType[t.ButtonType] + ". Resending...")
				broadcastOrderChannel <- OrderMessage{
					Floor:      t.Floor,
					ButtonType: t.ButtonType,
					AssignedTo: t.Order.AssignedTo,
					SenderIP:   localIP,
					OriginIP:   localIP,
					Event:      EventNewOrder,
				}

			case TimeoutAckOrderConfirmed:
				log.Println("[main]\t\t Not all elevators ACKed orderConfirmed at floor " + strconv.Itoa(t.Floor+1) + " for " + ButtonType[t.ButtonType] + ". Resending...")
				broadcastOrderChannel <- OrderMessage{
					Floor:      t.Floor,
					ButtonType: t.ButtonType,
					AssignedTo: t.Order.AssignedTo,
					SenderIP:   localIP,
					OriginIP:   t.OriginIP,
					Event:      EventOrderConfirmed,
				}

			case TimeoutOrderExecution:

				if t.Order.AssignedTo == localIP {
					motorChannel <- Stop
					time.Sleep(200 * time.Millisecond)
					fmt.Print(ColorRed)
					log.Println("[main]\t\t SUICIDE, could not complete order "+ButtonType[t.ButtonType]+" at floor "+strconv.Itoa(t.Floor+1)+". OriginIP: "+t.OriginIP+" AssignedTo: "+t.Order.AssignedTo, ColorNeutral)
					//os.Exit(1)
					restartElevator() //TODO: implement gracefull restart

				}
				broadcastOrderChannel <- OrderMessage{
					Floor:      t.Floor,
					ButtonType: t.ButtonType,
					AssignedTo: t.Order.AssignedTo,
					OriginIP:   localIP, //new OriginIP is first to timeout
					SenderIP:   localIP,
					Event:      EventReassignOrder,
				}
			}
		} // select
	} // for
} //function

func updateOnlineElevators(ElevatorStatus map[string]*Elevator, onlineElevators map[string]bool, localIP string, WatchdogLimit time.Duration) map[string]bool {
	for k := range ElevatorStatus {
		if time.Since(ElevatorStatus[k].Time) > WatchdogLimit { // remove elevator from 'onlineElevators' if watchdog timeout
			if onlineElevators[k] == true {
				delete(onlineElevators, k)
				//printSystemControl("Removed elevator " + ElevatorStatus[k].LocalIP + " in onlineElevators")
				log.Println("[main]\t \t All onlineElevators", onlineElevators, "Removed ", ElevatorStatus[k].LocalIP)
			}
		} else { // add elevator to 'onlineElevators' if watchdog not timeout
			if onlineElevators[k] != true {
				onlineElevators[k] = true
				//printSystemControl("Added elevator " + ElevatorStatus[k].LocalIP + " in onlineElevators")
				log.Println("[main]\t \t All onlineElevators", onlineElevators, "Added ", ElevatorStatus[k].LocalIP)
			}
		}
	}
	return onlineElevators
}

func setPanelLights(lightChannel chan ElevatorLight, localIP string) {
	var cabPanelLights [NumFloors]bool
	var hallPanelLights [NumFloors][2]bool
	for {
		for f := 0; f < NumFloors; f++ {
			if ElevatorStatus[localIP].CabOrders[f] == true && cabPanelLights[f] != true {
				lightChannel <- ElevatorLight{Floor: f, Kind: ButtonCommand, Active: true}
				cabPanelLights[f] = true
			} else if ElevatorStatus[localIP].CabOrders[f] == false && cabPanelLights[f] == true {
				lightChannel <- ElevatorLight{Floor: f, Kind: ButtonCommand, Active: false}
				cabPanelLights[f] = false
			}
			for k := ButtonCallUp; k <= ButtonCallDown; k++ {
				if (HallOrderMatrix[f][k].Status == Awaiting || HallOrderMatrix[f][k].Status == UnderExecution) && hallPanelLights[f][k] != true {
					lightChannel <- ElevatorLight{Floor: f, Kind: k, Active: true}
					hallPanelLights[f][k] = true
				} else if (HallOrderMatrix[f][k].Status == NotActive) && hallPanelLights[f][k] == true {
					lightChannel <- ElevatorLight{Floor: f, Kind: k, Active: false}
					hallPanelLights[f][k] = false
				}
			}
		}
		time.Sleep(PanelLightPollDelay)
	}
}

func allElevatorsHaveAcked(onlineElevators map[string]bool, HallOrderMatrix [NumFloors][2]HallOrder, order OrderMessage) bool {
	for ip, _ := range onlineElevators {
		if _, confirmedBy := HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[ip]; !confirmedBy {
			return false
		}
	}
	return true
}

func elevatorIsOnline(ip string, onlineElevators map[string]bool) bool {
	if onlineElevators[ip] == true {
		return true
	}
	return false
}

func restartElevator() {
	fmt.Print(ColorRed)
	log.Println("\t\t -----RESTARTING ELEVATOR-----", ColorNeutral)
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "pwd")
	err := cmd.Run()
	if err != nil {
		log.Println("[main] Failed to restart elevator")
		log.Fatal(err)
	}
}

// reset timer for all orders AssignetTo localIP
func resetTimerForAllAssignedOrders(orderTimeout time.Duration, ip string) {
	for f := 0; f < NumFloors; f++ {
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[f][k].AssignedTo == ip {
				HallOrderMatrixMutex.Lock()
				HallOrderMatrix[f][k].Timer.Reset(orderTimeout)
				HallOrderMatrixMutex.Unlock()
				//log.Println("[main]\t\t Reset timer on order " + ButtonType[k] + " at floor " + strconv.Itoa(f+1))
			}
		}
	}
}

func safeKill(safeKillChannel <-chan os.Signal, motorChannel chan int) {
	<-safeKillChannel
	motorChannel <- Stop
	time.Sleep(100 * time.Millisecond) // wait for motor stop too be processed
	fmt.Print(ColorWhite)
	log.Println("[main]\t User terminated program - MOTOR STOPPED", ColorNeutral)
	//os.Exit(1)
	restartElevator()
}

func printSystemControl(s string) {
	if debugSystemControl {
		log.Println("[main]\t\t", s)
	}
}

func printElevatorControl(s string) {
	if debugElevatorControl {
		log.Println("[main]\t\t", s)
	}
}
