package module

import (
	"flag"
	"github.com/jueyanyingyu/lsp/config"
	"os"
)

type FlagModule struct {
	OperateType config.OperateType
	Path        string
}

var compress *string
var decompress *string
var pack *string
var unpack *string
var compressWithPack *string
var decompressWithUnpack *string

func NewFlagModule() FlagModule {
	return FlagModule{}
}

func (module *FlagModule) InitPath() {
	if len(os.Args) != 3 {
		flag.Usage = func() {
			flag.PrintDefaults()
			return
		}
	}
	flag.StringVar(compress, "c", "", "compress the file")
	flag.StringVar(decompress, "d", "", "decompress this file")
	flag.StringVar(pack, "p", "", "pack the file")
	flag.StringVar(unpack, "u", "", "unpack the file")
	flag.StringVar(compressWithPack, "cp", "", "compress and pack the file or dictionary ")
	flag.StringVar(decompressWithUnpack, "uc", "", "unpack and decompress the file or dictionary")

	flag.Parse()

	if compress != nil {
		module.OperateType = config.OpCompress
		module.Path = *compress
	} else if decompress != nil {
		module.OperateType = config.OpDecompress
		module.Path = *decompress
	} else if pack != nil {
		module.OperateType = config.OpPack
		module.Path = *pack
	} else if unpack != nil {
		module.OperateType = config.OpUnpack
		module.Path = *unpack
	} else if compressWithPack != nil {
		module.OperateType = config.OpCompressWithPack
		module.Path = *compressWithPack
	} else if decompressWithUnpack != nil {
		module.OperateType = config.OpDecompressWithUnpack
		module.Path = *decompressWithUnpack
	}
	module.OperateType = config.OpUnDefine
	module.Path = ""
}
