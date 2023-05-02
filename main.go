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

// New creates a new Repository with the given options.
// It applies the flag configuration, generates a default file path if not disabled, and returns the Repository
func New(options ...OptionFunc) (Gonfig, error) {
	r := &Repository{osArgs: os.Args, fileSystem: afero.NewOsFs()}

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
	return func(r *Repository) {
		r.osArgs = osArgs
	}
}

// OptionSetFileSystem sets the file system.
func OptionSetFileSystem(fs afero.Fs) OptionFunc {
	return func(r *Repository) {
		r.fileSystem = fs
	}
}

// OptionAppendFlagConfig appends the given `flagConfig` to the flag configuration.
func OptionAppendFlagConfig(fc flagConfig) OptionFunc {
	return func(r *Repository) {
		r.flagConfiguration = append(r.flagConfiguration, fc)
	}
}

// OptionSetFlagConfiguration sets the flag configuration to the given `flagConfiguration`.
func OptionSetFlagConfiguration(fc flagConfiguration) OptionFunc {
	return func(r *Repository) {
		r.flagConfiguration = fc
	}
}

// OptionDisableDefaultFlagConfiguration disables the default flag configuration.
func OptionDisableDefaultFlagConfiguration(v bool) OptionFunc {
	return func(r *Repository) {
		r.disableDefaultFlagConfiguration = v
	}
}

// OptionSetConfigFilePathVariable sets the file path and disables default path generation.
func OptionSetConfigFilePathVariable(path string) OptionFunc {
	return func(r *Repository) {
		r.filePath = path
		r.disableDefaultFilePathGeneration = true
	}
}

// applyFlagConfiguration applies flag configurations to command line arguments.
func (r *Repository) applyFlagConfiguration() error {
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

// checkModel verifies that the provided model is a non-nil pointer to a struct type.
func (r *Repository) checkModel(model interface{}) error {
	if model == nil || reflect.ValueOf(model).Kind() != reflect.Ptr || reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
		return ErrInvalidConfigModel
	}

	return nil
}

// openFile opens the Repository's designated file using the file system.
func (r *Repository) openFile() (afero.File, closeFileFunc, error) {
	if len(r.filePath) == 0 {
		return nil, nil, ErrEmptyConfigFilePath
	}

	file, err := r.fileSystem.Open(r.filePath)
	deferFunc := func(f afero.File) {
		_ = f.Close()
	}

	return file, deferFunc, err
}

// Load loads configuration data from the file at the file path into the given model.
func (r *Repository) Load(model interface{}) error {
	if err := r.checkModel(model); err != nil {
		return err
	}

	file, deferFunc, err := r.openFile()
	if err != nil {
		return err
	}
	defer deferFunc(file)

	r.data = model

	return json.NewDecoder(file).Decode(&r.data)
}

// WriteSkeleton writes a JSON skeleton of the model to the file path.
func (r *Repository) WriteSkeleton(model interface{}) error {
	if err := r.checkModel(model); err != nil {
		return err
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

// IsEmpty checks if the config file is empty. Returns `true` if the file does not exist or is empty.
func (r *Repository) IsEmpty(model interface{}) (bool, error) {
	if err := r.checkModel(model); err != nil {
		return false, err
	}

	file, deferFunc, err := r.openFile()
	if os.IsNotExist(err) {
		return true, nil
	}

	if err != nil {
		return false, err
	}
	defer deferFunc(file)

	fileInfo, _ := file.Stat()
	if fileInfo.Size() > 0 {
		return false, nil
	}

	return true, nil
}
