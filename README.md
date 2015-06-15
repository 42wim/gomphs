# gomphs

Simple tool to ping multiple hosts at once with an overview

## usage 
Needs to be run as root (raw sockets for icmp)  
When using hostnames with ipv4 and ipv6 addresses, preference goes to ipv4.

```
# gomphs
usage:
  -hosts="": ip addresses/hosts to ping, space seperated (e.g "8.8.8.8 8.8.4.4 google.com 2a00:1450:400c:c07::66")
  -showrtt=false: show roundtrip time in ms
```

```
. host is up
! host is down
```

* With showrtt enabled 

When rtt > 1s it will just show ">1s"  
When rtt < 0ms it will just show 0  

After 20 pings the header will be repeated

## example

```
# gomphs -hosts "8.8.8.8 192.168.1.1 8.8.4.4 2a00:1450:400c:c07::66 facebook.com" -showrtt
	|   1 |   2 |   3 |   4 |   5 |
0000|  12 |   ! |  14 |  39 | 110 |
0001|  13 |   ! |  12 |  39 | 110 |
0002|  11 |   ! |  11 |  40 | 108 |
0003|  13 |   ! |  11 |  37 | 110 |
0004|  12 |   ! |  12 |  39 | 109 |
0005|  12 |   ! |  13 |  38 | 109 |

# gomphs -hosts "8.8.8.8 192.168.1.1 8.8.4.4 2a00:1450:400c:c07::66 facebook.com"
    | 1 | 2 | 3 | 4 | 5 |
0000| . | ! | . | . | . |
0001| . | ! | . | . | . |
0002| . | ! | . | . | . |
0003| . | ! | . | . | . |
0004| . | ! | . | . | . |
0005| . | ! | . | . | . |
```
