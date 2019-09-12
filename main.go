package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ywengineer/snowflake-golang/snowflake"
	"github.com/ywengineer/snowflake-golang/v2"
	"net/http"
	"strconv"
)

var idWorkerMap = make(map[int]*snowflake.Node)
var workerMap = make(map[string]*v2.Worker)

func main() {
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// Get ID
	r.GET("/worker/:id", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Params.ByName("id"))
		value, ok := idWorkerMap[id]
		if ok {
			nid := value.Generate()
			c.JSON(200, gin.H{"id": nid})
		} else {
			iw, err := snowflake.NewNode(int64(id))
			if err == nil {
				nid := value.Generate()
				idWorkerMap[id] = iw
				c.JSON(200, gin.H{"id": nid})
			} else {
				fmt.Println(err)
			}
		}
	})

	// Get ID
	r.GET("/v2/worker/:center/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Params.ByName("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		center, err := strconv.Atoi(c.Params.ByName("center"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		workerKey := fmt.Sprintf("%d_%d", center, id)
		value, ok := workerMap[workerKey]
		if ok {
			if nid, err := value.NextId(); err == nil {
				c.JSON(200, gin.H{"id": nid})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}
		} else {
			iw, err := v2.NewWorker(uint64(center), uint64(id))
			if err == nil {
				workerMap[workerKey] = iw
				if nid, err := value.NextId(); err == nil {
					c.JSON(200, gin.H{"id": nid})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}
		}
	})
	// Listen and Server in 0.0.0.0:8182
	_ = r.Run(":8182")
}
