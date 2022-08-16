package request

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
	"tinypng/cPrint"
)

type FileOptions struct {
	FilePath   string // 原路径
	TargetPath string // 目标路径 (如何是文件夹或根据文件夹去的远逻辑展示数据)
}

// 获取下载地址
func (info *FileOptions) GetTargetPath() string {
	// 如果没有指定下载路径 就不处理
	if info.TargetPath == "" {
		return info.FilePath
	}
	return path.Join(info.TargetPath, path.Base(info.FilePath))
}

// 获取随机Ip地址
func GEtRandomIP() string {
	// TODO:监听20次和今天是否已经使用
	rand.Seed(time.Now().UnixMicro())
	nums := make([]int, 4)
	for i := 0; i < 4; i++ {
		nums[i] = rand.Intn(254) + 1 // 1-255
	}
	return fmt.Sprintf("%d.%d.%d.%d", nums[0], nums[1], nums[2], nums[3])
}

type UploadResult struct {
	Input  FileInfo `json:"input" form:"input"`
	Output FileInfo `json:"output" form:"output"`
}

type FileInfo struct {
	Size   int64   `json:"size" form:"size"`
	Type   string  `json:"type" form:"type"`
	Width  uint    `json:"width" form:"width"`
	Height uint    `json:"height" form:"height"`
	Ratio  float64 `json:"ratio" form:"ratio"`
	Url    string  `json:"url" form:"url"`
}

// 上传文件并返回结果集
func UploadFileToTinyPng(filePath string) (*UploadResult, error) {
	client := &http.Client{}
	openFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer openFile.Close() // 关闭
	// 开始上传文件
	req, err := http.NewRequest("POST", "https://tinify.cn/web/shrink", openFile)
	if err != nil {
		return nil, err
	}
	// 获取随机ip
	ip := GEtRandomIP()
	// 整理header信息
	req.Header.Add("Postman-Token", strconv.FormatInt(time.Now().UnixMicro(), 10))
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")
	req.Header.Add("X-Forwarded-For", ip) // 设定随机ip
	// resp.Request.Host = "TinyPNG – Compress PNG images while preserving transparency"
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("上传图片失败")
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result UploadResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// 根据下载返回数据下载图片信息
func DownloadTinyPngFile(result *UploadResult, filePath string) error {
	// 生成文件  								   写			创建		覆盖
	openFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("下载文件-打开文件失败")
	}
	defer openFile.Close()
	// 获取文件名
	fileName := path.Base(filePath)
	res, err := http.Get(fmt.Sprintf("%s/%s", result.Output.Url, url.QueryEscape(fileName)))
	if err != nil {
		return fmt.Errorf("下载文件-下载文件失败,失败原因：%s", err.Error())
	}
	defer res.Body.Close()
	_, err = io.Copy(openFile, res.Body)
	if err != nil {
		return fmt.Errorf("下载文件-写入文件失败,失败原因：%s", err.Error())
	}
	return nil
}

// 压缩图片
func CompressionFile(option *FileOptions) error {
	if option.FilePath == "" {
		return fmt.Errorf("原文件地址不能为空")
	}
	// 上传文件
	result, err := UploadFileToTinyPng(option.FilePath)
	if err != nil {
		ShowError(option.FilePath, err)
		return err
	}
	err = DownloadTinyPngFile(result, option.GetTargetPath())
	if err != nil {
		ShowError(option.FilePath, err)
		return err
	}
	ShowSuccess(result)
	return nil
}

// 显示错误消息
func ShowError(filePath string, err error) {
	logList := []*cPrint.PrintStruct{
		{Message: "图片压缩失败:", ColorType: cPrint.ColorTypeRed, ShowType: cPrint.ShowTypeHigh},
		{Message: err.Error(), ShowType: cPrint.ShowTypeHigh},
	}
	cPrint.PrintList(logList)
}

// 显示成功的信息
func ShowSuccess(result *UploadResult) {
	logList := []*cPrint.PrintStruct{
		{Message: "压缩完成:", ColorType: cPrint.ColorTypeGreen, ShowType: cPrint.ShowTypeHigh},
		{Message: fmt.Sprintf("源文件大小：%s，压缩完大小：%s，压缩比：%.2f%%", ShowFileSize(result.Input.Size), ShowFileSize(result.Output.Size), result.Output.Ratio*100), ShowType: cPrint.ShowTypeHigh},
	}
	cPrint.PrintList(logList)
}

// 展示文件大小
func ShowFileSize(fileSize int64) string {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	}
	return ""
}

// 并发执行
func CompressionFiles(list []*FileOptions, size int) []error {
	if size == 0 {
		size = 10
	}
	var wg sync.WaitGroup
	ch := make(chan struct{}, 10)
	defer close(ch)
	errs := make([]error, 0)
	for i := 0; i < len(list); i++ {
		ch <- struct{}{}
		wg.Add(1)
		go func(options *FileOptions) {
			defer wg.Done()
			err := CompressionFile(options)
			if err != nil {
				errs = append(errs, err)
			}
			<-ch
		}(list[i])
	}
	wg.Wait()
	return errs
}
