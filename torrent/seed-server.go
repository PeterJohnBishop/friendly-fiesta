package torrent

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func SeedServer() {

	// Create necessary directories for torrent files, chunks, and metadata
	os.MkdirAll("torrent/files", os.ModePerm)
	os.MkdirAll("torrent/chunks", os.ModePerm)
	os.MkdirAll("torrent/metadata", os.ModePerm)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	limited := router.Group("/torrent")
	limited.Use(LimitConcurrentRequests(10)) // Limit to 10 concurrent requests
	addSeederRoutes(router)
	addLimitedConcurrencySeedRoutes(limited)

	fmt.Println("Seeding on port 8080")
	router.Run(":8080")
}
