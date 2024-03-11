package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

const childEnv = "BARB_CHILD"
const serviceIDArg = "BARB_SERVICE_ID"
const serviceNameArg = "BARB_SERVICE_NAME"

type Config struct {
	IsChild     bool
	ConfigFile  string
	LogCommands bool
	ServiceID   int

	Basic ConfigBasic
}

type Service struct {
	Name     string   // The name of the service
	Image    string   // The image to use
	Hostname string   // The hostname to bind to (must be unique)
	Root     string   // The root directory to chroot to
	Username string   // The user to run the service as
	Exec     []string // The target program to execute
	Env      []string // The environment variables to set

	CpuNs       int64 // The CPU quota in nanosecond
	MemoryLimit int64 // The memory limit in bytes

	MountPoints []struct {
		Src  string
		Dst  string
		Mode string
	} // Mount points from the rootfs

	ClockOffset int64
}

type Network struct {
	Name     string
	Gateway  string
	Subnet   string
	Services []struct {
		Name string
		Addr string
	}
}

type ConfigBasic struct {
	Services []Service
	Networks []Network
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

	service_id := os.Getenv(serviceIDArg)
	if service_id == "" && config.IsChild {
		return ErrServiceIDNotFound
	} else if service_id != "" {
		var err error
		config.ServiceID, err = strconv.Atoi(service_id)
		if err != nil {
			return fmt.Errorf("error paring service id: %s", err.Error())
		}
	}

	config.LogCommands = *logCmd

	if *basicConfigPath == "-" {
		return ErrConfigFileNotSuppiled
	}

	if !fileExists(*basicConfigPath) {
		return ErrConfigFileNotFound
	}

	file, err := os.Open(*basicConfigPath)
	if err != nil {
		return err
	}

	defer file.Close()

	err = config.parseYAML(file)
	if err != nil {
		return err
	}

	for _, service := range config.Basic.Services {
		if service.Image == "" {
			log.Println("Image not found for service: ", service.Name)
			return ErrImageNotFound
		}
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
	ErrServiceIDNotFound     = errors.New("Service ID not found")
	ErrImageNotFound         = errors.New("Image not found")
)

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		return false
	}
	return true
}
