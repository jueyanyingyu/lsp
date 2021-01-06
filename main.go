package main

import (
	"bufio"
	"github.com/jueyanyingyu/lsp/config"
	"github.com/jueyanyingyu/lsp/module"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

func main() {
	cliModule := module.NewCliModule()
	cliModule.Init()
	err := cliModule.App.Run(os.Args)
	if err != nil {
		log.Printf("cli cliModule err:%v", err)
		return
	}
	if cliModule.OperateType == config.OpUnDefine {
		log.Printf("wrong command")
		return
	}

	var inputName = cliModule.Path
	var outputName string
	switch cliModule.OperateType {
	case config.OpCompress:
		outputName = cliModule.Path + ".cp"
		err = compress(inputName, outputName)
	case config.OpDecompress:
		outputName = cliModule.Path[:strings.LastIndex(cliModule.Path, ".cp")]
		err = decompress(inputName, outputName)
	case config.OpPack:
		outputName = cliModule.Path + ".pk"
	case config.OpUnpack:
		outputName = cliModule.Path[:strings.LastIndex(cliModule.Path, ".pk")]
	case config.OpPackAndCompress:
		outputName = cliModule.Path + ".pk.cp"
	case config.OpDecompressAndUnpack:
		outputName = cliModule.Path[:strings.LastIndex(cliModule.Path, ".pk.cp")]
	}
}

func compress(inputName, outputName string) error {
	source, err := os.Open(inputName)
	if err != nil {
		log.Printf("open source err:%v", err)
		return err
	}
	defer func() {
		if err := source.Close(); err != nil {
			log.Printf("close source err:%v", err)
			return
		}
	}()
	target, err := os.Create(outputName)
	if err != nil {
		log.Printf("create target err:%v", err)
		return err
	}
	defer func() {
		if err := target.Close(); err != nil {
			log.Printf("close target err:%v", err)
			return
		}
	}()
	bufSource := bufio.NewReader(source)
	bufTarget := bufio.NewWriter(target)
	out1, in1 := io.Pipe()
	out2, in2 := io.Pipe()

	compressModule := module.NewCompressModule(out1, in2)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(source *bufio.Reader, in *io.PipeWriter) {
		defer func() {
			_ = in.Close()
			wg.Done()
		}()
		for {
			data := make([]uint8, 1024)
			n, err := source.Read(data)
			if n > 0 {
				_, err := in.Write(data[:n])
				if err != nil {
					log.Printf("in write err:%v", err)
					return
				}
			}
			if err != nil && err != io.EOF {
				log.Printf("source read err:%v", err)
				return
			}
			if err == io.EOF {
				break
			}
		}
	}(bufSource, in1)

	wg.Add(1)
	go func(module *module.CompressModule) {
		defer wg.Done()
		err := module.Compress()
		if err != nil {
			log.Printf("Compress err:%v", err)
			return
		}
	}(compressModule)

	wg.Add(1)
	go func(target *bufio.Writer, out *io.PipeReader) {
		defer func() {
			target.Flush()
			wg.Done()
		}()
		for {
			data := make([]uint8, 1024)
			n, err := out.Read(data)
			if n > 0 {
				_, err := target.Write(data[:n])
				if err != nil {
					log.Printf("targer write err:%v", err)
					return
				}
			}
			if err != nil && err != io.EOF {
				log.Printf("out read err:%v", err)
				return
			}
			if err == io.EOF {
				break
			}
		}
	}(bufTarget, out2)

	wg.Wait()

	return nil
}
func decompress(inputName, outputName string) error {
	source, err := os.Open(inputName)
	if err != nil {
		log.Printf("open source err:%v", err)
		return err
	}
	defer func() {
		if err := source.Close(); err != nil {
			log.Printf("close source err:%v", err)
			return
		}
	}()
	target, err := os.Create(outputName)
	if err != nil {
		log.Printf("create target err:%v", err)
		return err
	}
	defer func() {
		if err := target.Close(); err != nil {
			log.Printf("close target err:%v", err)
			return
		}
	}()
	bufSource := bufio.NewReader(source)
	bufTarget := bufio.NewWriter(target)
	out1, in1 := io.Pipe()
	out2, in2 := io.Pipe()

	compressModule := module.NewCompressModule(out1, in2)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(source *bufio.Reader, in *io.PipeWriter) {
		defer func() {
			_ = in.Close()
			wg.Done()
		}()
		for {
			data := make([]uint8, 1024)
			n, err := source.Read(data)
			if n > 0 {
				_, err := in.Write(data[:n])
				if err != nil {
					log.Printf("in write err:%v", err)
					return
				}
			}
			if err != nil && err != io.EOF {
				log.Printf("source read err:%v", err)
				return
			}
			if err == io.EOF {
				break
			}
		}
	}(bufSource, in1)

	wg.Add(1)
	go func(module *module.CompressModule) {
		defer wg.Done()
		err := module.Decompress()
		if err != nil {
			log.Printf("Decompress err:%v", err)
			return
		}
	}(compressModule)

	wg.Add(1)
	go func(target *bufio.Writer, out *io.PipeReader) {
		defer func() {
			target.Flush()
			wg.Done()
		}()
		for {
			data := make([]uint8, 1024)
			n, err := out.Read(data)
			if n > 0 {
				_, err := target.Write(data[:n])
				if err != nil {
					log.Printf("targer write err:%v", err)
					return
				}
			}
			if err != nil && err != io.EOF {
				log.Printf("out read err:%v", err)
				return
			}
			if err == io.EOF {
				break
			}
		}
	}(bufTarget, out2)

	wg.Wait()

	return nil
}
