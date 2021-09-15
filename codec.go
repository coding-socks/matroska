package matroska

import "strings"

const (
	CodecTypeVideo    = CodecType("Video")
	CodecTypeAudio    = CodecType("Audio")
	CodecTypeSubtitle = CodecType("Subtitle")
	CodecTypeButton   = CodecType("Button")
)

type CodecType string

const (
	AudioCodecAAC       = "A_AAC"
	AudioCodecAAC_2LC   = "A_AAC/MPEG2/LC"
	AudioCodecAAC_2MAIN = "A_AAC/MPEG2/MAIN"
	AudioCodecAAC_2SBR  = "A_AAC/MPEG2/LC/SBR"
	AudioCodecAAC_2SSR  = "A_AAC/MPEG2/SSR"
	AudioCodecAAC_4LC   = "A_AAC/MPEG4/LC"
	AudioCodecAAC_4LTP  = "A_AAC/MPEG4/LTP"
	AudioCodecAAC_4MAIN = "A_AAC/MPEG4/MAIN"
	AudioCodecAAC_4SBR  = "A_AAC/MPEG4/LC/SBR"
	AudioCodecAAC_4SSR  = "A_AAC/MPEG4/SSR"
	AudioCodecAC3       = "A_AC3"
	AudioCodecACM       = "A_MS/ACM"
	AudioCodecALAC      = "A_ALAC"
	AudioCodecDTS       = "A_DTS"
	AudioCodecEAC3      = "A_EAC3"
	AudioCodecFLAC      = "A_FLAC"
	AudioCodecMLP       = "A_MLP"
	AudioCodecMP2       = "A_MPEG/L2"
	AudioCodecMP3       = "A_MPEG/L3"
	AudioCodecOPUS      = "A_OPUS"
	AudioCodecPCM       = "A_PCM/INT/LIT"
	AudioCodecPCM_BE    = "A_PCM/INT/BIG"
	AudioCodecPCM_FLOAT = "A_PCM/FLOAT/IEEE"
	AudioCodecQDMC      = "A_QUICKTIME/QDMC"
	AudioCodecQDMC2     = "A_QUICKTIME/QDM2"
	AudioCodecQUICKTIME = "A_QUICKTIME"
	AudioCodecTRUEHD    = "A_TRUEHD"
	AudioCodecTTA       = "A_TTA1"
	AudioCodecVORBIS    = "A_VORBIS"
	AudioCodecWAVPACK4  = "A_WAVPACK4"

	VideoCodecAV1          = "V_AV1"
	VideoCodecDIRAC        = "V_DIRAC"
	VideoCodecMPEG1        = "V_MPEG1"
	VideoCodecMPEG2        = "V_MPEG2"
	VideoCodecMPEG4_AP     = "V_MPEG4/ISO/AP"
	VideoCodecMPEG4_ASP    = "V_MPEG4/ISO/ASP"
	VideoCodecMPEG4_AVC    = "V_MPEG4/ISO/AVC"
	VideoCodecMPEG4_SP     = "V_MPEG4/ISO/SP"
	VideoCodecMPEGH_HEVC   = "V_MPEGH/ISO/HEVC"
	VideoCodecMSCOMP       = "V_MS/VFW/FOURCC"
	VideoCodecPRORES       = "V_PRORES"
	VideoCodecQUICKTIME    = "V_QUICKTIME"
	VideoCodecREALV1       = "V_REAL/RV10"
	VideoCodecREALV2       = "V_REAL/RV20"
	VideoCodecREALV3       = "V_REAL/RV30"
	VideoCodecREALV4       = "V_REAL/RV40"
	VideoCodecTHEORA       = "V_THEORA"
	VideoCodecUNCOMPRESSED = "V_UNCOMPRESSED"
	VideoCodecVP8          = "V_VP8"
	VideoCodecVP9          = "V_VP9"

	SubtitleCodecDVBSUB      = "S_DVBSUB"
	SubtitleCodecHDMV_PGS    = "S_HDMV/PGS"
	SubtitleCodecHDMV_TEXTST = "S_HDMV/TEXTST"
	SubtitleCodecKATE        = "S_KATE"
	SubtitleCodecTEXTASCII   = "S_TEXT/ASCII"
	SubtitleCodecTEXTASS     = "S_TEXT/ASS"
	// Deprecated: use SubtitleCodecTEXTASS instead
	SubtitleCodecASS     = "S_ASS"
	SubtitleCodecTEXTSSA = "S_TEXT/SSA"
	// Deprecated: use SubtitleCodecTEXTSSA instead
	SubtitleCodecSSA        = "S_SSA"
	SubtitleCodecTEXTUSF    = "S_TEXT/USF"
	SubtitleCodecTEXTUTF8   = "S_TEXT/UTF8"
	SubtitleCodecTEXTWEBVTT = "S_TEXT/WEBVTT"
	SubtitleCodecVOBSUB     = "S_VOBSUB"
	SubtitleCodecVOBSUBZLIB = "S_VOBSUB/ZLIB"

	ButtonCodecVOBBTN = "B_VOBBTN"
)

func CodecID(s string) (prefix CodecType, major, suffix string) {
	if len(s) < 2 && s[1] != '_' {
		panic("invalid codec id")
	}
	switch s[:2] {
	case "V_":
		prefix = CodecTypeVideo
	case "A_":
		prefix = CodecTypeAudio
	case "S_":
		prefix = CodecTypeSubtitle
	case "B_":
		prefix = CodecTypeButton
	}
	j := strings.Index(s, "/")
	if j == -1 {
		return prefix, s[2:], ""
	}
	return prefix, s[2:j], s[j+1:]
}
