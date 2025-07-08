package torrent

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var baseTorrentPath = "/data"

const ChunkSize = 1 * 1024 * 1024 // 1MB

type ChunkMetadata struct {
	FileName    string   `json:"file_name"`
	ChunkSize   int      `json:"chunk_size"`
	NumChunks   int      `json:"num_chunks"`
	ChunkHashes []string `json:"chunk_hashes"`
}

func splitFile(filePath string) (*ChunkMetadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		println("Error opening file:", err.Error())
		return nil, err
	}
	defer f.Close()

	chunksDir := baseTorrentPath + "/chunks"
	os.MkdirAll(chunksDir, os.ModePerm)

	stat, _ := f.Stat()
	fileSize := stat.Size()
	numChunks := int((fileSize + ChunkSize - 1) / ChunkSize)

	hashes := []string{}
	for i := 0; i < numChunks; i++ {
		buf := make([]byte, ChunkSize)
		n, _ := f.Read(buf)

		chunk := buf[:n]
		hash := sha256.Sum256(chunk)
		hashes = append(hashes, fmt.Sprintf("%x", hash[:]))

		chunkPath := filepath.Join(chunksDir, fmt.Sprintf("%s.chunk.%d", stat.Name(), i))
		os.WriteFile(chunkPath, chunk, 0644)
	}

	meta := &ChunkMetadata{
		FileName:    stat.Name(),
		ChunkSize:   ChunkSize,
		NumChunks:   numChunks,
		ChunkHashes: hashes,
	}

	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	os.WriteFile(filepath.Join("metadata", stat.Name()+".meta.json"), metaBytes, 0644)

	return meta, nil
}

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

		seedFilePath := filepath.Join(baseTorrentPath, "files", filepath.Base(file.Filename))

		err = c.SaveUploadedFile(file, seedFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		metadata, err := splitFile(seedFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to split file"})
			return
		}

		metaFilePath := filepath.Join(baseTorrentPath, "metadata", file.Filename+".meta.json")
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
		metaFilePath := filepath.Join(baseTorrentPath, "metadata", file+".meta.json")
		c.File(metaFilePath)
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

}

func addLimitedConcurrencyRoutes(r *gin.RouterGroup) {

	r.GET("/chunk/:filename/:index", func(c *gin.Context) {
		file := c.Param("filename")
		index := c.Param("index")

		chunkPath := filepath.Join(baseTorrentPath, fmt.Sprintf("%s.chunk.%s", file, index))
		if _, err := os.Stat(chunkPath); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chunk not found"})
			return
		}

		c.File(chunkPath)
	})
}
