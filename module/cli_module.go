package module

import (
	"github.com/jueyanyingyu/lsp/config"
	"github.com/urfave/cli/v2"
	"strings"
)

type CliModule struct {
	App         *cli.App
	OperateType config.OperateType
	Path        string
}

func NewCliModule() CliModule {
	return CliModule{}
}

func (m *CliModule) Init() {
	var toCompress string
	var toDecompress string
	var toPack string
	var toUnpack string
	var toPackAndCompress string
	var toDecompressAndUnpack string

	m.App = &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "c",
				Value:       "",
				Usage:       "compress the file",
				Destination: &toCompress,
			},
			&cli.StringFlag{
				Name:        "d",
				Value:       "",
				Usage:       "decompress this file",
				Destination: &toDecompress,
			},
			&cli.StringFlag{
				Name:        "p",
				Value:       "",
				Usage:       "pack the file",
				Destination: &toPack,
			},
			&cli.StringFlag{
				Name:        "u",
				Value:       "",
				Usage:       "unpack the file",
				Destination: &toUnpack,
			},
			&cli.StringFlag{
				Name:        "pc",
				Value:       "",
				Usage:       "pack and compress the file or directory",
				Destination: &toPackAndCompress,
			},
			&cli.StringFlag{
				Name:        "du",
				Value:       "",
				Usage:       "decompress and unpack the file or directory",
				Destination: &toDecompressAndUnpack,
			},
		},
		Action: func(c *cli.Context) error {
			if toCompress != "" {
				m.OperateType = config.OpCompress
				m.Path = toCompress
			} else if toDecompress != "" && strings.HasSuffix(toDecompress, ".cp") {
				m.OperateType = config.OpDecompress
				m.Path = toDecompress
			} else if toPack != "" {
				m.OperateType = config.OpPack
				m.Path = toPack
			} else if toUnpack != "" && strings.HasSuffix(toUnpack, ".pk") {
				m.OperateType = config.OpUnpack
				m.Path = toUnpack
			} else if toPackAndCompress != "" {
				m.OperateType = config.OpPackAndCompress
				m.Path = toPackAndCompress
			} else if toDecompressAndUnpack != "" && strings.HasSuffix(toDecompressAndUnpack, ".pk.cp") {
				m.OperateType = config.OpDecompressAndUnpack
				m.Path = toDecompressAndUnpack
			} else {
				m.OperateType = config.OpUnDefine
				m.Path = ""
			}
			return nil
		},
	}
}
