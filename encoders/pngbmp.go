package encoders

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

type EncoderPNGBMP struct {
}

func (e *EncoderPNGBMP) Decode(data []byte) ([]byte, error) {
	if bytes.Equal(data[:4], []byte("\x89PNG")) {
		return e.DecodePNG(data)
	} else if bytes.Equal(data[:2], []byte("BM")) {
		return e.DecodeBMP(data)
	}
	return nil, errors.New("未知格式")
}

func (e *EncoderPNGBMP) DecodePNG(_data []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(_data))
	if err != nil {
		return nil, err
	}

	rect := img.Bounds()
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, img, rect.Min, draw.Src)
	data := make([]byte, 0)
	//no Alpha
	for i := 0; i < len(rgba.Pix)/4; i++ {
		data = append(data, rgba.Pix[i*4:i*4+3]...)
	}
	//read length
	l := binary.LittleEndian.Uint32(data[:4])
	return data[4 : 4+l], nil
}

func (e *EncoderPNGBMP) DecodeBMP(data []byte) ([]byte, error) {
	return data[62:], nil
}

//BMP为兼容旧版使用，这里只编码PNG
//TODO 改善内存
func (e *EncoderPNGBMP) Encode(data []byte) []byte {
	var buf bytes.Buffer

	//分辨率
	minSide, dep := 10, 3
	side := int(math.Ceil(math.Sqrt(float64(len(data)+4) / float64(dep)))) //v0.5以前忘记加上4，最多被吞掉4个字节（默认4M分块不会损坏，但是meta图片可能会）
	if side < minSide {
		side = minSide
	}
	total := side * side * dep

	//写入长度
	buf2 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf2, uint32(len(data)))
	buf.Write(buf2)

	//写入数据
	buf.Write(data)

	//填充
	if buflen := buf.Len(); buflen < total {
		buf.Write(bytes.Repeat([]byte{0}, total-buflen))
	}

	//RGBA<->RGB?????
	var pngbuf bytes.Buffer
	png.Encode(&pngbuf, &RGB{Bytes: buf.Bytes(), Side: side})

	defer buf.Reset()
	defer pngbuf.Reset()

	return pngbuf.Bytes()
}

type RGB struct {
	Bytes []byte
	Side  int
}

func (p *RGB) At(x, y int) color.Color {
	offset := x + (y * p.Side)
	offset *= 3
	return color.RGBA{
		R: p.Bytes[offset],
		G: p.Bytes[offset+1],
		B: p.Bytes[offset+2],
		A: 255,
	}
}

func (p *RGB) Bounds() image.Rectangle {
	return image.Rect(0, 0, p.Side, p.Side)
}

func (p *RGB) ColorModel() color.Model {
	return color.RGBAModel
}
