# Cloud Infrastructure Management Tool

This tool provides a microservice with a REST API for integrating compute resources from Oracle Cloud Infrastructure. It's implemented in Go using the Gin framework and interacts with MongoDB to store and retrieve instance details.


# Project Name

This project provides a Go-based solution for managing compute services. It is organized into different packages for clarity and modularity.

## Project Structure


- `compute-services`: Main directory for the compute services.
    - `cmd`: Command-line application entry points.
        - `main.go`: Entry point for the application.
- `pkg`: Package directory containing shared code.
    - `compute`: Package related to compute services.
        - `compute.go`: General compute-related functions.
        - `instance.go`: Functions related to compute instances.
- `go.mod` and `go.sum`: Go module files.


## Config for oci sdk
 - you can make the a config.toml file as shown below
```bash 
    [DEFAULT]
    user=ocid1.user.oc1..<unique_ID>
    fingerprint=<your_fingerprint>
    key_file=~/.oci/oci_api_key.pem
    tenancy=ocid1.tenancy.oc1..<unique_ID>
    region=us-ashburn-1
    compartment_id ="your-compartment-id"
```
 - OR
    you can install OCI CLI and configure it 
    more inforamtion on this [link](https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliconfigure.htm)


## Usage

1. **Run the Application:**

    ```bash
    go run compute-services/cmd/main.go
    ```

    This will start the compute services application.

## Packages

### `compute`

#### `compute.go`

This file contains general functions related to compute services.

#### `instance.go`

This file contains functions specifically related to compute instances.

# HTTP Routes

## Route: `/compute-instances`

**Description:** Retrieves the list of all compute instances available.

**Example URL:**

```bash
    GET /compute-instances
```

## Route: `/instance-details/:instanceType`
**Description:** It will  give you all the detials like  CPU type, GPU type, GPU count, memory, storage and pricing per hour. Store them in a
MongoDB database if they don’t exist, else update it

**Example URL:**

```bash
    GET /instance-details/:x3Large
```

## Route: `/start-instance/:instanceType`
**Description:** It will start the instace and it will require parmeters like 
* instanceType: The type of the compute instance.
* comaprtmentId: The ID of the compartment.
* ad: The availability domain.
* image_id: The ID of the image.

**Example URL:**

```bash
    POST /start-instance/A1-med?comaprtmentId=yourCompartmentID&ad=yourAvailabilityDomain&image_id=yourImageID

```

## Route: `/terminate-instance/:instanceID`

**Description:**  terminate the instance and mark the status as terminated in the
MongDB. Calculate the cost incurred and store it in the same collection as above in
MongoDB

**Example URL:**

```bash
    POST /terminate-instance/yourInstanceID

```


## Contributing

Feel free to contribute by opening issues or submitting pull requests. Follow the [Contributing Guidelines](CONTRIBUTING.md) for more details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
