package upload

import (
	"fmt"
	"log"
	"os"

	"github.com/ncw/swift/v2"
)

func main() {
	// Create a connection to OpenStack Swift
	connection := &swift.Connection{
		UserName:     os.Getenv("OS_USERNAME"),
		ApiKey:       os.Getenv("OS_PASSWORD"),
		AuthUrl:      os.Getenv("OS_AUTH_URL"),
		Domain:       os.Getenv("OS_DOMAIN_NAME"), // For Keystone V3
		Tenant:       os.Getenv("OS_TENANT_NAME"), // Project name
		Region:       os.Getenv("OS_REGION_NAME"),
		TenantId:     os.Getenv("OS_PROJECT_ID"),     // Optional: Project ID
		TenantDomain: os.Getenv("OS_PROJECT_DOMAIN"), // Optional: Project domain
	}

	// Authenticate
	err := connection.Authenticate()
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Local file to upload
	localFilePath := os.Getenv("LOCAL_FILE_PATH")
	containerName := os.Getenv("SWIFT_CONTAINER")
	objectName := os.Getenv("SWIFT_OBJECT_NAME")

	// Open local file
	file, err := os.Open(localFilePath)
	if err != nil {
		log.Fatalf("Failed to open local file: %v", err)
	}
	defer file.Close()

	// Upload file to Swift
	_, err = connection.ObjectPut(containerName, objectName, file, false, "", "", nil)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	fmt.Printf("Successfully uploaded %s to Swift container %s\n", objectName, containerName)
}
