package game

import "time"

type Debouncer struct {
	input    chan []byte
	output   chan<- []byte
	duration time.Duration
	running  bool
}

func (d *Debouncer) debounce() {
	d.running = true
	var msg []byte
	timer := time.NewTimer(d.duration)
	for {
		select {
		case msg = <-d.input:
			timer.Reset(d.duration)
		case <-timer.C:
			if len(msg) > 0 {
				d.output <- msg
			}
		}
	}
}
