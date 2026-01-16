package main

import (
	"fmt"
	"os"

	"github.com/MatusOllah/enczip/zip"

	"github.com/MatusOllah/gotau/voicebank"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/text/encoding/japanese"
)

func main() {
	/*
			s := `[#VERSION]
		UST Version1.2
		[#SETTING]
		Tempo=120
		Tracks=1
		Project=C:\Users\matus\Desktop\test.ust
		VoiceDir=C:\Users\matus\Documents\OpenUtau\Singers\重音テト OU用日本語統合ライブラリー
		CacheDir=C:\Users\matus\Documents\OpenUtau\Cache
		Mode2=True
		[#0000]
		Length=720
		Lyric=a
		NoteNum=69
		PreUtterance=
		Velocity=100
		Intensity=100
		Modulation=0
		Flags=g0B0H0P86
		PBS=-40;0
		PBW=65
		PBY=0
		PBM=,
		[#TRACKEND]`

			f, err := ust.Decode(strings.NewReader(s))
			if err != nil {
				panic(err)
			}

			fmt.Printf("f: %+v\n", f)
	*/

	zr, err := zip.OpenReader(os.Args[1], japanese.ShiftJIS)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := zr.Close(); err != nil {
			panic(err)
		}
	}()

	vb, err := voicebank.Open(zr, voicebank.WithFileEncoding(japanese.ShiftJIS), voicebank.WithDecodeAssets(false))
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
}
