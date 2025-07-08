package torrent

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
		// currently allowing all origins
		// restrict with r.Header.Get("Origin") == "http://torrent-service:8080"
	},
}

func addRoutes(r *gin.Engine) {

	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				break
			}
			fmt.Println("Received from peer:", string(msg))
		}
	})

	r.POST("/seed", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
			return
		}

		filePath := filepath.Join("torrent/files", filepath.Base(file.Filename))

		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		metadata, err := SplitFile(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to split file"})
			return
		}

		metaFilePath := filepath.Join("torrent/metadata", file.Filename+".meta.json")
		metaFile, err := os.Create(metaFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create metadata file"})
			return
		}
		defer metaFile.Close()

		err = json.NewEncoder(metaFile).Encode(metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write metadata file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "File uploaded and split successfully",
			"metadata": metadata,
		})
	})

	r.GET("/metadata/:filename", func(c *gin.Context) {
		file := c.Param("filename")
		path := filepath.Join("torrent/metadata", file+".meta.json")
		c.File(path)
	})

	r.GET("/metadata", func(c *gin.Context) {
		metaDir := "torrent/metadata"
		files, err := os.ReadDir(metaDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read metadata directory"})
			return
		}

		var allMetadata []ChunkMetadata

		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".json" && filepath.Ext(file.Name()) != ".meta.json" {
				continue
			}

			path := filepath.Join(metaDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue // skip unreadable files
			}

			var metadata ChunkMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				continue // skip invalid JSON
			}

			allMetadata = append(allMetadata, metadata)
		}

		c.JSON(http.StatusOK, allMetadata)
	})

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

func addLimitedConcurrencyRoutes(r *gin.RouterGroup) {

	r.GET("/chunk/:filename/:index", func(c *gin.Context) {
		file := c.Param("filename")
		index := c.Param("index")

		chunkPath := filepath.Join("torrent/chunks", fmt.Sprintf("%s.chunk.%s", file, index))
		if _, err := os.Stat(chunkPath); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chunk not found"})
			return
		}

		c.File(chunkPath)
	})
}
