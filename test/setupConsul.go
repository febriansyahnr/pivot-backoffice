package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupConsul(ctx context.Context) (testcontainers.Container, string, error) {

	req := testcontainers.ContainerRequest{
		Image:        "hashicorp/consul:latest",
		ExposedPorts: []string{"8500/tcp"},
		WaitingFor:   wait.ForLog("Synced node info"),
	}

	consulContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	host, err := consulContainer.Host(ctx)
	if err != nil {
		return nil, "", err
	}

	port, err := consulContainer.MappedPort(ctx, "8500")
	if err != nil {
		return nil, "", err
	}

	consulURL := fmt.Sprintf("http://%s:%s", host, port.Port())

	return consulContainer, consulURL, nil
}

func addKeyValueToConsul(client *api.Client, key, value string) error {
	kv := client.KV()
	p := &api.KVPair{Key: key, Value: []byte(value)}
	_, err := kv.Put(p, nil)

	return err
}

// findProjectRoot safely finds the project root directory
// with additional validation to prevent directory traversal
func findProjectRoot(startPath, projectName string) (string, error) {
	// Prevent abnormally long path traversal
	maxIterations := 10
	iterations := 0

	dir := startPath
	for {
		iterations++
		if iterations > maxIterations {
			return "", fmt.Errorf("exceeded maximum directory traversal depth")
		}

		if strings.HasSuffix(dir, projectName) {
			return dir, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", fmt.Errorf("project root not found")
		}
		dir = parentDir
	}
}

// validateFilename checks if the filename contains only alphanumeric characters,
// underscores, and hyphens to prevent path traversal attacks
func validateFilename(filename string) bool {
	// Only allow alphanumeric characters, underscores, and hyphens
	validFilename := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(filename)
	return validFilename
}

// isPathSafe verifies that a constructed path doesn't escape its expected parent directory
func isPathSafe(path, expectedParent string) bool {
	// Normalize both paths to handle any '..' or '.' components
	normalizedPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	normalizedParent, err := filepath.Abs(expectedParent)
	if err != nil {
		return false
	}

	// Check if the path is within the expected parent directory
	return strings.HasPrefix(normalizedPath, normalizedParent)
}

func SetupFeatureFlag(consulURL string) error {
	// Setup Feature Flag Yaml
	config := api.DefaultConfig()
	config.Address = consulURL

	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	projectName := "backend-portal"
	projectRoot, err := findProjectRoot(currentDir, projectName)
	if err != nil {
		fmt.Printf("Error finding project root: %v\n", err)
		return err
	}

	// Construct the path with a fixed filename
	consulDir := filepath.Join(projectRoot, "test", "consul", "backend-portal")
	filename := "feature-flag.yaml"
	targetPath := filepath.Join(consulDir, filename)

	// Ensure the path is within the expected directory structure
	if !isPathSafe(targetPath, consulDir) {
		return fmt.Errorf("unsafe file path detected")
	}

	// Check if the file exists before reading it
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filename)
	}

	// Path has been validated to prevent directory traversal
	validatedPath := targetPath
	yamlContent, err := os.ReadFile(validatedPath) // #nosec G304 - Path has been validated by isPathSafe function to prevent directory traversal attacks
	if err != nil {
		fmt.Printf("Error reading yaml file: %v\n", err)
		return err
	}

	// Add key-value pair to Consul
	err = addKeyValueToConsul(client, "backend-portal/feature-flag", string(yamlContent))
	if err != nil {
		fmt.Printf("Error adding key-value pair to Consul: %v\n", err)
		return err
	}

	// Verify the key-value pair in Consul
	kv := client.KV()
	pair, _, err := kv.Get("backend-portal/feature-flag", nil)
	if err != nil {
		fmt.Printf("Error getting key-value pair from Consul: %v\n", err)
		return err
	} else if pair == nil {
		fmt.Println("Key-value pair not found in Consul")
		return fmt.Errorf("key-value pair not found in Consul")
	} else if string(yamlContent) != string(pair.Value) {
		fmt.Println("Key-value pair value mismatch")
		return fmt.Errorf("key-value pair value mismatch")
	}

	return nil
}

func SetupConsulValue(consulURL string, filename string) error {
	// Validate filename to prevent path traversal attacks
	if !validateFilename(filename) {
		return fmt.Errorf("invalid filename format: %s", filename)
	}

	// Setup Feature Flag Yaml
	config := api.DefaultConfig()
	config.Address = consulURL

	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	projectName := "backend-portal"
	projectRoot, err := findProjectRoot(currentDir, projectName)
	if err != nil {
		fmt.Printf("Error finding project root: %v\n", err)
		return err
	}

	// Define the expected parent directory
	expectedDir := filepath.Join(projectRoot, "test", "consul", "backend-portal")

	// Construct the filepath with validated filename
	targetPath := filepath.Join(expectedDir, filename+".yaml")

	// Ensure the final path is within the expected directory using the new utility function
	if !isPathSafe(targetPath, expectedDir) {
		return fmt.Errorf("file path traversal detected")
	}

	// Check if the file exists before reading it
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filename+".yaml")
	}

	// Path has been validated to prevent directory traversal
	validatedPath := targetPath
	yamlContent, err := os.ReadFile(validatedPath) // #nosec G304 - Path has been validated by validateFilename function to prevent directory traversal attacks
	if err != nil {
		fmt.Printf("Error reading yaml file: %v\n", err)
		return err
	}

	// Add key-value pair to Consul
	err = addKeyValueToConsul(client, "backend-portal/"+filename, string(yamlContent))
	if err != nil {
		fmt.Printf("Error adding key-value pair to Consul: %v\n", err)
		return err
	}

	// Verify the key-value pair in Consul
	kv := client.KV()
	pair, _, err := kv.Get("backend-portal/"+filename, nil)
	if err != nil {
		fmt.Printf("Error getting key-value pair from Consul: %v\n", err)
		return err
	} else if pair == nil {
		fmt.Println("Key-value pair not found in Consul")
		return fmt.Errorf("key-value pair not found in Consul")
	} else if string(yamlContent) != string(pair.Value) {
		fmt.Println("Key-value pair value mismatch")
		return fmt.Errorf("key-value pair value mismatch")
	}

	return nil
}
