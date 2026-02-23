package env

import (
	"bytes"
	"embed"
	"log"

	"github.com/spf13/viper"
)

//go:embed *.toml
var configFiles embed.FS

type Env struct {
	Sqlite3        Sqlite3        `mapstructure:"sqlite3"`
	SupabaseConfig SupabaseConfig `mapstructure:"supabase"`
}

type Sqlite3 struct {
	Name string `mapstructure:"name"`
}

type SupabaseConfig struct {
	ProjectID string `mapstructure:"project_id"`
	AnonKey   string `mapstructure:"anon_key"`
}

// GetConfig returns entire project configuration
func GetEnv() *Env {
	return GetEnvFromFile("env")
}

// GetEnvFromFile returns configuration from specific file object
func GetEnvFromFile(fileName string) *Env {
	viper.SetConfigType("toml")

	// 1. Read the file dynamically from the embedded file system
	// This replaces all the viper.AddConfigPath calls
	fileBytes, err := configFiles.ReadFile(fileName + ".toml")
	if err != nil {
		log.Printf("failed to find %s.toml in embedded files: %v\n", fileName, err)
		return &Env{}
	}

	// 2. Pass the bytes directly to viper
	err = viper.ReadConfig(bytes.NewReader(fileBytes))
	if err != nil {
		log.Printf("failed to parse config: %s\n", err)
	}

	config := &Env{}
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Printf("couldn't unmarshal config: %s\n", err)
	}

	return config
}
