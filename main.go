package main

import (
	"fmt"
	"friendly-fiesta/seeder"
)

func main() {
	fmt.Println("Hello, World!")
	seeder.SeedServer()
	fmt.Println("Seeder server is running on port 8080")
}
