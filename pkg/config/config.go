package config

import (
	"bufio"
	"fmt"
	"h2blog_server/pkg/fileTool"
	"h2blog_server/pkg/utils"
	"net"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// ProjectConfig defines the structure for all configuration data
type ProjectConfig struct {
	User   UserConfigData   `yaml:"user"`   // User configuration
	Server ServerConfigData `yaml:"server"` // Server configuration
	Logger LoggerConfigData `yaml:"logger"` // Logger configuration
	MySQL  MySQLConfigData  `yaml:"mysql"`  // MySQL database configuration
	Oss    OssConfig        `yaml:"oss"`    // OSS configuration
	Cache  CacheConfig      `yaml:"cache"`  // Cache configuration
}

// UserConfigData defines user-specific configuration
type UserConfigData struct {
	Username        string `yaml:"username"`         // User's username
	Email           string `yaml:"email"`            // User's email address
	BackgroundImage string `yaml:"background_image"` // Background image name
}

// WebPConfigData defines WebP image conversion settings
type WebPConfigData struct {
	Enable  bool    `yaml:"enable"`  // Whether to enable WebP conversion
	Quality float32 `yaml:"quality"` // WebP image quality (1-100)
	Size    float32 `yaml:"size"`    // Maximum WebP image size in MB
}

// ServerConfigData defines server-related configuration
type ServerConfigData struct {
	Port                uint16         `yaml:"port"`                  // Server port number
	TokenKey            string         `yaml:"token_key"`             // JWT signing and verification key
	TokenExpireDuration uint8          `yaml:"token_expire_duration"` // Token expiration in days
	Cors                CorsConfigData `yaml:"cors"`                  // CORS configuration
}

// CorsConfigData defines Cross-Origin Resource Sharing configuration
type CorsConfigData struct {
	Origins []string `yaml:"origins"` // Allowed origins
	Headers []string `yaml:"headers"` // Allowed headers
	Methods []string `yaml:"methods"` // Allowed methods
}

// LoggerConfigData defines logging configuration
type LoggerConfigData struct {
	Level      string `yaml:"level"`       // Logging level
	Path       string `yaml:"path"`        // Log file path
	MaxAge     uint16 `yaml:"max_age"`     // Maximum days to retain log files
	MaxSize    uint16 `yaml:"max_size"`    // Maximum size of log files in MB
	MaxBackups uint16 `yaml:"max_backups"` // Maximum number of log backups
	Compress   bool   `yaml:"compress"`    // Whether to compress log files
}

// MySQLConfigData defines MySQL database configuration
type MySQLConfigData struct {
	User     string `yaml:"user"`     // Database username
	Password string `yaml:"password"` // Database password
	Host     string `yaml:"host"`     // Database host address
	Port     uint16 `yaml:"port"`     // Database port number
	DB       string `yaml:"database"` // Database name
	MaxOpen  uint16 `yaml:"max_open"` // Maximum open connections
	MaxIdle  uint16 `yaml:"max_idle"` // Maximum idle connections
}

// OssConfig defines Object Storage Service configuration
type OssConfig struct {
	Endpoint        string         `yaml:"endpoint"`
	Region          string         `yaml:"region"`
	AccessKeyId     string         `yaml:"access_key_id"`
	AccessKeySecret string         `yaml:"access_key_secret"`
	Bucket          string         `yaml:"bucket"`
	ImageOssPath    string         `yaml:"image_oss_path"` // Path for storing images
	BlogOssPath     string         `yaml:"blog_oss_path"`  // Path for storing blog content
	WebP            WebPConfigData `yaml:"webp"`           // WebP image configuration
}

// CacheConfig defines cache system configuration
type CacheConfig struct {
	Aof AofConfig `yaml:"aof"` // AOF persistence configuration
}

// AofConfig defines Append-Only File persistence configuration
type AofConfig struct {
	Enable   bool   `yaml:"enable"`   // Whether to enable AOF persistence
	Path     string `yaml:"path"`     // AOF file path
	MaxSize  uint16 `yaml:"max_size"` // Maximum AOF file size in MB
	Compress bool   `yaml:"compress"` // Whether to compress AOF files
}

// Global configuration variables
var (
	// loadConfigLock ensures configuration is loaded only once
	loadConfigLock sync.Once

	// User holds global user configuration
	User *UserConfigData = nil

	// Server holds global server configuration
	Server *ServerConfigData = nil

	// Logger holds global logger configuration
	Logger *LoggerConfigData = nil

	// MySQL holds global MySQL database configuration
	MySQL *MySQLConfigData = nil

	// Oss holds global OSS configuration
	Oss *OssConfig = nil

	// Cache holds global cache configuration
	Cache *CacheConfig = nil
)

// LoadConfig 加载配置文件
func LoadConfig() error {
	userHomePath, err := getH2BlogDir()
	if err != nil {
		panic(err)
	}

	if !fileTool.IsExist(path.Join(userHomePath, "config", "h2blog_config.yaml")) {
		return NewNoConfigFileErr("配置文件不存在")
	}

	loadConfigLock.Do(
		func() {
			err = loadConfigFromFile()
			if err != nil {
				err = loadConfigFromTerminal()
			}

			// If both loading from file and terminal failed, then panic
			if err != nil {
				panic(err)
			}
		},
	)

	return nil
}

// getH2BlogDir returns the base directory for h2blog configuration and data
// It uses the user's home directory as the base path
//
// Returns:
//   - string: The full path to the h2blog directory
//   - error: Any error encountered while getting the user's home directory
func getH2BlogDir() (string, error) {
	// Get current user information
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	// Join home directory with .h2blog
	return path.Join(usr.HomeDir, ".h2blog"), nil
}

// loadConfigFromFile Loading config data from file
func loadConfigFromFile() error {
	// Try user's home directory first
	h2blogDir, err := getH2BlogDir()
	if err != nil {
		return fmt.Errorf("failed to get h2blog directory: %w", err)
	}

	// Try home directory first
	configPath := path.Join(h2blogDir, "config", "h2blog_config.yaml")
	if fileTool.IsExist(configPath) {
		return loadConfigFromPath(configPath)
	}

	// If not found in home directory, try current directory
	currentDirPath := path.Join(".h2blog", "config", "h2blog_config.yaml")
	if fileTool.IsExist(currentDirPath) {
		return loadConfigFromPath(currentDirPath)
	}

	return fmt.Errorf("config file not found")
}

// loadConfigFromPath loads config from specified path
func loadConfigFromPath(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("load config file error: %w", err)
	}

	conf := &ProjectConfig{}
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return fmt.Errorf("reflect config to struct error: %w", err)
	}

	// Set global config
	User = &conf.User
	Server = &conf.Server
	Logger = &conf.Logger
	MySQL = &conf.MySQL
	Oss = &conf.Oss
	Cache = &conf.Cache

	return nil
}

