package server

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/kisunji/ebiten-poc/common"
)

type AIData struct {
	Id           int32
	UpPressed    bool
	DownPressed  bool
	LeftPressed  bool
	RightPressed bool
}

type AI struct {
	*common.Char

	id   int32
	stop bool
}

func (w *World) RunAI(ai *AI, aiChan chan AIData) {
	rand.Seed(int64(time.Now().Nanosecond()))
	for {
		if ai.stop {
			break
		}
		if !w.Running {
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)
		w.aiChan <- computeMovement(ai)
		// move for up to 5 seconds
		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
		// wait
		aiChan <- AIData{
			Id:           ai.id,
			UpPressed:    false,
			DownPressed:  false,
			LeftPressed:  false,
			RightPressed: false,
		}
	}
	log.Printf("stopping ai %d\n", ai.id)
}
func computeMovement(ai *AI) AIData {
	biasx := ai.Px/float64(common.ScreenWidth) - .5
	biasy := ai.Py/float64(common.ScreenHeight) - .5
	fx := 0
	if rawx := math.Round(rand.NormFloat64() - biasx); rawx < 0 {
		fx = -1
	} else if rawx > 0 {
		fx = 1
	}
	fy := 0
	if rawy := math.Round(rand.NormFloat64() - biasy); rawy < 0 {
		fy = -1
	} else if rawy > 0 {
		fy = 1
	}

	return AIData{
		Id:           ai.id,
		UpPressed:    fy == -1,
		DownPressed:  fy == 1,
		LeftPressed:  fx == -1,
		RightPressed: fx == 1,
	}
}
