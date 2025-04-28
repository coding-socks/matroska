package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/coding-socks/matroska/internal/avi"
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
	switch fileType {
	default:
		if err := printRIFFDetails(w, r, "  "); err != nil {
			return err
		}
	case avi.AVI:
		if err := printAVIDetails(w, r, "  "); err != nil {
			return err
		}
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
		buf.Reset()
		switch id {
		default:
			_, err := io.Copy(&buf, rr)
			if err != nil {
				return fmt.Errorf("unable to decode data: %w", err)
			}
			fmt.Fprintf(w, "%s%s ckSize %d\n", indent, id, l)
			fmt.Fprintf(w, "%s  %.32s\n", indent, hex.EncodeToString(buf.Bytes()))
		case riff.JUNK:
			_, err := io.Copy(io.Discard, rr)
			if err != nil {
				return fmt.Errorf("unable to decode data: %w", err)
			}
			fmt.Fprintf(w, "%s%s ckSize %d\n", indent, id, l)
		case riff.LIST:
			listType, rr, err := riff.NewListReader(l, rr)
			if err != nil {
				return fmt.Errorf("unable to decode list: %w", err)
			}
			fmt.Fprintf(w, "%s%s listSize %d listType %s\n", indent, id, l, listType)
			if err := printRIFFDetails(w, rr, indent+"  "); err != nil {
				return err
			}
		}
	}

	return nil
}

func printAVIDetails(w io.Writer, r *riff.Reader, indent string) error {
	for {
		id, l, rr, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		buf.Reset()
		switch id {
		default:
			_, err := io.Copy(&buf, rr)
			if err != nil {
				return fmt.Errorf("unable to decode data: %w", err)
			}
			fmt.Fprintf(w, "%s%s ckSize %d\n", indent, id, l)
			fmt.Fprintf(w, "%s  %.32s\n", indent, hex.EncodeToString(buf.Bytes()))
		case riff.JUNK:
			_, err := io.Copy(io.Discard, rr)
			if err != nil {
				return fmt.Errorf("unable to decode data: %w", err)
			}
			fmt.Fprintf(w, "%s%s ckSize %d\n", indent, id, l)
		case riff.LIST:
			listType, rr, err := riff.NewListReader(l, rr)
			if err != nil {
				return fmt.Errorf("unable to decode list: %w", err)
			}
			fmt.Fprintf(w, "%s%s listSize %d listType %s\n", indent, id, l, listType)
			if err := printAVIDetails(w, rr, indent+"  "); err != nil {
				return err
			}
		case avi.ChunkAVIH:
			_, err := io.Copy(&buf, rr)
			if err != nil {
				return fmt.Errorf("unable to decode data: %w", err)
			}
			fmt.Fprintf(w, "%s%s ckSize %d\n", indent, id, l)
			mh := avi.MainHeader(buf.Bytes())
			fmt.Fprintf(w, "%s  %d\n", indent, mh.MicroSecPerFrame())
			fmt.Fprintf(w, "%s  %s\n", indent, hex.EncodeToString(buf.Bytes()))
		case /*avi.ChunkAVIH, */ avi.ChunkSTRH, avi.ChunkSTRF, avi.ChunkSTRD, avi.ChunkSTRN:
			_, err := io.Copy(&buf, rr)
			if err != nil {
				return fmt.Errorf("unable to decode data: %w", err)
			}
			fmt.Fprintf(w, "%s%s ckSize %d\n", indent, id, l)
			fmt.Fprintf(w, "%s  %s\n", indent, hex.EncodeToString(buf.Bytes()))
		case avi.ChunkIDX1:
			_, err := io.Copy(&buf, rr)
			if err != nil {
				return fmt.Errorf("unable to decode data: %w", err)
			}
			fmt.Fprintf(w, "%s%s ckSize %d\n", indent, id, l)
			b := buf.Bytes()
			for len(b) >= 16 {
				fmt.Fprintf(w, "%s  %s\n", indent, riff.FourCC(b[0:4]))
				fmt.Fprintf(w, "%s    flags %x\n", indent, binary.LittleEndian.Uint32(b[4:8]))
				fmt.Fprintf(w, "%s    offset %d\n", indent, binary.LittleEndian.Uint32(b[8:12]))
				fmt.Fprintf(w, "%s    size %d\n", indent, binary.LittleEndian.Uint32(b[12:16]))
				b = b[16:]
			}
		case riff.FourCC{'I', 'S', 'F', 'T'}:
			_, err := io.Copy(&buf, rr)
			if err != nil {
				return fmt.Errorf("unable to decode data: %w", err)
			}
			fmt.Fprintf(w, "%s  %s\n", indent, buf.String())
		}
	}

	return nil
}
