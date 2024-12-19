package config

import (
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// Config stores all configuration of the application
// The values are read by viper from a config file or env variables
type Config struct {
	Environment         string `mapstructure:"ENVIRONMENT"`
	DBDriver            string `mapstructure:"DB_DRIVER"`
	DBSource            string `mapstructure:"DB_SOURCE"`
	MigrationsURL       string `mapstructure:"MIGRATIONS_URL"`
	HTTPServerAddress   string `mapstructure:"HTTP_SERVER_ADDRESS"`
	RedisAddress        string `mapstructure:"REDIS_ADDRESS"`
	EmailSenderName     string `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress  string `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword string `mapstructure:"EMAIL_SENDER_PASSWORD"`
	TestMerchantEmail   string `mapstructure:"TEST_MERCHANT_EMAIL"`
}

// LoadConfig read configuration from the file or environment variables
func LoadConfig(path, name string) (config Config, err error) {
	viper.SetConfigName(name)  // Set the configuration file name to ".env"
	viper.SetConfigType("env") // Specify that it's an environment file
	viper.AddConfigPath(path)  // Add the current working directory as a search path
	viper.AutomaticEnv()       // Load environment variables from the system

	err = viper.ReadInConfig() // Read the configuration from the .env file
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config) // Unmarshal the loaded configuration into the config struct
	return
}
