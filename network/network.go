package network

import (
	"encoding/json"
	"log"

	. "../utilities"
)

const debug = true

func Init(sendMessageChannel <-chan ElevatorOrderMessage,
	receiveOrderChannel chan<- ElevatorOrderMessage,
	sendBackupChannel <-chan ElevatorBackupMessage,
	receiveBackupChannel chan<- ElevatorBackupMessage) (localIP string, err error) {

	UDPSendChannel := make(chan UDPMessage, 10)
	UDPReceiveChannel := make(chan UDPMessage)

	go sendMessageHandler(sendMessageChannel, sendBackupChannel, UDPSendChannel)
	go receiveMessageHandler(receiveOrderChannel, receiveBackupChannel, UDPReceiveChannel)

	localIP, err = InitUDP(UDPSendChannel, UDPReceiveChannel)
	CheckError("", err)

	return localIP, nil
}

// Receive message from main.go, marshal and send down to udp.go
func sendMessageHandler(sendMessageChannel <-chan ElevatorOrderMessage,
	sendBackupChannel <-chan ElevatorBackupMessage,
	UDPSendChannel chan<- UDPMessage) {

	for {
		select {
		case message := <-sendMessageChannel:
			data, err := json.Marshal(message)
			if err != nil {
				log.Println("ERROR [network]: sendMessage marshal failed", err)
			} else {
				UDPSendChannel <- UDPMessage{Data: data}
				printDebug("Sent an ElevatorOrderMessage with " + EventType[message.Event])

			}
		case message := <-sendBackupChannel:
			data, err := json.Marshal(message)
			if err != nil {
				log.Println("ERROR [network]: sendBackup marshal failed", err)
			} else {
				UDPSendChannel <- UDPMessage{Data: data}
				printDebug("Sent an ElevatorBackupMessage " + EventType[message.Event])
			}
		}
	}
}

// Receive message from udp.go, unmarshal and send up to main
func receiveMessageHandler(
	receiveOrderChannel chan<- ElevatorOrderMessage,
	receiveBackupChannel chan<- ElevatorBackupMessage,
	UDPReceiveChannel <-chan UDPMessage) {

	for {
		select {
		case msg := <-UDPReceiveChannel:
			var f interface{}
			err := json.Unmarshal(msg.Data[:msg.Length], &f)
			if err != nil {
				log.Println("ERROR [network]: Unmarshal failed", err)
			} else {
				// TODO: revrite two next lines, probably go build in reflect package
				m := f.(map[string]interface{})
				event := int(m["Event"].(float64)) // type assertion, float64 because

				if event <= 3 && event >= 0 {
					var backupMessage = ElevatorBackupMessage{}
					if err := json.Unmarshal(msg.Data[:msg.Length], &backupMessage); err == nil { //unmarshal into correct message struct
						if backupMessage.IsValid() {
							receiveBackupChannel <- backupMessage
							printDebug("Recived an ElevatorBackupMessage with Event " + EventType[backupMessage.Event])
						} else {
							printDebug("Rejected an ElevatorBackupMessage with Event " + EventType[backupMessage.Event])
						}
					} else {
						log.Println("[network] Error unmarshaling BackupMessage")
					}
				} else if event >= 4 && event <= 10 {
					var orderMessage = ElevatorOrderMessage{}
					if err := json.Unmarshal(msg.Data[:msg.Length], &orderMessage); err == nil { //unmarshal into correct message struct
						if orderMessage.IsValid(){
							receiveOrderChannel <- orderMessage
							printDebug("Recived an ElevatorOrderMessage with Event " + EventType[orderMessage.Event])
						} else {
							printDebug("Rejected an ElevatorOrderMessage with Event " + EventType[orderMessage.Event])
						}
					}
				} else {
					printDebug("Recived an unknown message type")
				}
			}
		}
	}
}

func printDebug(s string){
	if debug {
		log.Println("[network]", s)
	}
}
