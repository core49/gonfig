[![codecov](https://codecov.io/gh/core49/gonfig/branch/main/graph/badge.svg?token=AO6U2S2I91)](https://codecov.io/gh/core49/gonfig)

# Gonfig

Gonfig is a Go library for loading, writing, and managing configuration data from a JSON file. It provides a variety of
options for custom configuration based on the needs of the user.

## Installation

Gonfig can be installed using go get:

```bash
go get github.com/core49/gonfig
```

## Usage

To use Gonfig, first create a repository instance with your desired options:

```golang
package main

import "github.com/core49/gonfig"

type YourConfigStruct struct {
	Var1 string
	var2 int
}

func main() {
	// Create a new repository instance with options
	// It will not use the os arguments and uses the provided file path
	conf, err := gonfig.New(
		gonfig.OptionDisableDefaultFlagConfiguration(true),
		gonfig.OptionSetConfigFilePathVariable("/path/to/config.json"),
	)

	if err != nil {
		// Handle error
	}

	var config YourConfigStruct

	// Write an empty JSON skeleton of the struct/model to the file
	if err := conf.WriteSkeleton(config); err != nil {
		// Handle error
	}

	// Load configuration data into a struct/model
	if err := conf.Load(&config); err != nil {
		// Handle error
	}
}

```

## Options

The following options are available when creating a new repository instance:

- ```OptionSetFileSystem(fs afero.Fs)``` - sets the file system to the given afero file system instance.
- ```OptionSetConfigFilePathVariable(path string)``` - sets the file path to the given string.
- ```OptionDisableDefaultFlagConfiguration(v bool)``` - disables the default flag configuration if set to true.
- ```OptionSetFlagConfiguration(fc gonfig.FlagConfiguration)``` - sets the flag configuration to the given flag
  configuration slice.
- ```OptionAppendFlagConfig(fc gonfig.FlagConfig)``` - appends the given flag configuration to the flag configuration
  slice.
- ```OptionSetOsArgs(osArgs []string)``` - sets the os arguments to the given string slice.

## Contributing

Contributions to the Gonfig project are welcome and encouraged! Please see
the [contributing guidelines](CONTRIBUTING.md) for more information.

## License

Gonfig is licensed under the [MIT License](LICENSE.md).