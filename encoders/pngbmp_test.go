package encoders

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	enc := &EncoderPNGBMP{}
	blockLen, blockN := 4*1024*1024, 64

	data := make([]byte, blockLen) //4MB
	rand.Read(data)

	start := time.Now()
	for i := 0; i < blockN; i++ { //256MB
		enc.Encode(data)
	}
	usedTime := time.Now().Sub(start).Seconds()
	fmt.Println("Time:", usedTime, "s", "Speed:", float64(blockLen*blockN)/usedTime/1048576, "M/s")
}
