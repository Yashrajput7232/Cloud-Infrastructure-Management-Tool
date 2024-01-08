package compute

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine) {
	r.GET("/compute-instances", getAllComputeInstances)
	r.GET("/instance-details/:instanceType", getInstanceDetails)
	r.POST("/start-instance/:instanceType", startComputeInstance)
	r.POST("/terminate-instance/:instanceID", terminateComputeInstance)
}

