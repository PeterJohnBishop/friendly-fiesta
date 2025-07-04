package torrent

import (
	"os"

	"github.com/gin-gonic/gin"
)

func TorrentServer() {

	os.MkdirAll("torrent/files", os.ModePerm)
	os.MkdirAll("torrent/chunks", os.ModePerm)
	os.MkdirAll("torrent/metadata", os.ModePerm)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	limited := router.Group("/torrent")
	limited.Use(LimitConcurrentRequests(10)) // Limit to 10 concurrent requests
	addSeederRoutes(router)
	addLimitedSeederRoutes(limited)

	router.Run(":8080")
}
