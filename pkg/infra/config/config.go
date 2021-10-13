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
func LoadConfig(configurationFile string) (config *ConfigurationProvider, err error) {
	viper.SetConfigFile(configurationFile)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read configuration file")
	}
	return &ConfigurationProvider{viper.GetViper()}, nil
}

// SubConfig returns new configuration instance representing a sub tree of the given instance.
// SubConfig is case-insensitive for a key.
// A wrapper method for viper.Sub method
func (config *ConfigurationProvider) SubConfig(key string) (newConfig *ConfigurationProvider) {
	newConfig = new(ConfigurationProvider)
	newConfig.viperConfig = config.viperConfig.Sub(key)
	return newConfig
}

// Unmarshal config into our runtime config struct
// A wrapper method for viper.Unmarshal method
func (config *ConfigurationProvider) Unmarshal(runTimeConfig interface{}) (err error) {
	err = config.viperConfig.Unmarshal(&runTimeConfig)
	if err != nil {
		return errors.Wrap(err, "failed to read configuration file")
	}
	return nil
}

// AllSettings merges all settings and returns them as a map[string]interface{}
// A wrapper method for viper.AllSettings method
func (config *ConfigurationProvider) AllSettings() map[string]interface{} {
	return config.viperConfig.AllSettings()
}

// CreateSubConfiguration Create new configuration object for each resource,
// based on it's values in the main configuration file
func CreateSubConfiguration(mainConfiguration *ConfigurationProvider, subConfigHierarchy string, configuration interface{}) (err error) {

	// If viper encounter with unknown configuration, it throws panic. we wrap it and return error.
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("failed to set configuration")
		}
	}()

	configValues := mainConfiguration.SubConfig(subConfigHierarchy)
	err = configValues.Unmarshal(&configuration)
	if err != nil {
		return errors.Wrapf(err, "Unable to decode the %v into struct", subConfigHierarchy)
	}
	return nil
}
