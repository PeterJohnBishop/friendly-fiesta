package torrent

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func addLeechRoutes(r *gin.Engine) {

	r.GET("/leech/:filename", func(c *gin.Context) {
		file := c.Param("filename")
		sourceServer := "http://localhost:8080" // seeder server
		url := fmt.Sprintf("%s/metadata/%s", sourceServer, file)

		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metadata from source"})
			return
		}
		defer resp.Body.Close()

		localDir := "torrent/metadata"
		if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create metadata directory"})
			return
		}

		localPath := filepath.Join(localDir, file+".meta.json")
		outFile, err := os.Create(localPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create local metadata file"})
			return
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, resp.Body); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Metadata file downloaded and saved",
			"localPath": localPath,
		})
	})

}

func addLimitedConcurrencyLeechRoutes(r *gin.RouterGroup) {
}
