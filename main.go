package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/keshavchand/barbossa/veth"
)

var parent_config Config

func main() {
	parent_config.Parse()
	parent_config.LogCommands = true

	log.SetFlags(log.Lshortfile)
	log.SetPrefix("Barbossa: ")

	if parent_config.IsChild {
		log.Println("Starting Barbossa as child")
		if parent_config.ServiceID >= len(parent_config.Basic.Services) {
			log.Fatal("ServiceID is out of range")
		}
		child(parent_config.ServiceID)
	} else {
		parent()
	}
}

func child(id int) {
	log.Println("Starting child process for service:", parent_config.Basic.Services[id].Name)
	config := parent_config.Basic.Services[id]

	err := WaitForParent()
	if err != nil {
		log.Fatal(err)
	}

	must(syscall.Chroot(config.Root))
	must(os.Chdir("/"))

	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		log.Println("/proc directory does not exist, creating it")
		must(os.Mkdir("/proc", 0755))
		defer os.RemoveAll("/proc")
	}

	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	defer must(syscall.Unmount("/proc", 0))

	execAll := func(context context.Context, cmds []string) {
		for _, command := range cmds {
			cmd_arr_all := strings.Split(command, " ")
			cmd_arr := cmd_arr_all[:]
			if cmd_arr_all[0] == "!!" {
				// !! means run the program and exit if the return code is not 0
				cmd_arr = cmd_arr[1:]
			}

			cmd := exec.CommandContext(context, cmd_arr[0], cmd_arr[1:]...)
			// TODO: Replace stdout to custom logger
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			// XXX: should we set the environment variables here?
			// Currently set this for testing purposes
			cmd.Args = append(cmd.Args, os.Environ()...)

			err := cmd.Start()
			if err != nil {
				log.Println("Error starting command:", command, "with error:", err)
				if cmd_arr_all[0] == "!!" {
					must(err)
				} else {
					continue
				}
			}

			if parent_config.LogCommands {
				pid := cmd.Process.Pid
				log.Println("Running command:", cmd_arr, "with pid:", pid)
			}

			err = cmd.Wait()
			if err != nil {
				log.Println("Error command in execution:", cmd_arr, "with error:", err)
				if cmd_arr_all[0] == "!!" {
					must(err)
				} else {
					continue
				}
			}
		}
	}

	execAll(context.Background(), config.ExecPre)
	execAll(context.Background(), config.Exec)
	execAll(context.Background(), config.ExecPost)
}

func spawnChild(config childConfig) (*exec.Cmd, error) {
	location := "/proc/self/exe"
	args := os.Args[1:]

	cmd := exec.CommandContext(context.TODO(), location, args...)
	cmd.Env = append(cmd.Env,
		childEnv+"=True",
		fmt.Sprintf("%s=%d", serviceIDArg, config.ServiceID),
		fmt.Sprintf("%s=%s", serviceNameArg, config.ServiceName),
	)

	// TODO: use custom logger to color the output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Why should a service have stdin?
	// TODO: use seperate TTY for interactive services
	cmd.Stdin = os.Stdin

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWTIME | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
	}

	err := cmd.Start()
	return cmd, err
}

func parent() {
	sem, err := OpenSem()
	if err != nil {
		log.Fatal(err)
	}

	services := map[string]*exec.Cmd{}

	for c, service := range parent_config.Basic.Services {
		log.Println("Spawning child process for service:", service.Name)
		cmd, err := spawnChild(childConfig{
			ServiceID:   c,
			ServiceName: service.Name,
		})

		if err != nil {
			log.Println("Error spawning child process for service:", service.Name, "with error:", err)
			break
		}
		services[service.Name] = cmd
	}

	defer func() {
		for _, cmd := range services {
			if cmd.Process != nil {
				cmd.Process.Signal(syscall.SIGINT)
			}
		}
	}()

	configuredNetwork, err := ConfigureNetwork(services)
	if err != nil {
		log.Println("Error configuring network:", err)
		return
	}

	defer func() {
		for _, net := range configuredNetwork {
			err := net.Close()
			if err != nil {
				log.Println("Error closing network:", err)
			}
		}
	}()

	for range services {
		err := sem.Signal()
		if err != nil {
			log.Println(err)
		}
	}

	for _, cmd := range services {
		err := cmd.Wait()
		if err != nil {
			log.Println(err)
		}
	}
}

func ConfigureNetwork(services map[string]*exec.Cmd) (NetMap, error) {
	networks := parent_config.Basic.Networks

	netMap := NetMap{}

	for _, network := range networks {
		virtualNetwork, err := veth.New()
		if err != nil {
			return nil, err
		}

		fromCmd, ok := services[network.From.Name]
		if !ok {
			return nil, ErrServiceNotFound(network.From.Name)
		}

		toCmd, ok := services[network.To.Name]
		if !ok {
			return nil, ErrServiceNotFound(network.To.Name)
		}

		err = virtualNetwork.ConnectPids(fromCmd.Process.Pid, toCmd.Process.Pid, network.From.Addr, network.To.Addr)
		if err != nil {
			return nil, err
		}

		log.Print("Connected ", network.From.Name, " to ", network.To.Name, " with ", network.From.Addr, " and ", network.To.Addr)
		netMap[NetKey{network.From.Name, network.To.Name}] = virtualNetwork
	}

	// TODO: Monitor changes
	return netMap, nil
}

type childConfig struct {
	ServiceID   int
	ServiceName string
}

type NetKey struct {
	From string
	To   string
}

type NetMap map[NetKey]veth.VirtualNetwork
