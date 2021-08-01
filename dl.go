package main

import (
	"CDNDrive/drivers"
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gookit/color"
)

//所有需要下载的 URL 过一遍
func _forcehttpsURL(url string, f bool) (retURL string) {
	retURL = url
	if _debug {
		fmt.Println(url, f)
	}
	if f {
		retURL = strings.Replace(retURL, "http://", "https://", 1)
		return
	}
	return
}

func HandlerDownload(args []string, forceHTTPS bool, threadN int, batch bool) {
	//TODO bdex://xxx+sgdrive:// 这样的格式，两个一起下载？
	if batch {
		txt_batchdl := "<fg=black;bg=green>批量下载模式：</>"
		color.Println(txt_batchdl, "请输入链接，可以复制整段文字，会自动匹配。输入一行 end 三个字母，或按 Ctrl+D 结束。")
		s := bufio.NewScanner(os.Stdin)
		metaurls := make([]string, 0)

		for s.Scan() {
			line := s.Text()
			if strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r") == "end" {
				break
			}

			driver := queryDriverByMetaLink(line)
			if driver == nil {
				continue
			}
			metaurls = append(metaurls, driver.Real2Meta(driver.Meta2Real(line))) //TODO 这样不兼容bdrive
		}
		if len(metaurls) > 0 {
			color.Println(txt_batchdl, "检测到", len(metaurls), "个链接，开始下载。")
		} else {
			color.Println(txt_batchdl, "没有检测到链接。")
			return
		}

		for i, metaurl := range metaurls {
			color.Println(txt_batchdl, "正在下载第", i+1, "/", len(metaurls), "个文件")
			download(metaurl, threadN, forceHTTPS)
		}

	} else {
		download(args[0], threadN, forceHTTPS)
	}
}

//下载一个文件？
func download(link string, threadN int, forceHTTPS bool) {
	//常用文本
	txt_CannotDownload := "<fg=black;bg=red>下载失败：</>"

	//开始
	time_start := time.Now()
	d := queryDriverByMetaLink(link)
	if d == nil {
		fmt.Println("未知的链接格式")
		return
	}
	colorLogger.Println("<fg=black;bg=green>正在下载链接：</>", link, "类型：", d.DisplayName())

	//获取 block dict 图片
	req, _ := http.NewRequest("GET", _forcehttpsURL(d.Meta2Real(link), forceHTTPS), nil)
	for k, v := range d.Headers() {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		colorLogger.Println(err.Error())
		return
	}
	if resp.Body == nil {
		colorLogger.Println(txt_CannotDownload)
		return
	}
	defer resp.Body.Close()

	//尝试解码获取 block dict
	data, err := readPhotoBytes(resp.Body, d.Encoder())
	if err != nil {
		colorLogger.Println(txt_CannotDownload, "readPhotoBytes:", err.Error())
		return
	}
	v := &metaJSON{}
	err = json.Unmarshal(data, v)
	if err != nil {
		colorLogger.Println(txt_CannotDownload, "json.Unmarshal:", err.Error())
		return
	}
	blockN := len(v.BlockDicts)
	colorLogger.Println("<fg=black;bg=green>发现文件：</><yellow>", v.FileName, "</>大小", ConvertFileSize(v.Size), "创建时间", FormatTime(v.Time), "分块数", blockN, "sha1:", v.Sha1)

	//本地文件
	f, err := os.OpenFile(v.FileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		colorLogger.Println(txt_CannotDownload, err.Error())
	}
	defer f.Close()

	// 续传
	finishMap := make([]bool, blockN)
	finishMap_read := false
	stat, _ := f.Stat()
	if stat.Size() == 0 {
	} else if stat.Size() < v.Size {
		finishMap2 := make([]bool, blockN)
		f, err := os.Open(v.FileName + ".cdndrive")
		if err == nil {
			err = gob.NewDecoder(f).Decode(&finishMap2)
			if err == nil {
				finishMap = finishMap2
				finishMap_read = true
			}
		} else if stat.Size() > 0 {
			colorLogger.Println(txt_CannotDownload, "文件已经存在")
			return
		}
	} else if stat.Size() == v.Size {
		colorLogger.Println(txt_CannotDownload, "文件已经存在，并且大小一致")
		return
		//TODO sha1校验看看？
	} else {
		colorLogger.Println(txt_CannotDownload, "文件已经存在")
		return
	}

	if finishMap_read {
		colorLogger.Println("检测到下载进度，继续下载。")
	}

	//添加任务，等待完成
	f.Seek(0, 0)
	chanTask := make(chan metaJSON_Block, blockN)
	chanStatus := make(chan int, 0)
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	lock := &sync.Mutex{}

	var _offset int64
	var downloadSize int64
	for i, task := range v.BlockDicts {
		task.i = i
		task.offset = _offset

		_offset += int64(task.Size)

		if finishMap[i] {
			continue //已经下载了的
		}

		downloadSize += int64(task.Size)
		chanTask <- task
	}

	for j := 0; j < threadN; j++ {
		go worker_dl(chanTask, chanStatus, ctx, j, forceHTTPS, d, f, lock, wg, finishMap)
	}

	go func() {
		//TODO 显示速度
		var finishedBlockCounter int
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case finishedBlockID := <-chanStatus:
				if finishedBlockID < 0 { //负数是出错代码，此时该driver退出
					colorLogger.Println(txt_CannotDownload, d.DisplayName(), v.FileName)
					cancel()
					return
				}

				finishMap[finishedBlockID] = true
				finishedBlockCounter++

				if finishedBlockCounter == blockN {
					os.Remove(v.FileName + ".cdndrive")
					seconds := time.Now().Sub(time_start).Seconds()
					colorLogger.Printf("<fg=black;bg=green>下载完成：</> <yellow>%s</> 用时 %f 秒，平均速度 %s\n", v.FileName, seconds, ConvertFileSize(int64(float64(downloadSize)/seconds)))
					return
				}

			}
		}
	}()

	wg.Add(1)
	wg.Wait()
	cancel()
}

