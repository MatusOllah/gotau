package main

import (
	"archive/zip"
	"os"

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

	zr, err := zip.OpenReader(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := zr.Close(); err != nil {
			panic(err)
		}
	}()

	vb, err := voicebank.Open(zr, voicebank.WithFileEncoding(japanese.ShiftJIS))
	if err != nil {
		panic(err)
	}

	spew.Dump(vb)

}
