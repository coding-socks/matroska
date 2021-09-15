package matroska

import "testing"

func TestCodecID(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name       string
		args       args
		wantPrefix CodecType
		wantMajor  string
		wantSuffix string
	}{
		{
			name:       "AAC_2LC",
			args:       args{s: AudioCodecAAC_2LC},
			wantPrefix: CodecTypeAudio, wantMajor: "AAC", wantSuffix: "MPEG2/LC",
		},
		{
			name:       "AV1",
			args:       args{s: VideoCodecAV1},
			wantPrefix: CodecTypeVideo, wantMajor: "AV1", wantSuffix: "",
		},
		{
			name:       "TEXTASS",
			args:       args{s: SubtitleCodecTEXTASS},
			wantPrefix: CodecTypeSubtitle, wantMajor: "TEXT", wantSuffix: "ASS",
		},
		{
			name:       "VOBBTN",
			args:       args{s: ButtonCodecVOBBTN},
			wantPrefix: CodecTypeButton, wantMajor: "VOBBTN", wantSuffix: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefix, gotMajor, gotSuffix := CodecID(tt.args.s)
			if gotPrefix != tt.wantPrefix {
				t.Errorf("CodecID() gotPrefix = %v, want %v", gotPrefix, tt.wantPrefix)
			}
			if gotMajor != tt.wantMajor {
				t.Errorf("CodecID() gotMajor = %v, want %v", gotMajor, tt.wantMajor)
			}
			if gotSuffix != tt.wantSuffix {
				t.Errorf("CodecID() gotSuffix = %v, want %v", gotSuffix, tt.wantSuffix)
			}
		})
	}
}
