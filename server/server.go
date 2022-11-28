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
func (d *dlnaServer) Run() {
	background := context.Background()
	go webhookHandlerOnFree(background, media)
	r := mux.NewRouter()
	r.Use(setResponseMiddleware)
	r.HandleFunc("/description.xml", descriptionHandler).Methods(http.MethodGet)
	r.HandleFunc("/dlna/AVTransport.xml", avTransportHandler).Methods(http.MethodGet)
	r.HandleFunc("/dlna/RenderingControl.xml", renderingControlHandler).Methods(http.MethodGet)
	r.HandleFunc("/RenderingControl/action", renderingControlActionHandler).Methods(http.MethodPost)
	r.HandleFunc("/AVTransport/action", avTransportActionHandler).Methods(http.MethodPost)
	r.HandleFunc("/dlna/ConnectionManager.xml", connectionManagerHandler).Methods(http.MethodPost)
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
			log.Errorf("DLNA Server shutdown failed, err: %v", err)
			os.Exit(1)
		}
		log.Info("DLNA Server gracefully shutdown")
		os.Remove("/tmp/YCD_mpvsocket")
		close(processed)
	}()
	log.Infof("DLNA Server run the http://%s", ip)

	err := srv.ListenAndServe()
	if http.ErrServerClosed != err {
		log.Errorf("DLNA Server not gracefully shutdown, err :%v", err)
		os.Exit(1)
	}

	<-processed

}