func worker_dl(chanTask chan metaJSON_Block, chanStatus chan int, ctx context.Context, workerID int, forceHTTPS bool, d drivers.Driver, f *os.File, lock *sync.Mutex, wg *sync.WaitGroup, finishMap []bool) {
	txt_CannotDownloadBlock := "<fg=black;bg=red>无法下载分块图片：</>"

	client := &http.Client{}
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-chanTask:
			try_max := 10
			for i := 0; i < try_max; i++ { //尝试10次
				err := func() (err error) {
					//下载分块图片
					req, _ := http.NewRequest("GET", _forcehttpsURL(task.URL, forceHTTPS), nil)
					for k, v := range d.Headers() {
						req.Header.Set(k, v)
					}

					resp, err := client.Do(req)
					if err != nil {
						colorLogger.Println(txt_CannotDownloadBlock, err.Error())
						return
					}
					if resp.Body == nil {
						colorLogger.Println(txt_CannotDownloadBlock, "resp.Body == nil")
						return
					}
					defer resp.Body.Close()

					//尝试解码，校验
					data, err := readPhotoBytes(resp.Body, d.Encoder())
					if err != nil {
						colorLogger.Println(txt_CannotDownloadBlock, "readPhotoBytes", err.Error())
						return
					}

					sha1sum := sha1.Sum(data)
					if fmt.Sprintf("%x", sha1sum) != task.Sha1 {
						err = errors.New("校验不通过，可能图片损坏。")
						return
					}

					//写入本地文件、保存进度
					lock.Lock()

					f.Seek(task.offset, 0)
					f.Write(data)

					finishMap[task.i] = true
					f2, err := os.OpenFile(f.Name()+".cdndrive", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
					if err == nil {
						gob.NewEncoder(f2).Encode(&finishMap)
						f2.Close()
					}
					lock.Unlock()

					//完成？
					colorLogger.Println(d.DisplayName(), "分块", task.i+1, "下载完毕。")
					chanStatus <- task.i
					return
				}()
				if err != nil {
					if i < try_max-1 {
						colorLogger.Println(d.DisplayName(), "分块", task.i+1, "第", i+1, "次下载失败，重试。")
					} else {
						colorLogger.Println(d.DisplayName(), "分块", task.i+1, "第", i+1, "次下载失败，不重试，文件下载失败。")
						chanStatus <- -1 //停止代码 -1 上传失败
					}
				} else {
					break
				}
			}
		}
	}
}
