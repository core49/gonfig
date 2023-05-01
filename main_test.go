package gonfig

import (
	"encoding/json"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
)

//go:generate mockgen --build_flags=--mod=mod -destination mock_afero_fs_test.go -package gonfig github.com/spf13/afero Fs

type configModel struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

var defaultOsArgs = []string{"gonfig", "-configDir", "config/", "-environment", "test"}
var configModelData = configModel{
	Name:    "test config value",
	Version: 49,
}

func TestNew(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		t.Run("WithoutOptions", func(t *testing.T) {
			r, err := New(OptionSetOsArgs(defaultOsArgs))

			assert.NoError(t, err)
			assert.NotEmpty(t, r)
		})
		t.Run("WithOptions", func(t *testing.T) {
			t.Run("OptionAppendFlagConfig", func(t *testing.T) {
				var testVar string
				r, err := New(OptionSetOsArgs(defaultOsArgs), OptionSetFileSystem(afero.NewMemMapFs()), OptionAppendFlagConfig(flagConfig{
					Type:         "string",
					Name:         "test",
					DefaultValue: "testValue",
					UsageMessage: "This flag is used as an append test",
					VarPointer:   &testVar,
				}))

				assert.NoError(t, err)
				assert.NotEmpty(t, r)
				assert.NotEmpty(t, testVar)
			})
			t.Run("OptionSetFlagConfiguration", func(t *testing.T) {
				var (
					testStringVar   string
					testIntVar      int
					testBoolVar     bool
					testDurationVar time.Duration
				)

				duration, err := time.ParseDuration("1h20m")
				require.NoError(t, err)

				r, err := New(OptionSetOsArgs(defaultOsArgs), OptionSetFileSystem(afero.NewMemMapFs()), OptionSetFlagConfiguration(flagConfiguration{
					flagConfig{
						Type:         "string",
						Name:         "testString",
						DefaultValue: "testValue",
						UsageMessage: "This flag is used as an string flag test",
						VarPointer:   &testStringVar,
					},
					flagConfig{
						Type:         "int",
						Name:         "testInt",
						DefaultValue: 1,
						UsageMessage: "This flag is used as an int flag test",
						VarPointer:   &testIntVar,
					},
					flagConfig{
						Type:         "bool",
						Name:         "testBool",
						DefaultValue: true,
						UsageMessage: "This flag is used as an bool flag test",
						VarPointer:   &testBoolVar,
					},
					flagConfig{
						Type:         "duration",
						Name:         "testDuration",
						DefaultValue: duration,
						UsageMessage: "This flag is used as an time.duration flag test",
						VarPointer:   &testDurationVar,
					},
				}))

				assert.NoError(t, err)
				assert.NotEmpty(t, r)
			})
		})
		t.Run("Fail", func(t *testing.T) {
			t.Run("WithoutOSArgs", func(t *testing.T) {
				r, err := New(OptionSetOsArgs([]string{}), OptionSetFileSystem(afero.NewMemMapFs()))

				assert.Error(t, err)
				assert.Empty(t, r)
			})
			t.Run("WithOptions", func(t *testing.T) {
				t.Run("OptionDisableDefaultFlagConfiguration", func(t *testing.T) {
					r, err := New(OptionSetFileSystem(afero.NewMemMapFs()), OptionDisableDefaultFlagConfiguration(true))

					assert.Error(t, err)
					assert.Empty(t, r)
				})
			})
		})
	})
}

func TestOptionSetFileSystem(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		r := &repository{}
		require.Empty(t, r.fileSystem)

		afs := afero.NewMemMapFs()

		optionFunc := OptionSetFileSystem(afs)
		optionFunc(r)

		assert.Equal(t, afs, r.fileSystem)
	})
}

func TestOptionAppendFlagConfig(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		r := &repository{}
		fc := flagConfig{
			Type:         "string",
			Name:         "test",
			DefaultValue: "test flag",
			UsageMessage: "test flag",
			VarPointer:   new(string),
		}

		optionFunc := OptionAppendFlagConfig(fc)
		optionFunc(r)

		assert.Equal(t, r.flagConfiguration[0], fc)
	})
}

