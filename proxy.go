package main

import (
	"io"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

var whitelistIps = []string{"127.0.0.1"} //ip whitelist
var redisBindAddress = "127.0.0.1:6379"  //redis bind address
var serverBindAddress = ":8080"          //server bind address

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	listen, err := net.Listen("tcp", serverBindAddress)
	if err != nil {
		log.Panicln(err)
	}
	defer listen.Close()
	var count int32
	for {
		proxy, err := listen.Accept()
		if err != nil {
			log.Panicln(err)
		}
		remoteAddr := proxy.RemoteAddr().String()
		if !IsContain(whitelistIps, strings.Split(remoteAddr, ":")[0]) {
			proxy.Close()
			continue
		}
		go handleConn(proxy, remoteAddr, &count)
	}
}

func IsContain(items []string, item string) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

func handleConn(proxy net.Conn, clientIp string, count *int32) {
	atomic.AddInt32(count, 1)
	defer func() {
		atomic.AddInt32(count, -1)
		log.Printf("current count %d\n", *count)
	}()
	defer proxy.Close()
	db, err := net.DialTimeout("tcp", redisBindAddress, time.Second*2)
	if err != nil {
		log.Panicln(err)
	}
	log.Printf("clint %s connected to, count %d\n", clientIp, *count)
	go io.Copy(proxy, db)
	io.Copy(db, proxy)
	db.Close()
}
