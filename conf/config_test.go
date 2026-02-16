package conf

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLoadConfig_TableDriven(t *testing.T) {
	dir := t.TempDir()

	// Prepare test YAML files
	successFile := filepath.Join(dir, "success.yaml")
	successYAML := `
database:
  path: test.db
  migrations:
    run_on_startup: true
`
	if err := os.WriteFile(successFile, []byte(successYAML), 0644); err != nil {
		t.Fatal(err)
	}

	badYAMLFile := filepath.Join(dir, "bad.yaml")
	badYAML := `database: [invalid yaml`
	if err := os.WriteFile(badYAMLFile, []byte(badYAML), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		filePath   string
		expectErr  bool
		expectPath string
		expectRun  bool
	}{
		{
			name:       "success path",
			filePath:   successFile,
			expectErr:  false,
			expectPath: "test.db",
			expectRun:  true,
		},
		{
			name:      "file not found",
			filePath:  "non-existent.yaml",
			expectErr: true,
		},
		{
			name:      "invalid yaml",
			filePath:  badYAMLFile,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectErr {
				// Use subprocess to safely test log.Fatalf
				if os.Getenv("BE_CRASHER") == "1" {
					LoadConfig(tt.filePath)
					return
				}

				cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
				cmd.Env = append(os.Environ(), "BE_CRASHER=1")
				err := cmd.Run()
				if err == nil {
					t.Fatal("expected process to exit with error")
				}
			} else {
				// normal success path
				cfg := LoadConfig(tt.filePath)
				if cfg.Database.Path != tt.expectPath {
					t.Fatalf("expected path %s, got %s", tt.expectPath, cfg.Database.Path)
				}
				if cfg.Database.Migrations.RunOnStartup != tt.expectRun {
					t.Fatalf("expected RunOnStartup %v, got %v", tt.expectRun, cfg.Database.Migrations.RunOnStartup)
				}
			}
		})
	}
}
