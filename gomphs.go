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

	"github.com/fatih/color"
	"github.com/tatsushid/go-fastping"
)

var pingIP, listenPort string
var flagExpandDNS, flagShowRTT, flagEnableWeb, flagNoColor bool
var width = "2"

var ipList []string
var ipListMap map[string][]string
var pingStats map[string]stats

func init() {
	flag.BoolVar(&flagNoColor, "nocolor", false, "disable color output")
	flag.BoolVar(&flagEnableWeb, "web", false, "enable webserver")
	flag.BoolVar(&flagExpandDNS, "expand", false, "use all available ip's (ipv4/ipv6) of a hostname (multiple A, AAAA)")
	flag.StringVar(&listenPort, "port", "8887", "port the webserver listens on")
	flag.StringVar(&pingIP, "hosts", "", "ip addresses/hosts to ping, space seperated (e.g \"8.8.8.8 8.8.4.4 google.com 2a00:1450:400c:c07::66\")")
	flag.BoolVar(&flagShowRTT, "showrtt", false, "show roundtrip time in ms")
	flag.Parse()
	if flag.NFlag() == 0 {
		fmt.Println("usage: ")
		flag.PrintDefaults()
		os.Exit(2)
	}
	if flagShowRTT {
		width = "3"
	}
	if flagNoColor {
		color.NoColor = true
	}
}

type milliDuration time.Duration

func (hd milliDuration) String() string {
	milliseconds := time.Duration(hd).Nanoseconds()
	milliseconds = milliseconds / 1000000
	if milliseconds > 1000 {
		return fmt.Sprintf(">1s")
	}
	return fmt.Sprintf("%"+width+"d", milliseconds)
}

func (hd milliDuration) Int() int {
	milliseconds := time.Duration(hd).Nanoseconds()
	milliseconds = milliseconds / 1000000
	return int(milliseconds)
}

type gomphs struct {
	latestEntry   []byte
	pingIP        string
	flagShowRTT   bool
	flagExpandDNS bool
	IPList        []string
	IPListMap     map[string][]string
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
	for _, key := range g.IPList {
		for _, value := range g.IPListMap[key] {
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
	var rowcounter int
	ipListMap = make(map[string][]string)
	pingStats = make(map[string]stats)
	g := &gomphs{}

	if flagEnableWeb {
		listener, err := net.Listen("tcp", ":"+listenPort)
		if err != nil {
			log.Fatal(err)
		}
		go http.Serve(listener, nil)
		http.HandleFunc("/read.json", webReadJSONHandler(g))
		http.HandleFunc("/stream", webStreamHandler)
	}

	result := make(map[string]string)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Printf("\n%-38s: %4s %4s %4s %5s\n", "source", "min", "max", "avg", "ploss")
			for key, stat := range pingStats {
				ploss := 0
				for _, entry := range stat.rtts[0:rowcounter] {
					if entry != -1 {
						if entry > stat.max {
							stat.max = entry
						}
						if entry < stat.min {
							stat.min = entry
						}
						stat.avg = stat.avg + entry
					} else {
						ploss++
					}
				}
				if ploss == rowcounter {
					stat.min = -1
					stat.max = -1
					stat.avg = -1
				} else {
					stat.avg = stat.avg / (rowcounter - ploss)
				}
				plosspct := float32(ploss) / float32(rowcounter) * 100
				fmt.Printf("%-38s: %4d %4d %4d %5d/%d (%.2f%%)\n", key, stat.min, stat.max, stat.avg, ploss, rowcounter, plosspct)
			}
			os.Exit(0)
		}
	}()
	p := fastping.NewPinger()

	for _, host := range strings.Fields(pingIP) {
		if flagExpandDNS {
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
	g.IPList = ipList
	g.IPListMap = ipListMap
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		result[addr.String()] = milliDuration(rtt).String()
		stats := pingStats[addr.String()]
		if stats.count == 0 {
			stats.min = 100000
		}
		stats.rtts = append(stats.rtts, milliDuration(rtt).Int())
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
				fmt.Printf(" ")
				if result[value] != "" {
					color.Set(color.BgGreen, color.FgYellow, color.Bold)
					if flagShowRTT {
						fmt.Printf("%"+width+"s", result[value])
					} else {
						fmt.Printf("%"+width+"s", ".")
					}
					color.Unset()
				} else {
					stats := pingStats[value]
					stats.rtts = append(stats.rtts, -1)
					pingStats[value] = stats
					color.Set(color.BgRed, color.FgYellow, color.Bold)
					fmt.Printf("%"+width+"s", "!")
					color.Unset()
				}
			}
		}
		fmt.Println(" ")
		result = make(map[string]string)
		rowcounter++
	}
	if flagExpandDNS {
		printFirstHeader()
	}
	for {
		err := p.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}
