package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/coding-socks/matroska/internal/riff"
	"io"
	"os"
)

func analyzeRIFFPackages(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fileType, r, err := riff.NewReader(f)
	if err != nil {
		return fmt.Errorf("unable to decode RIFF header: %w", err)
	}

	w := os.Stdout
	fmt.Fprintf(w, "%s\n", fileType)
	if err := printRIFFDetails(w, r, "  "); err != nil {
		return err
	}
	return nil
}

var buf bytes.Buffer

func printRIFFDetails(w io.Writer, r *riff.Reader, indent string) error {
	for {
		id, l, rr, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if id == riff.LIST {
			listType, rr, err := riff.NewListReader(l, rr)
			if err != nil {
				return fmt.Errorf("unable to decode list: %w", err)
			}
			fmt.Fprintf(w, "%s%s listSize %d listType %s\n", indent, id, l, listType)
			if err := printRIFFDetails(w, rr, indent+"  "); err != nil {
				return err
			}
		} else {
			buf.Reset()
			_, err := io.Copy(&buf, rr)
			if err != nil {
				return fmt.Errorf("unable to decode data: %w", err)
			}
			fmt.Fprintf(w, "%s%s ckSize %d\n", indent, id, l)
			fmt.Fprintf(w, "%s  %.30s\n", indent, hex.EncodeToString(buf.Bytes()))
		}
	}

	return nil
}
