package ssdp

import (
	"bufio"
	"bytes"
	"dlnaSpeaker/log"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"runtime"
	"time"
)

// 网络相关
const (
	netWork             = "udp4"
	MulticastAddrString = "239.255.255.250:1900"
	msgLength           = 4096
	ssdpInterfaceFlags  = net.FlagUp | net.FlagMulticast

	STTypeForSSDPALL        = "ssdp:all"
	STTypeForUPNPROOTDEVICE = "upnp:rootdevice"
)

var (
	MulticastAddr *net.UDPAddr
	UUID          string
	version       = "v1"
)

type requestType int

const (
	RequestTypeOther requestType = iota + 1
	RequestTypeSSDPDiscover
)

func init() {
	UUID = uuid.New().String()
	var err error
	MulticastAddr, err = net.ResolveUDPAddr(netWork, MulticastAddrString)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

type ssdpServer struct {
	version           string
	interfaceMap      map[string]string
	descriptionXMLUrl string
}

func NewSSDPServer(version, ifName, descriptionXMLUrl string) *ssdpServer {
	m, err := ssdpInterface(ifName)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	return &ssdpServer{
		version,
		m,
		descriptionXMLUrl,
	}
}

func (s *ssdpServer) makeUPnPMessage(device, ip string, age int) []byte {
	// todo 需要动态计算
	sendData := &bytes.Buffer{}
	fmt.Fprint(sendData, fmt.Sprintf("NOTIFY * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nNTS: ssdp:alive\r\nUSN: %s\r\nLOCATION: http://%s%s\r\nEXT: \r\nSERVER: %s UPnP/1.0 %s/%s\r\nCACHE-CONTROL: max-age=%d\r\nNT: urn:%s\r\n\r\n", device, ip, s.descriptionXMLUrl, runtime.GOARCH, os.Args[0], version, age, device))
	return sendData.Bytes()
}

func (s *ssdpServer) makeDevices() (devices []string) {
	devices = append(devices, fmt.Sprintf("uuid:%s::upnp:rootdevice", UUID))
	devices = append(devices, fmt.Sprintf("uuid:%s", UUID))
	devices = append(devices, fmt.Sprintf("uuid:%s::urn:schemas-upnp-org:device:MediaRenderer:1", UUID))
	devices = append(devices, fmt.Sprintf("uuid:%s::urn:schemas-upnp-org:service:RenderingControl:1", UUID))
	devices = append(devices, fmt.Sprintf("uuid:%s::urn:schemas-upnp-org:service:AVTransport:1", UUID))
	return devices
}

func (s *ssdpServer) handler(udpConn *net.UDPConn, ip string) {
	for {
		buf := make([]byte, msgLength)
		n, add, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Error(err)
			return
		}
		reqType, err := parseMessage(bufio.NewReader(bytes.NewReader(buf[:n])))
		if err != nil {
			log.Errorf("from %s Multicast,err:%s", add, err)
			return
		}
		switch reqType {
		case RequestTypeSSDPDiscover:
			devices := s.makeDevices()
			for _, device := range devices {
				message := s.makeUPnPMessage(device, ip, 30)
				err := s.sedMsg(udpConn, message, add)
				if err != nil {
					log.Error("send faild, ", err)
				}
			}
		}
	}
}

func (s *ssdpServer) RunServer(port int) {
	// 所有网卡
	for ifName, interfaceIP := range s.interfaceMap {
		if_, err := net.InterfaceByName(ifName)
		if err != nil {
			log.Error(err)
			return
		}

		udpConn, err := net.ListenMulticastUDP("udp4", if_, MulticastAddr)
		if err != nil {
			log.Error(err)
			return
		}
		p := ipv4.NewPacketConn(udpConn)
		if err := p.SetMulticastTTL(2); err != nil {
			log.Error(err)
		}
		ip := fmt.Sprintf("%s:%d", interfaceIP, port)

		// handler 处理 M-SEARCH UDP请求
		go s.handler(udpConn, ip)
		// 组播
		go s.multicastNotify(udpConn, ip)
		log.Infof("started SSDP on %s ip is %s", ifName, ip)

	}

}

func (s *ssdpServer) multicastNotify(udpConn *net.UDPConn, ip string) {
	for {
		devices := s.makeDevices()
		for _, device := range devices {
			message := s.makeUPnPMessage(device, ip, 60)
			err := s.sedMsg(udpConn, message, MulticastAddr)
			if err != nil {
				log.Error("send faild, ", err)
			}
			time.Sleep(time.Millisecond * 500)
		}
		time.Sleep(time.Second * 60)
	}
}
func (s *ssdpServer) sedMsg(ret *net.UDPConn, message []byte, sendAddr *net.UDPAddr) error {
	_, err := ret.WriteToUDP(message, sendAddr)
	return err
}
