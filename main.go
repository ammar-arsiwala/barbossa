package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

var parent_config Config

func main() {
	parent_config.Parse()
	parent_config.LogCommands = true

	log.SetFlags(log.Lshortfile)
	log.SetPrefix("Barbossa: ")

	if parent_config.IsChild {
		log.Println("Starting Barbossa as child")

		var wg sync.WaitGroup
		defer wg.Wait()
		for id := range parent_config.Basic.Services {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				child(id)
			}(id)
		}
	} else {
		parent()
	}
}

func child(id int) {
	config := parent_config.Basic.Services[id]
	pid := os.Getpid()
	cwd, err := os.Getwd()
	must(err)

	log.Printf("Running as child process with pid: %d @ %s", pid, cwd)
	log.Println("Rootfs set to ", config.Root)

	must(chroot(config.Root))
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
			if cmd_arr_all[0] == "!!" { // !! means run the program and exit if the return code is not 0
				cmd_arr = cmd_arr[1:]
			}

			cmd := exec.CommandContext(context, cmd_arr[0], cmd_arr[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr

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

func parent() {
	location := "/proc/self/exe"
	args := os.Args[1:]

	cmd := exec.CommandContext(context.TODO(), location, args...)
	cmd.Env = append(cmd.Env, childEnv+"=True")

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

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
	must(cmd.Start())
	must(cmd.Wait())
}

func chroot(newroot string) error {
	return syscall.Chroot(newroot)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func dirTree(start string, level int) {
	dirs, err := os.ReadDir(start)
	if err != nil {
		return
	}

	for _, dir := range dirs {
		for i := 0; i < level; i++ {
			fmt.Print("  ")
		}
		fmt.Println(dir.Name())
		dirTree(dir.Name(), level+1)
	}
}
