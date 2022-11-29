package server

import (
	"bytes"
	"context"
	"dlnaSpeaker/log"
	"dlnaSpeaker/ssdp"
	"encoding/json"
	"fmt"
	"github.com/beevik/etree"
	"io"
	"net/http"
	"strings"
	"time"
)

func descriptionHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, fmt.Sprintf(xmlDescription, ssdp.UUID, FriendlyName))
	return
}

func connectionManagerHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, xmlConnectionManager)
	return
}

func avTransportHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, xmlAVTransport)
	return
}
func renderingControlHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, xmlRenderingControl)
	return
}

func renderingControlActionHandler(writer http.ResponseWriter, request *http.Request) {
	body := request.Body
	defer body.Close()
	all, err := io.ReadAll(body)
	if err != nil {
		log.Error(err)
	}
	doc := etree.NewDocument()
	err = doc.ReadFromBytes(all)
	if err != nil {
		log.Error(err)
	}

	actions := request.Header["Soapaction"]
	if len(actions) > 0 {
		action := strings.ReplaceAll(actions[0], "\"", "")
		data := parserAction(request.Context(), action, doc, all)
		if data != "" {
			fmt.Fprintln(writer, data)
		}
	}

	return
}

func avTransportActionHandler(writer http.ResponseWriter, request *http.Request) {

	body := request.Body
	defer body.Close()
	all, err := io.ReadAll(body)
	if err != nil {
		log.Error(err)
	}
	doc := etree.NewDocument()
	err = doc.ReadFromBytes(all)
	if err != nil {
		log.Error(err)
	}

	actions := request.Header["Soapaction"]
	if len(actions) > 0 {
		action := strings.ReplaceAll(actions[0], "\"", "")
		data := parserAction(request.Context(), action, doc, all)
		if data != "" {
			fmt.Fprintln(writer, data)
		}
	}

}

func avTransportEventHandler(writer http.ResponseWriter, request *http.Request) {

	body := request.Body
	defer body.Close()
	all, err := io.ReadAll(body)
	if err != nil {
		log.Error(err)
	}
	fmt.Println("-----------------------avTransportEventHandler-----------------------------")
	fmt.Println(request.Header)
	fmt.Println(string(all))
	fmt.Println("-----------------------avTransportEventHandler-----------------------------")

}

func renderingControlEvenHandler(writer http.ResponseWriter, request *http.Request) {

	body := request.Body
	defer body.Close()
	all, err := io.ReadAll(body)
	if err != nil {
		log.Error(err)
	}
	fmt.Println("-----------------------renderingControlEvenHandler-----------------------------")
	fmt.Println(request.Header)
	fmt.Println(string(all))
	fmt.Println("-----------------------renderingControlEvenHandler-----------------------------")

}

type Data struct {
	IP    string
	Media string
	State string
}

func webhookHandler(ctx context.Context, media, state string) {
	if WebHook == "" {
		return
	}

	var d = Data{
		IP:    "none",
		Media: media,
		State: state,
	}
	marshal, err := json.Marshal(d)
	if err != nil {
		log.Error(err)
		return
	}

	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(time.Second*2))
	defer cancelFunc()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, WebHook, bytes.NewReader(marshal))
	if err != nil {
		log.Error(err)
		return
	}
	client := http.Client{}
	do, err := client.Do(request)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("web hook to %s", WebHook)
	defer do.Body.Close()
	all, err := io.ReadAll(do.Body)
	if err != nil {
		log.Error(err)
	}
	log.Info(string(all))
}

// webhookHandlerOnPlayer
//
//	@Description: 每当播放时 发送 webhook
//	@param ctx
//	@param media
//	@param state
func webhookHandlerOnPlayer(ctx context.Context, media, state string) {
	webhookHandler(ctx, media, state)
}

// webhookHandlerOnFree
//
//	@Description: render 空闲状态检查 默认空闲5分钟后发送 webhook
//	@param ctx
//	@param media
func webhookHandlerOnFree(ctx context.Context, media string) {
	closeChan := make(chan struct{})
	go func() {
		for {
			_, _, flag := render.Status()
			if !flag {
				closeChan <- struct{}{}
				continue
			}
			time.Sleep(10 * time.Second)
		}
	}()
	for {
		select {
		case <-closeChan:
			log.Info("close after 5 minutes render.")
			time.Sleep(300 * time.Second)
			_, _, flag := render.Status()
			if !flag {
				webhookHandler(ctx, media, "Shutdown")
			}
			continue
		}
	}
}
