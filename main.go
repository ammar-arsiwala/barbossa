package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

var config Config

func main() {
	config.Parse()
	config.LogCommands = true

	log.SetFlags(log.Lshortfile)
	log.SetPrefix("Barbossa: ")

	if config.IsChild {
		child()
	} else {
		parent()
	}
}

func child() {
	pid := os.Getpid()
	cwd, err := os.Getwd()
	must(err)
	log.Printf("Running as child process with pid: %d @ %s", pid, cwd)
	log.Println("Chrooting to ", config.TargetDir)
	must(syscall.Chroot(config.TargetDir))
	must(os.Chdir("/"))

	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		log.Println("/proc directory does not exist, creating it")
		must(os.Mkdir("/proc", 0755))
		defer os.RemoveAll("/proc")
	}

	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	// NOTE: the binary must be a static compiled binary
	// Go binaries should be compiled with
	// -tags netgo -ldflags '-extldflags "-static"'
	// XXX: Extrypoint should be provided by user
	cmd := exec.CommandContext(context.TODO(), config.TargetCli[0], config.TargetCli[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	must(cmd.Start())

	log.Println("Starting Process with pid: ", cmd.Process.Pid)
	must(cmd.Wait())
	must(syscall.Unmount("/proc", 0))
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

	childPid := cmd.Process.Pid

	setVETH := func(childPID int) {
		// Note: This thread will be going to the net namespace of the childfd
		// this is done so to configure the virtual ethernet
		defer func() { select {} }()
		runtime.LockOSThread()
		// defer runtime.UnlockOSThread()

		childFD, err := unix.PidfdOpen(childPID, 0)
		if err != nil {
			log.Fatal(err)
		}
		unix.Setns(childFD, syscall.CLONE_NEWNET)

		return
		cli := []string{
			"ip link add dev vm1 type veth peer name vm2",
			"ip link set dev vm1 up",
			"ip tuntap add tapm mode tap",
			"ip link set dev tapm up",
			"ip link add brm type bridge",
		}

		for _, c := range cli {
			cliArr := strings.Split(c, " ")
			cmd := exec.CommandContext(context.TODO(), cliArr[0], cliArr[1:]...)
			cmd.Run()
		}
	}

	go setVETH(childPid)
	must(cmd.Wait())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func runCmd(should_log bool, cmd string, args ...string) {
	if should_log {
		log.Println("Running command:", cmd, args)
	}
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	must(c.Run())
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