func TestOptionSetFlagConfiguration(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		r := &repository{}

		optionFunc := OptionSetFlagConfiguration(defaultFlagConfig)
		optionFunc(r)

		assert.Equal(t, r.flagConfiguration, defaultFlagConfig)
	})
}

func TestOptionDisableDefaultFlagConfiguration(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		r := &repository{}

		optionFunc := OptionDisableDefaultFlagConfiguration(true)
		optionFunc(r)

		assert.True(t, r.disableDefaultFlagConfiguration)
	})
}

func TestOptionSetConfigFilePathVariable(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		r := &repository{}

		optionFunc := OptionSetConfigFilePathVariable("/tmp/test.json")
		optionFunc(r)

		assert.Equal(t, r.filePath, "/tmp/test.json")
		assert.True(t, r.disableDefaultFilePathGeneration)
	})
}

func TestRepository_Load(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		afs := afero.NewMemMapFs()

		file, err := afs.Create("test.json")
		require.NoError(t, err)

		defer func() {
			_ = file.Close()
		}()

		j, err := json.Marshal(configModelData)
		require.NoError(t, err)

		_, err = file.Write(j)
		require.NoError(t, err)

		r := &repository{
			filePath:   "test.json",
			fileSystem: afs,
		}

		cm := &configModel{}
		err = r.Load(cm)

		assert.NoError(t, err)
		assert.Equal(t, &configModelData, cm)
	})
	t.Run("Fail", func(t *testing.T) {
		t.Run("InvalidConfigModel", func(t *testing.T) {
			t.Run("Nil", func(t *testing.T) {
				r := &repository{}
				err := r.Load(nil)

				assert.EqualError(t, err, ErrInvalidConfigModel.Error())
			})
			t.Run("NoneStruct", func(t *testing.T) {
				r := &repository{}
				err := r.Load("string")

				assert.EqualError(t, err, ErrInvalidConfigModel.Error())
			})
			t.Run("NonePointerStruct", func(t *testing.T) {
				r := &repository{}
				err := r.Load(configModel{})

				assert.EqualError(t, err, ErrInvalidConfigModel.Error())
			})
		})
		t.Run("EmptyFilePath", func(t *testing.T) {
			r := &repository{}
			err := r.Load(&configModel{})

			assert.EqualError(t, err, ErrEmptyConfigFilePath.Error())
		})
		t.Run("InvalidFilePath", func(t *testing.T) {
			afs := afero.NewMemMapFs()

			r := &repository{
				filePath:   "not.existing",
				fileSystem: afs,
			}
			err := r.Load(&configModel{})

			assert.Error(t, err)
		})
	})
}

func TestRepository_WriteSkeleton(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		afs := afero.NewMemMapFs()

		r := &repository{
			filePath:   "test.json",
			fileSystem: afs,
		}
		err := r.WriteSkeleton(&configModel{})
		require.NoError(t, err)

		file, err := r.fileSystem.Open(r.filePath)
		require.NoError(t, err)

		defer func() {
			_ = file.Close()
		}()

		b, err := io.ReadAll(file)
		require.NoError(t, err)

		assert.Equal(t, "{\n\t\"name\": \"\",\n\t\"version\": 0\n}", string(b))
	})
	t.Run("Fail", func(t *testing.T) {
		t.Run("ConfigFileExist", func(t *testing.T) {
			afs := afero.NewMemMapFs()
			file, err := afs.Create("test.json")
			require.NoError(t, err)
			require.NoError(t, file.Close())

			r := &repository{
				filePath:   "test.json",
				fileSystem: afs,
			}
			err = r.WriteSkeleton(&configModel{})

			assert.EqualError(t, ErrConfigFileExist, err.Error())
		})
		t.Run("InvalidConfigModel", func(t *testing.T) {
			t.Run("Nil", func(t *testing.T) {
				r := &repository{}
				err := r.WriteSkeleton(nil)

				assert.EqualError(t, err, ErrInvalidConfigModel.Error())
			})
			t.Run("NoneStruct", func(t *testing.T) {
				r := &repository{}
				err := r.WriteSkeleton("string")

				assert.EqualError(t, err, ErrInvalidConfigModel.Error())
			})
			t.Run("NonePointerStruct", func(t *testing.T) {
				r := &repository{}
				err := r.WriteSkeleton(configModel{})

				assert.EqualError(t, err, ErrInvalidConfigModel.Error())
			})
		})
		t.Run("MarshalIndent", func(t *testing.T) {
			r := &repository{}

			m := struct {
				X chan int
			}{}
			err := r.WriteSkeleton(&m)

			assert.Error(t, err)
		})
		t.Run("EmptyFilePath", func(t *testing.T) {
			r := &repository{}
			err := r.WriteSkeleton(&configModel{})

			assert.EqualError(t, err, ErrEmptyConfigFilePath.Error())
		})
		t.Run("InvalidFilePath", func(t *testing.T) {
			afs := afero.NewReadOnlyFs(afero.NewMemMapFs())

			r := &repository{
				filePath:   "test.json",
				fileSystem: afs,
			}

			err := r.WriteSkeleton(&configModel{})

			assert.Error(t, err)
		})
	})
}

