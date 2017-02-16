package network

import (
	"encoding/json"
	"log"

	. "../config"
)

const debug = false

func InitNetwork(sendMessageChannel chan ElevatorOrderMessage,
	receiveOrderChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	receiveBackupChannel chan ElevatorBackupMessage) (localIP string, err error) {

	udpSendDatagramChannel := make(chan UDPMessage, 10)
	udpReceiveDatagramChannel := make(chan UDPMessage, 5)

	go sendMessageHandler(sendMessageChannel, sendBackupChannel, udpSendDatagramChannel)
	go receiveMessageHandler(receiveOrderChannel, udpReceiveDatagramChannel)

	localIP, err = InitUDP(udpSendDatagramChannel, udpReceiveDatagramChannel)
	CheckError("", err)

	return localIP, nil
}

// Receive message from main.go, marshal and send down to udp.go
func sendMessageHandler(sendMessageChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	udpSendDatagramChannel chan UDPMessage) {

	for {
		select {
		case message := <-sendMessageChannel:
			networkPack, err := json.Marshal(message)
			//log.Println("[network] Message to be sendt", message)
			if err != nil {
				log.Println("ERROR [network]: sendMessage marshal failed", err)
			} else {
				udpSendDatagramChannel <- UDPMessage{Raddr: "32", Data: networkPack} // UDPMessage
			}

		case message := <-sendBackupChannel:
			networkPack, err := json.Marshal(message)
			if err != nil {
				log.Println("ERROR [network]: sendBackup marshal failed", err)
			} else {
				udpSendDatagramChannel <- UDPMessage{Raddr: "32", Data: networkPack} // UDPMessage
			}
		}
	}

}

// Receive message from udp.go, unmarshal and send up to main
func receiveMessageHandler(receiveOrderChannel chan ElevatorOrderMessage,
	udpReceiveDatagramChannel chan UDPMessage) {

	var receivedOrder ElevatorOrderMessage
	for {
		select {
		case message := <-udpReceiveDatagramChannel:
			err := json.Unmarshal(message.Data, &receivedOrder)
			if err != nil {
				log.Println("ERROR [network]: Unmarshal failed", err)
			} else {
				receiveOrderChannel <- receivedOrder

			}
		}
	}
}
