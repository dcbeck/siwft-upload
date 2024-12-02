package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/ncw/swift/v2"
)

var (
	// Use sync.Once to ensure .env is loaded only once
	envOnce sync.Once

	// Create a reusable Swift connection pool
	swiftConnPool = &sync.Pool{
		New: func() interface{} {

			return &swift.Connection{
				UserName:    os.Getenv("SWIFT_USERNAME"),
				ApiKey:      os.Getenv("SWIFT_PASSWORD"),
				AuthVersion: authVersion(),
				AuthUrl:     os.Getenv("SWIFT_AUTH_URL"),
				Domain:      os.Getenv("SWIFT_DOMAIN"),
				Retries:     3,
			}
		},
	}
	containerName = "test"    // Consider making this configurable
	bufferSize    = 32 * 1024 // 32KB buffer
	// Pre-compile logger for better performance
	logger = log.New(os.Stdout, "swift-server: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func init() {
	// Ensure .env is loaded safely
	envOnce.Do(func() {
		if err := godotenv.Load(); err != nil {
			logger.Println("Error loading .env file:", err)
		}
	})
}

func removeFirstSlash(s string) string {
	// Check if the string starts with a slash and if so, remove it
	if strings.HasPrefix(s, "/") {
		return s[1:]
	}
	return s
}

func authVersion() int {
	version, err := strconv.Atoi(os.Getenv("SWIFT_AUTH_VERSION"))
	if err != nil {
		return 3
	}
	return version
}

func main() {
	http.HandleFunc("/", fileHandler)

	const port = ":8080"
	logger.Printf("Server started on port %s\n", port[1:])

	server := &http.Server{
		Addr:              port,
		ReadHeaderTimeout: 3 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB max header size
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("ListenAndServe:", err)
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	// Early return for unsupported methods
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	objectName := strings.TrimSpace(removeFirstSlash(r.URL.String()))

	logger.Println(objectName)
	// Validate object name
	if objectName == "" || objectName == "/" {
		http.Error(w, "Invalid file name", http.StatusBadRequest)
		return
	}

	// Create context with timeout and cancellation
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get a connection from the pool
	conn := swiftConnPool.Get().(*swift.Connection)
	defer swiftConnPool.Put(conn)

	// Authenticate with a timeout
	if err := conn.Authenticate(ctx); err != nil {
		logger.Println("Failed to authenticate with Swift:", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Open object stream
	objectStream, _, err := conn.ObjectOpen(ctx, containerName, objectName, false, nil)
	if err != nil {
		logger.Println("Failed to open object:", err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer objectStream.Close()

	// Set efficient response headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", objectName))

	// Use efficient streaming with buffer

	if _, err = io.CopyBuffer(w, objectStream, make([]byte, bufferSize)); err != nil {
		logger.Println("Error streaming file:", err)
		http.Error(w, "Error streaming file", http.StatusInternalServerError)
	}
}
