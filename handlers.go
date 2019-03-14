package main

import (
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	logs.Info("a sample app log")
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// Your handlers goes here