// checkPortAvailable attempts to bind to a port to check its availability
// It returns an error if the port is already in use
// Parameters:
//   - port: The port number to check
//
// Returns:
//   - error: nil if port is available, error message if port is in use
func checkPortAvailable(port uint16) error {
	// Try to listen on the port
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is not available", port)
	}
	// Close the listener immediately to free the port
	err = listener.Close()
	if err != nil {
		return err
	}
	return nil
}

// loadConfigFromTerminal loads configuration from terminal input
// It prompts user for necessary configuration values and saves them to a config file
func loadConfigFromTerminal() error {
	fmt.Println("No configuration file found. Please enter configuration values:")

	// Get h2blog directory
	h2blogDir, err := getH2BlogDir()
	if err != nil {
		return fmt.Errorf("failed to get h2blog directory: %w", err)
	}

	// Update default paths for logs and aof
	defaultLogPath := path.Join(h2blogDir, "logs", "h2blog.log")
	defaultAofPath := path.Join(h2blogDir, "aof", "h2blog.aof")

	conf := &ProjectConfig{
		User: UserConfigData{
			Username: getInput("Username: "),
			Email:    getInput("Email: "),
		},

		Server: ServerConfigData{
			Port: func() uint16 {
				for {
					port := getUint16Input("Server port (press Enter to use default '2233'): ", 0, 65535, "2233")

					// Check if the specified port is available
					if err := checkPortAvailable(port); err != nil {
						fmt.Printf("Port %d is not available. Please choose a different port.\n", port)
						continue
					}

					return port
				}
			}(),
			TokenKey:            getInput("JWT token key (press Enter for generating random string as token key): ", utils.GenRandomString(32)),
			TokenExpireDuration: getUint8Input("Token expiration in days (press Enter for default 1 day): ", 1, 365, "1"),
			Cors: CorsConfigData{
				Origins: []string{"*"}, // Default values
				Headers: []string{"Content-Type", "Authorization", "Token"},
				Methods: []string{"GET", "POST", "PUT", "DELETE"},
			},
		},

		Logger: LoggerConfigData{
			Level:      "info",
			Path:       getInput(fmt.Sprintf("Log file path (press Enter to use default '%s'): ", defaultLogPath), defaultLogPath),
			MaxAge:     getUint16Input("Log max age in days (press Enter for default 1 day): ", 1, 365, "1"),
			MaxSize:    getUint16Input("Log max size in MB (press Enter for default 1): ", 1, 100, "1"),
			MaxBackups: getUint16Input("Max log backups (press Enter for default 10): ", 1, 100, "10"),
			Compress:   getBoolInput("Compress logs? (y/n) (press Enter for default yes): ", "y"),
		},

		MySQL: MySQLConfigData{
			User:     getInput("MySQL username: "),
			Password: getInput("MySQL password: "),
			Host:     getInput("MySQL host: "),
			Port:     getUint16Input("MySQL port (1-65535): ", 1, 65535),
			DB:       getInput("MySQL database name: "),
			MaxOpen:  getUint16Input("Max open connections (1-1000) (press Enter for default 10): ", 1, 1000, "10"),
			MaxIdle:  getUint16Input("Max idle connections (1-100) (press Enter for default 10): ", 1, 100, "10"),
		},

		Oss: OssConfig{
			Endpoint:        getInput("OSS endpoint: "),
			Region:          getInput("OSS region: "),
			AccessKeyId:     getInput("OSS access key ID: "),
			AccessKeySecret: getInput("OSS access key secret: "),
			Bucket:          getInput("OSS bucket name: "),
			ImageOssPath:    "images/", // Default values
			BlogOssPath:     "blogs/",
			WebP: WebPConfigData{
				Enable:  getBoolInput("Enable WebP conversion? (y/n) (press Enter for default yes): ", "y"),
				Quality: getFloatInput("WebP quality (1-100) (press Enter for default 75): ", 1, 100, "75"),
				Size:    getFloatInput("Maximum WebP size in MB (press Enter for default 1): ", 0.1, 10, "1"),
			},
		},

		Cache: CacheConfig{
			Aof: AofConfig{
				Enable:   getBoolInput("Enable AOF persistence? (y/n) (press Enter for default yes): ", "y"),
				Path:     getInput(fmt.Sprintf("AOF file path (press Enter to use default '%s'): ", defaultAofPath), defaultAofPath),
				MaxSize:  getUint16Input("AOF max size in MB (press Enter for default 1): ", 1, 10, "1"),
				Compress: getBoolInput("Compress AOF files? (y/n) (press Enter for default yes): ", "y"),
			},
		},
	}

	// Set global variables
	User = &conf.User
	Server = &conf.Server
	Logger = &conf.Logger
	MySQL = &conf.MySQL
	Oss = &conf.Oss
	Cache = &conf.Cache

	// Create config directory in user's home
	configDir := path.Join(h2blogDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save configuration to file
	data, err := yaml.Marshal(conf)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := path.Join(configDir, "h2blog_config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}

	fmt.Printf("Configuration saved to %s\n", configPath)
	return nil
}

// Helper functions for getting user input

// getInput prompts for user input and ensures non-empty response
func getInput(prompt string, defaultValue ...string) string {
	for {
		fmt.Print(prompt)
		var input string

		// Use bufio.NewReader to read full line including spaces
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v. Please try again.\n", err)
			continue
		}

		// Remove trailing newline and spaces
		input = strings.TrimSpace(line)

		// Ensure input is not empty
		if input != "" {
			return input
		}

		// If no input is provided and a default value is provided, use the default value
		if input == "" && len(defaultValue) > 0 {
			return defaultValue[0]
		}

		fmt.Println("Input cannot be empty. Please try again.")
	}
}

// getBoolInput prompts for a yes/no response and returns true for yes
func getBoolInput(prompt string, defaultValue ...string) bool {
	for {
		input := strings.ToLower(getInput(prompt, defaultValue...))
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Please enter 'y' or 'n'")
	}
}

func getUint8Input(prompt string, min, max uint16, defaultValue ...string) uint8 {
	for {
		input := getInput(prompt, defaultValue...)
		val, err := strconv.ParseUint(input, 10, 8)
		if err == nil && val >= uint64(min) && val <= uint64(max) {
			return uint8(val)
		}
		fmt.Printf("Please enter a number between %d and %d\n", min, max)
	}
}

// getUint16Input prompts for a uint16 within the specified range
func getUint16Input(prompt string, min, max uint16, defaultValue ...string) uint16 {
	for {
		input := getInput(prompt, defaultValue...)
		val, err := strconv.ParseUint(input, 10, 16)
		if err == nil && val >= uint64(min) && val <= uint64(max) {
			return uint16(val)
		}
		fmt.Printf("Please enter a number between %d and %d\n", min, max)
	}
}

// getFloatInput prompts for a float32 within the specified range
func getFloatInput(prompt string, min, max float32, defaultValue ...string) float32 {
	for {
		input := getInput(prompt, defaultValue...)

		val, err := strconv.ParseFloat(input, 32)
		if err == nil && float32(val) >= min && float32(val) <= max {
			return float32(val)
		}
		fmt.Printf("Please enter a number between %.1f and %.1f\n", min, max)
	}
}
