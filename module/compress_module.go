package module

import (
	"errors"
	"io"
	"log"
)

type CompressModule struct {
	reader *io.PipeReader
	writer *io.PipeWriter
}

func NewCompressModule(reader *io.PipeReader, writer *io.PipeWriter) *CompressModule {
	return &CompressModule{
		reader: reader,
		writer: writer,
	}
}

func (m *CompressModule) Compress() error {
	var buffer []uint8
	var encoder *lz4SequenceEncoder
	var noMore bool
	var finish bool
	for !noMore || !finish {
		if len(buffer) == 0 && !noMore {
			nb := make([]uint8, 64)
			n, err := m.reader.Read(nb)
			if n > 0 {
				buffer = append(buffer, nb[0:n]...)
			}
			if err != nil && err != io.EOF {
				log.Printf("read from stream err:%v", err)
				return err
			}
			if err == io.EOF {
				noMore = true
			}
		}

		if encoder == nil {
			encoder = newLz4SequenceEncoder()
		}
		if encoder.done {
			//fmt.Printf("seq:%v\n", encoder.getByByte())
			result := encoder.getByByte()
			//fmt.Printf("result:%v",result)
			if len(result) == 0 {
				finish = true
			}
			n, err := m.writer.Write(result)
			if n < len(result) || err != nil {
				log.Printf("write ot stream err:%v", err)
				return err
			}
			//fmt.Printf("badbuffer:%v\n",encoder.badBuffer)
			buffer = append(encoder.badBuffer, buffer...)
			encoder = nil
			continue
		}
		if len(buffer) == 0 {
			encoder.compress()
		} else {
			//fmt.Printf("%v",buffer[0])
			encoder.compressWithNewByte(buffer[0])
			buffer = buffer[1:]
		}
	}
	//fmt.Printf("compress close the writer\n")
	err := m.writer.Close()
	if err != nil {
		log.Printf("close writer err:%v", err)
		return err
	}
	return nil
}

func (m *CompressModule) Decompress() error {
	var buffer []uint8
	var decoder *lz4SequenceDecoder
	var noMore bool
	for {
		if len(buffer) == 0 && !noMore {
			nb := make([]uint8, 64)
			n, err := m.reader.Read(nb)
			if n > 0 {
				//fmt.Printf("%v",nb[0:n])
				buffer = append(buffer, nb[0:n]...)
			}
			if err != nil && err != io.EOF {
				log.Printf("read from stream err:%v", err)
				return err
			}
			if err == io.EOF {
				noMore = true
			}
		}
		if decoder == nil {
			decoder = newLz4SequenceDecoder()
		}
		if decoder.done {
			result := decoder.getByByte()
			//fmt.Printf("result:%v",result)
			n, err := m.writer.Write(result)
			if n < len(result) || err != nil {
				log.Printf("write ot stream err:%v", err)
				return err
			}
			decoder = nil
			if noMore {
				break
			}
			continue
		}
		//fmt.Printf("%v\n",decoder.sequence)
		if len(buffer) == 0 {
			log.Printf("broken data")
			return errors.New("broken data")
		} else {
			decoder.decompressWithNewByte(buffer[0])
			buffer = buffer[1:]
		}
	}
	//fmt.Printf("decompress close the writer\n")
	err := m.writer.Close()
	if err != nil {
		log.Printf("close writer err:%v", err)
		return err
	}
	return nil
}

type sequence struct {
	literalsLen uint8
	literals    []uint8
	offset      uint8
	matchLen    uint8
}

type lz4SequenceEncoder struct {
	done     bool
	anchor   uint8
	hashMap  map[uint32]uint8
	sequence sequence

	headerBuffer []uint8
	badBuffer    []uint8
}

func newLz4SequenceEncoder() *lz4SequenceEncoder {
	return &lz4SequenceEncoder{
		hashMap: make(map[uint32]uint8),
	}
}

