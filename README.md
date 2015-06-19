# gomphs

Simple tool to ping multiple hosts at once with a CLI and web-based overview.

## usage 
Needs to be run as root (raw sockets for icmp)  
When using hostnames with ipv4 and ipv6 addresses, preference goes to ipv4. (-expand option shows both)

```
# gomphs
Usage of ./gomphs:
  -expand=false: use all available ip's (ipv4/ipv6) of a hostname (multiple A, AAAA)
  -hosts="": ip addresses/hosts to ping, space seperated (e.g "8.8.8.8 8.8.4.4 google.com 2a00:1450:400c:c07::66")
  -port="8887": port the webserver listens on
  -showrtt=false: show roundtrip time in ms
  -web=false: enable webserver
```

### no options
```
. host is up
! host is down
```

### showrtt
Add -showrtt on commandline   

When rtt > 1s it will just show ">1s"  
When rtt < 0ms it will just show 0  

After 25 pings the header will be repeated

### expand 
Add -expand on commandline  

When using expand this also prints an extra (onetime) header so you know what ip addresses belong to what number.
See examples.

### webserver
Add -web on commandline  

The web GUI is by default available via http://<server-running-gomphs>:8887/stream  
Use -port to use another port.

When an IP address/host isn't reachable, this will drop to -10 on the Y-axis. 

## example
```
# gomphs -hosts="facebook.com slashdot.org www.linkedin.com" -showrtt -expand -web
```
 ![stream](http://i.snag.gy/Ow7kK.jpg)

```
# gomphs -hosts "8.8.8.8 192.168.1.1 8.8.4.4 2a00:1450:400c:c07::66 facebook.com" -showrtt
	#   1 #   2 #   3 #   4 #   5 #
0000|  12 |   ! |  14 |  39 | 110 |
0001|  13 |   ! |  12 |  39 | 110 |
0002|  11 |   ! |  11 |  40 | 108 |
0003|  13 |   ! |  11 |  37 | 110 |
0004|  12 |   ! |  12 |  39 | 109 |
0005|  12 |   ! |  13 |  38 | 109 |

# gomphs -hosts "8.8.8.8 192.168.1.1 8.8.4.4 2a00:1450:400c:c07::66 facebook.com"
    # 1 # 2 # 3 # 4 # 5 #
0000| . | ! | . | . | . |
0001| . | ! | . | . | . |
0002| . | ! | . | . | . |
0003| . | ! | . | . | . |
0004| . | ! | . | . | . |
0005| . | ! | . | . | . |
```

* Using expand
Facebook resolves into 2 addresses (ipv4/ipv6), see 5 and 6 in example.  
When using expand this also prints an extra (onetime) header so you know what ip addresses belong to what number

```
# gomphs -hosts "8.8.8.8 192.168.1.1 8.8.4.4 2a00:1450:400c:c07::66 facebook.com" -showrtt -expand
1=8.8.8.8
2=192.168.1.1
3=8.8.4.4
4=2a00:1450:400c:c07::66
5=2a03:2880:2130:cf05:face:b00c:0:1
6=173.252.120.6
    #   1 #   2 #   3 #   4 #   5 #   6 #
0000|  12 |   ! |  12 |  40 | 129 | 131 |
0001|  13 |   ! |  15 |  40 | 129 | 135 |
0002|  14 |   ! |  14 |  41 | 128 | 134 |
0003|  14 |   ! |  14 |  41 | 130 | 134 |
0004|  16 |   ! |  24 |  43 | 128 | 134 |
0005|  13 |   ! |  12 |  39 | 126 | 132 |
0006|  14 |   ! |  14 |  40 | 128 | 132 |
0007|  15 |   ! |  14 |  41 | 126 | 134 |
0008|  11 |   ! |  13 |  39 | 128 | 134 |

# gomphs -hosts "8.8.8.8 192.168.1.1 8.8.4.4 2a00:1450:400c:c07::66 facebook.com" -expand
1=8.8.8.8
2=192.168.1.1
3=8.8.4.4
4=2a00:1450:400c:c07::66
5=2a03:2880:2130:cf05:face:b00c:0:1
6=173.252.120.6
    # 1 # 2 # 3 # 4 # 5 # 6 #
0000| . | ! | . | . | . | . |
0001| . | ! | . | . | . | . |
0002| . | ! | . | . | . | . |
0003| . | ! | . | . | . | . |
0004| . | ! | . | . | . | . |
0005| . | ! | . | . | . | . |
0006| . | ! | . | . | . | . |
```
