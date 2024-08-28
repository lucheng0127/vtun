# vtun
## Description
Vtun is a cross platfrom vpn base on tun(Layer 3).

## Features
* Cross-platform: support Linux, MacOS and Windows
* AES encrypt the L3 traffic
* Assign IP to Endpoint according to login username hash
* Embed [wintun](https://www.wintun.net/) in vtun.exe for Windows user
* Allowed-ips supported, just like wireguard

## Build
```
➜  vtun git:(main) ✗ make
mkdir -p build/amd64 build/arm64
cp wintun/amd64/wintun.dll pkg/utils
GOOS=linux GOARCH=amd64 go build -o build/amd64/vtun main.go
GOOS=darwin GOARCH=amd64 go build -o build/amd64/vtun.darwin main.go
GOOS=windows GOARCH=amd64 go build -o build/amd64/vtun.exe main.go
cp wintun/arm64/wintun.dll pkg/utils
GOOS=linux GOARCH=arm64 go build -o build/arm64/vtun main.go
GOOS=darwin GOARCH=arm64 go build -o build/arm64/vtun.darwin main.go
GOOS=windows GOARCH=arm64 go build -o build/arm64/vtun.exe main.go
```

## Try it out
**Server**
```
➜  vtun git:(main) ✗ ./build/amd64/vtun server -d conf/server
INFO[0000] vtun server run on udp port 6123             
INFO[0015] remote 10.67.0.2:60761 login with user user2 asign ip 192.168.203.116/24 
INFO[0015] add allowed ip 10.68.0.0/24 for endpoint with ip 192.168.203.116/24 remote address 10.67.0.2:60761
DEBU[0015] hearbeat received from 10.67.0.2:60761 ip 192.168.203.116/24
```

**Client**
```
➜  vtun git:(main) ✗ ip netns exec c2 ./build/amd64/vtun client -c conf/client/config2.yaml
DEBU[0000] send allowed-ips 10.68.0.0/24                
INFO[0000] connect to server succeed, endpoint ip 192.168.203.116/24
DEBU[0000] Heartbeat sent
```

**Ping**
```
➜  vtun git:(main) ✗ ip l |grep tun
188: tun-KEmS: <POINTOPOINT,MULTICAST,NOARP,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UNKNOWN mode DEFAULT group default qlen 500
➜  vtun git:(main) ✗ ping -c 1 192.168.203.116   
PING 192.168.203.116 (192.168.203.116) 56(84) bytes of data.
64 bytes from 192.168.203.116: icmp_seq=1 ttl=64 time=0.205 ms

--- 192.168.203.116 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 0.205/0.205/0.205/0.000 ms
➜  vtun git:(main) ✗ ip r add 10.68.0.0/24 dev tun-KEmS                                                                                                                     
➜  vtun git:(main) ✗ ping -c 1 10.68.0.2
PING 10.68.0.2 (10.68.0.2) 56(84) bytes of data.
64 bytes from 10.68.0.2: icmp_seq=1 ttl=63 time=0.434 ms

--- 10.68.0.2 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 0.434/0.434/0.434/0.000 ms
```
