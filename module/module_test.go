package module

import (
	"fmt"
	"io"
	"log"
	"sync"
	"testing"
)

func init() {
	testing.Init()
}

//func TestCli(t *testing.T) {
//	module := NewCliModule()
//	module.Init()
//	err := module.App.Run(os.Args)
//	if err != nil {
//		log.Printf("cli module err:%v", err)
//		return
//	}
//	log.Printf("operateType:%v\n", module.OperateType)
//	log.Printf("path:%v\n", module.Path)
//}

func TestLz4(t *testing.T) {
	test := []uint8("abababababab")
	//var test []uint8
	//for i := uint8(0); i < 128; i++ {
	//	test = append(test, i)
	//}
	//test = append(test, uint8(128))
	//for i := uint8(0); i < 128; i++ {
	//	test = append(test, i)
	//}
	//test = append(test, uint8(128))
	encoder := newLz4SequenceEncoder()
	for _, v := range test {
		if !encoder.done {
			encoder.compressWithNewByte(v)
		}
	}
	if !encoder.done {
		encoder.compress()
	}
	fmt.Printf("\nliteralsLen:%v\n", encoder.sequence.literalsLen)
	fmt.Printf("\nliterals:%d\n", encoder.sequence.literals)
	fmt.Printf("\noffset:%v\n", encoder.sequence.offset)
	fmt.Printf("\nMatchLen:%v\n", encoder.sequence.matchLen)
	fmt.Printf("\nbadBuffer:%d\n", encoder.badBuffer)
	fmt.Printf("\nbyteList:%d\n", encoder.getByByte())

	decoder := newLz4SequenceDecoder()
	for _, v := range encoder.getByByte() {
		if !decoder.done {
			decoder.decompressWithNewByte(v)
		}
	}
	fmt.Printf("\nliterals:%d\n", decoder.sequence.literals)
	fmt.Printf("\nbyteList:%d\n", decoder.getByByte())
}

