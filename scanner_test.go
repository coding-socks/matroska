package matroska

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestScanner(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	tests := []struct {
		name     string
		filename string
		source   string
		wantErr  bool

		wantClusterLen int // Result of `mkvinfo -o matroska/test8.mkv | grep '|+ Cluster' | wc -l`
	}{
		{
			name:     "Basic file",
			filename: "test1.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test1.mkv?raw=true",

			wantClusterLen: 11,
		},
		{
			name:     "Non default timecodescale & aspect ratio",
			filename: "test2.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test2.mkv?raw=true",

			wantClusterLen: 47,
		},
		{
			name:     "Header stripping & standard block",
			filename: "test3.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test3.mkv?raw=true",

			wantClusterLen: 47,
		},
		{
			name:     "Live stream recording",
			filename: "test4.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test4.mkv?raw=true",

			wantClusterLen: 36,
		},
		{
			name:     "Multiple audio/subtitles",
			filename: "test5.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test5.mkv?raw=true",

			wantClusterLen: 25,
		},
		{
			name:     "Different EBML head sizes & cue-less seeking",
			filename: "test6.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test6.mkv?raw=true",

			wantClusterLen: 11,
		},
		{
			name:     "Extra unknown/junk elements & damaged",
			filename: "test7.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test7.mkv?raw=true",
			wantErr:  true,

			wantClusterLen: 37,
		},
		{
			name:     "Audio gap",
			filename: "test8.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test8.mkv?raw=true",

			wantClusterLen: 47,
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
			s, err := NewScanner(f)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("%+v\n", s.Info())
			fmt.Printf("%+v\n", s.Tracks())
			seekHead, ok := s.SeekHead()
			fmt.Printf("%+v, %+v\n", ok, seekHead)
			var cluster []Cluster
			for s.Next() {
				cluster = append(cluster, s.Cluster())
			}
			if tt.wantErr != (s.Err() != nil) {
				t.Error(s.Err())
			}
			if got := len(cluster); got != tt.wantClusterLen {
				t.Errorf("len(Cluster) got = %v, want %v", got, tt.wantClusterLen)
			}
			//fmt.Printf("%+v\n", len(b.Cluster))
		})
	}
}
