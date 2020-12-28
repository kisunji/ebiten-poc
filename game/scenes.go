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
			err := s.client.DialTLS("ws.chriskim.dev:3000")
			// err := s.client.Dial("localhost:8080")
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
	if x > common.ScreenWidth-120 && x < common.ScreenWidth-30 &&
		y > common.ScreenHeight-46 && y < common.ScreenHeight-30 {
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
			s = fmt.Sprintf("Player %d", i+1)
			if l.yourId == int32(i) {
				s = fmt.Sprintf("%s (you)", s)
			}
			if l.hostId == int32(i) {
				s = fmt.Sprintf("%s (host)", s)
			}
		} else {
			s = "Not connected"
		}
		text.Draw(screen, s, smallFont, 45, common.ScreenHeight/2+i*18, color.White)
	}
	if l.hostId == l.yourId {
		text.Draw(screen, l.startText, smallFont, common.ScreenWidth-120, common.ScreenHeight-30, color.White)
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
	Coins       []*common.Coin
	next        Scene
	Op          *ebiten.DrawImageOptions
	lastUpdated time.Time
	debouncer   *Debouncer
	EndMessage  string
}

func NewMainGame(c *Client, d *Debouncer) *MainGame {
	return &MainGame{
		Op:        &ebiten.DrawImageOptions{},
		Speed:     5,
		Client:    c,
		Chars:     make([]*common.Char, common.MaxChars),
		next:      SceneMainGame,
		debouncer: d,
	}
}

func (mg *MainGame) Update() {
	if mg.lastUpdated.IsZero() {
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
			case *pb.ServerMessage_NewCoin:
				coin := &common.Coin{
					Px:          buf.NewCoin.Px,
					Py:          buf.NewCoin.Py,
					FrameOffset: int(buf.NewCoin.FrameOffset),
				}
				mg.Coins = append(mg.Coins, coin)
			case *pb.ServerMessage_CoinGot:
				mg.Coins[buf.CoinGot.Index].PickedUp = true
			case *pb.ServerMessage_GameStart:
			case *pb.ServerMessage_GameEnd:
				var sb strings.Builder
				sb.WriteString("GAME OVER\n")
				for i, score := range buf.GameEnd.Score {
					if score > 0 {
						if i == int(buf.GameEnd.Survivor) {
							sb.WriteString(fmt.Sprintf("Player %d: %d (survivor)\n", i+1, score))
						} else {
							sb.WriteString(fmt.Sprintf("Player %d: %d\n", i+1, score))
						}
					}
				}
				mg.EndMessage = sb.String()
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
		if char.Attacking() {
			char.Attack()
		}
		char.Move()
	}
	mg.count++
}

func (mg *MainGame) parseInput() {
	var pi pb.Input
	var inputChanged bool
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) || rightTouched() {
		pi.RightPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) || leftTouched() {
		pi.LeftPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || upTouched() {
		pi.UpPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) || downTouched() {
		pi.DownPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		pi.ActionPressed = true
	}
	if mg.input.RightPressed != pi.RightPressed ||
		mg.input.LeftPressed != pi.LeftPressed ||
		mg.input.UpPressed != pi.UpPressed ||
		mg.input.DownPressed != pi.DownPressed ||
		mg.input.ActionPressed != pi.ActionPressed {
		inputChanged = true
	}
	mg.input.RightPressed = pi.RightPressed
	mg.input.LeftPressed = pi.LeftPressed
	mg.input.UpPressed = pi.UpPressed
	mg.input.DownPressed = pi.DownPressed
	mg.input.ActionPressed = pi.ActionPressed
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
	mg.debouncer.input <- b
}

func (mg *MainGame) Draw(screen *ebiten.Image) {
	for _, coin := range mg.Coins {
		if coin.PickedUp {
			continue
		}
		img := coinFrame(mg.count/15 + coin.FrameOffset)
		w, h := img.Size()
		mg.Op.GeoM.Reset()
		mg.Op.GeoM.Translate(
			coin.Px-float64(w)/2,
			coin.Py-float64(h)/2,
		)
		screen.DrawImage(img, mg.Op)
	}
	tmp := make([]*common.Char, len(mg.Chars))
	copy(tmp, mg.Chars)
	sort.Slice(tmp, func(i, j int) bool {
		if tmp[i] == nil || tmp[j] == nil {
			return true
		}
		return tmp[i].IsDead || tmp[i].Py < tmp[j].Py
	})
	for _, char := range tmp {
		if char == nil {
			continue
		}
		if char.IsDead {
			sprite := deadFrame()
			w, h := sprite.Size()
			mg.Op.GeoM.Reset()
			if char.Fx < 0 {
				mg.Op.GeoM.Scale(-1, 1)
				mg.Op.GeoM.Translate(float64(w), 0)
			}
			mg.Op.GeoM.Translate(
				char.Px-float64(w)/2,
				char.Py-float64(h)/2,
			)
			screen.DrawImage(sprite, mg.Op)
			continue
		}
		// fallback sprite if (fx,fy) = (0,0)
		sprite := downRestingFrame(0)
		clock := mg.count/mg.Speed + char.Offset
		if char.Vx == 0 && char.Vy == 0 {
			clock = 0
		}
		if char.Fx != 0 {
			if char.Fy < 0 {
				if char.Attacking() {
					sprite = upRightAttackingFrame(char.AttackFrame)
				} else {
					sprite = upRightRestingFrame(clock)
				}
			}
			if char.Fy > 0 {
				if char.Attacking() {
					sprite = downRightAttackingFrame(char.AttackFrame)
				} else {
					sprite = downRightRestingFrame(clock)
				}
			}
			if char.Fy == 0 {
				if char.Attacking() {
					sprite = rightAttackingFrame(char.AttackFrame)
				} else {
					sprite = rightRestingFrame(clock)
				}
			}
		} else {
			if char.Fy < 0 {
				if char.Attacking() {
					sprite = upAttackingFrame(char.AttackFrame)
				} else {
					sprite = upRestingFrame(clock)
				}
			}
			if char.Fy > 0 {
				if char.Attacking() {
					sprite = downAttackingFrame(char.AttackFrame)
				} else {
					sprite = downRestingFrame(clock)
				}
			}
		}

		w, h := sprite.Size()
		mg.Op.GeoM.Reset()
		if char.Fx < 0 {
			mg.Op.GeoM.Scale(-1, 1)
			mg.Op.GeoM.Translate(float64(w), 0)
		}
		mg.Op.GeoM.Translate(
			char.Px-float64(w)/2,
			char.Py-float64(h)/2,
		)
		screen.DrawImage(sprite, mg.Op)
	}
	if mg.EndMessage != "" {
		text.Draw(screen, mg.EndMessage, smallFont, 0, common.ScreenHeight/2, color.White)
	}
}

func (mg *MainGame) Next() Scene {
	return mg.next
}
