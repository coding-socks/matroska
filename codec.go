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
	AudioCodecAAC             = "A_AAC"
	AudioCodecAAC_2LC         = "A_AAC/MPEG2/LC"
	AudioCodecAAC_2MAIN       = "A_AAC/MPEG2/MAIN"
	AudioCodecAAC_2SBR        = "A_AAC/MPEG2/LC/SBR"
	AudioCodecAAC_2SSR        = "A_AAC/MPEG2/SSR"
	AudioCodecAAC_4LC         = "A_AAC/MPEG4/LC"
	AudioCodecAAC_4LTP        = "A_AAC/MPEG4/LTP"
	AudioCodecAAC_4MAIN       = "A_AAC/MPEG4/MAIN"
	AudioCodecAAC_4SBR        = "A_AAC/MPEG4/LC/SBR"
	AudioCodecAAC_4SSR        = "A_AAC/MPEG4/SSR"
	AudioCodecAC3             = "A_AC3"
	AudioCodecAC3_BSID9       = "A_AC3/BSID9"
	AudioCodecAC3_BSID10      = "A_AC3/BSID10"
	AudioCodecALAC            = "A_ALAC"
	AudioCodecATRAC_AT1       = "A_ATRAC/AT1"
	AudioCodecDTS             = "A_DTS"
	AudioCodecDTS_EXPRESS     = "A_DTS/EXPRESS"
	AudioCodecDTS_LOSSLESS    = "A_DTS/LOSSLESS"
	AudioCodecEAC3            = "A_EAC3"
	AudioCodecFLAC            = "A_FLAC"
	AudioCodecMLP             = "A_MLP"
	AudioCodecMPC             = "A_MPC"
	AudioCodecMP1             = "A_MPEG/L1"
	AudioCodecMP2             = "A_MPEG/L2"
	AudioCodecMP3             = "A_MPEG/L3"
	AudioCodecMS_ACM          = "A_MS/ACM"
	AudioCodecOPUS            = "A_OPUS"
	AudioCodecPCM             = "A_PCM/INT/LIT"
	AudioCodecPCM_BE          = "A_PCM/INT/BIG"
	AudioCodecPCM_FLOAT       = "A_PCM/FLOAT/IEEE"
	AudioCodecQUICKTIME       = "A_QUICKTIME"
	AudioCodecQUICKTIME_QDMC  = "A_QUICKTIME/QDMC"
	AudioCodecQUICKTIME_QDMC2 = "A_QUICKTIME/QDM2"
	AudioCodecREAL_14         = "A_REAL/14_4"
	AudioCodecREAL_28         = "A_REAL/28_8"
	AudioCodecREAL_COOK       = "A_REAL/COOK"
	AudioCodecREAL_SIPR       = "A_REAL/SIPR"
	AudioCodecREAL_RALF       = "A_REAL/RALF"
	AudioCodecREAL_ATRC       = "A_REAL/ATRC"
	AudioCodecTRUEHD          = "A_TRUEHD"
	AudioCodecTTA             = "A_TTA1"
	AudioCodecVORBIS          = "A_VORBIS"
	AudioCodecWAVPACK4        = "A_WAVPACK4"

	VideoCodecAV1            = "V_AV1"
	VideoCodecAVS2           = "V_AVS2"
	VideoCodecAVS3           = "V_AVS3"
	VideoCodecDIRAC          = "V_DIRAC"
	VideoCodecFFV1           = "V_FFV1"
	VideoCodecMPEG1          = "V_MPEG1"
	VideoCodecMPEG2          = "V_MPEG2"
	VideoCodecMPEG4_ISO_AP   = "V_MPEG4/ISO/AP"
	VideoCodecMPEG4_ISO_ASP  = "V_MPEG4/ISO/ASP"
	VideoCodecMPEG4_ISO_AVC  = "V_MPEG4/ISO/AVC"
	VideoCodecMPEG4_ISO_SP   = "V_MPEG4/ISO/SP"
	VideoCodecMPEG4_MS_V3    = "V_MPEG4/MS/V3"
	VideoCodecMPEGH_ISO_HEVC = "V_MPEGH/ISO/HEVC"
	VideoCodecMSCOMP         = "V_MS/VFW/FOURCC"
	VideoCodecPRORES         = "V_PRORES"
	VideoCodecQUICKTIME      = "V_QUICKTIME"
	VideoCodecREALV1         = "V_REAL/RV10"
	VideoCodecREALV2         = "V_REAL/RV20"
	VideoCodecREALV3         = "V_REAL/RV30"
	VideoCodecREALV4         = "V_REAL/RV40"
	VideoCodecTHEORA         = "V_THEORA"
	VideoCodecUNCOMPRESSED   = "V_UNCOMPRESSED"
	VideoCodecVP8            = "V_VP8"
	VideoCodecVP9            = "V_VP9"

	SubtitleCodecDVBSUB      = "S_DVBSUB"
	SubtitleCodecHDMV_PGS    = "S_HDMV/PGS"
	SubtitleCodecHDMV_TEXTST = "S_HDMV/TEXTST"
	SubtitleCodecIMAGE_BMP   = "S_IMAGE/BMP"
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
