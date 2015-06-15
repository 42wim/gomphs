package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/tatsushid/go-fastping"
)

var pingIP string
var showRTT bool
var expandDNS bool
var width string = "2"
var ipList []string

func init() {
	flag.StringVar(&pingIP, "hosts", "", "ip addresses/hosts to ping, space seperated (e.g \"8.8.8.8 8.8.4.4 google.com 2a00:1450:400c:c07::66\")")
	flag.BoolVar(&showRTT, "showrtt", false, "show roundtrip time in ms")
	flag.BoolVar(&expandDNS, "expand", false, "use all available ip's (ipv4/ipv6) of a hostname (multiple A, AAAA)")
	flag.Parse()
	if flag.NFlag() == 0 {
		fmt.Println("usage: ")
		flag.PrintDefaults()
		os.Exit(2)
	}
	if showRTT {
		width = "4"
	}
}

type MilliDuration time.Duration

func (hd MilliDuration) String() string {
	milliseconds := time.Duration(hd).Nanoseconds()
	milliseconds = milliseconds / 1000000
	if milliseconds > 1000 {
		return fmt.Sprintf(">1s")
	} else {
		return fmt.Sprintf("%4d", milliseconds)
	}
}

func printFirstHeader() {
	for index, content := range ipList {
		fmt.Printf("%d=%s\n", index+1, content)
	}
}

func printHeader() {
	fmt.Printf("   ")
	for index, _ := range ipList {
		fmt.Printf(" #%"+width+"d", index+1)
	}
	fmt.Println(" #")
}

func checkHostErr(host string, err error) {
	if err != nil {
		log.Fatal(host, ": ", err)
	}
}

func main() {
	var rowcounter int = 0
	result := make(map[string]string)
	p := fastping.NewPinger()

	for _, host := range strings.Fields(pingIP) {
		if expandDNS {
			lookups, err := net.LookupIP(host)
			checkHostErr(host, err)
			for _, ip := range lookups {
				ra := &net.IPAddr{IP: ip}
				p.AddIPAddr(ra)
				ipList = append(ipList, ra.String())
			}
		} else {
			ra, err := net.ResolveIPAddr("ip:icmp", host)
			checkHostErr(host, err)
			p.AddIPAddr(ra)
			ipList = append(ipList, ra.String())
		}
	}
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		result[addr.String()] = MilliDuration(rtt).String()
	}
	p.OnIdle = func() {
		if rowcounter%25 == 0 {
			printHeader()
			if rowcounter == 10000 {
				rowcounter = 0
			}
		}
		fmt.Printf("%04d", rowcounter)
		for _, value := range ipList {
			if result[value] != "" {
				if showRTT {
					fmt.Printf("|%"+width+"s ", result[value])
				} else {
					fmt.Printf("|%"+width+"s ", ".")
				}
			} else {
				fmt.Printf("|%"+width+"s ", "!")
			}
		}
		fmt.Println("|")
		result = make(map[string]string)
		rowcounter += 1
	}
	if expandDNS {
		printFirstHeader()
	}
	for {
		err := p.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}