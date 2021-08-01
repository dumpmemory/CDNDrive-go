package drivers

import (
	"CDNDrive/encoders"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
)

type DriverBilibili struct {
	app_key      string
	default_url  string
	default_hdrs map[string]string
}

func NewDriverBilibili() *DriverBilibili {
	d := &DriverBilibili{}
	d.app_key = "1d8b6e7d45233436"
	d.default_hdrs = map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
		"Accept":          "application/json, text/plain, */*",
		"Accept-Encoding": "gzip, deflate",
		"Origin":          "https://t.bilibili.com",
		"Referer":         "https://t.bilibili.com/",
	}
	d.default_url = "http://i0.hdslb.com/bfs/album/{sha1}.png"
	return d
}

func (d *DriverBilibili) Name() string {
	return "bili"
}

func (d *DriverBilibili) DisplayName() string {
	return "<cyan>BiliBiliDrive</>"
}

func (d *DriverBilibili) Headers() map[string]string {
	return d.default_hdrs
}

func (d *DriverBilibili) Encoder() encoders.Encoder {
	return &encoders.EncoderPNGBMP{}
}

func (d *DriverBilibili) Exist(sha1 string) (bool, error) {
	resp, err := http.Head(d.GenURL(sha1))
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 {
		return true, nil
	}
	return false, nil
}

//Hash->URL
func (d *DriverBilibili) GenURL(sha1 string) string {
	return strings.Replace(d.default_url, "{sha1}", sha1, 1)
}

func (d *DriverBilibili) Meta2Real(metaURL string) string {
	//这个有两种scheme...
	exp, _ := regexp.Compile("(bdex|bdrive)://([a-fA-F0-9]{40})")
	matchs := exp.FindStringSubmatch(metaURL)
	if len(matchs) < 3 {
		return ""
	}
	_scheme := matchs[1]
	_sha1 := matchs[2]

	if _scheme == "bdex" {
		return d.GenURL(_sha1)
	} else if _scheme == "bdrive" {
		return strings.ReplaceAll(d.GenURL(_sha1), ".png", ".x-ms-bmp")
	}
	return ""
}

func (d *DriverBilibili) Real2Meta(realURL string) string {
	exp, _ := regexp.Compile("/([a-fA-F0-9]{40})")
	matchs := exp.FindStringSubmatch(realURL)
	if len(matchs) < 2 {
		return ""
	}

	return "bdex://" + matchs[1]
}

func (d *DriverBilibili) CheckCookie(cookie string) (bool, error) {
	req, _ := http.NewRequest("GET", "https://api.bilibili.com/x/space/myinfo", nil)
	for k, v := range d.Headers() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", "https://space.bilibili.com/")
	req.Header.Set("Origin", "https://space.bilibili.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	data, _ := ReadAllDecompress(resp)
	if strings.Index(string(data), "level_exp") > 0 {
		return true, nil
	}
	return false, errors.New("bilibili: " + string(data))
}

func (d *DriverBilibili) Upload(_data []byte, ctx context.Context, client *http.Client, cookie, sha1sum string) (string, error) {
	//编码图片
	data := (&encoders.EncoderPNGBMP{}).Encode(_data)
	//查重
	sha1sum_photo := fmt.Sprintf("%x", sha1.Sum(data))
	if e, _ := d.Exist(sha1sum_photo); e {
		return d.GenURL(sha1sum_photo), nil
	}

	//表单上传
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w2, _ := w.CreateFormFile("file_up", sha1sum_photo+".png")
	w2.Write(data)
	w.WriteField("biz", "draw")
	w.WriteField("category", "daily")
	w.Close()

	req, _ := http.NewRequest("POST", "https://api.vc.bilibili.com/api/v1/drawImage/upload", &b)
	req = req.WithContext(ctx)
	for k, v := range d.Headers() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	//解析
	v := make(map[string]interface{})
	body, _ := ReadAllDecompress(resp)
	err = json.Unmarshal(body, &v)
	if err != nil {
		return "", err
	}
	if a, ok := v["code"].(float64); ok && a == 0 {
		exp, _ := regexp.Compile("/([a-fA-F0-9]{40})")
		matchs := exp.FindStringSubmatch(string(body))
		if len(matchs) < 2 {
			return "", errors.New("解析错误")
		}

		//校验
		if !strings.Contains(strings.ToLower(matchs[1]), sha1sum_photo) {
			return "", errors.New("校验值不一致")
		}
		return d.GenURL(matchs[1]), nil
	} else {
		return "", errors.New(fmt.Sprintf("bilibili %f: %s", v["code"], v["message"]))
	}

}

func ReadAllDecompress(resp *http.Response) ([]byte, error) {
	switch resp.Header.Get("Content-Encoding") {
	// case "br":
	// 	return ioutil.ReadAll(brotli.NewReader(resp.Body))
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(gr)
	case "deflate":
		zr := flate.NewReader(resp.Body)
		defer zr.Close()
		return ioutil.ReadAll(zr)

	default:
		return ioutil.ReadAll(resp.Body)
	}
}
