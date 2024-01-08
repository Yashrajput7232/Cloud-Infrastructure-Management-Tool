package compute

import (
	"context"
	"fmt"
	"time"
	"net/http"
    "math/rand"
	"io/ioutil"
	"github.com/gin-gonic/gin"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"log"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB connection details
var mongoURI = "mongodb+srv://username:passwordcluster0.kn66rfv.mongodb.net/?retryWrites=true&w=majority"
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
	LaunchedTime time.Time
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
	

// Implement getInstanceDetails to retrieve instance details from OCI
func getInstanceDetails(c *gin.Context) {
    instanceType := c.Query("instanceType")
	configProvider:=getConfigurationProvider()
	compute, err := core.NewComputeClientWithConfigurationProvider(configProvider)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create Compute client: %v", err)})
			return
		}

    // Retrieve instance details from OCI Compute
    shape, err := compute.GetShape(context.Background(), compute.GetShapeInput{
        ShapeName: &instanceType,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instance details"})
        return
    }

    // Extract relevant details
    cpuType := shape.ProcessorDescription
    gpuType := shape.GpuDescription
    gpuCount := shape.GpuCount
    memory := shape.MemoryInGBs
    storage := shape.BaseShape.StorageInGBs

    // Retrieve pricing per hour
    pricing, err := getPricing(compute,instanceType)
    if err != nil {
        log.Printf("Error retrieving pricing: %v", err)
        pricing = 0.0 // Set a placeholder in case of error
    }

    // Create InstanceDetails object
    details := InstanceDetails{
        InstanceType: instanceType,
        CPUType:      cpuType,
        GPUType:      gpuType,
        GPUCount:     gpuCount,
        Memory:       memory,
        Storage:      storage,
        Pricing:      pricing,
    }

    // Store details in MongoDB
    err = storeInstanceDetails(details)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store instance details in MongoDB"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Instance details retrieved and stored successfully"})
}

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
	compartmentID:=c.Query("comaprtmentId")
	availabilityDomain:=c.Query("ad")
	imageID:=c.Query("image_id")
	configPath := "/workspaces/Cloud-Infrastructure-Management-Tool/compute-service/config.toml" // Replace with your config path
	configProvider, err := common.ConfigurationProviderFromFile(configPath, "DEFAULT")
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to load config: %v", err)})
		return
	}
	
	port := 9000 + rand.Intn(101)

	request := core.LaunchInstanceRequest{
		LaunchInstanceDetails: core.LaunchInstanceDetails{
			CompartmentId:       &compartmentID,
			AvailabilityDomain:  &availabilityDomain,
			ImageId:             &imageID,
			Shape:               &instanceShape,
			// Add more parameters as needed
		},
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

	
	// Send request to Oracle Cloud API to launch an instance
	launchResponse, err := computeClient.LaunchInstance(context.Background(), request)
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
	collection := client.Database(dbName).Collection(collectionName) // Replace with your database and collection names
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
	cost := calculateCost(instanceID) // You need to implement the calculateCost function

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
func calculateCost(instanceID string) float64 {
    // Connect to MongoDB
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
    if err != nil {
        fmt.Println("Error connecting to MongoDB:", err)
        return 0 // Return a placeholder cost in case of error
    }
    defer client.Disconnect(context.Background())

    // Fetch instance details from MongoDB
    collection := client.Database(dbName).Collection(collectionName)
    filter := bson.M{"instanceId": instanceID}
    var instanceDetails InstanceDetails
    err = collection.FindOne(context.Background(), filter).Decode(&instanceDetails)
    if err != nil {
        fmt.Println("Error fetching instance details from MongoDB:", err)
        return 0 // Return a placeholder cost in case of error
    }

    // Calculate cost based on pricing and running time
    // runningTime := time.Now().Sub(time.Parse(time.RFC3339, instanceDetails.LaunchedTime))
    // launchedTime, err := time.Parse(time.RFC3339, instanceDetails.LaunchedTime)
	launchedTime := instanceDetails.LaunchedTime // directly use the time.Time value
	runningTime := time.Now().Sub(launchedTime)
	cost := instanceDetails.Pricing * runningTime.Hours()
	if err != nil {
		fmt.Println("Error parsing launched time:", err)
		// Handle the error appropriately, e.g., return a placeholder value or abort calculation
}
    return cost
}


func getPricing(computeClient core.ComputeClient, instanceType string) (float64, error) {
    // Check if pricing is already stored in MongoDB (optional optimization)
  

    // Make API call to fetch pricing
    client := http.Client{} // Create an HTTP client
    req, err := http.NewRequest("GET", "https://apexapps.oracle.com/pls/apex/cetools/api/v1/products/", nil)
    if err != nil {
        return 0, fmt.Errorf("failed to create API request: %w", err)
    }
    q := req.URL.Query()
    q.Add("partNumber", instanceType) // Filter by instance type
    req.URL.RawQuery = q.Encode()

    resp, err := client.Do(req)
    if err != nil {
        return 0, fmt.Errorf("failed to fetch pricing from API: %w", err)
    }
    defer resp.Body.Close()

    // Parse JSON response
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return 0, fmt.Errorf("failed to read API response body: %w", err)
    }
    var pricingData []map[string]interface{}
    err = json.Unmarshal(body, &pricingData)
    if err != nil {
        return 0, fmt.Errorf("failed to unmarshal API response: %w", err)
    }

    // Extract hourly price
    hourlyPrice := 0.0
    for _, item := range pricingData {
        if item["partNumber"] == instanceType {
            hourlyPrice = item["prices"].([]interface{})[0].(map[string]interface{})["value"].(float64)
            break
        }
    }

    if hourlyPrice == 0 {
        return 0, fmt.Errorf("pricing not found for instance type %s", instanceType)
    }

    return hourlyPrice, nil
}

