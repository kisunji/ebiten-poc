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

	id      int32
	stop    bool
	killSig chan struct{}
}

// RunAI should be run in a goroutine
func (w *World) RunAI(ai *AI, aiChan chan AIData) {
	moveTimer := time.NewTimer(time.Duration(rand.Intn(5000)) * time.Millisecond)
	sleepTimer := time.NewTimer(time.Duration(rand.Intn(10000)) * time.Millisecond)
	sleepTimer.Stop()
	for {
		select {
		case <-moveTimer.C:
			aiChan <- nextMovement(ai)
			sleepTimer.Reset(time.Duration(rand.Intn(5000)) * time.Millisecond)
		case <-sleepTimer.C:
			aiChan <- AIData{
				Id:           ai.id,
				UpPressed:    false,
				DownPressed:  false,
				LeftPressed:  false,
				RightPressed: false,
			}
			moveTimer.Reset(time.Duration(rand.Intn(5000)) * time.Millisecond)
		case <-ai.killSig:
			log.Printf("stopping ai %d\n", ai.id)
			return
		}
	}
}

func nextMovement(ai *AI) AIData {
	aiData := AIData{Id: ai.id}
	biasx := (ai.Px/float64(common.ScreenWidth) - .5) * 2
	biasy := (ai.Py/float64(common.ScreenHeight) - .5) * 2
	if x := math.Round(rand.NormFloat64()*.5 - biasx); x < 0 {
		aiData.LeftPressed = true
	} else if x > 0 {
		aiData.RightPressed = true
	}

	if y := math.Round(rand.NormFloat64()*.5 - biasy); y < 0 {
		aiData.UpPressed = true
	} else if y > 0 {
		aiData.DownPressed = true
	}
	return aiData
}
