package render

type Render interface {
	Run()
	Player(media string)
	Status() (durationTime, positionTime string, flag bool)
	GetVolume() (int, error)
	SetVolume(volume int) error
	Seek(positionStr string) error
	Play() error
	Pause() error
}
