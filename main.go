package gonfig

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/spf13/afero"
	"os"
	"reflect"
	"time"
)

// New creates a new repository with the given options.
// It applies the flag configuration, generates a default file path if not disabled,
// and returns the repository and any errors encountered during creation.
func New(options ...OptionFunc) (Repository, error) {
	r := &repository{osArgs: os.Args, fileSystem: afero.NewOsFs()}

	for _, optionFunc := range options {
		optionFunc(r)
	}

	err := r.applyFlagConfiguration()
	if err != nil {
		return nil, err
	}

	if !r.disableDefaultFilePathGeneration {
		r.filePath = configDir + environment + ".json"
	}

	return r, nil
}

// OptionSetOsArgs sets the os arguments.
func OptionSetOsArgs(osArgs []string) OptionFunc {
	return func(r *repository) {
		r.osArgs = osArgs
	}
}

// OptionSetFileSystem sets the file system.
func OptionSetFileSystem(fs afero.Fs) OptionFunc {
	return func(r *repository) {
		r.fileSystem = fs
	}
}

// OptionAppendFlagConfig appends the given `flagConfig` to the flag configuration.
func OptionAppendFlagConfig(fc flagConfig) OptionFunc {
	return func(r *repository) {
		r.flagConfiguration = append(r.flagConfiguration, fc)
	}
}

// OptionSetFlagConfiguration sets the flag configuration to the given `flagConfiguration`.
func OptionSetFlagConfiguration(fc flagConfiguration) OptionFunc {
	return func(r *repository) {
		r.flagConfiguration = fc
	}
}

// OptionDisableDefaultFlagConfiguration disables the default flag configuration.
func OptionDisableDefaultFlagConfiguration(v bool) OptionFunc {
	return func(r *repository) {
		r.disableDefaultFlagConfiguration = v
	}
}

// OptionSetConfigFilePathVariable sets the file path and disables default path generation.
func OptionSetConfigFilePathVariable(path string) OptionFunc {
	return func(r *repository) {
		r.filePath = path
		r.disableDefaultFilePathGeneration = true
	}
}

// applyFlagConfiguration applies flag configurations to command line arguments.
func (r *repository) applyFlagConfiguration() error {
	fmt.Println(r.osArgs)
	if len(r.osArgs) == 0 {
		return errors.New("os args should not be empty")
	}

	flagSet := flag.NewFlagSet(r.osArgs[0], flag.ContinueOnError)

	if !r.disableDefaultFlagConfiguration {
		r.flagConfiguration = append(r.flagConfiguration, defaultFlagConfig...)
	}

	for _, config := range r.flagConfiguration {
		switch config.Type {
		case "string":
			flagSet.StringVar(config.VarPointer.(*string), config.Name, config.DefaultValue.(string), config.UsageMessage)
		case "int":
			flagSet.IntVar(config.VarPointer.(*int), config.Name, config.DefaultValue.(int), config.UsageMessage)
		case "bool":
			flagSet.BoolVar(config.VarPointer.(*bool), config.Name, config.DefaultValue.(bool), config.UsageMessage)
		case "duration":
			flagSet.DurationVar(config.VarPointer.(*time.Duration), config.Name, config.DefaultValue.(time.Duration), config.UsageMessage)
		}
	}

	return flagSet.Parse(r.osArgs[1:])
}

// Load loads configuration data from the file at the file path into the given model.
// It returns an error if there was a problem with the file or the model is invalid.
func (r *repository) Load(model interface{}) error {
	if model == nil || reflect.ValueOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrInvalidConfigModel
	}

	if len(r.filePath) == 0 {
		return ErrEmptyConfigFilePath
	}

	file, err := r.fileSystem.Open(r.filePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	r.data = model

	return json.NewDecoder(file).Decode(&r.data)
}

// WriteSkeleton writes a JSON skeleton of the model to the file path.
// It returns an error if there was a problem with the model or the file.
func (r *repository) WriteSkeleton(model interface{}) error {
	if model == nil || reflect.ValueOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrInvalidConfigModel
	}

	b, err := json.MarshalIndent(model, "", "\t")
	if err != nil {
		return err
	}

	if len(r.filePath) == 0 {
		return ErrEmptyConfigFilePath
	}

	_, err = r.fileSystem.Stat(r.filePath)
	if !os.IsNotExist(err) {
		return ErrConfigFileExist
	}

	var file afero.File
	file, err = r.fileSystem.Create(r.filePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	_, err = file.Write(b)
	return err
}
