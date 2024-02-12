package veth

import (
	"errors"
	"log"
	"net"
	"runtime"
	"syscall"

	"github.com/milosgajdos/tenus"
	"golang.org/x/sys/unix"
)

type VirtualNetwork interface {
	Close() error
	ConnectPids(pid1, pid2 int, ip1_cidr, ip2_cidr string) error
}

type virtualNetworkInterfaceImpl struct {
	tenus.Vether
	inUse                bool
	linkPid, linkPeerPid int
}

func New() (VirtualNetwork, error) {
	veth, err := tenus.NewVethPair()
	if err != nil {
		return nil, err
	}

	vethImpl := virtualNetworkInterfaceImpl{
		Vether: veth,
	}

	return &vethImpl, nil
}

func (v *virtualNetworkInterfaceImpl) Close() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer unix.Setns(1, syscall.CLONE_NEWNET)

	err := unix.Setns(v.linkPid, syscall.CLONE_NEWNET)
	if err != nil {
		return err
	}

	err = v.Vether.DeleteLink()
	if err != nil {
		return err
	}

	err = unix.Setns(v.linkPeerPid, syscall.CLONE_NEWNET)
	if err != nil {
		return err
	}

	err = v.Vether.DeletePeerLink()
	// Note(keshav): deleting link might as well delete the pair interface too
	// Just ignore the error if that happens
	if err != nil && !errors.Is(err, syscall.ENODEV) {
		return err
	}

	return nil
}

func (v *virtualNetworkInterfaceImpl) ConnectPids(pid1, pid2 int, ip1, ip2 string) error {
	v.linkPid = pid1
	v.linkPeerPid = pid2
	v.inUse = true

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer enterPidNetNamespace(1)

	err := v.Vether.SetLinkNetNsPid(pid1)
	if err != nil {
		return err
	}

	err = v.Vether.SetPeerLinkNsPid(pid2)
	if err != nil {
		return err
	}

	err = enterPidNetNamespace(pid1)
	if err != nil {
		log.Println(err)
	}

	ip, ipnet, err := net.ParseCIDR(ip1)
	if err != nil {
		return err
	}

	err = v.Vether.SetLinkIp(ip, ipnet)
	if err != nil {
		return err
	}

	err = v.Vether.SetLinkUp()
	if err != nil {
		return err
	}

	err = enterPidNetNamespace(pid2)
	if err != nil {
		log.Println(err)
	}

	ip, ipnet, err = net.ParseCIDR(ip2)
	if err != nil {
		return err
	}

	err = v.Vether.SetPeerLinkIp(ip, ipnet)
	if err != nil {
		return err
	}

	err = v.Vether.SetPeerLinkUp()
	if err != nil {
		return err
	}

	return nil
}

func enterPidNetNamespace(pid int) error {
	fd, err := unix.PidfdOpen(pid, 0)
	if err != nil {
		return err
	}

	return unix.Setns(fd, syscall.CLONE_NEWNET)
}

var ErrUnimplemented = errors.New("unimplemented")
