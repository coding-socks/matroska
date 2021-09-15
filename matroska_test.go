package matroska

import (
	"github.com/coding-socks/ebml"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func Test_init(t *testing.T) {
	found := false
	for _, docType := range ebml.DocTypes() {
		if docType == "matroska" {
			found = true
			break
		}
	}
	if !found {
		t.Error("matroska doctype not found")
	}
}

func downloadTestFile(filename, source string) error {
	resp, err := http.Get(source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(filepath.Join(".", filename))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		source   string
		wantErr  bool
	}{
		{
			name:     "Basic file",
			filename: "test1.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test1.mkv?raw=true",
		},
		{
			name:     "Non default timecodescale & aspect ratio",
			filename: "test2.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test2.mkv?raw=true",
		},
		{
			name:     "Header stripping & standard block",
			filename: "test3.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test3.mkv?raw=true",
		},
		{
			name:     "Live stream recording",
			filename: "test4.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test4.mkv?raw=true",
		},
		{
			name:     "Multiple audio/subtitles",
			filename: "test5.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test5.mkv?raw=true",
		},
		{
			name:     "Different EBML head sizes & cue-less seeking",
			filename: "test6.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test6.mkv?raw=true",
		},
		{
			name:     "Extra unknown/junk elements & damaged",
			filename: "test7.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test7.mkv?raw=true",
		},
		{
			name:     "Audio gap",
			filename: "test8.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test8.mkv?raw=true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join(".", tt.filename))
			if err != nil {
				if !os.IsNotExist(err) {
					t.Fatal(err)
				}
				if err := downloadTestFile(tt.filename, tt.source); err != nil {
					t.Fatal(err)
				}
				if f, err = os.Open(filepath.Join(".", tt.filename)); err != nil {
					t.Fatal(err)
				}
			}
			defer f.Close()
			d := ebml.NewDecoder(f)
			header, err := d.DecodeHeader()
			if err != nil {
				t.Fatal(err)
			}
			log.Printf("%+v", header)
			var b Segment
			if err = d.DecodeBody(&b); err != nil && err != io.EOF {
				t.Error(err)
			}
			log.Printf("%+v", b.Info)
		})
	}
}
