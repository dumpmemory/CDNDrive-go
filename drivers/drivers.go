package drivers

import (
	"CDNDrive/encoders"
	"context"
	"net/http"
)

type Driver interface {
	//插件名、显示名
	Name() string
	DisplayName() string

	Headers() map[string]string //http要用
	Encoder() encoders.Encoder  //使用的Encoder

	Exist(url string) (bool, error)  //图片是否已存在
	Meta2Real(metaURL string) string //bdex:// -> URL （模糊匹配）
	Real2Meta(realURL string) string //URL -> bdex://
	CheckCookie(cookie string) (bool, error)
	Upload(_data []byte, ctx context.Context, client *http.Client, cookie, sha1sum string) (string, error) //data是分块
}
