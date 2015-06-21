package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/tatsushid/go-fastping"
)

var pingIP, listenPort string
var expandDNS, showRTT, enableWeb bool
var width string = "2"

var ipList []string
var ipListMap map[string][]string
var pingStats map[string]stats

func init() {
	flag.BoolVar(&enableWeb, "web", false, "enable webserver")
	flag.BoolVar(&expandDNS, "expand", false, "use all available ip's (ipv4/ipv6) of a hostname (multiple A, AAAA)")
	flag.StringVar(&listenPort, "port", "8887", "port the webserver listens on")
	flag.StringVar(&pingIP, "hosts", "", "ip addresses/hosts to ping, space seperated (e.g \"8.8.8.8 8.8.4.4 google.com 2a00:1450:400c:c07::66\")")
	flag.BoolVar(&showRTT, "showrtt", false, "show roundtrip time in ms")
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

func (hd MilliDuration) Int() int {
	milliseconds := time.Duration(hd).Nanoseconds()
	milliseconds = milliseconds / 1000000
	return int(milliseconds)
}

type gomphs struct {
	latestEntry []byte
	pingIP      string
	showRTT     bool
	expandDNS   bool
	IpList      []string
	IpListMap   map[string][]string
}

type stats struct {
	min   int
	max   int
	avg   int
	count int
	rtts  []int
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
	pingStats = make(map[string]stats)
	g := &gomphs{}

	if enableWeb {
		listener, err := net.Listen("tcp", ":"+listenPort)
		if err != nil {
			log.Fatal(err)
		}
		go http.Serve(listener, nil)
		http.HandleFunc("/read.json", webReadJsonHandler(g))
		http.HandleFunc("/stream", webStreamHandler)
	}

	result := make(map[string]string)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Printf("\n%-32s: %4s %4s %4s %5s\n", "source", "min", "max", "avg", "ploss")
			for key, stat := range pingStats {
				ploss := stat.count - len(stat.rtts)
				plosspct := float32(ploss) / float32(stat.count)
				fmt.Printf("%-32s: %4d %4d %4d %5d(%.2f%%)\n", key, stat.min, stat.max, stat.avg, ploss, plosspct)
			}
			os.Exit(0)
		}
	}()
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
		stats := pingStats[addr.String()]
		if stats.count == 0 {
			stats.min = 100000
		}
		if MilliDuration(rtt).Int() > stats.max {
			stats.max = MilliDuration(rtt).Int()
		}
		if MilliDuration(rtt).Int() < stats.min {
			stats.min = MilliDuration(rtt).Int()
		}
		stats.rtts = append(stats.rtts, MilliDuration(rtt).Int())
		pingStats[addr.String()] = stats
	}
	p.OnIdle = func() {
		if rowcounter%25 == 0 {
			printHeader()
			if rowcounter == 10000 {
				rowcounter = 0
			}
		}
		g.update(result)
		fmt.Printf("%04d", rowcounter+1)
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

		for _, key := range ipList {
			for _, value := range ipListMap[key] {
				if result[value] != "" {
					stats := pingStats[value]
					stats.count = rowcounter + 2
					stats.avg = 0
					i := 1
					for _, rtt := range stats.rtts {
						stats.avg = stats.avg + rtt
						i += 1
					}
					stats.avg = stats.avg / i
					pingStats[value] = stats
				}
			}
		}
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
