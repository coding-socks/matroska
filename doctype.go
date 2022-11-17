// Code generated by go run make_doctype.go. DO NOT EDIT.

package matroska

import (
	_ "embed"
	"time"
)

//go:embed ebml_matroska.xml
var docType []byte

var (
	IDSegment                     = "0x18538067"
	IDSeekHead                    = "0x114D9B74"
	IDSeek                        = "0x4DBB"
	IDSeekID                      = "0x53AB"
	IDSeekPosition                = "0x53AC"
	IDInfo                        = "0x1549A966"
	IDSegmentUUID                 = "0x73A4"
	IDSegmentFilename             = "0x7384"
	IDPrevUUID                    = "0x3CB923"
	IDPrevFilename                = "0x3C83AB"
	IDNextUUID                    = "0x3EB923"
	IDNextFilename                = "0x3E83BB"
	IDSegmentFamily               = "0x4444"
	IDChapterTranslate            = "0x6924"
	IDChapterTranslateID          = "0x69A5"
	IDChapterTranslateCodec       = "0x69BF"
	IDChapterTranslateEditionUID  = "0x69FC"
	IDTimestampScale              = "0x2AD7B1"
	IDDuration                    = "0x4489"
	IDDateUTC                     = "0x4461"
	IDTitle                       = "0x7BA9"
	IDMuxingApp                   = "0x4D80"
	IDWritingApp                  = "0x5741"
	IDCluster                     = "0x1F43B675"
	IDTimestamp                   = "0xE7"
	IDSilentTracks                = "0x5854"
	IDSilentTrackNumber           = "0x58D7"
	IDPosition                    = "0xA7"
	IDPrevSize                    = "0xAB"
	IDSimpleBlock                 = "0xA3"
	IDBlockGroup                  = "0xA0"
	IDBlock                       = "0xA1"
	IDBlockVirtual                = "0xA2"
	IDBlockAdditions              = "0x75A1"
	IDBlockMore                   = "0xA6"
	IDBlockAdditional             = "0xA5"
	IDBlockAddID                  = "0xEE"
	IDBlockDuration               = "0x9B"
	IDReferencePriority           = "0xFA"
	IDReferenceBlock              = "0xFB"
	IDReferenceVirtual            = "0xFD"
	IDCodecState                  = "0xA4"
	IDDiscardPadding              = "0x75A2"
	IDSlices                      = "0x8E"
	IDTimeSlice                   = "0xE8"
	IDLaceNumber                  = "0xCC"
	IDFrameNumber                 = "0xCD"
	IDBlockAdditionID             = "0xCB"
	IDDelay                       = "0xCE"
	IDSliceDuration               = "0xCF"
	IDReferenceFrame              = "0xC8"
	IDReferenceOffset             = "0xC9"
	IDReferenceTimestamp          = "0xCA"
	IDEncryptedBlock              = "0xAF"
	IDTracks                      = "0x1654AE6B"
	IDTrackEntry                  = "0xAE"
	IDTrackNumber                 = "0xD7"
	IDTrackUID                    = "0x73C5"
	IDTrackType                   = "0x83"
	IDFlagEnabled                 = "0xB9"
	IDFlagDefault                 = "0x88"
	IDFlagForced                  = "0x55AA"
	IDFlagHearingImpaired         = "0x55AB"
	IDFlagVisualImpaired          = "0x55AC"
	IDFlagTextDescriptions        = "0x55AD"
	IDFlagOriginal                = "0x55AE"
	IDFlagCommentary              = "0x55AF"
	IDFlagLacing                  = "0x9C"
	IDMinCache                    = "0x6DE7"
	IDMaxCache                    = "0x6DF8"
	IDDefaultDuration             = "0x23E383"
	IDDefaultDecodedFieldDuration = "0x234E7A"
	IDTrackTimestampScale         = "0x23314F"
	IDTrackOffset                 = "0x537F"
	IDMaxBlockAdditionID          = "0x55EE"
	IDBlockAdditionMapping        = "0x41E4"
	IDBlockAddIDValue             = "0x41F0"
	IDBlockAddIDName              = "0x41A4"
	IDBlockAddIDType              = "0x41E7"
	IDBlockAddIDExtraData         = "0x41ED"
	IDName                        = "0x536E"
	IDLanguage                    = "0x22B59C"
	IDLanguageBCP47               = "0x22B59D"
	IDCodecID                     = "0x86"
	IDCodecPrivate                = "0x63A2"
	IDCodecName                   = "0x258688"
	IDAttachmentLink              = "0x7446"
	IDCodecSettings               = "0x3A9697"
	IDCodecInfoURL                = "0x3B4040"
	IDCodecDownloadURL            = "0x26B240"
	IDCodecDecodeAll              = "0xAA"
	IDTrackOverlay                = "0x6FAB"
	IDCodecDelay                  = "0x56AA"
	IDSeekPreRoll                 = "0x56BB"
	IDTrackTranslate              = "0x6624"
	IDTrackTranslateTrackID       = "0x66A5"
	IDTrackTranslateCodec         = "0x66BF"
	IDTrackTranslateEditionUID    = "0x66FC"
	IDVideo                       = "0xE0"
	IDFlagInterlaced              = "0x9A"
	IDFieldOrder                  = "0x9D"
	IDStereoMode                  = "0x53B8"
	IDAlphaMode                   = "0x53C0"
	IDOldStereoMode               = "0x53B9"
	IDPixelWidth                  = "0xB0"
	IDPixelHeight                 = "0xBA"
	IDPixelCropBottom             = "0x54AA"
	IDPixelCropTop                = "0x54BB"
	IDPixelCropLeft               = "0x54CC"
	IDPixelCropRight              = "0x54DD"
	IDDisplayWidth                = "0x54B0"
	IDDisplayHeight               = "0x54BA"
	IDDisplayUnit                 = "0x54B2"
	IDAspectRatioType             = "0x54B3"
	IDUncompressedFourCC          = "0x2EB524"
	IDGammaValue                  = "0x2FB523"
	IDFrameRate                   = "0x2383E3"
	IDColour                      = "0x55B0"
	IDMatrixCoefficients          = "0x55B1"
	IDBitsPerChannel              = "0x55B2"
	IDChromaSubsamplingHorz       = "0x55B3"
	IDChromaSubsamplingVert       = "0x55B4"
	IDCbSubsamplingHorz           = "0x55B5"
	IDCbSubsamplingVert           = "0x55B6"
	IDChromaSitingHorz            = "0x55B7"
	IDChromaSitingVert            = "0x55B8"
	IDRange                       = "0x55B9"
	IDTransferCharacteristics     = "0x55BA"
	IDPrimaries                   = "0x55BB"
	IDMaxCLL                      = "0x55BC"
	IDMaxFALL                     = "0x55BD"
	IDMasteringMetadata           = "0x55D0"
	IDPrimaryRChromaticityX       = "0x55D1"
	IDPrimaryRChromaticityY       = "0x55D2"
	IDPrimaryGChromaticityX       = "0x55D3"
	IDPrimaryGChromaticityY       = "0x55D4"
	IDPrimaryBChromaticityX       = "0x55D5"
	IDPrimaryBChromaticityY       = "0x55D6"
	IDWhitePointChromaticityX     = "0x55D7"
	IDWhitePointChromaticityY     = "0x55D8"
	IDLuminanceMax                = "0x55D9"
	IDLuminanceMin                = "0x55DA"
	IDProjection                  = "0x7670"
	IDProjectionType              = "0x7671"
	IDProjectionPrivate           = "0x7672"
	IDProjectionPoseYaw           = "0x7673"
	IDProjectionPosePitch         = "0x7674"
	IDProjectionPoseRoll          = "0x7675"
	IDAudio                       = "0xE1"
	IDSamplingFrequency           = "0xB5"
	IDOutputSamplingFrequency     = "0x78B5"
	IDChannels                    = "0x9F"
	IDChannelPositions            = "0x7D7B"
	IDBitDepth                    = "0x6264"
	IDEmphasis                    = "0x52F1"
	IDTrackOperation              = "0xE2"
	IDTrackCombinePlanes          = "0xE3"
	IDTrackPlane                  = "0xE4"
	IDTrackPlaneUID               = "0xE5"
	IDTrackPlaneType              = "0xE6"
	IDTrackJoinBlocks             = "0xE9"
	IDTrackJoinUID                = "0xED"
	IDTrickTrackUID               = "0xC0"
	IDTrickTrackSegmentUID        = "0xC1"
	IDTrickTrackFlag              = "0xC6"
	IDTrickMasterTrackUID         = "0xC7"
	IDTrickMasterTrackSegmentUID  = "0xC4"
	IDContentEncodings            = "0x6D80"
	IDContentEncoding             = "0x6240"
	IDContentEncodingOrder        = "0x5031"
	IDContentEncodingScope        = "0x5032"
	IDContentEncodingType         = "0x5033"
	IDContentCompression          = "0x5034"
	IDContentCompAlgo             = "0x4254"
	IDContentCompSettings         = "0x4255"
	IDContentEncryption           = "0x5035"
	IDContentEncAlgo              = "0x47E1"
	IDContentEncKeyID             = "0x47E2"
	IDContentEncAESSettings       = "0x47E7"
	IDAESSettingsCipherMode       = "0x47E8"
	IDContentSignature            = "0x47E3"
	IDContentSigKeyID             = "0x47E4"
	IDContentSigAlgo              = "0x47E5"
	IDContentSigHashAlgo          = "0x47E6"
	IDCues                        = "0x1C53BB6B"
	IDCuePoint                    = "0xBB"
	IDCueTime                     = "0xB3"
	IDCueTrackPositions           = "0xB7"
	IDCueTrack                    = "0xF7"
	IDCueClusterPosition          = "0xF1"
	IDCueRelativePosition         = "0xF0"
	IDCueDuration                 = "0xB2"
	IDCueBlockNumber              = "0x5378"
	IDCueCodecState               = "0xEA"
	IDCueReference                = "0xDB"
	IDCueRefTime                  = "0x96"
	IDCueRefCluster               = "0x97"
	IDCueRefNumber                = "0x535F"
	IDCueRefCodecState            = "0xEB"
	IDAttachments                 = "0x1941A469"
	IDAttachedFile                = "0x61A7"
	IDFileDescription             = "0x467E"
	IDFileName                    = "0x466E"
	IDFileMediaType               = "0x4660"
	IDFileData                    = "0x465C"
	IDFileUID                     = "0x46AE"
	IDFileReferral                = "0x4675"
	IDFileUsedStartTime           = "0x4661"
	IDFileUsedEndTime             = "0x4662"
	IDChapters                    = "0x1043A770"
	IDEditionEntry                = "0x45B9"
	IDEditionUID                  = "0x45BC"
	IDEditionFlagHidden           = "0x45BD"
	IDEditionFlagDefault          = "0x45DB"
	IDEditionFlagOrdered          = "0x45DD"
	IDEditionDisplay              = "0x4520"
	IDEditionString               = "0x4521"
	IDEditionLanguageIETF         = "0x45E4"
	IDChapterAtom                 = "0xB6"
	IDChapterUID                  = "0x73C4"
	IDChapterStringUID            = "0x5654"
	IDChapterTimeStart            = "0x91"
	IDChapterTimeEnd              = "0x92"
	IDChapterFlagHidden           = "0x98"
	IDChapterFlagEnabled          = "0x4598"
	IDChapterSegmentUUID          = "0x6E67"
	IDChapterSkipType             = "0x4588"
	IDChapterSegmentEditionUID    = "0x6EBC"
	IDChapterPhysicalEquiv        = "0x63C3"
	IDChapterTrack                = "0x8F"
	IDChapterTrackUID             = "0x89"
	IDChapterDisplay              = "0x80"
	IDChapString                  = "0x85"
	IDChapLanguage                = "0x437C"
	IDChapLanguageBCP47           = "0x437D"
	IDChapCountry                 = "0x437E"
	IDChapProcess                 = "0x6944"
	IDChapProcessCodecID          = "0x6955"
	IDChapProcessPrivate          = "0x450D"
	IDChapProcessCommand          = "0x6911"
	IDChapProcessTime             = "0x6922"
	IDChapProcessData             = "0x6933"
	IDTags                        = "0x1254C367"
	IDTag                         = "0x7373"
	IDTargets                     = "0x63C0"
	IDTargetTypeValue             = "0x68CA"
	IDTargetType                  = "0x63CA"
	IDTagTrackUID                 = "0x63C5"
	IDTagEditionUID               = "0x63C9"
	IDTagChapterUID               = "0x63C4"
	IDTagAttachmentUID            = "0x63C6"
	IDSimpleTag                   = "0x67C8"
	IDTagName                     = "0x45A3"
	IDTagLanguage                 = "0x447A"
	IDTagLanguageBCP47            = "0x447B"
	IDTagDefault                  = "0x4484"
	IDTagDefaultBogus             = "0x44B4"
	IDTagString                   = "0x4487"
	IDTagBinary                   = "0x4485"
)

