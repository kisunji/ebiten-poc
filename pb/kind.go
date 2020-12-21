//go:generate protoc --go_out=. --go_opt=paths=source_relative connect_response.proto

package pb

type Kind byte

const (
	MsgUnknown Kind = iota
	MsgConnectResponse
	MsgUpdatePlayer
	MsgDisconnectPlayer
)

var kindToString = []string{
	MsgUnknown:          "MsgUnknown",
	MsgConnectResponse:  "MsgConnectResponse",
	MsgUpdatePlayer:     "MsgUpdatePlayer",
	MsgDisconnectPlayer: "MsgDisconnectPlayer",
}

func (kind Kind) String() string {
	kindAsInt := int(kind)
	if kindAsInt >= 0 && kindAsInt < len(kindToString) {
		return kindToString[kind]
	}
	return "MsgUnknown"
}
