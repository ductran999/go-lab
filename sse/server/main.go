package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Cấu trúc dữ liệu gửi qua SSE
type ProvisionEvent struct {
	Step               string `json:"step"`
	ProgressPercentage int    `json:"progress_percentage"`
	Message            string `json:"message"`
	Error              string `json:"error,omitempty"`
}

var computeStore = make(map[string]chan ProvisionEvent)

func main() {
	r := gin.Default()

	r.POST("/computes", func(c *gin.Context) {
		computeID := "cpt-12345"

		computeStore[computeID] = make(chan ProvisionEvent, 10)

		go simulateProvisioning(computeID)

		c.JSON(http.StatusAccepted, gin.H{
			"compute_id": computeID,
			"message":    "Provisioning started",
		})
	})

	r.GET("/computes/:id/stream", func(c *gin.Context) {
		computeID := c.Param("id")

		ch, exists := computeStore[computeID]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Compute not found or already finished"})
			return
		}

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		c.Stream(func(w io.Writer) bool {
			event, ok := <-ch
			if !ok {
				return false
			}

			data, err := json.Marshal(event)
			if err != nil {
				return false
			}

			payload := fmt.Appendf(nil, "data: %s\n\n", data)

			_, err = c.Writer.Write(payload)
			if err != nil {
				return false
			}
			c.Writer.Flush()

			if event.Step == "COMPLETED" || event.Step == "FAILED" {
				delete(computeStore, computeID)
				return false
			}

			return true
		})
	})

	r.Run(":8080")
}

func simulateProvisioning(computeID string) {
	ch := computeStore[computeID]
	if ch == nil {
		return
	}
	defer close(ch)

	time.Sleep(2 * time.Second)
	ch <- ProvisionEvent{
		Step:               "CREATING_EXEC_NAMESPACE",
		ProgressPercentage: 25,
		Message:            "Creating execution namespace and applying ResourceQuota...",
	}

	time.Sleep(2 * time.Second)
	ch <- ProvisionEvent{
		Step:               "CREATING_OPERATOR_NAMESPACE",
		ProgressPercentage: 50,
		Message:            "Creating operator namespace...",
	}

	time.Sleep(3 * time.Second)
	ch <- ProvisionEvent{
		Step:               "INSTALLING_OPERATOR",
		ProgressPercentage: 75,
		Message:            "Deploying Spark Operator helm chart...",
	}

	time.Sleep(2 * time.Second)
	ch <- ProvisionEvent{
		Step:               "COMPLETED",
		ProgressPercentage: 100,
		Message:            "Compute is ready to use!",
	}
}
