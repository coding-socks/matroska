package main

import (
	"github.com/coding-socks/matroska/internal/ogg"
	"io"
	"log"
	"os"
)

func analyzeOggPackages(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	d := ogg.NewDecoder(f)
	streams, err := d.DecodeStreams()
	if err != nil {
		return err
	}
	log.Println("Streams:", streams)
	for {
		page, err := d.DecodePage()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		log.Printf("%d GranPos %d, Len %d, SegTabLen %d", page.SequenceNum, page.GranulePosition, page.FullLen, len(page.SegmentTable))
	}
	return nil
}
