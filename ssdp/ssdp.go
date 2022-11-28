package ssdp

import (
	"bufio"
	"bytes"
	"dlnaSpeaker/log"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/net/ipv4"
	"io"
	"net"
	"net/textproto"
	"os"
	"runtime"
	"strings"
	"time"
)

// 网络相关
const (
	netWork             = "udp4"
	MulticastAddrString = "239.255.255.250:1900"
	msgLength           = 4096
)

/*
ST：设置服务查询的目标，它必须是下面的类型：
	ssdp:all 搜索所有设备和服务
	upnp:rootdevice 仅搜索网络中的根设备
	uuid:device-UUID 查询UUID标识的设备
	urn:schemas-upnp-org:device:device-
	type:version查询device-Type字段指定的设备类型，设备类型和版本由UPNP组织定义
	urn:schemas-upnp-org:service:service-
	type:version 查询service-Type字段指定的服务类型，服务类型和版本由UPNP组织定义
*/

/* UPnP 搜索 报文
   M-SEARCH * HTTP/1.1           // 固定
   ST: upnp:rootdevice			 // 指定了特殊类型设备 也可以是  ssdp:all [表示 搜索所有]
   MX: 3						 // 等待时长
   MAN: "ssdp:discover" 		 // 固定
   HOST: 239.255.255.250:1900    // 多播地址
   ===
   HOST: 239.255.255.250:1900
   NTS: ssdp:alive
   USN: uuid:ab69a6f5-c04f-4d12-ba58-4aa11381da13::urn:schemas-upnp-org:service:AVTransport:1
   LOCATION: http://192.168.1.13:4564/description.xml
   EXT:
   SERVER: Linux/6.0.9-arch1-1 UPnP/1.0 Macast/0.7
   CACHE-CONTROL: max-age=66
   NT: urn:schemas-upnp-org:service:AVTransport:1

 返回设备信息

	NOTIFY * HTTP/1.1
	HOST: 239.255.255.250:1900
	NTS: ssdp:alive
	USN: uuid:a11d2b86-60c2-42f4-9ab7-cd6c752d59ec::urn:schemas-upnp-org:device:MediaRenderer:1
	LOCATION: http://192.168.1.13:4564/description.xml
	EXT:
	SERVER: Linux/6.0.7-arch1-1 UPnP/1.0 Macast/0.7
	CACHE-CONTROL: max-age=66
	NT: urn:schemas-upnp-org:device:MediaRenderer:1


*/
// STType
const (
	STTypeForSSDPALL        = "ssdp:all"
	STTypeForUPNPROOTDEVICE = "upnp:rootdevice"
)

// NTSType  (Notification Sub Type)
type NTSType int

const (
	NTSTypeForAlive NTSType = iota + 1
	NTSTypeForByeBye
)

func (n NTSType) Str() string {
	switch n {
	case NTSTypeForAlive:
		return "ssdp:alive"
	case NTSTypeForByeBye:
		return "ssdp:byebye"
	}
	return ""
}

type requestType int

const (
	RequestTypeOther requestType = iota + 1
	RequestTypeSSDPDiscover
)

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

	// 拆解 HTTP 协议 version
	//var ok bool
	//if req.ProtoMajor, req.ProtoMinor, ok = http.ParseHTTPVersion(strings.TrimSpace(f[2])); !ok {
	//	return RequestTypeOther, errors.New("不支持的HTTP version")
	//}
	// 拆解剩下的报文
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return RequestTypeOther, err
	}

	//if len(mimeHeader["MAN"]) == 0 {
	//	fmt.Println("---------------------")
	//	fmt.Println(s)
	//	fmt.Println("---------------------")
	//	return RequestTypeOther, errors.New("不包含 MAN 协议")
	//}
	a := make(map[string][]string)
	for key, value := range mimeHeader {
		a[strings.ToLower(key)] = value
	}

	if len(a["man"]) == 0 {

		//fmt.Printf("flag: %v, MAN:%d,Man:%d,Data: %v,\n", len(mimeHeader["MAN"]) == 0 || len(mimeHeader["Man"]) == 0, len(mimeHeader["MAN"]), len(mimeHeader["Man"]), mimeHeader)
		return RequestTypeOther, errors.New("不包含 man 协议")
	}

	switch {
	case a["man"][0] == `"ssdp:discover"`:

		return RequestTypeSSDPDiscover, nil
	}

	return
}

func SSPDServer(port int, ifName string) {
	MulticastNotify(port, ifName)
}

func makeUPnPMessage(device, ip string, age int) []byte {
	// todo 需要动态计算
	sendData := &bytes.Buffer{}
	fmt.Fprint(sendData, fmt.Sprintf("NOTIFY * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nNTS: ssdp:alive\r\nUSN: %s\r\nLOCATION: http://%s/description.xml\r\nEXT: \r\nSERVER: %s UPnP/1.0 %s/%s\r\nCACHE-CONTROL: max-age=%d\r\nNT: urn:%s\r\n\r\n", device, ip, runtime.GOARCH, os.Args[0], version, age, device))
	return sendData.Bytes()
}

var (
	addr    *net.UDPAddr
	UUID    string
	version = "v1"
)

func init() {
	UUID = uuid.New().String()
	var err error
	addr, err = net.ResolveUDPAddr(netWork, MulticastAddrString)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func makeDevices() (devices []string) {
	devices = append(devices, fmt.Sprintf("uuid:%s::upnp:rootdevice", UUID))
	devices = append(devices, fmt.Sprintf("uuid:%s", UUID))
	devices = append(devices, fmt.Sprintf("uuid:%s::urn:schemas-upnp-org:device:MediaRenderer:1", UUID))
	devices = append(devices, fmt.Sprintf("uuid:%s::urn:schemas-upnp-org:service:RenderingControl:1", UUID))
	devices = append(devices, fmt.Sprintf("uuid:%s::urn:schemas-upnp-org:service:AVTransport:1", UUID))
	return devices
}

func MulticastNotify(port int, ifName string) {
	interfaceList, err := ssdpInterface(ifName)
	if err != nil {
		log.Error(err)
		return
	}
	for ifName, interfaceIP := range interfaceList {
		if_, err := net.InterfaceByName(ifName)
		if err != nil {
			log.Error(err)
			return
		}

		ret, err := net.ListenMulticastUDP("udp4", if_, addr)
		if err != nil {
			log.Error(err)
			return
		}
		p := ipv4.NewPacketConn(ret)
		if err := p.SetMulticastTTL(2); err != nil {
			log.Error(err)
		}

		ip := fmt.Sprintf("%s:%d", interfaceIP, port)

		go func(ifName string) {
			for {
				devices := makeDevices()
				for _, device := range devices {
					message := makeUPnPMessage(device, ip, 10)
					err := sedMsg(ret, message)
					if err != nil {
						log.Error("send faild, ", err)
					}
					time.Sleep(time.Millisecond * 500)
				}
				log.Infof("started SSDP on %s ip is %s", ifName, ip)
				time.Sleep(time.Second * 60)
			}
		}(ifName)
	}

}
func sedMsg(ret *net.UDPConn, message []byte) error {
	_, err := ret.WriteToUDP(message, addr)
	return err
}

const ssdpInterfaceFlags = net.FlagUp | net.FlagMulticast

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
