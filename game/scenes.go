package game

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/kisunji/ebiten-poc/common"
	"github.com/kisunji/ebiten-poc/pb"
	"google.golang.org/protobuf/proto"
)

type SceneHandler interface {
	Update()
	Draw(*ebiten.Image)
	Next() Scene
}

type StartMenu struct {
	client          *Client
	hostGameHovered bool
	scanningInput   bool
	startPressed    bool
	inputText       string
	startText       string
	next            Scene
}

func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}

func (s *StartMenu) Update() {
	s.next = SceneStartMenu
	x, y := ebiten.CursorPosition()
	if x > 40 && x < common.ScreenWidth/2 &&
		y > common.ScreenHeight/2+24 && y < common.ScreenHeight/2+48 {
		s.startText = ">START"
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			s.startText = "> START"
			s.startPressed = true
		} else if s.startPressed {
			// err := c.DialTLS("ws.chriskim.dev:3000")
			err := s.client.Dial("localhost:8080")
			if err != nil {
				s.next = SceneNotConnected
				return
			} else {
				go s.client.Listen(context.Background())
			}
			s.next = SceneLobby
			return
		} else {
			s.startPressed = false
		}
	} else {
		s.startPressed = false
		s.startText = "START"
	}

	if s.scanningInput {
		if len(s.inputText) < 4 {
			s.inputText += strings.ToUpper(string(ebiten.InputChars()))
		}
		if repeatingKeyPressed(ebiten.KeyBackspace) {
			if len(s.inputText) >= 1 {
				s.inputText = s.inputText[:len(s.inputText)-1]
			}
		}
	}
}

func (s *StartMenu) Draw(screen *ebiten.Image) {
	text.Draw(screen, common.GameTitle, titleFont, 40, common.ScreenHeight/2-50, color.White)
	text.Draw(screen, s.startText, menuFont, 45, common.ScreenHeight/2+50, color.White)
}

func (s *StartMenu) Next() Scene {
	return s.next
}

type NotConnected struct{}

func (n *NotConnected) Update() {}

func (n *NotConnected) Draw(screen *ebiten.Image) {
	text.Draw(screen, `cannot connect to server

try again later :(`, smallFont, 10, common.ScreenHeight/2, color.White)
}

func (n *NotConnected) Next() Scene {
	return SceneNotConnected
}

type Lobby struct {
	Players      []bool
	yourId       int32
	hostId       int32
	Client       *Client
	starting     bool
	next         Scene
	startText    string
	startPressed bool
}

func NewLobby(c *Client) *Lobby {
	return &Lobby{
		Client:  c,
		Players: make([]bool, common.MaxClients),
		next:    SceneLobby,
	}
}

func (l *Lobby) Update() {
	l.startText = "START"
	x, y := ebiten.CursorPosition()
	if x > common.ScreenWidth-50 && x < common.ScreenWidth &&
		y > common.ScreenHeight-50 && y < common.ScreenHeight {
		l.startText = ">START"
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			l.startText = ">START"
			l.startPressed = true
		} else if l.startPressed {
			b, err := proto.Marshal(&pb.ClientMessage{
				Content: &pb.ClientMessage_StartGame{},
			})
			if err != nil {
				log.Fatalln(err)
			}
			l.Client.Send <- b
			l.startPressed = false
		} else {
			l.startPressed = false
		}
	} else {
		l.startText = "START"
	}

outer:
	for {
		select {
		case bytes := <-l.Client.Recv:
			msg := &pb.ServerMessage{}
			err := proto.Unmarshal(bytes, msg)
			if err != nil {
				log.Fatalln(err)
			}
			switch buf := msg.Content.(type) {
			case *pb.ServerMessage_ConnectResponse:
				log.Println("connected")
				if buf.ConnectResponse.IsHost {
					log.Println("is host")
					l.hostId = buf.ConnectResponse.ClientSlot
				}
				l.yourId = buf.ConnectResponse.ClientSlot
			case *pb.ServerMessage_ConnectError:
				log.Println(buf.ConnectError.Message)
				l.next = SceneNotConnected
			case *pb.ServerMessage_UpdateLobby:
				log.Println("updated slots")
				l.Players = buf.UpdateLobby.ConnectedSlots
				l.hostId = buf.UpdateLobby.HostSlot
			case *pb.ServerMessage_GameStart:
				l.next = SceneMainGame
			case *pb.ServerMessage_PlayerDisconnected:
				l.Players[buf.PlayerDisconnected.Id] = false
			case *pb.ServerMessage_NewHost:
				l.hostId = buf.NewHost.Id
			default:
				log.Printf("Unknown message type %T\n", buf)
			}
		case <-l.Client.Disconnect:
			log.Println("lost connection to server")
			l.next = SceneNotConnected
			break outer
		default:
			// no more messages
			break outer
		}
	}
}

