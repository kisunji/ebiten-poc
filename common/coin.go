package common

import "math/rand"

type Coin struct {
	Px, Py       float64
	PickupRadius float64
	PickedUp     bool
	FrameOffset  int
}

func NewCoin() *Coin {
	return &Coin{
		Px:           float64(ScreenPadding + rand.Intn(ScreenWidth-ScreenPadding*5)),
		Py:           float64(ScreenPadding + rand.Intn(ScreenHeight-ScreenPadding*5)),
		PickupRadius: 10.0,
		FrameOffset:  rand.Intn(3),
	}
}
