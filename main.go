package main

import (
	"dlnaSpeaker/log"
	"dlnaSpeaker/server"
	"dlnaSpeaker/ssdp"
	"flag"
	"strings"
)

var (
	ifName       string
	friendlyName string
	port         int
	webhook      string
	whitelist    string
)

func init() {
	flag.StringVar(&ifName, "i", "wlo1", "指定网卡")
	flag.StringVar(&friendlyName, "n", "DLNA音箱", "指定音箱名字")
	flag.IntVar(&port, "p", 4564, "指定端口")
	flag.StringVar(&webhook, "web.hook", "", "指定webhook地址")
	flag.StringVar(&whitelist, "white.list", "", "指定ip白名单,多个请用,隔开")

}
func main() {
	flag.Parse()
	whitelistIP := []string{}
	if whitelist != "" {
		split := strings.Split(whitelist, ",")
		whitelistIP = append(whitelistIP, split...)
	}
	server.WhiteIPs = whitelistIP
	log.NewLogger("info")
	server.FriendlyName = friendlyName
	server.WebHook = webhook

	go ssdp.NewSSDPServer("v1.0", ifName, server.URLDescription).RunServer(port)

	server.NewDLNAServer("/tmp/dlnaSpeaker_mpvsocket", port).Run()
}
