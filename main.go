package main

import (
	"fmt"
	"friendly-fiesta/torrent"
)

func main() {
	fmt.Println("Hello, World!")
	torrent.TorrentServer()
	fmt.Println("Seeder server is running on port 8080")
}
