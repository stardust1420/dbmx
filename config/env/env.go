package env

import (
	"log"

	"github.com/spf13/viper"
)

type Env struct {
	Sqlite3        Sqlite3        `mapstructure:"sqlite3"`
	SupabaseConfig SupabaseConfig `mapstructure:"supabase_config"`
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
	// looking for filename `default` inside `src/server` dir with `.toml` extension
	viper.SetConfigName(fileName)
	viper.AddConfigPath("../config/env/")
	viper.AddConfigPath("../../config/env/")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/env/")
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("failed to load config")
	}
	config := &Env{}
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Printf("couldn't read config: %s", err)
	}

	return config
}
