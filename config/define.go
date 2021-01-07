package config

type OperateType int64

const (
	OpUnDefine = OperateType(iota)
	OpCompress
	OpDecompress
	OpPack
	OpUnpack
	OpPackAndCompress
	OpDecompressAndUnpack

	HeaderBufferSize  = 1 << 7
	SlidingWindowSize = 1 << 16

	BufferSize = 1024

	MinPrefixSize = 16
)
