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

	HeaderBufferSize = 128
	SlidingWindowSize = 65536

	BufferSize = 1024

	MinPrefixSize = 16
)
