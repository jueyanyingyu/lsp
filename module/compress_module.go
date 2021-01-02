package module

import (
	"github.com/jueyanyingyu/lsp/config"
	"io"
	"log"
	"strings"
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
	encoder := newLz77SequenceEncoder()
	for {
		var buffer []uint8
		nb := make([]uint8, config.BufferSize)
		n, err := m.reader.Read(nb)
		if n > 0 {
			buffer = append(buffer, nb[0:n]...)
			//fmt.Printf("buffer:%v",buffer)
			var result []uint8
			for _, v := range buffer {
				//fmt.Printf("v:%v",v)
				result = append(result, encoder.compressWithNewByte(v)...)
			}
			_, err := m.writer.Write(result)
			if err != nil {
				log.Printf("write to stream err:%v", err)
				return err
			}
		}
		if err != nil && err != io.EOF {
			log.Printf("read from stream err:%v", err)
			return err
		}
		if err == io.EOF {
			var result []uint8
			result = append(result, encoder.compress()...)
			//fmt.Printf("result:%v",result)
			_, err := m.writer.Write(result)
			if err != nil {
				log.Printf("write to stream err:%v", err)
				return err
			}
			break
		}
	}
	err := m.writer.Close()
	if err != nil {
		log.Printf("close write err:%v", err)
		return err
	}
	return nil
}

func (m *CompressModule) Decompress() error {
	decoder := newLz77SequenceDecoder()
	for {
		var buffer []uint8
		nb := make([]uint8, config.BufferSize)
		n, err := m.reader.Read(nb)
		if n > 0 {
			buffer = append(buffer, nb[0:n]...)
			//fmt.Printf("buffer:%v\n",buffer)
			var result []uint8
			for _, v := range buffer {
				result = append(result, decoder.decompressWithNewByte(v)...)
			}
			//fmt.Printf("%v",decoder.result)
			_, err := m.writer.Write(result)
			if err != nil {
				log.Printf("write to stream err:%v", err)
				return err
			}
		}
		if err != nil && err != io.EOF {
			log.Printf("read from stream err:%v", err)
			return err
		}
		if err == io.EOF {
			var result []uint8
			result = append(result, decoder.result...)
			//fmt.Printf("result:%v",result)
			_, err := m.writer.Write(result)
			if err != nil {
				log.Printf("write to stream err:%v", err)
				return err
			}
			break
		}
	}
	err := m.writer.Close()
	if err != nil {
		log.Printf("close write err:%v", err)
		return err
	}
	return nil
}

func getLongestPrefix(slidingWindow []uint8, headBuffer []uint8) (uint16, uint16, uint8) {
	var maxLength uint16
	var maxLengthOffset uint16
	var next uint8
	for i := len(headBuffer) - 1; i > 0; i-- {
		index := strings.Index(string(slidingWindow), string(headBuffer[:i]))
		if index >= 0 {
			maxLength = uint16(i)
			maxLengthOffset = uint16(index)
			next = headBuffer[i]
			break
		}
	}
	if maxLength == 0 {
		next = headBuffer[0]
	}
	return maxLengthOffset, maxLength, next
}

type lz77SequenceEncoder struct {
	slidingWindow []uint8
	headerBuffer  []uint8

	matchStatus bool
	literalNum  uint16
	matchNum    uint16
	result      []uint8
}

func newLz77SequenceEncoder() *lz77SequenceEncoder {
	return &lz77SequenceEncoder{}
}

