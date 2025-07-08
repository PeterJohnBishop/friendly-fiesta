package main

import (
	"fmt"
	"friendly-fiesta/torrent"
	"sync"
)

func main() {

	fmt.Println("Starting the torrent distribution server...")
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		torrent.Server()
	}()

	wg.Wait()

}