func TestNewCompressModule(t *testing.T) {
	test := []uint8("I am happy to join with you today in what will go down in history as the greatest demonstration for freedom in the history of our nation.\n\nFive score years ago, a great American, in whose symbolic shadow we stand today, signed the Emancipation Proclamation. This momentous decree came as a great beacon light of hope to millions of Negro slaves who had been seared in the flames of withering injustice. It came as a joyous daybreak to end the long night of their captivity.\n\nBut one hundred years later, the Negro still is not free. One hundred years later, the life of the Negro is still sadly crippled by the manacles of segregation and the chains of discrimination. One hundred years later, the Negro lives on a lonely island of poverty in the midst of a vast ocean of material prosperity. One hundred years later, the Negro is still languished in the corners of American society and finds himself an exile in his own land. And so we've come here today to dramatize a shameful condition.\n\nIn a sense we've come to our nation's capital to cash a check. When the architects of our republic wrote the magnificent words of the Constitution and the Declaration of Independence, they were signing a promissory note to which every American was to fall heir. This note was a promise that all men, yes, black men as well as white men, would be guaranteed the \"unalienable Rights\" of \"Life, Liberty and the pursuit of Happiness.\" It is obvious today that America has defaulted on this promissory note, insofar as her citizens of color are concerned. Instead of honoring this sacred obligation, America has given the Negro people a bad check, a check which has come back marked \"insufficient funds.\"\n\nBut we refuse to believe that the bank of justice is bankrupt. We refuse to believe that there are insufficient funds in the great vaults of opportunity of this nation. And so, we've come to cash this check, a check that will give us upon demand the riches of freedom and the security of justice.\n\nWe have also come to this hallowed spot to remind America of the fierce urgency of Now. This is no time to engage in the luxury of cooling off or to take the tranquilizing drug of gradualism. Now is the time to make real the promises of democracy. Now is the time to rise from the dark and desolate valley of segregation to the sunlit path of racial justice. Now is the time to lift our nation from the quicksands of racial injustice to the solid rock of brotherhood. Now is the time to make justice a reality for all of God's children.\n\nIt would be fatal for the nation to overlook the urgency of the moment. This sweltering summer of the Negro's legitimate discontent will not pass until there is an invigorating autumn of freedom and equality. Nineteen sixty-three is not an end, but a beginning. And those who hope that the Negro needed to blow off steam and will now be content will have a rude awakening if the nation returns to business as usual. And there will be neither rest nor tranquility in America until the Negro is granted his citizenship rights. The whirlwinds of revolt will continue to shake the foundations of our nation until the bright day of justice emerges.\n\nBut there is something that I must say to my people, who stand on the warm threshold which leads into the palace of justice: In the process of gaining our rightful place, we must not be guilty of wrongful deeds. Let us not seek to satisfy our thirst for freedom by drinking from the cup of bitterness and hatred. We must forever conduct our struggle on the high plane of dignity and discipline. We must not allow our creative protest to degenerate into physical violence. Again and again, we must rise to the majestic heights of meeting physical force with soul force.\n\nThe marvelous new militancy which has engulfed the Negro community must not lead us to a distrust of all white people, for many of our white brothers, as evidenced by their presence here today, have come to realize that their destiny is tied up with our destiny. And they have come to realize that their freedom is inextricably bound to our freedom.\n\nWe cannot walk alone.\n\nAnd as we walk, we must make the pledge that we shall always march ahead.\n\nWe cannot turn back.\n\nThere are those who are asking the devotees of civil rights, \"When will you be satisfied?\" We can never be satisfied as long as the Negro is the victim of the unspeakable horrors of police brutality. We can never be satisfied as long as our bodies, heavy with the fatigue of travel, cannot gain lodging in the motels of the highways and the hotels of the cities. **We cannot be satisfied as long as the negro's basic mobility is from a smaller ghetto to a larger one. We can never be satisfied as long as our children are stripped of their self-hood and robbed of their dignity by signs stating: \"For Whites Only.\"** We cannot be satisfied as long as a Negro in Mississippi cannot vote and a Negro in New York believes he has nothing for which to vote. No, no, we are not satisfied, and we will not be satisfied until \"justice rolls down like waters, and righteousness like a mighty stream.\"")
	//test := []uint8("abababab")
	//var test []uint8
	//for i := uint8(0); i < 128; i++ {
	//	test = append(test, i)
	//}
	//test = append(test, uint8(128))
	//for i := uint8(0); i < 128; i++ {
	//	test = append(test, i)
	//}
	//test = append(test, uint8(128))
	//for i := 0; i < 1000; i++ {
	//	test = append(test, uint8(i%16))
	//}
	//fmt.Printf("test:%v\n", test)
	fmt.Printf("压缩前:%v\n", len(test))
	out1, in1 := io.Pipe()
	out2, in2 := io.Pipe()
	out3, in3 := io.Pipe()
	out4, in4 := io.Pipe()
	compressModule := NewCompressModule(out1, in2)
	decompressModule := NewCompressModule(out3, in4)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(in *io.PipeWriter, test []uint8) {
		defer wg.Done()
		_, err := in.Write(test)
		if err != nil {
			log.Printf("in write err:%v", err)
			return
		}
		//fmt.Printf("input close the writer\n")
		_ = in.Close()
	}(in1, test)

	wg.Add(1)
	go func(module *CompressModule) {
		defer wg.Done()
		err := module.Compress()
		if err != nil {
			log.Printf("Compress err:%v", err)
			return
		}
	}(compressModule)

	wg.Add(1)
	go func(out *io.PipeReader,in *io.PipeWriter) {
		defer wg.Done()
		byteCount := 0
		for {
			data := make([]uint8, 64)
			n, err := out.Read(data)
			if n > 0 {
				byteCount = byteCount + n
				_, err = in.Write(data[0:n])
				if err != nil {
					log.Printf("in write err:%v", err)
					return
				}
			}
			if err != nil && err != io.EOF {
				log.Printf("out read err:%v", err)
				return
			}
			if err == io.EOF || n == 0 {
				break
			}
			//fmt.Printf("seq:%v\n", data[0:n])
		}
		fmt.Printf("压缩后:%v\n", byteCount)
		_ = in.Close()
	}(out2,in3)

	wg.Add(1)
	go func(module *CompressModule) {
		defer wg.Done()
		err := module.Decompress()
		if err != nil {
			log.Printf("Decompress err:%v", err)
			return
		}
	}(decompressModule)

	wg.Add(1)
	go func(out *io.PipeReader) {
		defer wg.Done()
		byteCount := 0
		for {
			data := make([]uint8, 64)
			n, err := out.Read(data)
			if n > 0 {
				byteCount = byteCount + n
				//fmt.Print(data[0:n])
			}
			if err != nil && err != io.EOF {
				log.Printf("out read err:%v", err)
				return
			}
			if err == io.EOF || n == 0 {
				break
			}
			//fmt.Printf("seq:%v\n", data[0:n])
		}
		fmt.Printf("解压后:%v\n", byteCount)
	}(out4)

	wg.Wait()

}
