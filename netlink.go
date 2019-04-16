package main

import (
	"fmt"
	"syscall"
)

type NetlinkListener struct {
	fd int
	sa *syscall.SockaddrNetlink
}

func ListenNetlink() (*NetlinkListener, error) {
	groups := (1 << (syscall.RTNLGRP_LINK - 1)) |
		(1 << (syscall.RTNLGRP_IPV4_IFADDR - 1)) |
		(1 << (syscall.RTNLGRP_IPV6_IFADDR - 1))

	s, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM,
		syscall.NETLINK_ROUTE)
	if err != nil {
		return nil, fmt.Errorf("socket: %s", err)
	}

	saddr := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Pid:    uint32(0),
		Groups: uint32(groups),
	}

	err = syscall.Bind(s, saddr)
	if err != nil {
		return nil, fmt.Errorf("bind: %s", err)
	}

	return &NetlinkListener{fd: s, sa: saddr}, nil
}

func (l *NetlinkListener) ReadMsgs() ([]syscall.NetlinkMessage, error) {
	defer func() {
		recover()
	}()

	pkt := make([]byte, 2048)

	n, err := syscall.Read(l.fd, pkt)
	if err != nil {
		return nil, fmt.Errorf("read: %s", err)
	}

	msgs, err := syscall.ParseNetlinkMessage(pkt[:n])
	if err != nil {
		return nil, fmt.Errorf("parse: %s", err)
	}

	return msgs, nil
}

func IsNewAddr(msg *syscall.NetlinkMessage) bool {
	if msg.Header.Type == syscall.RTM_NEWADDR {
		return true
	}

	return false
}

func IsDelAddr(msg *syscall.NetlinkMessage) bool {
	if msg.Header.Type == syscall.RTM_DELADDR {
		return true
	}

	return false
}

func IsRelevant(msg *syscall.IfAddrmsg) bool {
	if msg.Scope == syscall.RT_SCOPE_UNIVERSE ||
		msg.Scope == syscall.RT_SCOPE_SITE {
		return true
	}

	return false
}
