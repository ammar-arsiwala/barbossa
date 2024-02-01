package main

import (
	"errors"
	"flag"
	"io"
	"os"

	yaml "gopkg.in/yaml.v2"
)

var childEnv string = "BARB_CHILD"

type Config struct {
	IsChild     bool
	ConfigFile  string
	LogCommands bool

	TargetDir string
	TargetCli []string

	MountPoints []Mount
}

type yamlConfig struct {
	TargetCli   []string `yaml:"cli,flow"`
	MountPoints []Mount  `yaml:"mount,flow"`
}

type Mount struct {
	Source      string `yaml:"src"`
	Destination string `yaml:"dst"`
	Permissions string `yaml:"perm"`
}

func (config *Config) Parse() error {
	targetDir := flag.String("targetdir", "rootfs", "targetdir directory")
	logIP := flag.Bool("logcmd", false, "log all commands")
	configFilepath := flag.String("config", "-", "configuration file path")
	flag.Parse()

	childFlag := os.Getenv(childEnv)
	if childFlag == "" {
		config.IsChild = false
	} else {
		config.IsChild = true
	}

	config.TargetDir = *targetDir
	config.LogCommands = *logIP
	config.ConfigFile = *configFilepath

	if *configFilepath != "-" {
		if !fileExists(*configFilepath) {
			return ErrConfigFileNotFound
		}

		file, err := os.Open(*configFilepath)
		if err != nil {
			return err
		}
		defer file.Close()
		config.parseYAML(file)
	} else {
		config.TargetCli = append(config.TargetCli, "./main")
	}
	return nil
}

func (config *Config) parseYAML(reader io.Reader) error {
	var yamlConfig yamlConfig
	err := yaml.NewDecoder(reader).Decode(&yamlConfig)
	if err != nil {
		return err
	}

	config.MountPoints = append(config.MountPoints, yamlConfig.MountPoints...)
	config.TargetCli = yamlConfig.TargetCli
	return nil
}

var (
	ErrConfigFileNotFound = errors.New("Config file not found")
)

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		return false
	}

	return true
}
