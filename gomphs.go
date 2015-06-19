package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
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
var ipListMap map[string][]string

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

type gomphs struct {
	latestEntry []byte
	pingIP      string
	showRTT     bool
	expandDNS   bool
	IpList      []string
	IpListMap   map[string][]string
}

func (g *gomphs) update(result map[string]string) {
	record := []string{time.Now().Format("2006/01/02 15:04:05")}
	for _, key := range g.IpList {
		for _, value := range g.IpListMap[key] {
			if result[value] != "" {
				res := strings.Replace(result[value], " ", "", -1)
				record = append(record, res)
			} else {
				record = append(record, "-10")
			}
		}
	}
	g.latestEntry, _ = json.Marshal(record)
}

func main() {
	var rowcounter int = 0
	ipListMap = make(map[string][]string)
	g := &gomphs{}
	listener, err := net.Listen("tcp", ":8887")
	if err != nil {
		log.Fatal(err)
	}
	go http.Serve(listener, nil)
	http.HandleFunc("/read.json", webReadJsonHandler(g))
	http.HandleFunc("/stream", webStreamHandler)

	result := make(map[string]string)
	p := fastping.NewPinger()

	for _, host := range strings.Fields(pingIP) {
		if expandDNS {
			lookups, err := net.LookupIP(host)
			checkHostErr(host, err)
			ipList = append(ipList, host)
			for _, ip := range lookups {
				ra := &net.IPAddr{IP: ip}
				p.AddIPAddr(ra)
				ipListMap[host] = append(ipListMap[host], ra.String())
			}
		} else {
			ra, err := net.ResolveIPAddr("ip:icmp", host)
			checkHostErr(host, err)
			p.AddIPAddr(ra)
			ipList = append(ipList, ra.String())
			ipListMap[ra.String()] = append(ipListMap[ra.String()], ra.String())
		}
	}
	g.IpList = ipList
	g.IpListMap = ipListMap
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
		g.update(result)
		fmt.Printf("%04d", rowcounter)
		for _, key := range ipList {
			for _, value := range ipListMap[key] {
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
