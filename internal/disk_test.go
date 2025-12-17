package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiskFreeSpaceInBytes(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "valid path - current directory",
			path:      ".",
			wantError: false,
		},
		{
			name:      "valid path - temp directory",
			path:      os.TempDir(),
			wantError: false,
		},
		{
			name:      "invalid path",
			path:      "/this/path/definitely/does/not/exist/anywhere",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DiskFreeSpaceInBytes(tt.path)

			if tt.wantError {
				if err == nil {
					t.Errorf("DiskFreeSpaceInBytes() expected error for path %s, got nil", tt.path)
				}
				if got != nil {
					t.Errorf("DiskFreeSpaceInBytes() expected nil result on error, got %v", *got)
				}
			} else {
				if err != nil {
					t.Errorf("DiskFreeSpaceInBytes() unexpected error: %v", err)
				}
				if got == nil {
					t.Error("DiskFreeSpaceInBytes() returned nil, expected value")
				} else {
					// Verify the returned value is reasonable (> 0 bytes)
					if *got == 0 {
						t.Error("DiskFreeSpaceInBytes() returned 0 bytes, expected positive value")
					}
					// Log the value for debugging purposes
					t.Logf("Free space for %s: %d bytes", tt.path, *got)
				}
			}
		})
	}
}

func TestDiskFreeSpaceInBytes_WithRealDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	freeSpace, err := DiskFreeSpaceInBytes(tempDir)

	if err != nil {
		t.Fatalf("DiskFreeSpaceInBytes() error = %v, want nil", err)
	}

	if freeSpace == nil {
		t.Fatal("DiskFreeSpaceInBytes() returned nil, expected value")
	}

	if *freeSpace == 0 {
		t.Error("DiskFreeSpaceInBytes() returned 0, expected positive value")
	}

	t.Logf("Temp directory %s has %d bytes free", tempDir, *freeSpace)
}

func TestDiskFreeSpaceInBytes_ConsistentResults(t *testing.T) {
	path := "."

	// Call the function twice in succession
	result1, err1 := DiskFreeSpaceInBytes(path)
	result2, err2 := DiskFreeSpaceInBytes(path)

	if err1 != nil || err2 != nil {
		t.Fatalf("DiskFreeSpaceInBytes() errors: err1=%v, err2=%v", err1, err2)
	}

	if result1 == nil || result2 == nil {
		t.Fatal("DiskFreeSpaceInBytes() returned nil")
	}

	// Results should be very close (within 10% - allows for concurrent disk activity)
	diff := int64(*result1) - int64(*result2)
	if diff < 0 {
		diff = -diff
	}

	maxDiff := int64(float64(*result1) * 0.1)
	if diff > maxDiff {
		t.Errorf("DiskFreeSpaceInBytes() results differ too much: %d vs %d (diff: %d)", *result1, *result2, diff)
	}
}

func TestAppConfig_FreeSpaceOfDiskPathInBytes(t *testing.T) {
	tests := []struct {
		name      string
		diskPath  string
		wantError bool
	}{
		{
			name:      "valid path",
			diskPath:  ".",
			wantError: false,
		},
		{
			name:      "temp directory",
			diskPath:  os.TempDir(),
			wantError: false,
		},
		{
			name:      "invalid path",
			diskPath:  "/nonexistent/path/that/does/not/exist",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := AppConfig{DiskPath: tt.diskPath}
			got, err := config.FreeSpaceOfDiskPathInBytes()

			if tt.wantError {
				if err == nil {
					t.Errorf("FreeSpaceOfDiskPathInBytes() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("FreeSpaceOfDiskPathInBytes() error = %v, want nil", err)
				}
				if got == nil {
					t.Error("FreeSpaceOfDiskPathInBytes() returned nil, expected value")
				} else if *got == 0 {
					t.Error("FreeSpaceOfDiskPathInBytes() returned 0, expected positive value")
				}
			}
		})
	}
}

func TestDiskFreeSpaceInBytes_WithFile(t *testing.T) {
	// Create a temporary file
	tempFile := filepath.Join(t.TempDir(), "testfile.txt")
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// syscall.Statfs should work on files too (returns filesystem stats)
	freeSpace, err := DiskFreeSpaceInBytes(tempFile)

	if err != nil {
		t.Errorf("DiskFreeSpaceInBytes() error = %v, want nil", err)
	}

	if freeSpace == nil {
		t.Error("DiskFreeSpaceInBytes() returned nil for file path")
	} else if *freeSpace == 0 {
		t.Error("DiskFreeSpaceInBytes() returned 0 for file path")
	}
}

func BenchmarkDiskFreeSpaceInBytes(b *testing.B) {
	path := "."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DiskFreeSpaceInBytes(path)
		if err != nil {
			b.Fatalf("DiskFreeSpaceInBytes() error = %v", err)
		}
	}
}
