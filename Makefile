default: build

build:
	go build .
	mv ./mydocker /usr/local/bin/
	mkdir -p /root/mydockerspace/images

init:
	cp -rf ./images/* /root/mydockerspace/images

clean_bridge:
	ip link set mybridge down
	brctl delbr mybridge
	ip link set testbridge down
	brctl delbr testbridge
	rm -rf /root/mydockerspace/network


