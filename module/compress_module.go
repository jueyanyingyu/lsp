package module

import (
	"bufio"
	"github.com/jueyanyingyu/lsp/config"
	"io"
	"log"
	"math"
)

type CompressModule struct {
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewCompressModule(reader *bufio.Reader, writer *bufio.Writer) *CompressModule {
	return &CompressModule{
		reader: reader,
		writer: writer,
	}
}

func (m *CompressModule) Compress() error {
	encoder := newLz77SequenceEncoder()
	for {
		nb, err := m.reader.ReadByte()
		if err != nil && err != io.EOF {
			log.Printf("read from stream err:%v", err)
			return err
		}
		if err == io.EOF {
			_, err := m.writer.Write(encoder.compress())
			if err != nil {
				log.Printf("write to stream err:%v", err)
				return err
			}
			break
		}
		_, err = m.writer.Write(encoder.compressWithNewByte(nb))
		if err != nil {
			log.Printf("write to stream err:%v", err)
			return err
		}
	}
	err := m.writer.Flush()
	if err != nil {
		log.Printf("flush write err:%v", err)
		return err
	}
	return nil
}

func (m *CompressModule) Decompress() error {
	decoder := newLz77SequenceDecoder()
	for {
		nb, err := m.reader.ReadByte()
		if err != nil && err != io.EOF {
			log.Printf("read from stream err:%v", err)
			return err
		}
		if err == io.EOF {
			_, err := m.writer.Write(decoder.result)
			if err != nil {
				log.Printf("write to stream err:%v", err)
				return err
			}
			break
		}
		_, err = m.writer.Write(decoder.decompressWithNewByte(nb))
		if err != nil {
			log.Printf("write to stream err:%v", err)
			return err
		}
	}
	err := m.writer.Flush()
	if err != nil {
		log.Printf("flush write err:%v", err)
		return err
	}
	return nil
}

//func getNextOptimize(p []byte) []int {
//	pLen := len(p)
//	next := make([]int, pLen, pLen)
//	next[0] = -1
//	next[1] = 0
//	i := 0
//	j := 1
//	for j < pLen-1 { //因为next[pLen-1]由s[i] == s[pLen-2]算出
//		if i == -1 || p[i] == p[j] { //-1代表了起始位不匹配，i=0,s[0]!=s[j]=>i=next[0]=-1
//			i++
//			j++
//			if p[i] != p[j] { //因为出现在j位置不匹配的话会跳到next[j]=i位置去匹配,p[i] == p[j]肯定又是不匹配（优化核心点）
//				next[j] = i
//			} else {
//				next[j] = next[i]
//			}
//
//		} else {
//			i = next[i]
//		}
//	}
//	return next
//}
//
//func kmpSearch(s, p []byte) int {
//	i, j := 0, 0
//	pLen := len(p)
//	sLen := len(s)
//	next := getNextOptimize(p)
//	for i < sLen && j < pLen {
//		if j == -1 || s[i] == p[j] { //s[i]!=s[0]=>j=next[0]=-1,第0位不匹配所以i++，j++;j=0
//			i++
//			j++
//		} else {
//			j = next[j]
//		}
//	}
//	if j == pLen {
//		return i - j
//	} else {
//		return -1
//	}
//}

type lz77SequenceEncoder struct {
	slidingWindow []uint8
	headerBuffer  []uint8

	matchStatus    bool
	literalNum     uint8
	matchNum       uint8
	literalNumList []uint8
	matchNumList   []uint8
	result         []uint8

	hashMap             map[uint32][]uint64
	slidingWindowOffset uint64
}

func newLz77SequenceEncoder() *lz77SequenceEncoder {
	return &lz77SequenceEncoder{
		hashMap: make(map[uint32][]uint64),
	}
}

func (e *lz77SequenceEncoder) updateHash(matchLen uint8) {
	if len(e.slidingWindow) < 4 {
		return
	}
	iBegin := len(e.slidingWindow) - int(matchLen)
	if iBegin <= 3 {
		iBegin = 3
	}
	//fmt.Printf("len(e.slidingWindow):%v,int(matchLen):%v\n",len(e.slidingWindow),int(matchLen))
	for i := iBegin; i < len(e.slidingWindow); i++ {
		k := uint32(e.slidingWindow[i]) + uint32(e.slidingWindow[i-1])<<8 + uint32(e.slidingWindow[i-2])<<16 + uint32(e.slidingWindow[i-3])<<24
		e.hashMap[k] = append(e.hashMap[k], e.slidingWindowOffset+uint64(i-3))
	}
}
func (e *lz77SequenceEncoder) getHashOffset(k uint32) []uint64 {
	if v, ok := e.hashMap[k]; ok {
		var newV []uint64
		for _, h := range v {
			if h >= e.slidingWindowOffset {
				newV = append(newV, h)
			}
		}
		e.hashMap[k] = newV
		//fmt.Printf("\nnewV:%v\n",newV)
		return newV
	}
	return nil
}
func (e *lz77SequenceEncoder) getLongestPrefix(slidingWindow []uint8, headBuffer []uint8) (uint16, uint8, uint8) {
	var maxLength uint8
	var maxLengthOffset uint16
	var next uint8
	if len(headBuffer) < 4 {
		next = headBuffer[0]
		return maxLengthOffset, maxLength, next
	}
	k := uint32(headBuffer[0]) + uint32(headBuffer[1])<<8 + uint32(headBuffer[2])<<16 + uint32(headBuffer[3])<<24
	indexList := e.getHashOffset(k)
	for _, index := range indexList {
		if index >= 0 {
			length := uint8(0)
			//fmt.Printf("index:%v,slidingWindowOffset:%v\n",index,e.slidingWindowOffset)
			iBegin := index - e.slidingWindowOffset
			i := iBegin
			j := 0
			for ; int(i) < len(slidingWindow) && j < len(headBuffer)-1 && slidingWindow[i] == headBuffer[j]; {
				length++
				i++
				j++
			}
			if length > config.MinPrefixSize && length > maxLength {
				maxLength = length
				maxLengthOffset = uint16(iBegin)
				next = headBuffer[j]
			}
		}
	}
	if maxLength == 0 {
		next = headBuffer[0]
	}
	return maxLengthOffset, maxLength, next
}
func (e *lz77SequenceEncoder) compressWithNewByte(nb uint8) []uint8 {

	//fmt.Printf("hashMap:%v\n",e.hashMap)
	//缓冲区不足补入
	if len(e.headerBuffer) < config.HeaderBufferSize {
		//fmt.Printf("nb:%v",nb)
		e.headerBuffer = append(e.headerBuffer, nb)
		return nil
	}
	//fmt.Printf("\n\n\nslidingWindow:%v  headerBuffer:%v\n",e.slidingWindow,e.headerBuffer)
	offset, matchLen, next := e.getLongestPrefix(e.slidingWindow, e.headerBuffer)
	//fmt.Printf("offset:%v,matchLen:%v,next:%v\n",offset,matchLen,next)
	//补偿因为滑动窗口未满导致的偏移误差
	if matchLen != 0 {
		//fmt.Printf("offset:%v,matchLen:%v,next:%v\n", offset, matchLen, next)
		offset = offset + uint16(config.SlidingWindowSize-len(e.slidingWindow))
	}
	//fmt.Printf("slidingWindow:%v  headerBuffer:%v\n",e.slidingWindow,e.headerBuffer)
	//匹配部分加入滑动窗口
	e.slidingWindow = append(e.slidingWindow, e.headerBuffer[:matchLen+1]...)
	//fmt.Printf("slidingWindowOffset:%v\n",e.slidingWindowOffset)
	//滑动窗口舍弃旧数据使得大小不超过上限
	if len(e.slidingWindow) > config.SlidingWindowSize {
		dropLen := len(e.slidingWindow) - config.SlidingWindowSize
		e.slidingWindow = e.slidingWindow[dropLen:]
		e.slidingWindowOffset = e.slidingWindowOffset + uint64(dropLen)
	}
	//fmt.Printf("slidingWindowOffset:%v\n",e.slidingWindowOffset)
	//fmt.Printf("hashMap:%v\n",e.hashMap)
	e.updateHash(matchLen + 1)
	//fmt.Printf("hashMap:%v\n",e.hashMap)

	e.headerBuffer = e.headerBuffer[matchLen+1:]
	//fmt.Printf("slidingWindow:%v  headerBuffer:%v\n",e.slidingWindow,e.headerBuffer)
	if len(e.headerBuffer) < config.HeaderBufferSize {
		//fmt.Printf("nb:%v",nb)
		e.headerBuffer = append(e.headerBuffer, nb)
	}

	var toReturn []uint8
	if !e.matchStatus {
		//字面量状态
		if matchLen == 0 {
			//新增字面量 //超过上限就扩展位来记录
			if e.literalNum == math.MaxUint8 {
				e.literalNumList = append(e.literalNumList, math.MaxUint8)
				e.literalNum = 0
			}
			e.literalNum++
			e.result = append(e.result, next)
		} else {
			//出现匹配项，字面量数据输出
			toReturn = append(toReturn, e.literalNumList...)
			toReturn = append(toReturn, e.literalNum)
			if e.literalNum == math.MaxUint8 {
				toReturn = append(toReturn, 0)
			}
			toReturn = append(toReturn, e.result...)
			e.literalNumList = nil
			e.literalNum = 0
			e.result = nil
			e.matchStatus = true

			//新增匹配项
			e.matchNum++
			e.result = append(e.result, uint8(offset>>8))
			e.result = append(e.result, uint8(offset))
			e.result = append(e.result, matchLen)
			e.result = append(e.result, next)
		}
	} else {
		if matchLen == 0 {
			//出现字面量，匹配项数据输出
			//fmt.Printf("e.matchNum:%v",e.matchNum)
			toReturn = append(toReturn, e.matchNumList...)
			toReturn = append(toReturn, e.matchNum)
			if e.matchNum == math.MaxUint8 {
				toReturn = append(toReturn, 0)
			}
			toReturn = append(toReturn, e.result...)
			e.matchNumList = nil
			e.matchNum = 0
			e.result = nil
			e.matchStatus = false

			//新增字面量
			e.literalNum++
			e.result = append(e.result, next)
		} else {
			//新增匹配项 超过就扩展
			if e.matchNum == math.MaxUint8 {
				e.matchNumList = append(e.matchNumList, math.MaxUint8)
				e.matchNum = 0
			}
			e.matchNum++
			e.result = append(e.result, uint8(offset>>8))
			e.result = append(e.result, uint8(offset))
			e.result = append(e.result, matchLen)
			e.result = append(e.result, next)
		}
	}
	return toReturn
}

func (e *lz77SequenceEncoder) compress() []uint8 {
	var totalResult []uint8
	for len(e.headerBuffer) > 0 {
		offset, matchLen, next := e.getLongestPrefix(e.slidingWindow, e.headerBuffer)
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
			dropLen := len(e.slidingWindow) - config.SlidingWindowSize
			e.slidingWindow = e.slidingWindow[dropLen:]
			e.slidingWindowOffset = e.slidingWindowOffset + uint64(dropLen)
		}
		e.updateHash(matchLen + 1)
		e.headerBuffer = e.headerBuffer[matchLen+1:]
		//fmt.Printf("slidingWindow:%v",e.slidingWindow)
		//fmt.Printf("headerBuffer:%v\n",e.headerBuffer)
		var toReturn []uint8
		if !e.matchStatus {
			//字面量状态
			if matchLen == 0 {
				//新增字面量 //超过上限就扩展位来记录
				if e.literalNum == math.MaxUint8 {
					e.literalNumList = append(e.literalNumList, math.MaxUint8)
					e.literalNum = 0
				}
				e.literalNum++
				e.result = append(e.result, next)
			} else {
				//出现匹配项，字面量数据输出
				toReturn = append(toReturn, e.literalNumList...)
				toReturn = append(toReturn, e.literalNum)
				if e.literalNum == math.MaxUint8 {
					toReturn = append(toReturn, 0)
				}
				toReturn = append(toReturn, e.result...)
				e.literalNumList = nil
				e.literalNum = 0
				e.result = nil
				e.matchStatus = true

				//新增匹配项
				e.matchNum++
				e.result = append(e.result, uint8(offset>>8))
				e.result = append(e.result, uint8(offset))
				e.result = append(e.result, matchLen)
				e.result = append(e.result, next)
			}
		} else {
			if matchLen == 0 {
				//出现字面量，匹配项数据输出
				//fmt.Printf("e.matchNum:%v",e.matchNum)
				toReturn = append(toReturn, e.matchNumList...)
				toReturn = append(toReturn, e.matchNum)
				if e.matchNum == math.MaxUint8 {
					toReturn = append(toReturn, 0)
				}
				toReturn = append(toReturn, e.result...)
				e.matchNumList = nil
				e.matchNum = 0
				e.result = nil
				e.matchStatus = false

				//新增字面量
				e.literalNum++
				e.result = append(e.result, next)
			} else {
				//新增匹配项 超过就扩展
				if e.matchNum == math.MaxUint8 {
					e.matchNumList = append(e.matchNumList, math.MaxUint8)
					e.matchNum = 0
				}
				e.matchNum++
				e.result = append(e.result, uint8(offset>>8))
				e.result = append(e.result, uint8(offset))
				e.result = append(e.result, matchLen)
				e.result = append(e.result, next)
			}
		}
		totalResult = append(totalResult, toReturn...)
	}
	if e.matchStatus {
		totalResult = append(totalResult, e.matchNumList...)
		totalResult = append(totalResult, e.matchNum)
		if e.matchNum == math.MaxUint8 {
			totalResult = append(totalResult, 0)
		}
	} else {
		totalResult = append(totalResult, e.literalNumList...)
		totalResult = append(totalResult, e.literalNum)
		if e.literalNum == math.MaxUint8 {
			totalResult = append(totalResult, 0)
		}
	}
	totalResult = append(totalResult, e.result...)
	//fmt.Printf("result:%v",totalResult)
	return totalResult
}

