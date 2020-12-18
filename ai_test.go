package main

import (
	"testing"
)

func Test_computeMovement(t *testing.T) {
	type args struct {
		px float64
		py float64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "top left corner",
			args: args{px: 0, py: 0},
		},
		{
			name: "top right corner",
			args: args{px: screenWidth, py:0},
		},
		{
			name: "bottom left corner",
			args: args{px: 0, py: screenHeight},
		},
		{
			name: "bottom right corner",
			args: args{px: screenWidth, py: screenHeight},
		},
		{
			name: "centre",
			args: args{px: screenWidth/2, py: screenHeight/2},
		},
	}
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i:=0; i < 10; i++ {
				computeMovement(tt.args.px, tt.args.py)
			}
		})
	}
}
