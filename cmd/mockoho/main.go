package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/mock"
	"github.com/mockoho/mockoho/internal/proxy"
	"github.com/mockoho/mockoho/internal/server"
	"github.com/mockoho/mockoho/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Version is the version of the application
	Version = "0.1.0"
	
	// ConfigDir is the directory containing mock configurations
	ConfigDir string
)

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:     "mockoho",
		Short:   "Mockoho - A mock server system",
		Version: Version,
		Run:     runUI,
	}
	
	// Add flags
	rootCmd.PersistentFlags().StringVarP(&ConfigDir, "config", "c", "mocks", "Directory containing mock configurations")
	
	// Add subcommands
	rootCmd.AddCommand(serverCmd())
	
	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// runUI runs the UI
func runUI(cmd *cobra.Command, args []string) {
	// Ensure config directory exists
	if err := ensureConfigDir(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	// Create config
	cfg := config.New(ConfigDir)
	if err := cfg.Load(); err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	
	// Create mock manager
	mockManager := mock.New(cfg)
	
	// Create proxy manager
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		fmt.Printf("Error creating proxy manager: %v\n", err)
		os.Exit(1)
	}
	
	// Create server
	srv := server.New(cfg, mockManager, proxyManager)
	
	// Create UI model
	model := ui.New(cfg, mockManager, proxyManager, srv)
	
	// Run UI with additional options for better terminal handling
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
		os.Exit(1)
	}
}

// serverCmd returns the server subcommand
func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the mock server without the UI",
		Run:   runServer,
	}
	
	return cmd
}

// runServer runs the server without the UI
func runServer(cmd *cobra.Command, args []string) {
	// Ensure config directory exists
	if err := ensureConfigDir(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	// Create config
	cfg := config.New(ConfigDir)
	if err := cfg.Load(); err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	
	// Create mock manager
	mockManager := mock.New(cfg)
	
	// Create proxy manager
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		fmt.Printf("Error creating proxy manager: %v\n", err)
		os.Exit(1)
	}
	
	// Create server
	srv := server.New(cfg, mockManager, proxyManager)
	
	// Start server
	if err := srv.Start(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Server started at %s\n", srv.GetAddress())
	fmt.Println("Press Ctrl+C to stop")
	
	// Wait for interrupt
	<-make(chan struct{})
}

// ensureConfigDir ensures the config directory exists
func ensureConfigDir() error {
	// Get absolute path
	absPath, err := filepath.Abs(ConfigDir)
	if err != nil {
		return err
	}
	
	// Update ConfigDir to absolute path
	ConfigDir = absPath
	
	// Check if directory exists
	info, err := os.Stat(ConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory
			if err := os.MkdirAll(ConfigDir, 0755); err != nil {
				return err
			}
			
			// Create default config files
			if err := createDefaultConfigs(); err != nil {
				return err
			}
			
			return nil
		}
		
		return err
	}
	
	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", ConfigDir)
	}
	
	return nil
}

// createDefaultConfigs creates default configuration files
func createDefaultConfigs() error {
	// Create config.json
	configPath := filepath.Join(ConfigDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configContent := `{
  "proxyConfig": {
    "target": "https://api.real-server.com",
    "changeOrigin": true,
    "pathRewrite": {
      "^/api": ""
    }
  },
  "serverConfig": {
    "port": 3000,
    "host": "localhost"
  },
  "editor": {
    "command": "code",
    "args": ["-g", "{file}:{line}"]
  }
}`
		
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			return err
		}
	}
	
	// Create example.json
	examplePath := filepath.Join(ConfigDir, "example.json")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		exampleContent := `{
  "feature": "example",
  "endpoints": [
    {
      "id": "hello-world",
      "method": "GET",
      "path": "/api/hello",
      "active": true,
      "defaultResponse": "standard",
      "responses": {
        "standard": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "message": "Hello, World!",
            "timestamp": "{{now}}"
          },
          "delay": 0
        }
      }
    }
  ]
}`
		
		if err := os.WriteFile(examplePath, []byte(exampleContent), 0644); err != nil {
			return err
		}
	}
	
	return nil
}