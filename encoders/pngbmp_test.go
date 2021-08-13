package encoders

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkEncode(t *testing.B) {
	enc := &EncoderPNGBMP{}
	blockLen, blockN := 4*1024*1024, 64

	data := make([]byte, blockLen) //4MB
	rand.Read(data)

	var photo []byte
	start := time.Now()
	for i := 0; i < blockN; i++ { //256MB
		photo = enc.Encode(data)
	}

	usedTime := time.Now().Sub(start).Seconds()
	fmt.Println("Time:", usedTime, "s", "Speed:", float64(blockLen*blockN)/usedTime/1048576, "M/s")
	fmt.Println("Photo Size:", len(photo))
	//原始数据 4194304
	//打开压缩 4198451 ( +4.0KB
	//关闭压缩 4201579 ( +7.1KB
}
