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
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	fastping "github.com/tatsushid/go-fastping"
)

var pingIP, listenPort, pingLabel string
var flagExpandDNS, flagShowRTT, flagEnableWeb, flagNoColor, flagTimestamp bool
var width = "2"
var rowcounter, maxPingCount int
var interval int

var ipList, ipLabelList, inputHosts, inputLabels []string
var ipListMap map[string][]string
var pingStats map[string]stats

func init() {
	flag.BoolVar(&flagNoColor, "nocolor", false, "disable color output")
	flag.BoolVar(&flagEnableWeb, "web", false, "enable webserver")
	flag.BoolVar(&flagExpandDNS, "expand", false, "use all available ip's (ipv4/ipv6) of a hostname (multiple A, AAAA)")
	flag.BoolVar(&flagTimestamp, "timestamp", false, "enables timestamp instead of ping count")
	flag.StringVar(&listenPort, "port", "8887", "port the webserver listens on")
	flag.StringVar(&pingIP, "hosts", "", "ip addresses/hosts to ping, space seperated (e.g \"8.8.8.8 8.8.4.4 google.com 2a00:1450:400c:c07::66\")")
	flag.StringVar(&pingLabel, "labels", "", "labels matching the hosts, must be the same amount of values as the hosts")
	flag.BoolVar(&flagShowRTT, "showrtt", false, "show roundtrip time in ms")
	flag.IntVar(&maxPingCount, "c", 99999, "packets to send")
	flag.IntVar(&interval, "i", 1000, "Ping interval in Milliseconds")
	flag.Parse()
	if flag.NFlag() == 0 {
		fmt.Println("usage: ")
		flag.PrintDefaults()
		os.Exit(2)
	}
	if flagShowRTT {
		width = "5"
	}
	inputHosts = strings.Fields(pingIP)
	if len(pingLabel) > 0 {
		inputLabels = strings.Fields(pingLabel)
		if len(inputHosts) != len(inputLabels) {
			fmt.Println("ERROR: The number of hosts vs the number of labels does not match. Usage:")
			flag.PrintDefaults()
			os.Exit(2)
		}
	}
	if flagNoColor {
		color.NoColor = true
	}
	if runtime.GOOS != "windows" {
		if os.Geteuid() != 0 {
			fmt.Println("Please start gomphs as root or use sudo!")
			fmt.Println("This is required for raw socket access.")
			os.Exit(1)
		}
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
	printcounter := rowcounter
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
			printStat()
			os.Exit(0)
		}
	}()
	p := fastping.NewPinger()
	p.MaxRTT = time.Millisecond * time.Duration(interval)

	labelIndex := 0
	for _, host := range inputHosts {
		if flagExpandDNS {
			lookups, err := net.LookupIP(host)
			checkHostErr(host, err)
			ipList = append(ipList, host)
			hostLabelIndex := 1
			for _, ip := range lookups {
				ra := &net.IPAddr{IP: ip}
				p.AddIPAddr(ra)
				ipListMap[host] = append(ipListMap[host], ra.String())
				if len(inputLabels) > 0 {
					ipLabelList = append(ipLabelList, inputLabels[labelIndex]+"-"+strconv.Itoa(hostLabelIndex))
				}
				hostLabelIndex++
			}
		} else {
			ra, err := net.ResolveIPAddr("ip:icmp", host)
			checkHostErr(host, err)
			p.AddIPAddr(ra)
			ipList = append(ipList, ra.String())
			ipListMap[ra.String()] = append(ipListMap[ra.String()], ra.String())
			if len(inputLabels) > 0 {
				ipLabelList = append(ipLabelList, inputLabels[labelIndex])
			}
		}
		labelIndex++
	}
	for _, label := range ipLabelList {
		widthInt, _ := strconv.Atoi(width)
		if len(label) > widthInt {
			width = strconv.Itoa(len(label))
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
			if flagTimestamp {
				printHeader("25", ipLabelList)
			} else {
				printHeader("4", ipLabelList)
			}
			if rowcounter%10000 == 0 {
				printcounter = 0
			}
		}
		g.update(result)
		if flagTimestamp {
			t := time.Now()
			fmt.Printf("%24s", t.Format(time.RFC3339))
		} else {
			fmt.Printf("%04d", printcounter+1)
		}
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
		printcounter++
	}
	if flagExpandDNS {
		printFirstHeader(ipLabelList)
	}
	for {
		if rowcounter == maxPingCount {
			break
		}
		err := p.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
	printStat()
}
