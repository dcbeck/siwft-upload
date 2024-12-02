package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/ncw/swift/v2"
)

func authVersion() int {
	version, err := strconv.Atoi(os.Getenv("SWIFT_AUTH_VERSION"))
	if err != nil {
		return 3
	}
	return version
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
	}
	ctx := context.Background()
	// Create a connection to OpenStack Swift
	connection := &swift.Connection{
		UserName:    os.Getenv("SWIFT_USERNAME"),
		ApiKey:      os.Getenv("SWIFT_PASSWORD"),
		AuthVersion: authVersion(),
		AuthUrl:     os.Getenv("SWIFT_AUTH_URL"),
		Domain:      os.Getenv("SWIFT_DOMAIN"),
		Retries:     3,
	}

	// Authenticate
	err := connection.Authenticate(ctx)
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
	_, err = connection.ObjectPut(ctx, containerName, objectName, file, false, "", "", nil)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	fmt.Printf("Successfully uploaded %s to Swift container %s\n", objectName, containerName)
}
