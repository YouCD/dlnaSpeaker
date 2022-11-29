package ssdp

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
