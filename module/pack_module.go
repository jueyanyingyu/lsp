package module

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
)

type PackModule struct {
	sourceNameToPack string
	writer           *bufio.Writer
}

func NewPackModule(sourceNameToPack string, writer *bufio.Writer) *PackModule {
	return &PackModule{
		sourceNameToPack: sourceNameToPack,
		writer:           writer,
	}
}

type packedNode struct {
	nameSize uint16
	name     []uint8
	dataSize uint32

	nameSizeDone1 bool
	nameSizeDone2 bool

	dataSizeDone1 bool
	dataSizeDone2 bool
	dataSizeDone3 bool
	dataSizeDone4 bool
	done          bool
}

func newPackedNode() *packedNode {
	return &packedNode{}
}

func (n *packedNode) setByte(nb uint8) {
	//fmt.Printf("%v",nb)
	if !n.nameSizeDone1 {
		n.nameSize = uint16(nb) << 8
		n.nameSizeDone1 = true
		return
	}
	if !n.nameSizeDone2 {
		n.nameSize = n.nameSize + uint16(nb)
		n.nameSizeDone2 = true
		return
	}
	if len(n.name) < int(n.nameSize) {
		n.name = append(n.name, nb)
		return
	}
	if !n.dataSizeDone1 {
		n.dataSize = uint32(nb) << 24
		n.dataSizeDone1 = true
		return
	}
	if !n.dataSizeDone2 {
		n.dataSize = n.dataSize + uint32(nb)<<16
		n.dataSizeDone2 = true
		return
	}
	if !n.dataSizeDone3 {
		n.dataSize = n.dataSize + uint32(nb)<<8
		n.dataSizeDone3 = true
		return
	}
	if !n.dataSizeDone4 {
		n.dataSize = n.dataSize + uint32(nb)
		n.dataSizeDone4 = true
		n.done = true
		return
	}
}

func (n *packedNode) getByte() []uint8 {
	var result []uint8
	result = append(result, uint8(n.nameSize>>8))
	result = append(result, uint8(n.nameSize))
	result = append(result, n.name...)
	result = append(result, uint8(n.dataSize>>24))
	result = append(result, uint8(n.dataSize>>16))
	result = append(result, uint8(n.dataSize>>8))
	result = append(result, uint8(n.dataSize))
	return result
}

func (m *PackModule) Pack() error {
	var filePathList []string
	infoMap := make(map[string]os.FileInfo)
	err := filepath.Walk(m.sourceNameToPack, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.Mode().IsRegular() {
			filePathList = append(filePathList, path)
			infoMap[path] = f
		}
		return nil
	})
	if err != nil {
		log.Printf("walk err:%v", err)
		return err
	}
	for _, path := range filePathList {
		node := newPackedNode()
		node.nameSize = uint16(len(infoMap[path].Name()))
		node.name = []uint8(infoMap[path].Name())
		node.dataSize = uint32(infoMap[path].Size())
		headerByte := node.getByte()
		//fmt.Printf("%v",headerByte)
		_, err = m.writer.Write(headerByte)
		if err != nil {
			log.Printf("in write err:%v", err)
			return err
		}

		source, err := os.Open(path)
		if err != nil {
			log.Printf("open source err:%v", err)
			return err
		}
		bufSource := bufio.NewReader(source)
		for {
			nb, err := bufSource.ReadByte()
			if err != nil && err != io.EOF {
				log.Printf("source read err:%v", err)
				return err
			}
			if err == io.EOF {
				break
			}
			//fmt.Printf("%c",nb)
			err = m.writer.WriteByte(nb)
			if err != nil {
				log.Printf("in write err:%v", err)
				return err
			}
		}
		_ = m.writer.Flush()
		err = source.Close()
		if err != nil {
			log.Printf("close source err:%v", err)
			return err
		}
	}
	return nil
}

type UnpackModule struct {
	targetNameToUnpack string
	reader             *bufio.Reader
}

func NewUnpackModule(targetNameToUnpack string, reader *bufio.Reader) *UnpackModule {
	return &UnpackModule{
		targetNameToUnpack: targetNameToUnpack,
		reader:             reader,
	}
}

func (m *UnpackModule) Unpack() error {
	err := os.Mkdir(m.targetNameToUnpack, os.ModePerm)
	if err != nil {
		log.Printf("mkdir  err:%v", err)
		return err
	}
	node := newPackedNode()
	reader := bufio.NewReader(m.reader)
	for {
		nb, err := reader.ReadByte()
		//fmt.Printf("%v ",nb)
		if err != nil && err != io.EOF {
			log.Printf("read from stream err:%v", err)
			return err
		}
		//fmt.Printf("%v %v %v;",node.nameSize,node.name,node.dataSize)
		if !node.done {
			node.setByte(nb)
		} else {
			file, err := os.Create(filepath.Join(m.targetNameToUnpack, string(node.name)))
			if err != nil {
				log.Printf("create file err:%v", err)
				return err
			}
			writer := bufio.NewWriter(file)
			err = writer.WriteByte(nb)
			if err != nil {
				log.Printf("write file err:%v", err)
				return err
			}
			dataSize := node.dataSize - 1

			for ; dataSize > 0; dataSize-- {
				dataNb, err := reader.ReadByte()
				if err != nil && err != io.EOF {
					log.Printf("read from stream err:%v", err)
					return err
				}
				if err == io.EOF {
					break
				}
				//fmt.Printf("%c",dataNb)
				err = writer.WriteByte(dataNb)
				if err != nil {
					log.Printf("write file err:%v", err)
					return err
				}
			}
			_ = writer.Flush()
			node = newPackedNode()
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}
