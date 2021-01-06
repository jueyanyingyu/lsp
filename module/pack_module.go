package module

import (
	"io"
)

type PackModule struct {
	sourceNameToPack   string
	writer             *io.PipeWriter
}

func NewPackModule(sourceNameToPack string, writer *io.PipeWriter) *PackModule {
	return &PackModule{
		sourceNameToPack:   sourceNameToPack,
		writer:             writer,
	}
}

func (m *PackModule) Pack() error {

	return nil
}

type packer struct {

}
















type UnpackModule struct {
	targetNameToUnpack string
	reader             *io.PipeReader
}

func NewUnpackModule(targetNameToUnpack string, reader *io.PipeReader) *UnpackModule {
	return &UnpackModule{
		targetNameToUnpack: targetNameToUnpack,
		reader:             reader,
	}
}

func (m *UnpackModule) Unpack() error {

	return nil
}



