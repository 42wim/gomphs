# gomphs

Simple tool to ping multiple hosts at once with a CLI and web-based overview.

## Download
[Binaries] (https://github.com/42wim/gomphs/releases)

## Usage 
Needs to be run as root (raw sockets for icmp)  
When using hostnames with ipv4 and ipv6 addresses, preference goes to ipv4. (-expand option shows both)

```
# gomphs
Usage of ./gomphs:
  -c int
        packets to send (default 99999)
  -expand
        use all available ip's (ipv4/ipv6) of a hostname (multiple A, AAAA)
  -file string
        ip addresses/hosts file to ping, space seperated (e.g "8.8.8.8 google.com 2a00:1450:400c:c07::66")
  -hosts string
        ip addresses/hosts to ping, space seperated (e.g "8.8.8.8 google.com 2a00:1450:400c:c07::66")
  -i int
        Ping interval in Milliseconds (default 1000)
  -labels string
        labels matching the hosts, must be the same amount of values as the hosts
  -nocolor
        disable color output
  -port string
        port the webserver listens on (default "8887")
  -showrtt
        show roundtrip time in ms
  -timestamp
        enables timestamp instead of ping count
  -web
        enable webserver
```

##### no options
```
. host is up
! host is down
```

##### showrtt
Add -showrtt on commandline   

When rtt > 1s it will just show ">1s"  
When rtt < 0ms it will just show 0  

After 25 pings the header will be repeated

##### expand 
Add -expand on commandline  

When using expand this also prints an extra (onetime) header so you know what ip addresses belong to what number.
See examples.

##### webserver
Add -web on commandline  

The web GUI is by default available via http://<server-running-gomphs>:8887/stream  
Use -port to use another port.

When an IP address/host isn't reachable, this will drop to -10 on the Y-axis. 

## Examples

##### using web, expand and showrtt
```
# gomphs -hosts="facebook.com slashdot.org www.linkedin.com" -showrtt -expand -web
```
 ![stream](http://i.snag.gy/Ow7kK.jpg)

##### using expand
Facebook resolves into 2 addresses (ipv4/ipv6), see 5 and 6 in example.  
When using expand this also prints an extra (onetime) header so you know what ip addresses belong to what number

[![asciicast](https://asciinema.org/a/4hh0lgl8j23ibycubz60vko3r.png)](https://asciinema.org/a/4hh0lgl8j23ibycubz60vko3r)

## Building
Make sure you have [Go](https://golang.org/doc/install) properly installed, including setting up your [GOPATH](https://golang.org/doc/code.html#GOPATH)

Next, clone this repository into $GOPATH/src/github.com/42wim/gomphs

``` bash
$ git clone https://github.com/42wim/gomphs.git
$ cd gomphs
$ go build
```

