package compute

import (
	"context"
	"fmt"
	"time"
	"net/http"
    "math/rand"
	"github.com/gin-gonic/gin"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"

	// "github.com/oracle/oci-go-sdk/v49/identity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "github.com/pelletier/go-toml"
)

// MongoDB connection details
var mongoURI = "mongodb+srv://yash7232:Yash%407232@cluster0.kn66rfv.mongodb.net/?retryWrites=true&w=majority"
var dbName = "computeServiceDB"
var collectionName = "instanceDetails"

// InstanceDetails represents details about a compute instance.
type InstanceDetails struct {
	InstanceType string  `bson:"instance_type"`
	CPUType      string  `bson:"cpu_type"`
	GPUType      string  `bson:"gpu_type"`
	GPUCount     int     `bson:"gpu_count"`
	Memory       float64 `bson:"memory"`
	Storage      string  `bson:"storage"`
	Pricing      float64 `bson:"pricing"`
}
func getConfigurationProvider() (common.ConfigurationProvider) {
    // ... TOML loading logic
	configPath := "/workspaces/Cloud-Infrastructure-Management-Tool/compute-service/config.toml" // Replace with your config path
	configProvider, err := common.ConfigurationProviderFromFile(configPath, "DEFAULT")
	if err!=nil{
		println(err)}
    return configProvider // configProvider is of type common.ConfigurationProvider
}
func getAllComputeInstances(c *gin.Context) {
	
		// Use the specified configuration file
		configPath := "/workspaces/Cloud-Infrastructure-Management-Tool/compute-service/config.toml" // Replace with your config path
		configProvider, err := common.ConfigurationProviderFromFile(configPath, "DEFAULT")
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("failed to load config: %v", err)})
			return
		}
	
		// Create compute client (without loop for ADs)
		computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create Compute client: %v", err)})
			return
		}
	
		// List instances directly (no AD loop)
		request := core.ListInstancesRequest{}
		response, err := computeClient.ListInstances(context.Background(), request)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("failed to list instances: %v", err)})
			return
		}
	
		var instances []map[string]string
		for _, instance := range response.Items {
			shape := *instance.Shape
			availabilityDomain := *instance.AvailabilityDomain
			instances = append(instances, map[string]string{"availabilityDomain": availabilityDomain, "shape": shape})
		}
	
		c.JSON(http.StatusOK, instances)
	}
	
// getInstanceDetails gets instance details based on instance type and stores/upserts them in MongoDB.
func getInstanceDetailsAndStore(c *gin.Context) {
    // Get instance type from query parameter
    // (assuming a parameter named "instanceType" is provided)
    instanceType := c.Query("instanceType")

    // Create compute client

    computeClient, err := core.NewComputeClientWithConfigurationProvider(getConfigurationProvider())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create compute client"})
        return
    }

    // Retrieve instance details
    instanceDetails, err := getInstanceDetails(computeClient, instanceType)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to retrieve instance details: %v", err)})
        return
    }

    // Retrieve pricing (optional)
    // pricing, err := getPricing(instanceType) // Implement this function for pricing retrieval
    // if err != nil {
    //     log.Println("Failed to retrieve pricing:", err) // Log error without failing the request
    // } else {
    //     // instanceDetails["pricingPerHour"] = pricing
    // }

    // Connect to MongoDB
    mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to MongoDB"})
        return
    }
    defer mongoClient.Disconnect(context.Background())

    collection := mongoClient.Database(dbName).Collection(collectionName)

    // Check if instance exists in MongoDB
    filter := bson.M{"instanceType": instanceType}
    result := collection.FindOne(context.Background(), filter)

    if result.Err() == mongo.ErrNoDocuments {
        // Insert new instance details
        insertResult, err := collection.InsertOne(context.Background(), instanceDetails)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert instance details"})
            return
        }
        c.JSON(http.StatusCreated, gin.H{"message": "Instance details stored successfully", "insertedID": insertResult.InsertedID})
    } else {
        // Update existing instance details
        updateResult, err := collection.UpdateOne(context.Background(), filter, bson.M{"$set": instanceDetails})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance details"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Instance details updated successfully", "matchedCount": updateResult.MatchedCount, "modifiedCount": updateResult.ModifiedCount})
    }
}

