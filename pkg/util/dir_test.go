package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureTmpDir(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test_ensure_dir")

	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Fatalf("Test directory should not exist before testing: %s", testDir)
	}

	err := EnsureTmpDir(testDir)
	if err != nil {
		t.Fatalf("Failed to ensure directory exists: %v", err)
	}

	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Fatalf("Directory was not created: %s", testDir)
	}

	defer os.RemoveAll(testDir)

	err = EnsureTmpDir(testDir)
	if err != nil {
		t.Fatalf("Failed when ensuring an already existing directory: %v", err)
	}
}
