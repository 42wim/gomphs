# gomphs

Simple tool to ping multiple hosts at once with an overview

## usage 
Needs to be run as root (raw sockets for icmp)

```
# gomphs
usage:
  -hosts="": ip addresses to ping, space seperated (e.g "8.8.8.8 8.8.4.4")
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
# gomphs -hosts "8.8.8.8 192.168.1.1 8.8.4.4"
	| 1 | 2 | 3 |
0000| . | ! | . |
0001| . | ! | . |
0002| . | ! | . |
0003| . | ! | . |
0004| . | ! | . |
0005| . | ! | . |
```

```
# gomphs -hosts "8.8.8.8 192.168.1.1 8.8.4.4" -showrtt
    |   1 |   2 |   3 |
0000|  13 |   ! |  12 |
0001|  12 |   ! |  12 |
0002|  11 |   ! |  13 |
0003|  11 |   ! |  13 |
0004|  12 |   ! |  12 |
0005|  12 |   ! |  12 |
```