type Segment struct {
	SeekHead    []SeekHead
	Info        Info
	Cluster     []Cluster
	Tracks      *Tracks
	Cues        *Cues
	Attachments *Attachments
	Chapters    *Chapters
	Tags        []Tags
}

type SeekHead struct {
	Seek []Seek
}

type Seek struct {
	SeekID       []byte
	SeekPosition uint
}

type Info struct {
	SegmentUUID      *[]byte
	SegmentFilename  *string
	PrevUUID         *[]byte
	PrevFilename     *string
	NextUUID         *[]byte
	NextFilename     *string
	SegmentFamily    [][]byte
	ChapterTranslate []ChapterTranslate
	TimestampScale   time.Duration
	Duration         *float64
	DateUTC          *time.Time
	Title            *string
	MuxingApp        string
	WritingApp       string
}

type ChapterTranslate struct {
	ChapterTranslateID         []byte
	ChapterTranslateCodec      uint
	ChapterTranslateEditionUID []uint
}

type Cluster struct {
	Timestamp      time.Duration
	SilentTracks   *SilentTracks
	Position       *uint
	PrevSize       *uint
	SimpleBlock    [][]byte
	BlockGroup     []BlockGroup
	EncryptedBlock [][]byte
}

type SilentTracks struct {
	SilentTrackNumber []uint
}

