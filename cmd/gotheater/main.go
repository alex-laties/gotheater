package main

import "github.com/gin-gonic/gin"

/**
A webserver that provides access to:
 - User management
 - Room management
 - Content management
 - Transcode management
*/
func main() {
	router := gin.Default()
	router.Static("/", "/var/lib/gotheater/frontend")
	router.Run(":8080")
}
