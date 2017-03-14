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

const debugSystemControl = true

const watchdogKickTime = 100 * time.Millisecond
const watchdogLimit = 3*watchdogKickTime + 10*time.Millisecond
const ackTimeLimit = 500 * time.Millisecond

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

	safeKillChannel := make(chan os.Signal, 10)
	floorReached := make(chan int, 5)

	var onlineElevators = make(map[string]bool)

	var localIP string
	var err error
	localIP, err = network.Init(broadcastOrderChannel, receiveOrderChannel, broadcastBackupChannel, receiveBackupChannel)
	CheckError("ERROR [main]: Could not initiate network", err)

	driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, ElevatorPollDelay)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var orderTimeout = OrderTimeout*time.Second + time.Duration(r.Intn(2000))*time.Millisecond // random timeout to prevent all elevator from timing out at the same time

	// Timers
	watchdogTimer := time.NewTicker(watchdogLimit)
	defer watchdogTimer.Stop()
	watchdogKickTimer := time.NewTicker(watchdogKickTime)
	defer watchdogKickTimer.Stop()

	ElevatorStatus[localIP] = ResolveElevator(Elevator{LocalIP: localIP})
	onlineElevators = updateOnlineElevators(ElevatorStatus, onlineElevators, localIP, watchdogLimit)

	go control.MessageLoop(
		newOrder,
		floorReached,
		buttonChannel,
		lightChannel,
		motorChannel,
		floorChannel,
		broadcastOrderChannel,
		receiveOrderChannel,
		broadcastBackupChannel,
		receiveBackupChannel,
		orderCompleteChannel,
		onlineElevators,
		ElevatorStatus,
		HallOrderMatrix,
		localIP)

	floor := driver.GoToFloorBelow(localIP, motorChannel, PollDelay)
	floorReached <- floor
	signal.Notify(safeKillChannel, os.Interrupt)
	go safeKill(safeKillChannel, motorChannel)
	go setPanelLights(lightChannel, localIP)

	broadcastBackupChannel <- BackupMessage{
		AskerIP: localIP,
		Event:   EventRequestBackup,
	}

	for {
		select {
		// Watchdog
		case <-watchdogKickTimer.C:
			broadcastBackupChannel <- ResolveWatchdogKickMessage(ElevatorStatus[localIP])
			//log.Printf("[systemControl] Watchdog send IAmAlive from %v \n", localIP)

		case <-watchdogTimer.C:
			onlineElevators = updateOnlineElevators(ElevatorStatus, onlineElevators, localIP, watchdogLimit)
			//log.Println("[systemControl] Active Elevators", onlineElevators)

		case backup := <-receiveBackupChannel:
			//log.Printf("[systemControl] receivedBackupChannel with event %v from %v]", EventType[backup.Event], backup.AskerIP)
			switch backup.Event {
			case EventElevatorOnline:
				if _, ok := ElevatorStatus[backup.ResponderIP]; ok { // check if a value exsist for ResponderIP
					ElevatorStatus[backup.ResponderIP].Time = time.Now() //update time for known elevator
				} else {
					log.Println("[systemControl]\t Received EventElevatorOnline from a new elevator with IP " + backup.ResponderIP)
					ElevatorStatus[backup.ResponderIP] = ResolveElevator(backup.State)
				}
				onlineElevators = updateOnlineElevators(ElevatorStatus, onlineElevators, localIP, watchdogLimit)
			default:
				log.Println("[systemControl]\tReceived invalid BackupMessage from", backup.ResponderIP)
			}

		case order := <-receiveOrderChannel:
			//printSystemControl("Received an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)
			switch order.Event {
			case EventNewOrder:
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // create new instance of ConfirmedBy map
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
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(ackTimeLimit, func() {
						log.Println("[systemControl]\t ACK-TIMEOUT\t newOrder at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " not ACK'ed by all ")
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
				//printSystemControl("case: EventAckNewOrder")
				// OriginIP is responsible for registering ack from other elevators
				if order.OriginIP == localIP {
					HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
					if allElevatorsHaveAcked(onlineElevators, HallOrderMatrix, order) {
						//printSystemControl("All elevators have ack'ed NewOrder at Floor " + strconv.Itoa(order.Floor+1) + " of  type " + ButtonType[order.ButtonType])
						HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()
						HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy()

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
				log.Println("[systemControl]\t Order " + ButtonType[order.ButtonType] + "\ton floor " + strconv.Itoa(order.Floor+1) + " is assigned to " + order.AssignedTo + ColorNeutral)
				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = order.AssignedTo
				HallOrderMatrix[order.Floor][order.ButtonType].Status = Awaiting
				if order.AssignedTo == localIP {
					newOrder <- true
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
				// The elevator which the order is assigned to must timeout before the others.
				if order.OriginIP != localIP {
					//printSystemControl("Start execution timeout on order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					timeout := orderTimeout
					if order.AssignedTo != localIP {
						timeout = 2 * orderTimeout
					}
					//printSystemControl("Starting execution timer [EventOrderConfirmed] on order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(timeout, func() {
						fmt.Print(ColorDarkGrey)
						log.Println("[systemControl]\t Order "+ButtonType[order.ButtonType]+"\t on floor "+strconv.Itoa(order.Floor+1)+" could not be completed by "+order.AssignedTo+". OriginIP "+order.OriginIP+" (OriginIP != localIP) ", ColorNeutral)
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
					HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
					if allElevatorsHaveAcked(onlineElevators, HallOrderMatrix, order) {
						//printSystemControl("All elevators have ack'ed OrderConfirmed at Floor " + strconv.Itoa(order.Floor+1) + " of  type " + ButtonType[order.ButtonType])
						HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()        // stop ackTimeout timer
						HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)

						// Timehout handler on order when there is only one elevator on the network
						timeout := orderTimeout
						if order.AssignedTo != localIP {
							timeout = 2 * orderTimeout
						}
						//log.Println("[systemConrtol]\t OriginIP start execution timer [EventOrderConfirmed] on order "+ButtonType[order.ButtonType]+" on floor "+strconv.Itoa(order.Floor+1)+" Timer: ", HallOrderMatrix[order.Floor][order.ButtonType].Timer)
						HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(timeout, func() {
							fmt.Print(ColorDarkGrey)
							log.Println("[systemControl]\t Order "+ButtonType[order.ButtonType]+"\t on floor "+strconv.Itoa(order.Floor+1)+" could not be completed by "+order.AssignedTo+". OriginIP "+order.OriginIP+" (OriginIP == localIP) ", ColorNeutral)
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
				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = ""
				HallOrderMatrix[order.Floor][order.ButtonType].Status = NotActive
				HallOrderMatrix[order.Floor][order.ButtonType].StopTimer() // stops timer set in EventAckOrderConfirmed
				//log.Println("[systemControl]\t EventOrderComplete stop timer at order  "+ButtonType[order.ButtonType]+" on floor "+strconv.Itoa(order.Floor+1)+" Timer: ", HallOrderMatrix[order.Floor][order.ButtonType].Timer)
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)

				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: order.AssignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventAckOrderCompleted,
				}

				if order.AssignedTo == localIP {
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(ackTimeLimit, func() {
						log.Println("[systemControl]\t ACK-TIMEOUT\t orderCompleted at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " not ACK'ed by all ")
						broadcastOrderChannel <- OrderMessage{ // Should we send to timeoutChannel - or just resend OrderMessage?
							Floor:      order.Floor,
							ButtonType: order.ButtonType,
							AssignedTo: order.AssignedTo,
							OriginIP:   order.OriginIP,
							SenderIP:   localIP,
							Event:      EventAckOrderCompleted,
						}
					})
				}

			case EventAckOrderCompleted: // delete order from matrix and timer functions
				//printSystemControl("case: EventAckOrderCompleted")
				HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
				if allElevatorsHaveAcked(onlineElevators, HallOrderMatrix, order) {
					fmt.Printf(ColorBlue)
					log.Println("[systemControl]\t Order "+ButtonType[order.ButtonType]+"\ton floor "+strconv.Itoa(order.Floor+1)+" is completed and ack'ed by all", ColorNeutral)
					HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()        // stop ackTimeout timer
					HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)
				}

			case EventReassignOrder:
				//printSystemControl("case: EventReassignOrder")
				HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()        // stop ackTimeout timer
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)
				HallOrderMatrix[order.Floor][order.ButtonType].Status = NotActive
				assignedTo, _ := orders.AssignOrderToElevator(order.Floor, order.ButtonType, onlineElevators, ElevatorStatus)

				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: assignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventNewOrder,
				}
				fmt.Printf(ColorWhite)
				log.Println("[systemControl]\t Order "+ButtonType[order.ButtonType]+"\ton floor "+strconv.Itoa(order.Floor+1)+" is reassigned from "+order.AssignedTo+" to "+assignedTo+" with OriginIP "+order.OriginIP, ColorNeutral)

			default:
				log.Println("[systemControl]\t Received an invalid OrderMessage from " + order.SenderIP)

			}

		// set HallOrders to NotActive when there is not network connection
		case order := <-orderCompleteChannel:
			//printSystemControl("case: orderCompleteChannel at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " for " + order.AssignedTo)
			if !elevatorIsOnline(order.AssignedTo, onlineElevators) {
				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = ""
				HallOrderMatrix[order.Floor][order.ButtonType].Status = NotActive
				log.Println("[systemControl]\t Order " + ButtonType[order.ButtonType] + "\t at floor " + strconv.Itoa(order.Floor+1) + " set to NotActive")
				HallOrderMatrix[order.Floor][order.ButtonType].StopTimer() // stops timer set in EventAckOrderConfirmed
				fmt.Printf(ColorBlue)
				log.Println("[systemControl]\t Order "+ButtonType[order.ButtonType]+"\ton floor "+strconv.Itoa(order.Floor+1)+" is completed while elevator is offline", ColorNeutral)
			}
		case t := <-timeoutChannel:
			switch t.TimeoutState {
			case TimeoutAckNewOrder:
				log.Println("[systemControl]\t Not all elevators ACKed newOrder at floor " + strconv.Itoa(t.Floor+1) + " for " + ButtonType[t.ButtonType] + ". Resending...")
				broadcastOrderChannel <- OrderMessage{
					Floor:      t.Floor,
					ButtonType: t.ButtonType,
					AssignedTo: t.Order.AssignedTo,
					SenderIP:   localIP,
					OriginIP:   localIP,
					Event:      EventNewOrder,
				}

			case TimeoutAckOrderConfirmed:
				log.Println("[systemControl]\t Not all elevators ACKed orderConfirmed at floor " + strconv.Itoa(t.Floor+1) + " for " + ButtonType[t.ButtonType] + ". Resending...")
				broadcastOrderChannel <- OrderMessage{
					Floor:      t.Floor,
					ButtonType: t.ButtonType,
					AssignedTo: t.Order.AssignedTo,
					SenderIP:   localIP,
					OriginIP:   t.OriginIP,
					Event:      EventOrderConfirmed,
				}

			case TimeoutOrderExecution: // EventAckOrderCompleted failed
				// kill self
				if t.Order.AssignedTo == localIP {
					motorChannel <- Stop
					time.Sleep(200 * time.Millisecond)
					fmt.Print(ColorRed)
					log.Println("[systemControl]\t SUICIDE, could not complete order "+ButtonType[t.ButtonType]+" at floor "+strconv.Itoa(t.Floor+1)+". OriginIP: "+t.OriginIP+" AssignedTo: "+t.Order.AssignedTo, ColorNeutral)
					os.Exit(1)
					//restartElevator() //TODO: implement gracefull restart

				}
				if t.OriginIP == localIP {
					//printSystemControl("Order timeout I am OriginIP ")
					broadcastOrderChannel <- OrderMessage{
						Floor:      t.Floor,
						ButtonType: t.ButtonType,
						AssignedTo: t.Order.AssignedTo,
						OriginIP:   localIP, //randomly assign OriginIP for new order
						SenderIP:   localIP,
						Event:      EventReassignOrder,
					}

				} else {
					//printSystemControl("Order timeout I am NOT OriginIP ")
					broadcastOrderChannel <- OrderMessage{
						Floor:      t.Floor,
						ButtonType: t.ButtonType,
						AssignedTo: t.Order.AssignedTo,
						OriginIP:   localIP, //randomly assign OriginIP for new order
						SenderIP:   localIP,
						Event:      EventReassignOrder,
					}
				}
			}
		} // select
	} // for
} //function

// removes elevator from 'onlineElevators' if watchdog timeout
// adds elevator to 'onlineElevators' if watchdog not timeout
func updateOnlineElevators(ElevatorStatus map[string]*Elevator, onlineElevators map[string]bool, localIP string, watchdogLimit time.Duration) map[string]bool {
	for k := range ElevatorStatus {
		if time.Since(ElevatorStatus[k].Time) > watchdogLimit { //watchdog timeout
			if onlineElevators[k] == true {
				delete(onlineElevators, k)
				//printSystemControl("Removed elevator " + ElevatorStatus[k].LocalIP + " in onlineElevators")
				log.Println("[systemControl] \t All onlineElevators", onlineElevators, "Removed ", ElevatorStatus[k].LocalIP)
			}
		} else { // watchdog not timed out
			if onlineElevators[k] != true {
				onlineElevators[k] = true
				//printSystemControl("Added elevator " + ElevatorStatus[k].LocalIP + " in onlineElevators")
				log.Println("[systemControl] \t All onlineElevators", onlineElevators, "Added ", ElevatorStatus[k].LocalIP)
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
				//printElevatorControl("Set CabOrder light on floor " + strconv.Itoa(f+1) + " on elevator " + localIP)
			} else if ElevatorStatus[localIP].CabOrders[f] == false && cabPanelLights[f] == true {
				lightChannel <- ElevatorLight{Floor: f, Kind: ButtonCommand, Active: false}
				cabPanelLights[f] = false
				//printElevatorControl("Clear CabOrder light on floor " + strconv.Itoa(f+1) + " on elevator " + localIP)
			}
			for k := ButtonCallUp; k <= ButtonCallDown; k++ {
				if (HallOrderMatrix[f][k].Status == Awaiting || HallOrderMatrix[f][k].Status == UnderExecution) && hallPanelLights[f][k] != true {
					lightChannel <- ElevatorLight{Floor: f, Kind: k, Active: true}
					hallPanelLights[f][k] = true
					//printElevatorControl("Set HallOrder light on floor " + strconv.Itoa(f+1) + " of kind " + MotorStatus[] + " on elevator " + localIP)
				} else if (HallOrderMatrix[f][k].Status == NotActive) && hallPanelLights[f][k] == true {
					lightChannel <- ElevatorLight{Floor: f, Kind: k, Active: false}
					hallPanelLights[f][k] = false
					//printElevatorControl("Clear HallOrder light on floor " + strconv.Itoa(f+1) + " of kind " + MotorStatus[k+1] + " on elevator " + localIP)
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
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
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func safeKill(safeKillChannel <-chan os.Signal, motorChannel chan int) {
	<-safeKillChannel
	motorChannel <- Stop
	time.Sleep(10 * time.Millisecond) // wait for motor stop too be processed
	log.Fatal(ColorWhite, "\nUser terminated program\nMOTOR STOPPED\n", ColorNeutral)
	os.Exit(1)
}

func printSystemControl(s string) {
	if debugSystemControl {
		log.Println("[systemControl]\t", s)
	}
}