type BlockGroup struct {
	Block             []byte
	BlockVirtual      *[]byte
	BlockAdditions    *BlockAdditions
	BlockDuration     *uint
	ReferencePriority uint
	ReferenceBlock    []int
	ReferenceVirtual  *int
	CodecState        *[]byte
	DiscardPadding    *int
	Slices            *Slices
	ReferenceFrame    *ReferenceFrame
}

type BlockAdditions struct {
	BlockMore []BlockMore
}

type BlockMore struct {
	BlockAdditional []byte
	BlockAddID      uint
}

type Slices struct {
	TimeSlice []TimeSlice
}

type TimeSlice struct {
	LaceNumber      *uint
	FrameNumber     uint
	BlockAdditionID uint
	Delay           uint
	SliceDuration   uint
}

type ReferenceFrame struct {
	ReferenceOffset    uint
	ReferenceTimestamp uint
}

type Tracks struct {
	TrackEntry []TrackEntry
}

type TrackEntry struct {
	TrackNumber                 uint
	TrackUID                    uint
	TrackType                   uint
	FlagEnabled                 uint
	FlagDefault                 uint
	FlagForced                  uint
	FlagHearingImpaired         *uint
	FlagVisualImpaired          *uint
	FlagTextDescriptions        *uint
	FlagOriginal                *uint
	FlagCommentary              *uint
	FlagLacing                  uint
	MinCache                    uint
	MaxCache                    *uint
	DefaultDuration             *uint
	DefaultDecodedFieldDuration *uint
	TrackTimestampScale         float64
	TrackOffset                 int
	MaxBlockAdditionID          uint
	BlockAdditionMapping        []BlockAdditionMapping
	Name                        *string
	Language                    string
	LanguageBCP47               *string
	CodecID                     string
	CodecPrivate                *[]byte
	CodecName                   *string
	AttachmentLink              *uint
	CodecSettings               *string
	CodecInfoURL                []string
	CodecDownloadURL            []string
	CodecDecodeAll              uint
	TrackOverlay                []uint
	CodecDelay                  uint
	SeekPreRoll                 uint
	TrackTranslate              []TrackTranslate
	Video                       *Video
	Audio                       *Audio
	TrackOperation              *TrackOperation
	TrickTrackUID               *uint
	TrickTrackSegmentUID        *[]byte
	TrickTrackFlag              uint
	TrickMasterTrackUID         *uint
	TrickMasterTrackSegmentUID  *[]byte
	ContentEncodings            *ContentEncodings
}