func (e *lz77SequenceEncoder) compressWithNewByte(nb uint8) []uint8 {
	//fmt.Printf("nb:%v",nb)
	//缓冲区不足补入
	if len(e.headerBuffer) < config.HeaderBufferSize {
		//fmt.Printf("nb:%v",nb)
		e.headerBuffer = append(e.headerBuffer, nb)
		return nil
	}
	offset, matchLen, next := getLongestPrefix(e.slidingWindow, e.headerBuffer)
	//fmt.Printf("offset:%v,matchLen:%v,next:%v",offset,matchLen,next)
	//补偿因为滑动窗口未满导致的偏移误差
	if matchLen != 0 {
		offset = offset + uint16(config.SlidingWindowSize-len(e.slidingWindow))
	}
	//匹配部分加入滑动窗口
	e.slidingWindow = append(e.slidingWindow, e.headerBuffer[:matchLen+1]...)
	//滑动窗口舍弃旧数据使得大小不超过上限
	if len(e.slidingWindow) > config.SlidingWindowSize {
		e.slidingWindow = e.slidingWindow[len(e.slidingWindow)-config.SlidingWindowSize:]
	}
	e.headerBuffer = e.headerBuffer[matchLen+1:]
	if len(e.headerBuffer) < config.HeaderBufferSize {
		//fmt.Printf("nb:%v",nb)
		e.headerBuffer = append(e.headerBuffer, nb)
	}

	var toReturn []uint8
	if !e.matchStatus {
		//字面量状态
		if matchLen == 0 {
			//新增字面量
			e.literalNum++
			e.result = append(e.result, next)
		} else {
			//出现匹配项，字面量数据输出
			toReturn = append(toReturn, uint8(e.literalNum>>8))
			toReturn = append(toReturn, uint8(e.literalNum))
			toReturn = append(toReturn, e.result...)
			e.literalNum = 0
			e.result = nil
			e.matchStatus = true

			//新增匹配项
			e.matchNum++
			e.result = append(e.result, uint8(offset>>8))
			e.result = append(e.result, uint8(offset))
			e.result = append(e.result, uint8(matchLen>>8))
			e.result = append(e.result, uint8(matchLen))
			e.result = append(e.result, next)
		}
	} else {
		if matchLen == 0 {
			//出现字面量，匹配项数据输出
			//fmt.Printf("e.matchNum:%v",e.matchNum)
			toReturn = append(toReturn, uint8(e.matchNum>>8))
			toReturn = append(toReturn, uint8(e.matchNum))
			toReturn = append(toReturn, e.result...)
			e.matchNum = 0
			e.result = nil
			e.matchStatus = false

			//新增字面量
			e.literalNum++
			e.result = append(e.result, next)
		} else {
			//新增匹配项
			e.matchNum++
			e.result = append(e.result, uint8(offset>>8))
			e.result = append(e.result, uint8(offset))
			e.result = append(e.result, uint8(matchLen>>8))
			e.result = append(e.result, uint8(matchLen))
			e.result = append(e.result, next)
		}
	}
	return toReturn
}

func (e *lz77SequenceEncoder) compress() []uint8 {
	var totalResult []uint8
	for len(e.headerBuffer) > 0 {
		offset, matchLen, next := getLongestPrefix(e.slidingWindow, e.headerBuffer)
		//补偿因为滑动窗口未满导致的偏移误差
		if matchLen != 0 {
			offset = offset + uint16(config.SlidingWindowSize-len(e.slidingWindow))
		}
		//fmt.Printf("slidingWindow:%v\n",e.slidingWindow)
		//fmt.Printf("headerBuffer:%v\n",e.headerBuffer)
		//fmt.Printf("offset:%v,matchLen:%v,next:%v\n",offset,matchLen,next)
		//匹配部分加入滑动窗口
		e.slidingWindow = append(e.slidingWindow, e.headerBuffer[:matchLen+1]...)
		//滑动窗口舍弃旧数据使得大小不超过上限
		if len(e.slidingWindow) > config.SlidingWindowSize {
			e.slidingWindow = e.slidingWindow[len(e.slidingWindow)-config.SlidingWindowSize:]
		}
		e.headerBuffer = e.headerBuffer[matchLen+1:]
		//fmt.Printf("slidingWindow:%v",e.slidingWindow)
		//fmt.Printf("headerBuffer:%v\n",e.headerBuffer)
		var toReturn []uint8
		if !e.matchStatus {
			//字面量状态
			if matchLen == 0 {
				//新增字面量
				e.literalNum++
				e.result = append(e.result, next)
			} else {
				//出现匹配项，字面量数据输出
				toReturn = append(toReturn, uint8(e.literalNum>>8))
				toReturn = append(toReturn, uint8(e.literalNum))
				toReturn = append(toReturn, e.result...)
				e.literalNum = 0
				e.result = nil
				e.matchStatus = true

				//新增匹配项
				e.matchNum++
				e.result = append(e.result, uint8(offset>>8))
				e.result = append(e.result, uint8(offset))
				e.result = append(e.result, uint8(matchLen>>8))
				e.result = append(e.result, uint8(matchLen))
				e.result = append(e.result, next)
			}
		} else {
			if matchLen == 0 {
				//出现字面量，匹配项数据输出
				//fmt.Printf("e.matchNum:%v",e.matchNum)
				toReturn = append(toReturn, uint8(e.matchNum>>8))
				toReturn = append(toReturn, uint8(e.matchNum))
				toReturn = append(toReturn, e.result...)
				e.matchNum = 0
				e.result = nil
				e.matchStatus = false

				//新增字面量
				e.literalNum++
				e.result = append(e.result, next)
			} else {
				//新增匹配项
				e.matchNum++
				e.result = append(e.result, uint8(offset>>8))
				e.result = append(e.result, uint8(offset))
				e.result = append(e.result, uint8(matchLen>>8))
				e.result = append(e.result, uint8(matchLen))
				e.result = append(e.result, next)
			}
		}
		totalResult = append(totalResult, toReturn...)
	}
	totalResult = append(totalResult, uint8(e.matchNum>>8))
	totalResult = append(totalResult, uint8(e.matchNum))
	totalResult = append(totalResult, e.result...)
	//fmt.Printf("result:%v",totalResult)
	return totalResult
}

