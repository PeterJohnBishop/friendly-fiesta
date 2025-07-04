package seeder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func SeedServer() {

	os.MkdirAll("seeder/files", os.ModePerm)
	os.MkdirAll("seeder/chunks", os.ModePerm)
	os.MkdirAll("seeder/metadata", os.ModePerm)

	router := gin.Default()

	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
			return
		}

		filePath := filepath.Join("seeder/files", filepath.Base(file.Filename))

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

		metaFilePath := filepath.Join("seeder/metadata", file.Filename+".meta.json")
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

	router.GET("/metadata/:filename", func(c *gin.Context) {
		file := c.Param("filename")
		path := filepath.Join("seeder/metadata", file+".meta.json")
		c.File(path)
	})

	router.GET("/chunk/:filename/:index", func(c *gin.Context) {
		file := c.Param("filename")
		index := c.Param("index")

		chunkPath := filepath.Join("seeder/chunks", fmt.Sprintf("%s.chunk.%s", file, index))
		if _, err := os.Stat(chunkPath); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chunk not found"})
			return
		}

		c.File(chunkPath)
	})

	router.Run(":8080")
}
