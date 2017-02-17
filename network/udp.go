package network

import (
	"log"
	"net"
	"strconv"
	"strings"

	. "../utilities"
)

// Maximum allowed UDP datagram size in bytes: 65,507 (imposed by the IPv4 protocol)
const messageSize = 1024
const localListenPort = 1
const broadcastListenPort = 6667

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
	udpSendDatagramChannel chan UDPMessage,
	udpReceiveDatagramChannel chan UDPMessage) (localIP string, err error) {

	broadcastAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255"+":"+strconv.Itoa(broadcastListenPort))
	CheckError("ERROR [udp]: Failed to resolve remote addr", err)

	// Local listen connection
	listenAddr, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(broadcastListenPort))
	CheckError("ERROR [udp]: Failed to resolve broadcastListenPort: ", err)

	// Get local IP address
	localIP, err = resolveLocalIP(broadcastAddr)
	CheckError("ERROR [udp]: Failed to get local addr: ", err)
	log.Println("[udp] LocalIP: ", localIP)

	// Local broadcast connection
	conn, err := net.DialUDP("udp4", nil, broadcastAddr)
	CheckError("ERROR [udp]: DialUDP failed", err)

	listen, err := net.ListenUDP("udp4", listenAddr)
	CheckError("ERROR [udp] ListenUDP failed", err)

	udpReceiveBufferChannel := make(chan UDPMessage)

	go udpTransmit(conn, udpSendDatagramChannel)
	go udpReceive(udpReceiveDatagramChannel, udpReceiveBufferChannel)
	go listenUDPStream(listen, udpReceiveBufferChannel)

	return localIP, nil
}

func resolveLocalIP(broadcastAddr *net.UDPAddr) (string, error) {
	if localIP == "" {
		conn, err := net.DialUDP("udp4", nil, broadcastAddr)
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil

}

func udpTransmit(conn *net.UDPConn, udpSendDatagramChannel chan UDPMessage) {
	defer conn.Close()
	for {
		select {
		case message := <-udpSendDatagramChannel:
			n, err := conn.Write(message.Data)
			if debug {
				log.Println("[udp] Number of bytes written:", n)
			}
			if err != nil {
				log.Println("ERROR [udp] udpTransmit: write UDP datagram error: ", err)
			}
		}
	}

}

func udpReceive(udpReceiveDatagramChannel chan UDPMessage, udpReceiveBufferChannel chan UDPMessage) {

	for {
		select {
		case u := <-udpReceiveBufferChannel:
			udpReceiveDatagramChannel <- u
		}
	}
}

func listenUDPStream(listen *net.UDPConn,
	udpReceiveBufferChannel chan UDPMessage) {
	defer listen.Close()

	buf := make([]byte, messageSize)
	for {
		n, raddr, err := listen.ReadFromUDP(buf)
		//log.Println("[udp] Number of bytes received:", n, " from:", raddr)
		if err != nil {
			log.Println("[udp] Failed to ReadFromUDP stream")
			return
		}
		udpReceiveBufferChannel <- UDPMessage{Raddr: raddr.String(), Data: buf[:n], Length: n}
	}

}
