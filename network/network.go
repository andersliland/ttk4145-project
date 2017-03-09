package network

import (
	"encoding/json"
	"log"

	. "../utilities"
)

const debugNetwork = false

func Init(sendBroadcastChannel <-chan OrderMessage,
	receiveOrderChannel chan<- OrderMessage,
	sendBackupChannel <-chan ElevatorBackupMessage,
	receiveBackupChannel chan<- ElevatorBackupMessage) (localIP string, err error) {

	UDPSendChannel := make(chan UDPMessage, 10)
	UDPReceiveChannel := make(chan UDPMessage)

	go sendMessageHandler(sendBroadcastChannel, sendBackupChannel, UDPSendChannel)
	go receiveMessageHandler(receiveOrderChannel, receiveBackupChannel, UDPReceiveChannel)

	localIP, err = InitUDP(UDPSendChannel, UDPReceiveChannel)
	CheckError("", err)

	return localIP, nil
}

// Receive message from main.go, marshal and send down to udp.go
func sendMessageHandler(sendBroadcastChannel <-chan OrderMessage,
	sendBackupChannel <-chan ElevatorBackupMessage,
	UDPSendChannel chan<- UDPMessage) {

	for {
		select {
		case message := <-sendBroadcastChannel:
			data, err := json.Marshal(message)
			if err != nil {
				log.Println("ERROR [network]: sendMessage marshal failed", err)
			} else {
				UDPSendChannel <- UDPMessage{Data: data}
				printNetwork("Sent an OrderMessage with " + EventType[message.Event])

			}
		case message := <-sendBackupChannel:
			data, err := json.Marshal(message)
			if err != nil {
				log.Println("ERROR [network]: sendBackup marshal failed", err)
			} else {
				UDPSendChannel <- UDPMessage{Data: data}
				printNetwork("Sent an ElevatorBackupMessage " + EventType[message.Event])
			}
		}
	}
}

// Receive message from udp.go, unmarshal and send up to main
func receiveMessageHandler(
	receiveOrderChannel chan<- OrderMessage,
	receiveBackupChannel chan<- ElevatorBackupMessage,
	UDPReceiveChannel <-chan UDPMessage) {

	for {
		select {
		case msg := <-UDPReceiveChannel:
			var f interface{}
			err := json.Unmarshal(msg.Data[:msg.Length], &f)
			if err != nil {
				log.Println("[network] First Unmarshal failed", err)
			} else {
				printNetwork(" New UDP datagram received, first Unmarshal sucess")

				// TODO: revrite two next lines, probably go build in reflect package
				m := f.(map[string]interface{})
				event := int(m["Event"].(float64)) // type assertion, float64 because

				if event <= 5 && event >= 0 {
					var backupMessage = ElevatorBackupMessage{}
					if err := json.Unmarshal(msg.Data[:msg.Length], &backupMessage); err == nil { //unmarshal into correct message struct
						printNetwork("ElevatorBackupMessage Unmarshal sucess")
						if backupMessage.IsValid() {
							receiveBackupChannel <- backupMessage
							printNetwork("Recived an ElevatorBackupMessage with Event " + EventType[backupMessage.Event])
						} else {
							printNetwork("Rejected an ElevatorBackupMessage with Event " + EventType[backupMessage.Event])
						}
					} else {
						log.Print("[network] ElevatorBackupMessage Unmarshal failed", err)
					}
				} else if event >= 6 && event <= 12 {
					var orderMessage = OrderMessage{}
					if err := json.Unmarshal(msg.Data[:msg.Length], &orderMessage); err == nil { //unmarshal into correct message struct
						printNetwork("[network] OrderMessage Unmarshal sucess")
						if orderMessage.IsValid() {
							receiveOrderChannel <- orderMessage
							printNetwork("Recived an OrderMessage with Event " + EventType[orderMessage.Event])
						} else {
							printNetwork("Rejected an OrderMessage with Event " + EventType[orderMessage.Event])
						}
					} else {
						log.Print("[network] OrderMessage Unmarshal failed")
					}
				} else {
					log.Println("[network] Recived an unknown message type")
				}
			}

		}
	}
}

func printNetwork(s string) {
	if debugNetwork {
		log.Println("[network]", s)
	}
}
