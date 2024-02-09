package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/aka-mj/go-semaphore"
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
	sem := semaphore.Semaphore{}
	err := sem.Open(semaphore_name, 0666, 0)
	if err != nil {
		log.Fatal(err)
	}
	services := map[string]*exec.Cmd{}
	failedBootUpProcess := false

	defer func() {
		if failedBootUpProcess {
			for _, cmd := range services {
				if cmd.Process != nil {
					// XXX: should we kill all the process (after some time) instead of giving them a sigint
					// 	Handle situation in which process deadlocks
					// 	What if the process has graceful exit builtin
					cmd.Process.Signal(syscall.SIGINT)
				}
			}
		} else {
			for _, cmd := range services {
				cmd.Wait()
			}
		}
	}()

	for c, service := range parent_config.Basic.Services {
		cmd, err := spawnChild(childConfig{
			ServiceID:   c,
			ServiceName: service.Name,
		})

		if err != nil {
			log.Println("Error spawning child process for service:", service.Name, "with error:", err)
			failedBootUpProcess = true
			break
		}
		services[service.Name] = cmd
	}

	for range services {
		err := sem.Post()
		if err != nil {
			log.Println(err)
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type childConfig struct {
	ServiceID   int
	ServiceName string
}

func WaitForParent() error {
	sem := semaphore.Semaphore{}
	err := sem.Open(semaphore_name, 0666, 0)
	if err != nil {
		return err
	}
	err = sem.Wait()
	if err != nil {
		return err
	}

	return nil
}
