package gonfig

import (
	"errors"
	"github.com/spf13/afero"
)

// flagConfiguration is a type representing a slice of flag configurations.
type flagConfiguration []flagConfig

// OptionFunc is a function that takes a repository pointer and modifies it based on the function's implementation.
type OptionFunc func(r *repository)

// closeFileFunc is a function type for closing a file object.
type closeFileFunc func(f afero.File)

// Repository interface contains all the functions that are available to public use.
type Repository interface {
	// Load loads configuration data from the file at the file path into the given model.
	Load(model interface{}) error
	// WriteSkeleton writes a JSON skeleton of the model to the file path.
	WriteSkeleton(model interface{}) error
	// IsEmpty checks if the config file is empty.  Returns `true` if the file does not exist or is empty.
	IsEmpty(model interface{}) (bool, error)
}

// repository struct is the initialized gonfig.
type repository struct {
	data                             interface{}
	disableDefaultFlagConfiguration  bool
	disableDefaultFilePathGeneration bool
	flagConfiguration                flagConfiguration
	filePath                         string
	fileSystem                       afero.Fs
	osArgs                           []string
}

// flagConfiguration struct is used for flagConfiguration that the config will parse at the start.
type flagConfig struct {
	Type         string
	Name         string
	DefaultValue interface{}
	UsageMessage string
	VarPointer   interface{}
}

var (
	// configDir contains the path to the config directory.
	configDir string
	// environment contains the current environment in which the system is running.
	environment string
)

var (
	// ErrEmptyConfigFilePath is an error indicating that a file path is empty.
	ErrEmptyConfigFilePath = errors.New("filepath is empty")
	// ErrInvalidConfigModel is an error indicating that a configuration model is invalid.
	ErrInvalidConfigModel = errors.New("model is empty or no valid struct")
	// ErrConfigFileExist is an error indicating that the configuration file exists and an empty skeleton can not be generated
	ErrConfigFileExist = errors.New("unable to generate skeleton. file already exists")
)

// defaultFlagConfig contains the default configuration flags which can also be disabled.
var defaultFlagConfig = flagConfiguration{
	{
		Type:         "string",
		Name:         "configDir",
		DefaultValue: "./config/",
		UsageMessage: "full path to the config directory with slash at the end.",
		VarPointer:   &configDir,
	},
	{
		Type:         "string",
		Name:         "environment",
		DefaultValue: "prod",
		UsageMessage: "the system environment",
		VarPointer:   &environment,
	},
}