type lz77SequenceDecoder struct {
	slidingWindow []uint8
	offset        uint16
	matchLen      uint16
	offsetDone1   bool
	offsetDone2   bool
	matchLenDone1 bool
	matchLenDone2 bool

	matchStatus     bool
	literalNum      uint16
	literalNumDone1 bool
	literalNumDone2 bool
	matchNum        uint16
	matchNumDone1   bool
	matchNumDone2   bool
	result          []uint8
}

func newLz77SequenceDecoder() *lz77SequenceDecoder {
	return &lz77SequenceDecoder{}
}

func (d *lz77SequenceDecoder) decompressWithNewByte(nb uint8) []uint8 {
	//fmt.Printf("%v ",nb)
	var toReturn []uint8
	if !d.matchStatus {
		if !d.literalNumDone1 {
			d.literalNum = uint16(nb) << 8
			d.literalNumDone1 = true
			return nil
		}
		if !d.literalNumDone2 {
			d.literalNum = d.literalNum + uint16(nb)
			d.literalNumDone2 = true
			return nil
		}
		if d.literalNum > 0 {
			d.result = append(d.result, nb)
			d.slidingWindow = append(d.slidingWindow, nb)
			d.literalNum--
		} else {
			toReturn = append(toReturn, d.result...)

			d.matchStatus = true
			d.literalNumDone1 = false
			d.literalNumDone2 = false
			d.result = nil

			//fmt.Printf("nb:%v",nb)
			d.matchNum = uint16(nb) << 8
			d.matchNumDone1 = true
		}
	} else {
		if !d.matchNumDone2 {
			//fmt.Printf("nb:%v",nb)
			d.matchNum = d.matchNum + uint16(nb)
			d.matchNumDone2 = true
			return nil
		}
		//fmt.Printf("matchNum:%v",d.matchNum)
		if d.matchNum > 0 {
			if !d.offsetDone1 {
				d.offset = uint16(nb) << 8
				d.offsetDone1 = true
				return nil
			}
			if !d.offsetDone2 {
				d.offset = d.offset + uint16(nb)
				d.offsetDone2 = true
				return nil
			}
			if !d.matchLenDone1 {
				d.matchLen = uint16(nb) << 8
				d.matchLenDone1 = true
				return nil
			}
			if !d.matchLenDone2 {
				d.matchLen = d.matchLen + uint16(nb)
				d.matchLenDone2 = true
				return nil
			}
			//fmt.Printf("slidingWindow:%v\n",d.slidingWindow)
			//fmt.Printf("offset:%v,matchLen:%v,next:%v\n",d.offset,d.matchLen,nb)
			//fmt.Printf("len(d.slidingWindow):%v\n",len(d.slidingWindow))
			d.offset = d.offset - uint16(config.SlidingWindowSize-len(d.slidingWindow))
			//fmt.Printf("offset:%v,matchLen:%v,next:%v\n",d.offset,d.matchLen,nb)
			d.result = append(d.result, d.slidingWindow[d.offset:d.offset+d.matchLen]...)
			d.result = append(d.result, nb)
			d.matchNum--
			d.offsetDone1 = false
			d.offsetDone2 = false
			d.matchLenDone1 = false
			d.matchLenDone2 = false

			d.slidingWindow = append(d.slidingWindow, d.slidingWindow[d.offset:d.offset+d.matchLen]...)
			d.slidingWindow = append(d.slidingWindow, nb)
		} else {
			toReturn = append(toReturn, d.result...)

			d.matchStatus = false
			d.matchNumDone1 = false
			d.matchNumDone2 = false
			d.offsetDone1 = false
			d.offsetDone2 = false
			d.matchLenDone1 = false
			d.matchLenDone2 = false
			d.result = nil

			d.literalNum = uint16(nb) << 8
			d.literalNumDone1 = true
		}
	}
	//fmt.Printf("slidingWindow:%v\n",d.slidingWindow)
	//滑动窗口舍弃旧数据使得大小不超过上限
	if len(d.slidingWindow) > config.SlidingWindowSize {
		d.slidingWindow = d.slidingWindow[len(d.slidingWindow)-config.SlidingWindowSize:]
	}
	//fmt.Printf("toReturn:%v",toReturn)
	return toReturn
}
