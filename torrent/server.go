package torrent

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func Server() {
	// Create directories
	os.MkdirAll(filepath.Join(baseTorrentPath, "files"), os.ModePerm)
	os.MkdirAll(filepath.Join(baseTorrentPath, "metadata"), os.ModePerm)
	os.MkdirAll(filepath.Join(baseTorrentPath, "chunks"), os.ModePerm)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	concurrent := router.Group("/concurrent")
	concurrent.Use(LimitConcurrentRequests(10))

	addRoutes(router)
	addLimitedConcurrencyRoutes(concurrent)

	// Start connecting to peers in background
	go func() {
		for {
			peerAddrs := getCurrentPeerList()
			connectToPeers(peerAddrs)
			time.Sleep(30 * time.Second) // re-check every 30s
		}
	}()

	fmt.Println("Listening on port 8080")
	router.Run(":8080")
}

func connectToPeers(peerAddrs []string) {
	for _, addr := range peerAddrs {
		go func(addr string) {
			url := fmt.Sprintf("ws://%s/ws", addr)
			conn, _, err := websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				log.Println("Dial error to", addr, ":", err)
				return
			}
			defer conn.Close()

			conn.WriteMessage(websocket.TextMessage, []byte("Peer connected"))
		}(addr)
	}
}

func getCurrentPeerList() []string {
	ips, err := net.LookupIP("torrent-service.default.svc.cluster.local")
	if err != nil {
		log.Println("DNS lookup failed:", err)
		return nil
	}

	var addrs []string
	for _, ip := range ips {
		// Exclude the local pod IP to avoid connecting to itself
		if ip.String() != getLocalPodIP() {
			addrs = append(addrs, fmt.Sprintf("%s:8080", ip.String()))
		}
	}
	return addrs
}

func getLocalPodIP() string {
	host, _ := os.Hostname()
	ip, _ := net.LookupIP(host)
	return ip[0].String()
}
