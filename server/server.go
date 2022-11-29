package server

import (
	"context"
	"dlnaSpeaker/log"
	renderInterface "dlnaSpeaker/render"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	FriendlyName string
	WebHook      string
	WhiteIPs     []string
	render       renderInterface.Render
)

type dlnaServer struct {
	port int
}

func NewDLNAServer(renderSocket string, port int) *dlnaServer {
	render = renderInterface.NewMPVRender(renderSocket)
	render.Run()
	return &dlnaServer{
		port: port,
	}
}

const (
	URLDescription            = "/description.xml"
	URLAVTransport            = "/dlna/AVTransport.xml"
	URLAVTransportAction      = "/dlna/AVTransport/action"
	URLAVTransportEvent       = "/dlna/AVTransport/event"
	URLRenderingControl       = "/dlna/RenderingControl.xml"
	URLRenderingControlAction = "/dlna/RenderingControl/action"
	URLRenderingControlEven   = "/dlna/RenderingControl/event"
	URLConnectionManager      = "/dlna/ConnectionManager.xml"
)

func (d *dlnaServer) Run() {
	background := context.Background()
	go webhookHandlerOnFree(background, media)
	r := mux.NewRouter()
	r.Use(whiteIPsMiddleware, setResponseMiddleware)
	r.HandleFunc(URLDescription, descriptionHandler).Methods(http.MethodGet)
	r.HandleFunc(URLAVTransport, avTransportHandler).Methods(http.MethodGet)
	r.HandleFunc(URLAVTransportAction, avTransportActionHandler).Methods(http.MethodPost)
	r.HandleFunc(URLAVTransportEvent, avTransportEventHandler).Methods(http.MethodPost)
	r.HandleFunc(URLRenderingControl, renderingControlHandler).Methods(http.MethodGet)
	r.HandleFunc(URLRenderingControlAction, renderingControlActionHandler).Methods(http.MethodPost)
	r.HandleFunc(URLRenderingControlEven, renderingControlEvenHandler).Methods(http.MethodPost)
	r.HandleFunc(URLConnectionManager, connectionManagerHandler).Methods(http.MethodPost)

	ip := fmt.Sprintf("%s:%d", "0.0.0.0", d.port)

	processed := make(chan struct{})
	srv := http.Server{
		Addr:    ip,
		Handler: r,
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); nil != err {
			log.Errorf("DLNA Speaker shutdown failed, err: %v", err)
			os.Exit(1)
		}
		log.Info("DLNA Speaker gracefully shutdown")
		os.Remove("/tmp/YCD_mpvsocket")
		close(processed)
	}()
	log.Infof("DLNA Speaker run the http://%s", ip)

	err := srv.ListenAndServe()
	if http.ErrServerClosed != err {
		log.Errorf("DLNA Speaker not gracefully shutdown, err :%v", err)
		os.Exit(1)
	}

	<-processed

}
