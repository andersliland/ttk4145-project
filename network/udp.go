package network

import (
	"log"
	"net"
	"strconv"
"strings"
	. "../config"
)

// Maximum allowed UDP datagram size in bytes: 65,507 (imposed by the IPv4 protocol)
const messageSize = 1024
const localListenPort = 1
const broadcastListenPort = 6666

type UDPMessage struct {
	Raddr  string
	Data   []byte
	Length int
}

var broadcastAddr *net.UDPAddr
var laddr *net.UDPAddr
var listenAddr *net.UDPAddr
var localIP string



func InitUDP(
	udpSendChannel chan UDPMessage,
	udpReceiveChannel chan UDPMessage) (localIP string, err error) {

	broadcastAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255"+":"+strconv.Itoa(broadcastListenPort)) // increment port by 1 for each new connection
	CheckError("ERROR [udp] Failed to resolve remote addr", err)

	// Local listen conneciton
	listenAddr, err := net.ResolveUDPAddr("udp4", ":6666")
	CheckError("ERROR [udp] Failed to resolve broadcastListenPort: ", err)

	// Get  local IP adress
	localIP, err = resolveLocalIP(broadcastAddr)
	CheckError("ERROR [udp] Failed to get local addr: ", err)
	log.Println("LocalIP: ", localIP)

	conn, err := net.DialUDP("udp4", nil, broadcastAddr)
	CheckError("ERROR [udp] DialUDP failed", err)
	//defer conn.Close() // Close connection when function collapses, shoud be moved to other function

	listen, err := net.ListenUDP("udp4", listenAddr)
	CheckError("ERROR [udp] ListenUDP failed", err)

	udpReceiveCh := make(chan UDPMessage)

	go udpTransmit(conn, udpSendChannel)
	go udpReceive(udpReceiveCh, udpReceiveChannel)
	go listenUDPStream(listen, udpReceiveCh)

	return localIP, nil
}

func resolveLocalIP(broadcastAddr *net.UDPAddr)( string, error) {
	if localIP == "" {
		conn, err := net.DialUDP("udp4", nil, broadcastAddr)
		if err != nil {
			return "",err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil

}

func udpTransmit(conn *net.UDPConn,
	udpSendChannel chan UDPMessage) {

	for {
		select {
		case message := <-udpSendChannel:
			_, err := conn.Write(message.Data)
			if err != nil {
				log.Println("ERROR [udp] udpTransmit: write UDP datagram error: ", err)
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

	buf := make([]byte, messageSize)
	for {

		n, raddr, err := listen.ReadFromUDP(buf)
		if err != nil {
			return
		}
		udpReceiveCh <- UDPMessage{Raddr: raddr.String(), Data: buf[:n], Length: n}

	}

}
