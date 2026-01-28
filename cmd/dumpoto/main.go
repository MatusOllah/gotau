package main

import (
	"fmt"
	"os"

	"github.com/MatusOllah/enczip/zip"

	"github.com/MatusOllah/gotau/sequence/ust"
	"github.com/MatusOllah/gotau/voicebank"
	"golang.org/x/text/encoding/japanese"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <voicebank zip> <input ust>\n", os.Args[0])
		return
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

	vb, err := voicebank.Open(zr, voicebank.WithFileEncoding(japanese.ShiftJIS), voicebank.WithDecodeAssets(false))
	if err != nil {
		panic(err)
	}

	ustFile, err := os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := ustFile.Close(); err != nil {
			panic(err)
		}
	}()

	seq, err := ust.Decode(ustFile)
	if err != nil {
		panic(err)
	}

	var outOto voicebank.Oto
	for i := range seq.Notes {
		cfg := voicebank.LookupConfig{
			Lyric: seq.Notes[i].Lyric,
			Note:  seq.Notes[i].NoteNum,
		}
		if i > 0 {
			cfg.PrevLyric = seq.Notes[i-1].Lyric
		}
		entry, ok := vb.Lookup(cfg)
		if !ok {
			fmt.Fprintf(os.Stderr, "warning: no oto entry found for note #%4d lyric %q\n", i, seq.Notes[i].Lyric)
			continue
		}
		outOto = append(outOto, entry)
	}

	if err := outOto.Encode(os.Stdout); err != nil {
		panic(err)
	}
}
