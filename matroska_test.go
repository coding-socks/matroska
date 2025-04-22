package matroska

import (
	"bytes"
	"flag"
	"github.com/coding-socks/ebml"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
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

type diagnosisVisitor struct{ t *testing.T }

func (v diagnosisVisitor) Visit(el ebml.Element, offset int64, headerSize int, val any) (w ebml.Visitor) {
	v.t.Logf("%s %s at %d, size %d+%d", el.Schema.Name, el.ID, offset, headerSize, el.DataSize)
	return v
}

var flagDiagnosis = flag.Bool("diagnosis", false, "")

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
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	tests := []struct {
		name     string
		filename string
		source   string
		wantErr  bool

		wantHeader ebml.EBML
		wantInfo   Info

		wantClusterLen int // Result of `mkvinfo -o matroska/test8.mkv | grep '|+ Cluster' | wc -l`
	}{
		{
			name:     "Basic file",
			filename: "test1.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test1.mkv?raw=true",

			wantClusterLen: 11,
			wantHeader: ebml.EBML{
				EBMLVersion:        1,
				EBMLReadVersion:    1,
				EBMLMaxIDLength:    4,
				EBMLMaxSizeLength:  8,
				DocType:            "matroska",
				DocTypeVersion:     2,
				DocTypeReadVersion: 2,
				DocTypeExtension:   nil,
			},
			wantInfo: Info{
				SegmentUUID:    &[]byte{0x92, 0x2d, 0x19, 0x32, 0x0f, 0x1e, 0x13, 0xc5, 0xb5, 0x5, 0x63, 0xa, 0xaf, 0xd8, 0x53, 0x36},
				TimestampScale: time.Millisecond,
				Duration: func() *float64 {
					f := float64((1*time.Minute + 27*time.Second + 336*time.Millisecond) / time.Millisecond)
					return &f
				}(),
				DateUTC:    func() *time.Time { t := time.Date(2010, 8, 21, 07, 23, 03, 0, time.UTC); return &t }(),
				MuxingApp:  "libebml2 v0.10.0 + libmatroska2 v0.10.1",
				WritingApp: "mkclean 0.5.5 ru from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.1.1 ('Bouncin' Back') built on Jul  3 2010 22:54:08",
			},
		},
		{
			name:     "Non default timecodescale & aspect ratio",
			filename: "test2.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test2.mkv?raw=true",

			wantClusterLen: 47,
			wantHeader: ebml.EBML{
				EBMLVersion:        1,
				EBMLReadVersion:    1,
				EBMLMaxIDLength:    4,
				EBMLMaxSizeLength:  8,
				DocType:            "matroska",
				DocTypeVersion:     2,
				DocTypeReadVersion: 2,
				DocTypeExtension:   nil,
			},
			wantInfo: Info{
				SegmentUUID:    &[]byte{0x92, 0xb2, 0xce, 0x31, 0x8a, 0x96, 0x50, 0x03, 0x9c, 0x48, 0x2d, 0x67, 0xaa, 0x55, 0xcb, 0x7b},
				TimestampScale: time.Millisecond / 10,
				Duration: func() *float64 {
					f := float64((47*time.Second + 509*time.Millisecond) / time.Millisecond * 10)
					return &f
				}(),
				DateUTC:    func() *time.Time { t := time.Date(2011, 6, 2, 12, 45, 20, 0, time.UTC); return &t }(),
				MuxingApp:  "libebml2 v0.21.0 + libmatroska2 v0.22.1",
				WritingApp: "mkclean 0.8.3 ru from libebml2 v0.10.0 + libmatroska2 v0.10.1 + mkclean 0.5.5 ru from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.1.1 ('Bouncin' Back') built on Jul  3 2010 22:54:08",
			},
		},
		{
			name:     "Header stripping & standard block",
			filename: "test3.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test3.mkv?raw=true",

			wantClusterLen: 47,
			wantHeader: ebml.EBML{
				EBMLVersion:        1,
				EBMLReadVersion:    1,
				EBMLMaxIDLength:    4,
				EBMLMaxSizeLength:  8,
				DocType:            "matroska",
				DocTypeVersion:     2,
				DocTypeReadVersion: 2,
				DocTypeExtension:   nil,
			},
			wantInfo: Info{
				SegmentUUID:    &[]byte{0x99, 0x49, 0x95, 0x82, 0xdf, 0xaf, 0x06, 0xd4, 0xa6, 0x76, 0xd2, 0xe6, 0x4c, 0x02, 0xa5, 0x07},
				TimestampScale: time.Millisecond,
				Duration: func() *float64 {
					f := float64((49*time.Second + 64*time.Millisecond) / time.Millisecond)
					return &f
				}(),
				DateUTC:    func() *time.Time { t := time.Date(2010, 8, 21, 21, 43, 25, 0, time.UTC); return &t }(),
				MuxingApp:  "libebml2 v0.11.0 + libmatroska2 v0.10.1",
				WritingApp: "mkclean 0.5.5 ro from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.1.1 ('Bouncin' Back') built on Jul  3 2010 22:54:08",
			},
		},
		{
			name:     "Live stream recording",
			filename: "test4.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test4.mkv?raw=true",

			wantClusterLen: 36,
			wantHeader: ebml.EBML{
				EBMLVersion:        1,
				EBMLReadVersion:    1,
				EBMLMaxIDLength:    4,
				EBMLMaxSizeLength:  8,
				DocType:            "matroska",
				DocTypeVersion:     1,
				DocTypeReadVersion: 1,
				DocTypeExtension:   nil,
			},
			wantInfo: Info{
				SegmentUUID:    &[]byte{0x8a, 0x70, 0x2f, 0x08, 0x8c, 0xaf, 0x3c, 0x67, 0xba, 0xc9, 0xf5, 0xe6, 0x13, 0x51, 0x2b, 0x09},
				TimestampScale: time.Millisecond,
				DateUTC:        func() *time.Time { t := time.Date(2010, 8, 21, 8, 42, 15, 0, time.UTC); return &t }(),
				MuxingApp:      "libebml2 v0.10.1 + libmatroska2 v0.10.1",
				WritingApp:     "mkclean 0.5.5 l from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.1.1 ('Bouncin' Back') built on Jul  3 2010 22:54:08",
			},
		},
		{
			name:     "Multiple audio/subtitles",
			filename: "test5.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test5.mkv?raw=true",

			wantClusterLen: 25,
			wantHeader: ebml.EBML{
				EBMLVersion:        1,
				EBMLReadVersion:    1,
				EBMLMaxIDLength:    4,
				EBMLMaxSizeLength:  8,
				DocType:            "matroska",
				DocTypeVersion:     2,
				DocTypeReadVersion: 2,
				DocTypeExtension:   nil,
			},
			wantInfo: Info{
				SegmentUUID:    &[]byte{0x9d, 0x51, 0x6a, 0x0f, 0x92, 0x7a, 0x12, 0xd2, 0x86, 0xe1, 0x50, 0x2d, 0x23, 0xd0, 0xfd, 0xb0},
				TimestampScale: time.Millisecond,
				Duration: func() *float64 {
					f := float64((46*time.Second + 665*time.Millisecond) / time.Millisecond)
					return &f
				}(),
				DateUTC:    func() *time.Time { t := time.Date(2010, 8, 21, 18, 6, 43, 0, time.UTC); return &t }(),
				MuxingApp:  "libebml v1.0.0 + libmatroska v1.0.0",
				WritingApp: "mkvmerge v4.0.0 ('The Stars were mine') built on Jun  6 2010 16:18:42",
			},
		},
		{
			name:     "Different EBML head sizes & cue-less seeking",
			filename: "test6.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test6.mkv?raw=true",

			wantClusterLen: 11,
			wantHeader: ebml.EBML{
				EBMLVersion:        1,
				EBMLReadVersion:    1,
				EBMLMaxIDLength:    4,
				EBMLMaxSizeLength:  8,
				DocType:            "matroska",
				DocTypeVersion:     2,
				DocTypeReadVersion: 2,
				DocTypeExtension:   nil,
			},
			wantInfo: Info{
				SegmentUUID:    &[]byte{0x84, 0xfa, 0x5b, 0x60, 0x97, 0x2a, 0x16, 0x5b, 0x85, 0x27, 0x66, 0xe7, 0xe5, 0xb0, 0xa2, 0x83},
				TimestampScale: time.Millisecond,
				Duration: func() *float64 {
					f := float64((1*time.Minute + 27*time.Second + 336*time.Millisecond) / time.Millisecond)
					return &f
				}(),
				DateUTC:    func() *time.Time { t := time.Date(2010, 8, 21, 16, 31, 55, 0, time.UTC); return &t }(),
				MuxingApp:  "libebml2 v0.10.1 + libmatroska2 v0.10.1",
				WritingApp: "mkclean 0.5.5 r from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.0.0 ('The Stars were mine') built on Jun  6 2010 16:18:42",
			},
		},
		{
			name:     "Extra unknown/junk elements & damaged",
			filename: "test7.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test7.mkv?raw=true",
			wantErr:  true,

			wantClusterLen: 37,
			wantHeader: ebml.EBML{
				EBMLVersion:        1,
				EBMLReadVersion:    1,
				EBMLMaxIDLength:    4,
				EBMLMaxSizeLength:  8,
				DocType:            "matroska",
				DocTypeVersion:     2,
				DocTypeReadVersion: 2,
				DocTypeExtension:   nil,
			},
			wantInfo: Info{
				SegmentUUID:    &[]byte{0xb9, 0x82, 0x1f, 0xa6, 0x51, 0xb1, 0xe2, 0x47, 0xb6, 0x79, 0x26, 0x0d, 0xd2, 0xe7, 0xe3, 0x71},
				TimestampScale: time.Millisecond,
				Duration: func() *float64 {
					f := float64((37*time.Second + 43*time.Millisecond) / time.Millisecond)
					return &f
				}(),
				DateUTC:    func() *time.Time { t := time.Date(2010, 8, 21, 17, 0, 23, 0, time.UTC); return &t }(),
				MuxingApp:  "libebml2 v0.10.1 + libmatroska2 v0.10.1",
				WritingApp: "mkclean 0.5.5 r from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.0.0 ('The Stars were mine') built on Jun  6 2010 16:18:42",
			},
		},
		{
			name:     "Audio gap",
			filename: "test8.mkv",
			source:   "https://github.com/Matroska-Org/matroska-test-files/blob/master/test_files/test8.mkv?raw=true",

			wantClusterLen: 47,
			wantHeader: ebml.EBML{
				EBMLVersion:        1,
				EBMLReadVersion:    1,
				EBMLMaxIDLength:    4,
				EBMLMaxSizeLength:  8,
				DocType:            "matroska",
				DocTypeVersion:     2,
				DocTypeReadVersion: 2,
				DocTypeExtension:   nil,
			},
			wantInfo: Info{
				SegmentUUID:    &[]byte{0x8a, 0x1e, 0x00, 0xbb, 0x51, 0x66, 0x13, 0x80, 0xaf, 0x10, 0xd1, 0xfe, 0x09, 0x97, 0x0b, 0x5d},
				TimestampScale: time.Millisecond,
				Duration: func() *float64 {
					f := float64((47*time.Second + 341*time.Millisecond) / time.Millisecond)
					return &f
				}(),
				DateUTC:    func() *time.Time { t := time.Date(2010, 8, 21, 17, 22, 14, 0, time.UTC); return &t }(),
				MuxingApp:  "libebml2 v0.10.1 + libmatroska2 v0.10.1",
				WritingApp: "mkclean 0.5.5 r from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.0.0 ('The Stars were mine') built on Jun  6 2010 16:18:42",
			},
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
			if *flagDiagnosis {
				d.SetVisitor(diagnosisVisitor{t: t})
			}
			header, err := d.DecodeHeader()
			if err != nil {
				t.Fatal(err)
			}
			if got := *header; !reflect.DeepEqual(got, tt.wantHeader) {
				t.Errorf("DecodeHeader(): got %#v, want %#v", got, tt.wantHeader)
			}
			var b Segment
			if err = d.DecodeBody(&b); tt.wantErr != (err != nil && err != io.EOF) {
				t.Errorf("DecodeBody() error = %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr {
				t.Log(err)
			}
			if got := b.Info; !reflect.DeepEqual(got, tt.wantInfo) {
				t.Errorf("DecodeBody().Info: got %#v, want %#v", got, tt.wantInfo)
			}
			if got := len(b.Cluster); got != tt.wantClusterLen {
				t.Errorf("len(DecodeBody().Cluster) got = %v, want %v", got, tt.wantClusterLen)
			}
		})
	}
}

