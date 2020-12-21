//go:generate protoc --go_out=. --go_opt=paths=source_relative connect_response.proto input.proto update_entity.proto update_all.proto

package pb

type Kind byte

const (
	MsgUnknown Kind = iota
	MsgConnectResponse
	MsgPlayerInput
	MsgUpdateEntity
	MsgUpdateAll
	MsgDisconnectPlayer
)

var kindToString = []string{
	MsgUnknown:          "MsgUnknown",
	MsgConnectResponse:  "MsgConnectResponse",
	MsgPlayerInput:      "MsgPlayerInput",
	MsgUpdateEntity:     "MsgUpdateEntity",
	MsgUpdateAll:        "MsgUpdateAll",
	MsgDisconnectPlayer: "MsgDisconnectPlayer",
}

func (kind Kind) String() string {
	kindAsInt := int(kind)
	if kindAsInt >= 0 && kindAsInt < len(kindToString) {
		return kindToString[kind]
	}
	return "MsgUnknown"
}

func AddHeader(data []byte, kind Kind) []byte {
	packetData := make([]byte, 1, len(data)+1)
	packetData[0] = byte(kind)
	packetData = append(packetData, data...)
	return packetData
}