func (e *lz4SequenceEncoder) compressWithNewByte(nb uint8) {
	//fmt.Printf("\ncompress:nb:%v\n",nb)
	//fmt.Printf("\nsequence:%v", e.sequence)
	//已经完成本块的压缩，正常不应为true
	if e.done {
		return
	}
	e.headerBuffer = append(e.headerBuffer, nb)
	//缓冲区不足补入
	if len(e.headerBuffer) < 4 {
		return
	}
	//if e.sequence.literalsLen < 3 {
	//	e.sequence.literalsLen++
	//	e.sequence.literals = append(e.sequence.literals, e.headerBuffer[0])
	//	e.anchor++
	//	e.headerBuffer = e.headerBuffer[1:]
	//	return
	//}
	var b1, b2, b3, b4 uint8
	b1 = e.headerBuffer[0]
	b2 = e.headerBuffer[1]
	b3 = e.headerBuffer[2]
	b4 = e.headerBuffer[3]
	b := uint32(b1)<<24 + uint32(b2)<<16 + uint32(b3)<<8 + uint32(b4)
	//fmt.Printf("index:%v\n", e.hashMap[b])
	if index, ok := e.hashMap[b]; !ok || index+4 > e.sequence.literalsLen {
		//之前已经压缩,新字节无法继续压缩
		if e.sequence.offset != 0 {
			e.badBuffer = append(e.badBuffer, nb)
			e.done = true
		} else {
			//从未遇到的组合，记录hash表中
			if !ok {
				e.hashMap[b] = e.anchor
			}
			e.sequence.literalsLen++
			e.sequence.literals = append(e.sequence.literals, e.headerBuffer[0])
			e.anchor++
			e.headerBuffer = e.headerBuffer[1:]
			//检查字面数组是否已达上限
			if e.sequence.literalsLen == 255 {
				e.done = true
				e.badBuffer = append(e.badBuffer, e.headerBuffer...)
			}
		}
	} else {
		//之前已经试图压缩
		if e.sequence.offset != 0 {
			//尝试扩展
			//fmt.Printf("b4:%c,nb:%c,anchor:%v,index:%v,offset:%v\n", b4, nb, e.anchor, index, e.sequence.offset)
			if e.anchor-index == e.sequence.offset {
				e.sequence.matchLen++
				e.anchor++
				e.headerBuffer = e.headerBuffer[1:]
				//匹配字符串已达上限
				if e.sequence.matchLen == 255 {
					e.done = true
				}
			} else {
				//无法扩展
				e.badBuffer = append(e.badBuffer, nb)
				e.done = true
			}
		} else {
			e.sequence.offset = e.anchor - index
			e.sequence.matchLen = 4
			e.headerBuffer = e.headerBuffer[1:]
			e.anchor++
		}
	}
}

func (e *lz4SequenceEncoder) compress() {
	if e.done {
		return
	}
	for e.sequence.offset == 0 && len(e.headerBuffer) > 0 {
		//fmt.Printf("offset:%v,len:%v\n",e.sequence.offset,len(e.headerBuffer))
		if e.sequence.literalsLen == 255 {
			e.badBuffer = append(e.badBuffer, e.headerBuffer...)
			e.headerBuffer = nil
		} else {
			e.sequence.literalsLen++
			e.sequence.literals = append(e.sequence.literals, e.headerBuffer[0])
			e.headerBuffer = e.headerBuffer[1:]
			e.anchor++
		}
	}
	e.done = true
}

func (e *lz4SequenceEncoder) getByByte() []uint8 {
	if !e.done {
		return nil
	}
	var result []uint8
	if e.sequence.literalsLen == 0 {
		return nil
	}
	result = append(result, e.sequence.literalsLen)
	result = append(result, e.sequence.literals...)
	result = append(result, e.sequence.offset)
	result = append(result, e.sequence.matchLen)
	return result
}

type lz4SequenceDecoder struct {
	done       bool
	sequence   sequence
	offsetDone bool
	matchDone  bool
}

func newLz4SequenceDecoder() *lz4SequenceDecoder {
	return &lz4SequenceDecoder{}
}

func (d *lz4SequenceDecoder) decompressWithNewByte(nb uint8) {
	if d.done {
		return
	}
	if d.sequence.literalsLen == 0 {
		d.sequence.literalsLen = nb
		return
	}
	if uint8(len(d.sequence.literals)) < d.sequence.literalsLen {
		d.sequence.literals = append(d.sequence.literals, nb)
		return
	}
	if !d.offsetDone {
		d.sequence.offset = nb
		d.offsetDone = true
		return
	}
	if !d.matchDone {
		d.sequence.matchLen = nb
		d.matchDone = true
	}
	d.done = true
}

func (d *lz4SequenceDecoder) getByByte() []uint8 {
	//fmt.Printf("\ndone:%v\n", d.done)
	if !d.done {
		return nil
	}
	var result []uint8
	result = append(result, d.sequence.literals...)
	if d.sequence.offset != 0 {
		from := d.sequence.literalsLen - d.sequence.offset
		to := from + d.sequence.matchLen
		for i := from; i < to; i++ {
			result = append(result, d.sequence.literals[i])
		}
	}
	return result
}
