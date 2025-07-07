package main

import (
	"fmt"
	"friendly-fiesta/torrent"
	"sync"
)

func main() {

	fmt.Println("Starting the torrent servers...")
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		torrent.SeedServer()
	}()

	go func() {
		defer wg.Done()
		torrent.LeechServer()
	}()

	wg.Wait()

}
