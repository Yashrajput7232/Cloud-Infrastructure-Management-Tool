package main

import (
	"github.com/gin-gonic/gin"
	"compute-service/pkg/compute"
)

func main() {
	r := gin.Default()

	compute.SetupRoutes(r) // Setup compute package routes

	r.Run(":8000")
}
