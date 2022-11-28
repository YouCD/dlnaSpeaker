package render

import (
	"dlnaSpeaker/log"
	"dlnaSpeaker/pkg/mpv"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type mpvRender struct {
	client *mpv.Client
	socket string
}

func NewMPVRender(socket string) Render {
	return &mpvRender{
		socket: socket,
	}
}
func (m *mpvRender) Run() {
	socketChan := make(chan bool)
	go func() {
		os.Remove("/tmp/YCD_mpvsocket")
		cmd := exec.Command("mpv",
			fmt.Sprintf("--input-ipc-server=%s", m.socket),
			"--image-display-duration=inf",
			"--idle=yes",
			"--no-terminal",
			"--on-all-workspaces",
			"--hwdec=yes",
			"--save-position-on-quit=yes",
			"--script-opts=osc-timetotal=yes,osc-layout=bottombar,osc-title=${title},osc-showwindowed=yes,osc-seekbarstyle=bar,osc-visibility=auto",
			"--ontop",
			"--geometry=98%:5%",
			"--autofit=20%",
		)

		err := cmd.Run()
		if err != nil {
			os.Remove("/tmp/YCD_mpvsocket")
			log.Errorf("cmd.Run() failed with %s\n", err)
		}
	}()

	go func() {
		for {
			_, err := os.Stat(m.socket)
			if err != nil {
				log.Warnf("mpv socket file not found: %s", m.socket)
				time.Sleep(time.Millisecond * 500)
				continue
			}
			socketChan <- true
			return
		}
	}()
	flag := <-socketChan
	if flag {
		ipcc := mpv.NewIPCClient(m.socket) // Lowlevel client
		c := mpv.NewClient(ipcc)           // Highlevel client, can also use RPCClient
		m.client = c
	}

}

func (m *mpvRender) Player(media string) {
	m.client.Loadfile(media, mpv.LoadFileModeReplace)
}

func (m *mpvRender) Status() (durationTime, positionTime string, flag bool) {
	filename, err := m.client.Filename()
	if err != nil {
		return "", "", false
	}
	if filename == "" || filename == "<nil>" {
		return "", "", false
	}
	duration, err := m.client.Duration()
	if err != nil {
		return "", "", false
	}
	durationTime = parserTime(duration)

	position, err := m.client.Position()
	if err != nil {
		return "", "", false
	}
	positionTime = parserTime(position)

	return durationTime, positionTime, true
}

func (m *mpvRender) Play() error {
	return m.client.SetPause(false)
}
func (m *mpvRender) Pause() error {
	return m.client.SetPause(true)
}

func (m *mpvRender) Seek(positionStr string) error {
	s := strings.Split(positionStr, ":")
	h, err := strconv.Atoi(s[0])
	if err != nil {
		return err
	}
	var position int
	position += h * 60 * 60

	mm, err := strconv.Atoi(s[1])
	if err != nil {
		return err
	}
	position += mm * 60

	s1, err := strconv.Atoi(s[2])
	if err != nil {
		return err
	}
	position += s1

	return m.client.Seek(position, mpv.SeekModeAbsolute)

}

func (m *mpvRender) SetVolume(volume int) error {
	return m.client.SetVolume(volume)
}

func (m *mpvRender) GetVolume() (int, error) {
	volume, err := m.client.Volume()
	if err != nil {
		return 0, err
	}
	return int(volume), nil
}

func parserTime(duration float64) string {
	split := strings.Split(fmt.Sprintf("%.2f", duration/60), ".")
	float, err := strconv.ParseFloat(fmt.Sprintf("0.%s", split[1]), 64)
	if err != nil {
		log.Error(err)
		return ""
	}
	Second := float * 60
	Minute := split[0]
	return fmt.Sprintf("0:%s:%.0f", Minute, Second)
}