type BlockAdditionMapping struct {
	BlockAddIDValue     *uint
	BlockAddIDName      *string
	BlockAddIDType      uint
	BlockAddIDExtraData *[]byte
}

type TrackTranslate struct {
	TrackTranslateTrackID    []byte
	TrackTranslateCodec      uint
	TrackTranslateEditionUID []uint
}

type Video struct {
	FlagInterlaced     uint
	FieldOrder         uint
	StereoMode         uint
	AlphaMode          uint
	OldStereoMode      *uint
	PixelWidth         uint
	PixelHeight        uint
	PixelCropBottom    uint
	PixelCropTop       uint
	PixelCropLeft      uint
	PixelCropRight     uint
	DisplayWidth       *uint
	DisplayHeight      *uint
	DisplayUnit        uint
	AspectRatioType    uint
	UncompressedFourCC *[]byte
	GammaValue         *float64
	FrameRate          *float64
	Colour             *Colour
	Projection         *Projection
}

type Colour struct {
	MatrixCoefficients      uint
	BitsPerChannel          uint
	ChromaSubsamplingHorz   *uint
	ChromaSubsamplingVert   *uint
	CbSubsamplingHorz       *uint
	CbSubsamplingVert       *uint
	ChromaSitingHorz        uint
	ChromaSitingVert        uint
	Range                   uint
	TransferCharacteristics uint
	Primaries               uint
	MaxCLL                  *uint
	MaxFALL                 *uint
	MasteringMetadata       *MasteringMetadata
}

