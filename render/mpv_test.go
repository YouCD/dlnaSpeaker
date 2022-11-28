package render

import (
	"dlnaSpeaker/log"
	"dlnaSpeaker/pkg/mpv"
	"fmt"
	"testing"
)

var (
	mpvR = mpvRender{
		func() *mpv.Client {
			ipcc := mpv.NewIPCClient("/tmp/YCD_mpvsocket")
			c := mpv.NewClient(ipcc)
			return c
		}(),
		"/tmp/YCD_mpvsocket",
	}
)

func Test_mpvRender_Status(t *testing.T) {
	log.NewLogger("info")

	time, positionTime, flag := mpvR.Status()
	fmt.Println(time, positionTime, flag)
}

func Test_mpvRender_SetVolume(t *testing.T) {
	err := mpvR.SetVolume(90)
	if err != nil {
		t.Error(err)
	}

	volume, err := mpvR.client.Volume()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(volume)

}

func Test_mpvRender_Seek(t *testing.T) {
	err := mpvR.Seek("00:00:30")
	if err != nil {
		t.Error(err)
	}

}
