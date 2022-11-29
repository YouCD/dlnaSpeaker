package ssdp

import (
	"bufio"
	"dlnaSpeaker/log"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/textproto"
	"strings"
)

func checkIP(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ip, 4
		case ':':
			return ip, 6
		}
	}
	return nil, 0
}

func checkIPWithIfName(Interface net.Interface) (interfaceMap map[string]string) {
	interfaceMap = make(map[string]string)

	addrs, err := Interface.Addrs()
	if err != nil {
		log.Errorf("Interface %s not have ip address", Interface.Name)
		return
	}
	for _, a := range addrs {
		ip, t := checkIP(strings.SplitN(a.String(), "/", 2)[0])
		if t == 4 {
			// 排除回环口
			if ip.String() == "127.0.0.1" {
				continue
			}
			// 排除不支持组播的网卡
			if Interface.Flags&ssdpInterfaceFlags != ssdpInterfaceFlags {
				continue
			}
			interfaceMap[Interface.Name] = ip.String()
		}
	}
	return
}
func ssdpInterface(ifName string) (interfaceMap map[string]string, err error) {
	interfaceMap = make(map[string]string)
	if ifName != "" {
		if_, err := net.InterfaceByName(ifName)
		if err != nil {
			return nil, err
		}
		withIfName := checkIPWithIfName(*if_)
		return withIfName, nil
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range interfaces {
		withIfName := checkIPWithIfName(i)
		for s, s2 := range withIfName {
			interfaceMap[s] = s2
		}
	}
	return
}
func parseMessage(b *bufio.Reader) (reqType requestType, err error) {
	tp := textproto.NewReader(b)
	var s string
	if s, err = tp.ReadLine(); err != nil {
		return RequestTypeOther, err
	}
	defer func() {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()

	// 拆解第一行报文
	var f []string
	if f = strings.SplitN(s, " ", 3); len(f) < 3 {
		return RequestTypeOther, errors.New("不支持的协议")
	}

	if f[1] != "*" {
		return RequestTypeOther, errors.New("不支持的请求")
	}

	// 拆解剩下的报文
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return RequestTypeOther, err
	}

	a := make(map[string][]string)
	for key, value := range mimeHeader {
		a[strings.ToLower(key)] = value
	}

	if len(a["man"]) == 0 {
		return RequestTypeOther, errors.New("不包含 man 协议")
	}

	switch {
	case a["man"][0] == `"ssdp:discover"`:
		return RequestTypeSSDPDiscover, nil
	}

	return
}