type lz77SequenceDecoder struct {
	slidingWindow []uint8
	offset        uint16
	matchLen      uint8
	offsetDone1   bool
	offsetDone2   bool
	matchLenDone  bool

	matchStatus    bool
	literalNum     uint8
	literalNumList []uint8
	literalNumDone bool
	matchNum       uint8
	matchNumList   []uint8
	matchNumDone   bool
	result         []uint8
}

func newLz77SequenceDecoder() *lz77SequenceDecoder {
	return &lz77SequenceDecoder{}
}

func (d *lz77SequenceDecoder) decompressWithNewByte(nb uint8) []uint8 {
	//fmt.Printf("slidingWindow:%v",d.slidingWindow)
	//fmt.Printf("\nnb:%v ",nb)
	//fmt.Printf("literalNum:%v ",d.literalNum)
	//fmt.Printf("matchNum:%v\n",d.matchNum)
	//fmt.Printf("literalNum:%v,list:%v,matchNum:%v,list:%v,offset:%v,matchLen:%v,next:%v\n", d.literalNum,d.literalNumList, d.matchNum,d.matchNumList, d.offset, d.matchLen, nb)
	var toReturn []uint8
	if !d.matchStatus {
		if !d.literalNumDone {
			if nb == math.MaxUint8 {
				d.literalNumList = append(d.literalNumList, math.MaxUint8)
			} else {
				d.literalNum = nb
				d.literalNumDone = true
			}
			return nil
		}
		if d.literalNum == 0 && len(d.literalNumList) > 0 {
			d.literalNum = d.literalNumList[0]
			d.literalNumList = d.literalNumList[1:]
		}
		//fmt.Printf("num:%v,list:%v\n",d.literalNum,d.literalNumList)
		if d.literalNum > 0 {
			d.result = append(d.result, nb)
			d.slidingWindow = append(d.slidingWindow, nb)
			d.literalNum--
		} else {
			toReturn = append(toReturn, d.result...)

			d.matchStatus = true
			d.literalNumDone = false
			d.literalNumList = nil
			d.result = nil

			//fmt.Printf("nb:%v",nb)
			if nb == math.MaxUint8 {
				d.matchNumList = append(d.matchNumList, math.MaxUint8)
			} else {
				d.matchNum = nb
				d.matchNumDone = true
			}
		}
	} else {
		//fmt.Printf("matchNum:%v",d.matchNum)
		if !d.matchNumDone {
			if nb == math.MaxUint8 {
				d.matchNumList = append(d.matchNumList, math.MaxUint8)
			} else {
				d.matchNum = nb
				d.matchNumDone = true
			}
			return nil
		}
		if d.matchNum == 0 && len(d.matchNumList) > 0 {
			d.matchNum = d.matchNumList[0]
			d.matchNumList = d.matchNumList[1:]
		}
		if d.matchNum > 0 {
			if !d.offsetDone1 {
				//fmt.Printf("%v\n",nb)
				d.offset = uint16(nb) << 8
				d.offsetDone1 = true
				return nil
			}
			if !d.offsetDone2 {
				//fmt.Printf("%v\n",nb)
				d.offset = d.offset + uint16(nb)
				d.offsetDone2 = true
				return nil
			}
			if !d.matchLenDone {
				d.matchLen = nb
				d.matchLenDone = true
				return nil
			}
			//fmt.Printf("slidingWindow:%v\n",d.slidingWindow)
			//fmt.Printf("offset:%v,matchLen:%v,next:%v\n",d.offset,d.matchLen,nb)
			d.offset = d.offset - uint16(config.SlidingWindowSize-len(d.slidingWindow))
			d.result = append(d.result, d.slidingWindow[d.offset:d.offset+uint16(d.matchLen)]...)
			d.result = append(d.result, nb)
			d.matchNum--
			d.offsetDone1 = false
			d.offsetDone2 = false
			d.matchLenDone = false

			d.slidingWindow = append(d.slidingWindow, d.slidingWindow[d.offset:d.offset+uint16(d.matchLen)]...)
			d.slidingWindow = append(d.slidingWindow, nb)
		} else {
			toReturn = append(toReturn, d.result...)

			d.matchStatus = false
			d.matchNumDone = false
			d.matchNumList = nil
			d.offsetDone1 = false
			d.offsetDone2 = false
			d.matchLenDone = false
			d.result = nil

			if nb == math.MaxUint8 {
				d.literalNumList = append(d.literalNumList, math.MaxUint8)
			} else {
				d.literalNum = nb
				d.literalNumDone = true
			}
		}
	}
	//fmt.Printf("slidingWindow:%v\n",d.slidingWindow)
	//滑动窗口舍弃旧数据使得大小不超过上限
	if len(d.slidingWindow) > config.SlidingWindowSize {
		d.slidingWindow = d.slidingWindow[len(d.slidingWindow)-config.SlidingWindowSize:]
	}
	//fmt.Printf("len(d.slidingWindow):%v\n",len(d.slidingWindow))
	//fmt.Printf("toReturn:%v",toReturn)
	return toReturn
}
