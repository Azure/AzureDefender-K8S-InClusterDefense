package configs

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

// Configuration stores all configurations of the application
// The values are read by Viper from a config file or from environment variables
type Configuration struct {
	viperConfig *viper.Viper
}

// NewConfiguration initiate a new configuration object and read configuration from file
func NewConfiguration(configurationName string, configurationType string, configurationPath string, readEnv bool) (Config *Configuration, err error){
	Config = new(Configuration)
	viper.SetConfigName(configurationName)
	viper.SetConfigType(configurationType)
	viper.AddConfigPath("./" + configurationPath)
	viper.AddConfigPath(configurationPath)
	err = viper.ReadInConfig()
	if readEnv{
		viper.AutomaticEnv()
	}
	Config.viperConfig = viper.GetViper()
	return Config, err
}

// SubConfig returns new configuration instance representing a sub tree of the given instance.
// SubConfig is case-insensitive for a key.
// A wrapper method for viper.Sub method
 func (config *Configuration) SubConfig(key string) (NewConfig *Configuration){
 	NewConfig = new(Configuration)
 	NewConfig.viperConfig = config.viperConfig.Sub(key)
 	return NewConfig
 }

// Unmarshal config into our runtime config struct
// A wrapper method for viper.Unmarshal method
func (config *Configuration) Unmarshal(runTimeConfig interface{}) (err error){
	err = config.viperConfig.Unmarshal(&runTimeConfig)
	return err
}

// BindEnvVariable binds a key to env variable.
// If env variable doesn't exist bind the key to a default value (if given).
// The method uses os.Getenv, viper.BindEnv and viper.SetDefault methods
func (config Configuration) BindEnvVariable(input... string) (err error) {
	if len(input) < 2 {
		err = fmt.Errorf("not enough arguments were given")
		return
	}

	if len(input) < 3 {
		err = config.viperConfig.BindEnv(input...)
		return
	}

	envValue := os.Getenv(input[1])
	if len(envValue) > 0 {
		config.viperConfig.SetDefault(input[0], input[1])
	} else {
		config.viperConfig.SetDefault(input[0], input[2])
	}
	return nil
}
