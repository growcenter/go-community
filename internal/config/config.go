package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type (
	Configuration struct {
		Application Application       `mapstructure:"app"`
		Frontend    Frontend          `mapstructure:"frontend"`
		PostgreSQL  PostgreSQL        `mapstructure:"psql"`
		Google      Google            `mapstructure:"google"`
		Auth        Auth              `mapstructure:"auth"`
		Department  map[string]string `mapstructure:"department"`
		Campus      map[string]string `mapstructure:"campus"`
	}
	Application struct {
		Name        string        `mapstructure:"name"`
		Version     string        `mapstructure:"version"`
		Port        int           `mapstructure:"port"`
		Environment string        `mapstructure:"environment"`
		Host        string        `mapstructure:"host"`
		Timeout     time.Duration `mapstructure:"timeout"`
		LogOption   string        `mapstructure:"log_option"`
		LogLevel    string        `mapstructure:"log_level"`
	}
	Frontend struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}
	PostgreSQL struct {
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"host"`
		Name     string `mapstructure:"name"`
		Port     int    `mapstructure:"port"`
		SSLMode  string `mapstructure:"ssl_mode"`
	}
	Google struct {
		ClientID     string `mapstructure:"client_id"`
		ClientSecret string `mapstructure:"client_secret"`
		Redirect     string `mapstructure:"redirect"`
		State        string `mapstructure:"state"`
	}
	Auth struct {
		BearerSecret    string `mapstructure:"bearer_secret"`
		BearerDuration  int    `mapstructure:"bearer_duration"`
		APIKey          string `mapstructure:"api_key"`
		RefreshSecret   string `mapstructure:"refresh_secret"`
		RefreshDuration int    `mapstructure:"refresh_duration"`
	}
)

func New(ctx context.Context) (*Configuration, error) {
	var config Configuration

	viper.AutomaticEnv()
	environment := strings.ToLower(viper.GetString("env"))
	configName := fmt.Sprintf("config.%s", environment)

	viper.AddConfigPath("./config")
	viper.AddConfigPath("./../config")
	viper.AddConfigPath("./../../config")
	viper.AddConfigPath("./../../../config")
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	fmt.Println(viper.AllSettings())
	fmt.Println(viper.AllKeys())
	fmt.Println("----")
	fmt.Printf("Unmarshalled config: %+v\n", config)

	if config.Department != nil {
		fmt.Println("Settings:", config.Department)
		if value, exists := config.Department["tc"]; exists {
			fmt.Println("Key1 Value:", value)
		} else {
			fmt.Println("Key1 not found")
		}
	} else {
		fmt.Println("Settings is nil or not found in the config file.")
	}

	return &config, nil
}
