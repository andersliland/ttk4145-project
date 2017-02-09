package network

import (
	"log"
	"net"
	"strconv"

	. "../config"
)

// Maximum allowed UDP datagram size in bytes: 65,507 (imposed by the IPv4 protocol)
const messageSize = 1024
const broadcastListenPort = 6666

type UDPMessage struct {
	Raddr  string
	Data   []byte
	Length int
}

var broadcastAddr net.UDPAddr
var laddr net.UDPAddr
var listenAddr net.UDPAddr

func InitUDP(
	udpSendChannel chan UDPMessage,
	udpReceiveChannel chan UDPMessage) {

	broadcastAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255"+":"+strconv.Itoa(broadcastListenPort)) // increment port by 1 for each new connection
	CheckError("Failed to resolve remote addr", err)

	//laddr, err := net.ResolveUDPAddr("udp4", "129.241.187.50:20014")
	CheckError("Failed to resolve local addr", err)

	listenAddr, err := net.ResolveUDPAddr("udp4", ":6666")
	CheckError("Failed to resolve listen port", err)

	conn, err := net.DialUDP("udp4", nil, broadcastAddr) //TODO: add laddr to DialUp
	CheckError("DialUDP failed", err)
	//defer conn.Close() // Close connection when function collapses, shoud be moved to other function

	listen, err := net.ListenUDP("udp4", listenAddr)
	CheckError("ListenUDP failed", err)

	udpReceiveCh := make(chan UDPMessage)

	go udpTransmit(conn, udpSendChannel)
	go udpReceive(udpReceiveCh, udpReceiveChannel)
	go listenUDPStream(listen, udpReceiveCh)
}

func udpTransmit(conn *net.UDPConn,
	udpSendChannel chan UDPMessage) {

	for {
		select {
		case message := <-udpSendChannel:
			_, err := conn.Write(message.Data)
			if err != nil {
				log.Println("udpTransmit: write UDP datagram error: ", err)
			}
		}
	}

}

func udpReceive(udpReceiveCh chan UDPMessage,
	udpReceiveChannel chan UDPMessage) {

	for {
		select {
		case message := <-udpReceiveCh:
			udpReceiveChannel <- message
		}
	}

}

func listenUDPStream(listen *net.UDPConn,
	udpReceiveCh chan UDPMessage) {

	receiveBuffer := make([]byte, messageSize)
	for {

		n, raddr, err := listen.ReadFromUDP(receiveBuffer)
		if err != nil {
			return
		}
		udpReceiveCh <- UDPMessage{Raddr: raddr.String(), Data: receiveBuffer[:n], Length: n}

	}

}
