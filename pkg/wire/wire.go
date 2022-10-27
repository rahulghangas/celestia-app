package wire

// Enumerate all valid MsgVersion values.
const (
	MsgVersion1 = uint16(1)
)

// Enumerate all valid MsgType values.
const (
	MsgTypePush = uint16(1)
	MsgTypePull = uint16(2)
	MsgTypeSync = uint16(3)
)

type Msg struct {
	Version  uint16 `json:"version"`
	Type     uint16 `json:"type"`
	To       []byte `json:"to"`
	Data     []byte `json:"data"`
	SyncData []byte `json:"syncData"`
}

type Packet struct {
	Msg  Msg
	Addr []byte
}
