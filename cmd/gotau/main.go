package main

import (
	"fmt"
	"os"

	"github.com/SladkyCitron/enczip/zip"
	"github.com/davecgh/go-spew/spew"

	"github.com/SladkyCitron/gotau/voicebank"
	"golang.org/x/text/encoding/japanese"
)

func main() {
	zr, err := zip.OpenReader(os.Args[1], japanese.ShiftJIS)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := zr.Close(); err != nil {
			panic(err)
		}
	}()

	vb, err := voicebank.Open(zr, voicebank.WithFileEncoding(japanese.ShiftJIS), voicebank.WithDecodeAssets(true))
	if err != nil {
		panic(err)
	}

	spew.Dump(vb)

	// render image
	if vb.CharacterInfo.Image.Image != nil {
		fmt.Println()
		width := vb.CharacterInfo.Image.Image.Bounds().Dx()
		height := vb.CharacterInfo.Image.Image.Bounds().Dy()
		img := vb.CharacterInfo.Image.Image
		buf := make([]byte, 0, height*(width+1))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				buf = append(buf, fmt.Appendf(nil, "\033[48;2;%d;%d;%dm \033[0m", uint8(r>>8), uint8(g>>8), uint8(b>>8))...)
			}
			buf = append(buf, '\n')
		}
		_, err := os.Stdout.Write(buf)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println()

	// print readme
	fmt.Println(vb.Readme)
}