func TestFrames(t *testing.T) {
	type args struct {
		lacing uint8
		data   []byte
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{
			name: "Lacing No",
			args: args{lacing: LacingFlagNo, data: []byte{0x01, 0x02, 0x03, 0x04}},
			want: [][]byte{{0x01, 0x02, 0x03, 0x04}},
		},
		{
			name: "Lacing Xiph",
			args: args{lacing: LacingFlagXiph, data: bytes.Join([][]byte{
				{0x02},
				{0xFF, 0xFF, 0xFF, 0x23},
				{0xFF, 0xF5},
				bytes.Repeat([]byte{0xFF}, 800),
				bytes.Repeat([]byte{0xFE}, 500),
				bytes.Repeat([]byte{0xFD}, 1000),
			}, nil)},
			want: [][]byte{
				bytes.Repeat([]byte{0xFF}, 800),
				bytes.Repeat([]byte{0xFE}, 500),
				bytes.Repeat([]byte{0xFD}, 1000),
			},
		},
		{
			name: "Lacing EBML",
			args: args{lacing: LacingFlagEBML, data: bytes.Join([][]byte{
				{0x02},
				{0x43, 0x20},
				{0x5E, 0xD3},
				bytes.Repeat([]byte{0xFF}, 800),
				bytes.Repeat([]byte{0xFE}, 500),
				bytes.Repeat([]byte{0xFD}, 1000),
			}, nil)},
			want: [][]byte{
				bytes.Repeat([]byte{0xFF}, 800),
				bytes.Repeat([]byte{0xFE}, 500),
				bytes.Repeat([]byte{0xFD}, 1000),
			},
		},
		{
			name: "Lacing Fixed",
			args: args{lacing: LacingFlagFixedSize, data: bytes.Join([][]byte{
				{0x02},
				bytes.Repeat([]byte{0xFF}, 800),
				bytes.Repeat([]byte{0xFE}, 800),
				bytes.Repeat([]byte{0xFD}, 800),
			}, nil)},
			want: [][]byte{
				bytes.Repeat([]byte{0xFF}, 800),
				bytes.Repeat([]byte{0xFE}, 800),
				bytes.Repeat([]byte{0xFD}, 800),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Frames(tt.args.lacing, tt.args.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Frames() = %x,\nwant %x", got, tt.want)
			}
		})
	}
}
