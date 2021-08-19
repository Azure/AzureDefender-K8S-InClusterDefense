package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)


// ConfigurationProvider stores all configurations of the application
// The values are read by Viper from a config file or from environment variables
type ConfigurationProvider struct {
	// Viper object containing the configuration values
	viperConfig *viper.Viper
}


// LoadConfig return a new configuration object.
// The object contains configuration values the were read from a config file and from env variables (if needed)
func LoadConfig(configurationName string, configurationType string, configurationPath string, isLocalDevelopment bool) (config *ConfigurationProvider, err error){
	config = new(ConfigurationProvider)
	viper.SetConfigName(configurationName)
	viper.SetConfigType(configurationType)
	if isLocalDevelopment {
		viper.AddConfigPath("./" + configurationPath)
	} else {
		viper.AddConfigPath(configurationPath)
	}
	err = viper.ReadInConfig()
	if err != nil{
		return nil, errors.Wrap(err, "failed to read configuration file")
	}
	config.viperConfig = viper.GetViper()
	return config, nil
}

// SubConfig returns new configuration instance representing a sub tree of the given instance.
// SubConfig is case-insensitive for a key.
// A wrapper method for viper.Sub method
 func (config *ConfigurationProvider) SubConfig(key string) (NewConfig *ConfigurationProvider){
 	NewConfig = new(ConfigurationProvider)
 	NewConfig.viperConfig = config.viperConfig.Sub(key)
 	return NewConfig
 }

// Unmarshal config into our runtime config struct
// A wrapper method for viper.Unmarshal method
func (config *ConfigurationProvider) Unmarshal(runTimeConfig interface{}) (err error){
	err = config.viperConfig.Unmarshal(&runTimeConfig)
	if err != nil{
		return errors.Wrap(err, "failed to read configuration file")
	}
	return nil
}

// AllSettings merges all settings and returns them as a map[string]interface{}
// A wrapper method for viper.AllSettings method
func (config *ConfigurationProvider) AllSettings() map[string] interface{} {
	return config.viperConfig.AllSettings()
}