type MasteringMetadata struct {
	PrimaryRChromaticityX   *float64
	PrimaryRChromaticityY   *float64
	PrimaryGChromaticityX   *float64
	PrimaryGChromaticityY   *float64
	PrimaryBChromaticityX   *float64
	PrimaryBChromaticityY   *float64
	WhitePointChromaticityX *float64
	WhitePointChromaticityY *float64
	LuminanceMax            *float64
	LuminanceMin            *float64
}

type Projection struct {
	ProjectionType      uint
	ProjectionPrivate   *[]byte
	ProjectionPoseYaw   float64
	ProjectionPosePitch float64
	ProjectionPoseRoll  float64
}

type Audio struct {
	SamplingFrequency       float64
	OutputSamplingFrequency *float64
	Channels                uint
	ChannelPositions        *[]byte
	BitDepth                *uint
	Emphasis                uint
}

type TrackOperation struct {
	TrackCombinePlanes *TrackCombinePlanes
	TrackJoinBlocks    *TrackJoinBlocks
}

type TrackCombinePlanes struct {
	TrackPlane []TrackPlane
}

type TrackPlane struct {
	TrackPlaneUID  uint
	TrackPlaneType uint
}

type TrackJoinBlocks struct {
	TrackJoinUID []uint
}

type ContentEncodings struct {
	ContentEncoding []ContentEncoding
}

