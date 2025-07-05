package torrent

import "github.com/gin-gonic/gin"

func LimitConcurrentRequests(maxConcurrent int) gin.HandlerFunc {
	semaphore := make(chan struct{}, maxConcurrent)

	return func(c *gin.Context) {
		semaphore <- struct{}{} // acquire slot
		defer func() {
			<-semaphore // release slot
		}()
		c.Next()
	}
}
