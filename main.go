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
		err = pack(inputName, outputName)
	case config.OpUnpack:
		outputName = cliModule.Path[:strings.LastIndex(cliModule.Path, ".pk")]
		err = unpack(inputName, outputName)
	case config.OpPackAndCompress:
		outputName = cliModule.Path + ".pk.cp"
		err = packAndCompress(inputName, outputName)
	case config.OpDecompressAndUnpack:
		outputName = cliModule.Path[:strings.LastIndex(cliModule.Path, ".pk.cp")]
		err = decompressAndUnpack(inputName, outputName)
	}
	if err != nil {
		log.Printf("cli cliModule err:%v", err)
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

	compressModule := module.NewCompressModule(bufSource, bufTarget)
	err = compressModule.Compress()
	if err != nil {
		log.Printf("compressModule Compress err:%v", err)
		return err
	}

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

	compressModule := module.NewCompressModule(bufSource, bufTarget)
	err = compressModule.Decompress()
	if err != nil {
		log.Printf("compressModule Decompress err:%v", err)
		return err
	}
	return nil
}
func pack(inputName, outputName string) error {
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
	bufTarget := bufio.NewWriter(target)
	packModule := module.NewPackModule(inputName, bufTarget)
	err = packModule.Pack()
	if err != nil {
		log.Printf("packModule Pack err:%v", err)
		return err
	}

	return nil
}
func unpack(inputName, outputName string) error {
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
	bufSource := bufio.NewReader(source)

	unpackModule := module.NewUnpackModule(outputName, bufSource)
	err = unpackModule.Unpack()
	if err != nil {
		log.Printf("unpackModule Unpack err:%v", err)
		return err
	}

	return nil
}
func packAndCompress(inputName, outputName string) error {
	out, in := io.Pipe()
	bufOut := bufio.NewReader(out)
	bufIn := bufio.NewWriter(in)
	target, err := os.Create(outputName)
	if err != nil {
		log.Printf("create target err:%v", err)
		return err
	}
	bufTarget := bufio.NewWriter(target)

	packModule := module.NewPackModule(inputName, bufIn)
	compressModule := module.NewCompressModule(bufOut, bufTarget)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer in.Close()
		defer wg.Done()
		packModule.Pack()
	}()

	wg.Add(1)
	go func() {
		defer target.Close()
		defer wg.Done()
		compressModule.Compress()
	}()
	wg.Wait()
	return nil
}
func decompressAndUnpack(inputName, outputName string) error {
	out, in := io.Pipe()
	bufOut := bufio.NewReader(out)
	bufIn := bufio.NewWriter(in)
	source, err := os.Open(inputName)
	if err != nil {
		log.Printf("open source err:%v", err)
		return err
	}
	bufSource := bufio.NewReader(source)

	compressModule := module.NewCompressModule(bufSource, bufIn)
	packModule := module.NewUnpackModule(outputName, bufOut)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer in.Close()
		defer wg.Done()
		compressModule.Decompress()
	}()

	wg.Add(1)
	go func() {
		defer out.Close()
		defer wg.Done()
		packModule.Unpack()
	}()


	wg.Wait()
	return nil
}
