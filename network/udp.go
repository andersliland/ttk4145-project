package network

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

func Udp() {
	fmt.Println("Hello from Udp.go")
}

type ClientJob struct {
	name string
	conn net.Conn
}

func InitUDP() {

	broadcastListenPort := 6666

	// Resolve broadcast addr
	var raddr = "255.255.255.255"
	broadcastAddr, err := net.ResolveUDPAddr("udp4", raddr+":"+strconv.Itoa(broadcastListenPort))
	checkError(err)

	// Resolve local addr
	//laddr, err := net.ResolveUDPAddr("udp4", "129.241.187.142:20014")

	// Dial
	conn, err := net.DialUDP("udp4", nil, broadcastAddr)
	checkError(err)
	defer conn.Close()

	listenAddr, err := net.ResolveUDPAddr("udp4", ":6666")
	checkError(err)

	listen, err := net.ListenUDP("udp4", listenAddr)
	checkError(err)

	clientJobs := make(chan ClientJob)

	var wg sync.WaitGroup
	wg.Add(3)

	go udpTransmit(conn, wg)
	go udpRecieve(listen, wg, clientJobs)
	go generateResponse(clientJobs, wg)

	wg.Wait()

}

func generateResponse(clientJobs chan ClientJob, wg sync.WaitGroup) {
	defer wg.Done()

	for {

		clientJob := <-clientJobs

		clientJob.conn.Write([]byte("Hello, " + clientJob.name))
	}
}

func udpTransmit(conn *net.UDPConn, wg sync.WaitGroup) {
	defer wg.Done()

	for {

		daytime := time.Now().String()
		_, err := conn.Write([]byte(daytime))
		checkError(err)
		time.Sleep(time.Second * 1)
		fmt.Println("Transmitted message to", conn.RemoteAddr())
	}

}

func udpRecieve(listen *net.UDPConn, wg sync.WaitGroup, clientJobs chan ClientJob) {
	defer wg.Done()

	for {

		var recieveBuffer [1024]byte
		_, addr, err := listen.ReadFromUDP(recieveBuffer[0:])
		if err != nil {
			return
		}
		fmt.Printf("Recieved from:", addr, "\n")

		clientJobs <- ClientJob{"Anders", listen}

	}

}

func checkError(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}
