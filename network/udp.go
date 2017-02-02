package network

import (
	"fmt"
	"net"
	"os"
	"time"
)

func Udp() {
	fmt.Println("Hello from Udp.go")
}

func InitUDP() {

	// Resolve broadcast addr
	var raddr = "255.255.255.255"
	broadcastAddr, err := net.ResolveUDPAddr("udp4", raddr)
	checkError(err)

	// Resolve local addr
	laddr, err := net.ResolveUDPAddr("udp4", "129.241.187.142:20014")

	// Dial
	conn, err := net.DialUDP("udp4", laddr, broadcastAddr)
	checkError(err)

	go udpTransmit(conn)
	go udpRecieve(conn)

}

func udpTransmit(conn *net.UDPConn) {

	defer conn.Close()

	for {

		daytime := time.Now().String()

		_, err := conn.Write([]byte(daytime))
		checkError(err)

		time.Sleep(time.Second * 1)
	}

}

func udpRecieve(conn *net.UDPConn) {

	for {

		var recieveBuffer [1024]byte
		_, addr, err := conn.ReadFromUDP(recieveBuffer[0:])
		if err != nil {
			return
		}

	}

}

func checkError(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}
