package server

import (
	"context"
	"dlnaSpeaker/log"
	"fmt"
	"github.com/beevik/etree"
	"strconv"
	"strings"
	"sync"
)

const (
	actionSetavtransporturi = "setavtransporturi"
	actionGetpositioninfo   = "getpositioninfo"
	actionPlay              = "play"
	actionSeek              = "seek"
	actionGetvolume         = "getvolume"
	actionSetVolume         = "setvolume"
	actionGetTransportInfo  = "gettransportinfo"
	actionPause             = "pause"
	actionStop              = "stop"
)

var (
	media  string
	locker sync.Mutex
)

func parserAction(ctx context.Context, actionHeader string, document *etree.Document, b []byte) string {
	log.Debug("action header is ", actionHeader)
	a := strings.Split(strings.ToLower(strings.ReplaceAll(actionHeader, "\"", "")), "#")[1]
	ip := ctx.Value("ip")
	// 设置播放内容
	switch a {
	case actionSetavtransporturi:
		locker.Lock()
		media = document.SelectElement("s:Envelope").SelectElement("s:Body").SelectElement("u:SetAVTransportURI").SelectElement("CurrentURI").Text()
		defer locker.Unlock()
		render.Player(media)
		webhookHandlerOnPlayer(ctx, media, "Player")
		log.Infof("from %s [%s] request. media is %s", ip, a, media)
	case actionGetpositioninfo:
		durationTime, positionTime, flag := render.Status()
		if !flag {
			log.Infof("from %s [%s] request. not on player", ip, a)
			return ""
		}
		log.Infof("from %s [%s] request. duration time: %s,  position time: %s ", ip, a, durationTime, positionTime)
		return fmt.Sprintf(xmlGetPositionInfo, durationTime, "TrackMetaData", media, positionTime, positionTime)

	case actionSeek:
		Seek := document.SelectElement("s:Envelope").SelectElement("s:Body").SelectElement("u:Seek").SelectElement("Target").Text()
		err := render.Seek(Seek)
		if err != nil {
			log.Errorf("from %s [%s] request. seek error %s", ip, a, err)
		}
		log.Infof("from %s [%s] request. seek to %s", ip, a, Seek)
	case actionGetvolume:
		volume, err := render.GetVolume()
		if err != nil {
			log.Errorf("from %s get volume. error %s", ip, err)
			return ""
		}
		log.Infof("from %s [%s] request. current volume is [%d%%]", ip, a, volume)
		return fmt.Sprintf(xmlGetCurrentVolume, volume)

	case actionSetVolume:
		volume := document.SelectElement("s:Envelope").SelectElement("s:Body").SelectElement("u:SetVolume").SelectElement("DesiredVolume").Text()
		atoi, err := strconv.Atoi(volume)
		if err != nil {
			log.Error(err)
			return ""
		}
		err = render.SetVolume(atoi)
		if err != nil {
			log.Error(err)
			return ""
		}
		log.Infof("from %s [%s] request.set volume is [%s]", ip, a, volume)
		return ""
	case actionGetTransportInfo:
		_, _, flag := render.Status()
		if !flag {
			return ""
		}
		log.Infof("from %s [%s] request. ", ip, a)
		return fmt.Sprintf(xmlGetTransportInfo, "PLAYING")

	case actionPlay:
		render.Play()
		log.Infof("from %s [%s] request. ", ip, a)
		return ""
	case actionPause:
		render.Pause()
		log.Infof("from %s [%s] request. ", ip, a)
		return ""
	case actionStop:
		// TODO stop action
		//webhookHandlerOnFree(ctx, media)
		log.Infof("from %s [%s] request. ", ip, a)
		return ""

	default:
		log.Info("-------------default-------------")
		fmt.Println(actionHeader)
		fmt.Println(string(b))
		log.Info("-------------default-------------")
	}
	return ""
}
