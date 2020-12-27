//go:generate protoc --go_out=./pb --go_opt=paths=source_relative message.proto
//go:generate file2byteslice -input sprites.png -output game/sprite_png.go -package game -var sprite_png

package main