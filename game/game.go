package game

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
	"github.com/kisunji/ebiten-poc/pb"
	"google.golang.org/protobuf/proto"
)

const (
	ScreenWidth   = 640
	ScreenHeight  = 480
	ScreenPadding = 10

	MaxChars   = 64
	MaxClients = 8
)

var lastUpdated time.Time

func init() {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)
}

type Game struct {
	Client *Client
	count  int
	Speed  int
	Chars  Chars
	inited bool
	input  input
	Op     *ebiten.DrawImageOptions
}

type input struct {
	UpPressed     bool
	DownPressed   bool
	LeftPressed   bool
	RightPressed  bool
	ActionPressed bool
}

func (g *Game) Update() error {
outer:
	for {
		select {
		case buf := <-g.Client.recv:
			kind := pb.Kind(buf[0])
			buf = buf[1:]
			switch kind {
			case pb.MsgConnectResponse:
				resp := &pb.ConnectResponse{}
				err := proto.Unmarshal(buf, resp)
				if err != nil {
					log.Fatal("marshaling error: ", err)
				}
				// Last time we received an update about the world
				lastUpdated = time.Now()

				log.Printf("%s: received login data: %v\n", kind, resp.ClientSlot)
				char := NewCharAt(resp.Px, resp.Py)
				g.Chars[resp.ClientSlot] = char
			case pb.MsgDisconnectPlayer:
				log.Println("disconnected by server")
			case pb.MsgUpdateEntity:
				resp := &pb.UpdateEntity{}
				err := proto.Unmarshal(buf, resp)
				if err != nil {
					log.Fatal("marshaling error: ", err)
				}
				// Last time we received an update about the world
				lastUpdated = time.Now()
				g.Chars.UpdateFromData(resp)
			case pb.MsgUpdateAll:
				resp := &pb.UpdateAll{}
				err := proto.Unmarshal(buf, resp)
				if err != nil {
					log.Fatal("marshaling error: ", err)
				}
				// Last time we received an update about the world
				lastUpdated = time.Now()
				for _, ue := range resp.Updates {
					g.Chars.UpdateFromData(ue)
				}
			default:
				log.Printf("Unhandled netmsg kind: %s, with data: %v", kind.String(), buf)
			}
		case <-g.Client.disconnect:
			log.Println("lost connection to server")
			break outer
		default:
			// no more messages
			break outer
		}
	}

	g.parseInput()
	for _, char := range g.Chars {
		if char == nil {
			continue
		}
		char.Move()
	}

	g.count++
	return nil
}

func (g *Game) parseInput() {
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
	if g.input.RightPressed != pi.RightPressed ||
		g.input.LeftPressed != pi.LeftPressed ||
		g.input.UpPressed != pi.UpPressed ||
		g.input.DownPressed != pi.DownPressed {
		inputChanged = true
	}
	g.input.RightPressed = pi.RightPressed
	g.input.LeftPressed = pi.LeftPressed
	g.input.UpPressed = pi.UpPressed
	g.input.DownPressed = pi.DownPressed
	if !inputChanged {
		return
	}
	data, err := proto.Marshal(&pi)
	if err != nil {
		log.Println(err)
	}
	g.Client.send <- pb.AddHeader(data, pb.MsgPlayerInput)
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, char := range g.Chars {
		if char == nil {
			continue
		}
		sprite := runnerWaitingFrame
		if char.Vx != 0 || char.Vy != 0 {
			sprite = runnerWalkingFrame
		}
		g.Op.GeoM.Reset()
		if char.Fx < 0 {
			g.Op.GeoM.Scale(-1, 1)
			g.Op.GeoM.Translate(frameWidth, 0)
		}
		g.Op.GeoM.Translate(
			char.Px-frameWidth/2,
			char.Py-frameHeight/2,
		)
		screen.DrawImage(sprite(g.count/g.Speed+char.clockOffset), g.Op)
	}
	msg := fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f\nPing: %dms\n",
		ebiten.CurrentTPS(),
		ebiten.CurrentFPS(),
		g.Client.latency,
	)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
