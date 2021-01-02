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

	HeaderBufferSize = 32
	SlidingWindowSize = 4096

	BufferSize = 1024
)
