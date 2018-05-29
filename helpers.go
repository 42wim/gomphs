package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

var webStream = `<html>

        <script src="http://cdnjs.cloudflare.com/ajax/libs/dygraph/1.1.0/dygraph-combined.js"></script>
        <script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.3/jquery.min.js"></script>
        <script src="http://cdnjs.cloudflare.com/ajax/libs/jquery-csv/0.71/jquery.csv-0.71.min.js"></script>
        <body>
		<div id="status" style="width:800px;"></div>
        <div id="graphdiv" style="width:800px; height:600px;"></div>
        <script type="text/javascript">
        var d = [];

        g = new Dygraph(
            document.getElementById("graphdiv"),
            d, {
			legend: 'always',
            showRoller: false,
			yLabelWidth: 40,
		    labelsDiv: document.getElementById('status'),
			hideOverlayOnMouseOut: false,
            #labelrewrite#
            });

        window.intervalId = setInterval(function () {
                        now = new Date();
                        $.getJSON('/read.json?t='+now.toString(), function(data) {
						#pushrewrite#
                    g.updateOptions( { 'file': d } );
                });
              }, 1000);
        </script>
        </body>
        </html>`

func webReadJSONHandler(g *gomphs) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(g.latestEntry))
	}
}
func webStreamHandler(w http.ResponseWriter, r *http.Request) {
	push := "d.push([now"
	labels := "labels: [\"time\""
	index := 1
	for _, key := range ipList {
		for _, value := range ipListMap[key] {
			push = fmt.Sprintf("%s,data[%d]", push, index)
			if key != value {
				labels = fmt.Sprintf("%s,\"%s\"", labels, key+"("+value+")")
			} else {
				labels = fmt.Sprintf("%s,\"%s\"", labels, key)
			}
			index++
		}
	}
	push = push + "])"
	labels = labels + "]"

	webStream = strings.Replace(webStream, "#pushrewrite#", push, -1)
	webStream = strings.Replace(webStream, "#labelrewrite#", labels, -1)
	fmt.Fprintf(w, webStream)
}

func printFirstHeader(labels []string) {
	index := 1
	for _, key := range ipList {
		for _, content := range ipListMap[key] {
			if len(labels) > 0 {
				fmt.Printf("%s=%s\n", labels[index-1], content)
			} else {
				fmt.Printf("%d=%s\n", index, content)
			}
			index++
		}
	}
}

func printHeader(hw string, labels []string) {
	fmt.Printf("%"+hw+"s", " ")
	index := 0
	for _, key := range ipList {
		for range ipListMap[key] {
			index++
			if len(labels) > 0 {
				fmt.Printf(" %"+width+"s", labels[index-1])
			} else {
				fmt.Printf(" %"+width+"d", index)
			}
		}
	}
	fmt.Println(" ")
}

func checkHostErr(host string, err error) {
	if err != nil {
		log.Fatal(host, ": ", err)
	}
}

func printHostIPStat(host string) {
	pingStatsHost := make(map[string]stats)
	hoststat := pingStatsHost[host]
	ipcount := len(ipListMap[host])
	hoststat.min = 10000
	hostploss := 0
	for _, hostip := range ipListMap[host] {
		stat := pingStats[hostip]
		ploss := 0
		for _, entry := range stat.rtts[0:rowcounter] {
			if entry != -1 {
				if entry > hoststat.max {
					hoststat.max = entry
				}
				if entry > stat.max {
					stat.max = entry
				}
				if entry < hoststat.min {
					hoststat.min = entry
				}
				if entry < stat.min {
					stat.min = entry
				}
				stat.avg = stat.avg + entry
				hoststat.avg = hoststat.avg + entry
			} else {
				ploss++
				hostploss++
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
		if len(ipListMap[host]) == 1 && hostip == host {
			fmt.Printf("+")
		} else {
			fmt.Printf("|")
		}
		fmt.Printf("%-38s: %4d %4d %4d %5d/%d (%.2f%%)\n", hostip, stat.min, stat.max, stat.avg, ploss, rowcounter, plosspct)
	}
	if hostploss == rowcounter*ipcount {
		hoststat.min = -1
		hoststat.max = -1
		hoststat.avg = -1
	} else {
		hoststat.avg = hoststat.avg / ((rowcounter * ipcount) - hostploss)
	}

	plosspct := float32(hostploss) / float32(rowcounter*ipcount) * 100
	pingStatsHost[host] = hoststat
	if ipcount == 1 && ipListMap[host][0] == host {
	} else {
		fmt.Printf("+%-38s: %4d %4d %4d %5d/%d (%.2f%%)\n", host, hoststat.min, hoststat.max, hoststat.avg, hostploss, rowcounter*ipcount, plosspct)
	}
}

func printStat() {
	fmt.Printf("\n%-39s: %4s %4s %4s   %5s\n", "source", "min", "max", "avg", "ploss")
	for _, host := range ipList {
		printHostIPStat(host)
		fmt.Println()
	}
}