type ContentEncoding struct {
	ContentEncodingOrder uint
	ContentEncodingScope uint
	ContentEncodingType  uint
	ContentCompression   *ContentCompression
	ContentEncryption    *ContentEncryption
}

type ContentCompression struct {
	ContentCompAlgo     uint
	ContentCompSettings *[]byte
}

type ContentEncryption struct {
	ContentEncAlgo        uint
	ContentEncKeyID       *[]byte
	ContentEncAESSettings *ContentEncAESSettings
	ContentSignature      *[]byte
	ContentSigKeyID       *[]byte
	ContentSigAlgo        uint
	ContentSigHashAlgo    uint
}

type ContentEncAESSettings struct {
	AESSettingsCipherMode uint
}

type Cues struct {
	CuePoint []CuePoint
}

type CuePoint struct {
	CueTime           uint
	CueTrackPositions []CueTrackPositions
}

type CueTrackPositions struct {
	CueTrack            uint
	CueClusterPosition  uint
	CueRelativePosition *uint
	CueDuration         *uint
	CueBlockNumber      *uint
	CueCodecState       uint
	CueReference        []CueReference
}

type CueReference struct {
	CueRefTime       uint
	CueRefCluster    uint
	CueRefNumber     uint
	CueRefCodecState uint
}

type Attachments struct {
	AttachedFile []AttachedFile
}

type AttachedFile struct {
	FileDescription   *string
	FileName          string
	FileMediaType     string
	FileData          []byte
	FileUID           uint
	FileReferral      *[]byte
	FileUsedStartTime *uint
	FileUsedEndTime   *uint
}

type Chapters struct {
	EditionEntry []EditionEntry
}

type EditionEntry struct {
	EditionUID         *uint
	EditionFlagHidden  uint
	EditionFlagDefault uint
	EditionFlagOrdered uint
	EditionDisplay     []EditionDisplay
	ChapterAtom        []ChapterAtom
}

type EditionDisplay struct {
	EditionString       string
	EditionLanguageIETF []string
}

type ChapterAtom struct {
	ChapterAtom              *ChapterAtom
	ChapterUID               uint
	ChapterStringUID         *string
	ChapterTimeStart         uint
	ChapterTimeEnd           *uint
	ChapterFlagHidden        uint
	ChapterFlagEnabled       uint
	ChapterSegmentUUID       *[]byte
	ChapterSkipType          *uint
	ChapterSegmentEditionUID *uint
	ChapterPhysicalEquiv     *uint
	ChapterTrack             *ChapterTrack
	ChapterDisplay           []ChapterDisplay
	ChapProcess              []ChapProcess
}

type ChapterTrack struct {
	ChapterTrackUID []uint
}

type ChapterDisplay struct {
	ChapString        string
	ChapLanguage      []string
	ChapLanguageBCP47 []string
	ChapCountry       []string
}

type ChapProcess struct {
	ChapProcessCodecID uint
	ChapProcessPrivate *[]byte
	ChapProcessCommand []ChapProcessCommand
}

type ChapProcessCommand struct {
	ChapProcessTime uint
	ChapProcessData []byte
}

type Tags struct {
	Tag []Tag
}

type Tag struct {
	Targets   Targets
	SimpleTag []SimpleTag
}

type Targets struct {
	TargetTypeValue  uint
	TargetType       *string
	TagTrackUID      []uint
	TagEditionUID    []uint
	TagChapterUID    []uint
	TagAttachmentUID []uint
}

type SimpleTag struct {
	SimpleTag        *SimpleTag
	TagName          string
	TagLanguage      string
	TagLanguageBCP47 *string
	TagDefault       uint
	TagDefaultBogus  uint
	TagString        *string
	TagBinary        *[]byte
}