func TestRepository_IsEmptyConfig(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		t.Run("FileDoesNotExist", func(t *testing.T) {
			afs := afero.NewMemMapFs()

			r := &repository{
				filePath:   "test.json",
				fileSystem: afs,
			}

			exists, err := r.IsEmpty(&configModel{})

			assert.NoError(t, err)
			assert.True(t, exists)
		})
		t.Run("FileExists", func(t *testing.T) {
			t.Run("Empty", func(t *testing.T) {
				afs := afero.NewMemMapFs()

				r := &repository{
					filePath:   "test.json",
					fileSystem: afs,
				}

				file, err := r.fileSystem.Create(r.filePath)
				require.NoError(t, err)

				defer func() {
					_ = file.Close()
				}()

				exists, err := r.IsEmpty(&configModel{})

				assert.NoError(t, err)
				assert.True(t, exists)
			})
			t.Run("ContainsEmptyJSON", func(t *testing.T) {
				afs := afero.NewMemMapFs()

				r := &repository{
					filePath:   "test.json",
					fileSystem: afs,
				}

				file, err := r.fileSystem.Create(r.filePath)
				require.NoError(t, err)

				_, err = file.Write([]byte("{}"))
				require.NoError(t, err)

				defer func() {
					_ = file.Close()
				}()

				exists, err := r.IsEmpty(&configModel{})

				assert.NoError(t, err)
				assert.False(t, exists)
			})
			t.Run("ContainsValidContent", func(t *testing.T) {
				afs := afero.NewMemMapFs()

				r := &repository{
					filePath:   "test.json",
					fileSystem: afs,
				}
				err := r.WriteSkeleton(&configModel{})
				require.NoError(t, err)

				file, err := r.fileSystem.Open(r.filePath)
				require.NoError(t, err)

				defer func() {
					_ = file.Close()
				}()

				b, err := io.ReadAll(file)
				require.NoError(t, err)
				require.Equal(t, "{\n\t\"name\": \"\",\n\t\"version\": 0\n}", string(b))

				exists, err := r.IsEmpty(&configModel{})

				assert.NoError(t, err)
				assert.False(t, exists)
			})
		})
	})
	t.Run("Fail", func(t *testing.T) {
		t.Run("InvalidConfigModel", func(t *testing.T) {
			t.Run("Nil", func(t *testing.T) {
				r := &repository{}
				exists, err := r.IsEmpty(nil)

				assert.EqualError(t, err, ErrInvalidConfigModel.Error())
				assert.False(t, exists)
			})
			t.Run("NoneStruct", func(t *testing.T) {
				r := &repository{}
				exists, err := r.IsEmpty("string")

				assert.EqualError(t, err, ErrInvalidConfigModel.Error())
				assert.False(t, exists)
			})
			t.Run("NonePointerStruct", func(t *testing.T) {
				r := &repository{}
				exists, err := r.IsEmpty(configModel{})

				assert.EqualError(t, err, ErrInvalidConfigModel.Error())
				assert.False(t, exists)
			})
		})
		t.Run("EmptyFilePath", func(t *testing.T) {
			r := &repository{}
			exists, err := r.IsEmpty(&configModel{})

			assert.EqualError(t, err, ErrEmptyConfigFilePath.Error())
			assert.False(t, exists)
		})
	})
}