func (l *Lobby) Draw(screen *ebiten.Image) {
	text.Draw(screen, "Lobby", titleFont, 40, common.ScreenHeight/2-50, color.White)
	for i, p := range l.Players {
		var s string
		if p {
			if l.hostId == int32(i) {
				s = fmt.Sprintf("Player %d (host)", i)
			} else {
				s = fmt.Sprintf("Player %d", i)
			}
		} else {
			s = "Not connected"
		}
		text.Draw(screen, s, smallFont, 45, common.ScreenHeight/2+i*18, color.White)
	}
	if l.hostId == l.yourId {
		text.Draw(screen, l.startText, smallFont, common.ScreenWidth-50, common.ScreenHeight-50, color.White)
	}
}

func (l *Lobby) Next() Scene {
	return l.next
}

type MainGame struct {
	Client      *Client
	input       input
	count       int
	Speed       int
	Chars       common.Chars
	next        Scene
	Op          *ebiten.DrawImageOptions
	lastUpdated time.Time
}

func (mg *MainGame) Update() {
	if mg.lastUpdated.IsZero() {
		log.Println("lastUpdated is Zero!")
		b, err := proto.Marshal(&pb.ClientMessage{
			Content: &pb.ClientMessage_WorldUpdate{},
		})
		if err != nil {
			log.Fatalln(err)
		}
		mg.Client.Send <- b
		mg.lastUpdated = time.Now()
	}
outer:
	for {
		select {
		case bytes := <-mg.Client.Recv:
			msg := &pb.ServerMessage{}
			err := proto.Unmarshal(bytes, msg)
			if err != nil {
				log.Fatalln(err)
			}
			switch buf := msg.Content.(type) {
			case *pb.ServerMessage_UpdateEntity:
				mg.Chars.UpdateFromData(buf.UpdateEntity)
			case *pb.ServerMessage_UpdateEntities:
				for _, ue := range buf.UpdateEntities.UpdateEntity {
					mg.Chars.UpdateFromData(ue)
				}
			case *pb.ServerMessage_GameStart:
			case *pb.ServerMessage_PlayerDisconnected:
				// maybe kill animation?
				mg.Chars[buf.PlayerDisconnected.Id] = nil
			default:
				log.Printf("Unknown message type %T\n", buf)
			}
		case <-mg.Client.Disconnect:
			log.Println("lost connection to server")
			mg.next = SceneNotConnected
			break outer
		default:
			// no more messages
			break outer
		}
	}

	mg.parseInput()
	for _, char := range mg.Chars {
		if char == nil {
			continue
		}
		char.Move()
	}
	mg.count++
}

func (mg *MainGame) parseInput() {
	var pi pb.Input
	var inputChanged bool
	if ebiten.IsKeyPressed(ebiten.KeyD) || rightTouched() {
		pi.RightPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || leftTouched() {
		pi.LeftPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || upTouched() {
		pi.UpPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || downTouched() {
		pi.DownPressed = true
	}
	if mg.input.RightPressed != pi.RightPressed ||
		mg.input.LeftPressed != pi.LeftPressed ||
		mg.input.UpPressed != pi.UpPressed ||
		mg.input.DownPressed != pi.DownPressed {
		inputChanged = true
	}
	mg.input.RightPressed = pi.RightPressed
	mg.input.LeftPressed = pi.LeftPressed
	mg.input.UpPressed = pi.UpPressed
	mg.input.DownPressed = pi.DownPressed
	if !inputChanged {
		return
	}
	b, err := proto.Marshal(&pb.ClientMessage{
		Content: &pb.ClientMessage_Input{
			Input: &pi,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	mg.Client.Send <- b
}

func (mg MainGame) Draw(screen *ebiten.Image) {
	tmp := make([]*common.Char, len(mg.Chars))
	copy(tmp, mg.Chars)
	sort.Slice(tmp, func(i, j int) bool {
		if tmp[i] == nil || tmp[j] == nil {
			return true
		}
		return tmp[i].Py < tmp[j].Py
	})
	for _, char := range tmp {
		if char == nil {
			continue
		}
		sprite := runnerWaitingFrame
		if char.Vx != 0 || char.Vy != 0 {
			sprite = runnerWalkingFrame
		}
		mg.Op.GeoM.Reset()
		if char.Fx < 0 {
			mg.Op.GeoM.Scale(-1, 1)
			mg.Op.GeoM.Translate(frameWidth, 0)
		}
		mg.Op.GeoM.Translate(
			char.Px-frameWidth/2,
			char.Py-frameHeight/2,
		)
		screen.DrawImage(sprite(mg.count/mg.Speed+char.Offset), mg.Op)
	}
}

func (mg MainGame) Next() Scene {
	return SceneMainGame
}
