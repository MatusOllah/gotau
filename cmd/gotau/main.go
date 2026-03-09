package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/SladkyCitron/enczip/zip"
	"github.com/SladkyCitron/gotau"
	"github.com/SladkyCitron/resona/afmt"
	"github.com/SladkyCitron/resona/aio"
	"github.com/SladkyCitron/resona/codec/wav"
	"github.com/SladkyCitron/resona/freq"

	"github.com/SladkyCitron/gotau/sequence/ust"
	"github.com/SladkyCitron/gotau/voicebank"
	"golang.org/x/text/encoding/japanese"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s voicebank.zip song.ust output.wav", os.Args[0])
		os.Exit(1)
	}

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

	inFile, err := os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := inFile.Close(); err != nil {
			panic(err)
		}
	}()

	ustFile, err := ust.Decode(inFile)
	if err != nil {
		panic(err)
	}

	seq := ustFile.Sequence()

	synth := gotau.New(vb, seq)

	outFile, err := os.Create(os.Args[3])
	if err != nil {
		panic(err)
	}

	enc, err := wav.NewEncoder(outFile, afmt.Format{44100 * freq.Hertz, 1}, afmt.SampleFormat{16, afmt.SampleEncodingInt, binary.LittleEndian}, wav.FormatInt)
	if err != nil {
		panic(err)
	}

	if _, err := aio.Copy(enc, synth); err != nil {
		panic(err)
	}

	if err := enc.Close(); err != nil {
		panic(err)
	}

	if err := outFile.Close(); err != nil {
		panic(err)
	}

	// not getting rid of this, I'm very proud of this one :)
	/*
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
			fmt.Println()
		}
	*/
}
