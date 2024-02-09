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

	Basic ConfigBasic
}

type ConfigBasic struct {
	Services []struct {
		Name     string   // The name of the service
		Hostname string   // The hostname to bind to (must be unique)
		Root     string   // The root directory to chroot to
		Username string   // The user to run the service as
		ExecPre  []string // Before the target program is executed
		Exec     []string // The target program to execute
		ExecPost []string // After the target program is executed

		CpuNs       int64 // The CPU quota in nanosecond
		MemoryLimit int64 // The memory limit in bytes

		MountPoints []struct {
			Src  string
			Dst  string
			Mode string
		} // Mount points from the rootfs

		ClockOffset int64
	}
}

func (config *Config) Parse() error {
	logCmd := flag.Bool("logcmd", false, "log all commands")
	basicConfigPath := flag.String("basic", "-", "basic configuration file path")
	flag.Parse()

	childFlag := os.Getenv(childEnv)
	if childFlag == "" {
		config.IsChild = false
	} else {
		config.IsChild = true
	}
	config.LogCommands = *logCmd

	if *basicConfigPath != "-" {
		if !fileExists(*basicConfigPath) {
			return ErrConfigFileNotFound
		}

		file, err := os.Open(*basicConfigPath)
		if err != nil {
			return err
		}
		defer file.Close()
		config.parseYAML(file)
	} else {
		return ErrConfigFileNotSuppiled
	}
	return nil
}

func (config *Config) parseYAML(reader io.Reader) error {
	var configBasic ConfigBasic
	err := yaml.NewDecoder(reader).Decode(&configBasic)
	if err != nil {
		return err
	}

	config.Basic = configBasic
	return nil
}

var (
	ErrConfigFileNotFound    = errors.New("Config file not found")
	ErrConfigFileNotSuppiled = errors.New("Config file not supplied")
)

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		return false
	}
	return true
}
