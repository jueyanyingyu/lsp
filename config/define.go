package config

type OperateType int64

const (
	OpUnDefine = OperateType(iota)
	OpCompress
	OpDecompress
	OpPack
	OpUnpack
	OpCompressWithPack
	OpDecompressWithUnpack
)
