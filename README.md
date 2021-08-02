# CDNDrive-go

<h4 align="center">☁️ 喵喵喵喵，喵喵喵喵！ ☁️</h4>
<h4 align="center">☁️ 支持任意文件的全速上传与下载 ☁️</h4>

## 特色

轻量：无复杂依赖，资源占用少

自由：无文件格式与大小限制，无容量限制

稳定：带有分块校验与超时重试机制，在较差的网络环境中依然能确保文件的完整性

快速：支持多线程传输与断点续传，同时借助各个站点的 CDN 资源，能最大化地利用网络环境进行上传与下载

相对与 Python 版 apachecn/CDNDrive 本项目优势：

1. 内存占用更小
2. Go 语言编写，跨平台便携性好
3. 支持从多个源下载同一个文件，减小链接失效风险
4. 批量下载模式，全集动画一瞬下载
5. 下载支持断点续传
6. 彩色输出

## 安装

下载 release 运行即可。

## 下载文件

```
NAME:
   CDNDrive download - 下载文件

USAGE:
   CDNDrive download [command options] [arguments...]

OPTIONS:
   --https                   强制使用https (default: false)
   --batch                   批量下载模式 (default: false)
   --thread value, -t value  并发连接数 (default: 4)
   --help, -h                show help (default: false)
```

### 使用例：单个文件下载

`CDNDrive download -t 线程数 链接`

链接可以是 `bdex://64ec4075dc3cfa28cf12f147da8f4282d635657b` 这样的单个链接，也可以是 `bdex://7fc4accd6cafa0cdd9168cf5ee81a407cabe89a1+sgdrive://100520146/5C0E029D88D39A6FB795AD8D92CBF101+bjdrive://17a65fbb83c249699a9256c3bcd98a6f` 这样的英文加号分割链接，表示这个文件可以从多个来源下载。

### 使用例：多个文件下载

`CDNDrive download --batch`

然后按照提示操作即可。

## 性能指标

测试时间 20210802 版本 v0.2

下面数字均为本机（普通的笔记本）测试结果

### 上传

1. 图片编码能力

TestEncode Speed: 10.73M/s

CPU 时间大部分消耗在 png.Encode() 上

由于内存限制，图片使用单线程编码，可能无法完全发挥 CPU 性能。

2. 内存占用

单 driver 16 线程上传，Python CDNDrive 占用约 3GB，CDNDrive-go 占用约 300MB

三 driver 16 线程同时上传，网络顺畅时占用约 400MB ，网络不顺畅时会内存泄露，可以通过调整参数缓解，默认条件下大概在 1GB 左右。

可能还存在其他内存泄露的问题。

### 下载

默认参数下 (4线程)

内存占用在 200M 以内，CPU 占用在 100% 以内。

一般情况下国内网络均可跑满。

## 支持情况

现在有以下 Driver

|名称|链接|上传是否需要登录|备注|
|----|----|----|----|
|BiliBiliDrive |bdex://     |需要登录   |陈叔叔家的，推荐！
|BaijiaDrive   |bjdrive://  |无需登录   |
|SogouDrive    |sgdrive://  |无需登录   |好像服务器会自动清理文件

## 免责声明

+   请自行对重要文件做好本地备份。
+   请不要上传含有个人隐私的文件，因为无法删除。

## 致谢

原作 [apachecn/CDNDrive](https://github.com/apachecn/CDNDrive)

原作的原作 [Hsury/BiliDrive](https://github.com/Hsury/BiliDrive)

## 猫猫很可爱，请给猫猫打钱

https://nekoquq.github.io/about/
