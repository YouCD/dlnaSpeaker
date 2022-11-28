package main

import (
	"dlnaSpeaker/log"
	"dlnaSpeaker/server"
	"dlnaSpeaker/ssdp"
	"flag"
)

var (
	ifName       string
	friendlyName string
	port         int
	webhook      string
)

func init() {
	flag.StringVar(&ifName, "i", "wlo1", "指定网卡")
	flag.StringVar(&friendlyName, "n", "DLNA音箱", "指定音箱名字")
	flag.IntVar(&port, "p", 4564, "指定端口")
	flag.StringVar(&webhook, "hook", "", "指定webhook地址")

}
func main() {
	flag.Parse()
	log.NewLogger("info")
	server.FriendlyName = friendlyName
	server.WebHook = webhook

	go ssdp.SSPDServer(port, ifName)
	server.NewDLNAServer("/tmp/dlnaSpeaker_mpvsocket", port).Run()
}
