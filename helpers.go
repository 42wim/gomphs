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

func webReadJsonHandler(g *gomphs) http.HandlerFunc {
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
			index += 1
		}
	}
	push = push + "])"
	labels = labels + "]"

	webStream = strings.Replace(webStream, "#pushrewrite#", push, -1)
	webStream = strings.Replace(webStream, "#labelrewrite#", labels, -1)
	fmt.Fprintf(w, webStream)
}

func printFirstHeader() {
	index := 1
	for _, key := range ipList {
		for _, content := range ipListMap[key] {
			fmt.Printf("%d=%s\n", index, content)
			index += 1
		}
	}
}

func printHeader() {
	fmt.Printf("   ")
	index := 0
	for _, key := range ipList {
		for _, _ = range ipListMap[key] {
			index += 1
			fmt.Printf(" #%"+width+"d", index)
		}
	}
	fmt.Println(" #")
}

func checkHostErr(host string, err error) {
	if err != nil {
		log.Fatal(host, ": ", err)
	}
}
