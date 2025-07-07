package torrent

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func LeechServer() {

	// Create necessary directories for torrent files, chunks, and metadata
	os.MkdirAll("torrent/files", os.ModePerm)
	os.MkdirAll("torrent/chunks", os.ModePerm)
	os.MkdirAll("torrent/metadata", os.ModePerm)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	limited := router.Group("/torrent")
	limited.Use(LimitConcurrentRequests(10)) // Limit to 10 concurrent requests
	addLeechRoutes(router)
	addLimitedConcurrencyLeechRoutes(limited)

	fmt.Println("Leeching on port 8081")
	router.Run(":8081")
}