// Implement getInstanceDetails to retrieve instance details from OCI
func getInstanceDetails(computeClient core.ComputeClient, instanceType string) (*InstanceDetails, error) {
    // Fetch shape details
	// shapeRequest := core.ListShapesRequest{}

    // shapeRequest := core.ListShapesRequest{}
    // shapeResponse, err := computeClient.ListShapes(context.Background(), shapeRequest)
	// shapeRequest := core.GetShapeRequest{ShapeName: &instanceType}
    // shapeResponse, err := computeClient.GetShape(context.Background(), shapeRequest)
    // if err != nil {
    //     return nil, fmt.Errorf("failed to get shape details: %w", err)
    // }

    // Map fields to InstanceDetails struct
    instanceDetails := &InstanceDetails{
        // InstanceType: shapeResponse.Shape.Name,
        // CPUType:      shapeResponse.Shape.Ocpus ,
        // GPUType:      shapeResponse.Shape.Gpu.GPUType,
        // GPUCount:     shapeResponse.Shape.Gpu.GPUCount,
        // Memory:       shapeResponse.Shape.MemoryInGBs,
        // Storage:      fmt.Sprintf("%.2f GB", shapeResponse.Shape.StorageInGBs), // Format storage as string
    }

    // Get pricing (optional)
    // pricing, err := getPricing(computeClient , instanceType)
    // if err != nil {
    //     log.Println("Failed to retrieve pricing:", err) // Log error without failing the request
    // } else {
    //     instanceDetails.Pricing = pricing
    // }

    return instanceDetails, nil
}


// Implement getPricing to retrieve pricing information (optional)


// storeInstanceDetails stores or updates instance details in MongoDB.
func storeInstanceDetails(details InstanceDetails) error {
	// Create a MongoDB client
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			fmt.Println("Error disconnecting MongoDB client:", err)
		}
	}()

	// Get the MongoDB collection
	collection := client.Database(dbName).Collection(collectionName)

	// Define a filter to check if the instance type already exists in the collection
	filter := bson.M{"instance_type": details.InstanceType}

	// Define an update to upsert the instance details
	update := bson.M{
		"$set": bson.M{
			"cpu_type":  details.CPUType,
			"gpu_type":  details.GPUType,
			"gpu_count": details.GPUCount,
			"memory":    details.Memory,
			"storage":   details.Storage,
			"pricing":   details.Pricing,
		},
		"$currentDate": bson.M{"last_modified": true},
	}

	// Perform the upsert operation
	_, err = collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}


func startComputeInstance(c *gin.Context) {
	instanceShape:=c.Query("instanceType")
	configPath := "/workspaces/Cloud-Infrastructure-Management-Tool/compute-service/config.toml" // Replace with your config path
	configProvider, err := common.ConfigurationProviderFromFile(configPath, "DEFAULT")
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to load config: %v", err)})
		return
	}

	// Create Compute client
	computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create Compute client: %v", err)})
		return
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to connect to MongoDB: %v", err)})
		return
	}
	defer client.Disconnect(context.Background())

	// Generate a random port number between 9000 and 9100
	port := 9000 + rand.Intn(101)

	// Set up instance launch request
	launchRequest := core.LaunchInstanceRequest{
		Shape : instanceShape, // Replace with your instance shape
		// ... configure other request fields, including network security rules
		// ... and environment variable setup
	}

	// Send request to Oracle Cloud API to launch an instance
	launchResponse, err := computeClient.LaunchInstance(context.Background(), launchRequest)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to launch instance: %v", err)})
		return
	}

	// Extract relevant information about the launched instance
	instanceInfo := map[string]interface{}{
		"InstanceId":     *launchResponse.Id,
		"InstanceType":   "your-instance-type", // Replace with your instance type
		"InstanceDetails": "your-instance-details", // Replace with your instance details
		"LaunchedTime":   time.Now().Format(time.RFC3339),
		"Status":         "launched",
		"Port":           port,
	}

	// Store instance information in MongoDB
	collection := client.Database("your-database-name").Collection("your-collection-name") // Replace with your database and collection names
	_, err = collection.InsertOne(context.Background(), instanceInfo)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to insert into MongoDB: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": "Instance launched successfully", "instanceInfo": instanceInfo})
}


func terminateComputeInstance(c *gin.Context) {
	configPath := "/workspaces/Cloud-Infrastructure-Management-Tool/compute-service/config.toml" // Replace with your config path
	configProvider, err := common.ConfigurationProviderFromFile(configPath, "DEFAULT")
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to load config: %v", err)})
		return
	}

	// Create Compute client
	computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create Compute client: %v", err)})
		return
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to connect to MongoDB: %v", err)})
		return
	}
	defer client.Disconnect(context.Background())

	// Extract instance ID from the request parameters
	instanceID := c.Param("instanceID")

	// Terminate the compute instance
	terminateRequest := core.TerminateInstanceRequest{
		InstanceId: common.String(instanceID),
	}

	_, err = computeClient.TerminateInstance(context.Background(), terminateRequest)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to terminate instance: %v", err)})
		return
	}

	// Mark the status as terminated and calculate cost
	status := "terminated"
	// cost := calculateCost(instanceID) // You need to implement the calculateCost function

	// Update the instance information in MongoDB
	collection := client.Database("your-database-name").Collection("your-collection-name") // Replace with your database and collection names
	update := bson.M{"$set": bson.M{"status": status, "cost": cost}}
	_, err = collection.UpdateOne(context.Background(), bson.M{"instanceId": instanceID}, update)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update MongoDB: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": "Instance terminated successfully", "status": status, "cost": cost})
}

// Calculate cost function (implement your logic here)


// Other compute-related functions
