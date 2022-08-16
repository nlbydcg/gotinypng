package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"tinypng/cPrint"
	"tinypng/request"
)

// 可以转换的图片格式
var ImgExt = []string{"webp", "jpeg", "jpg", "png"}

const FileMax = 1024 * 1024 * 5 // 最大支持文件大小

func main() {
	dirPath := flag.String("d", "", "请输入要压缩的文件地址,按逗号分割 暂只支持（webp,jpeg,jpg,png）")
	filePath := flag.String("f", "", "请输入要压缩的文件名,按逗号分割 暂只支持（webp,jpeg,jpg,png）")
	newFile := flag.String("n", "", "是否新创建文件，如果是新创建文件这需要指定文件夹")
	flag.Parse()

	if *dirPath == "" && *filePath == "" {
		cPrint.Error("请输入图片地址或者文件夹地址")
		return
	}
	if *newFile != "" {
		CreateDir(*newFile)
	}
	var files []*request.FileOptions
	var err error
	if *dirPath != "" {
		files, err = InitDirPath(*dirPath, *newFile)
		if err != nil {
			cPrint.Error(fmt.Sprintf("获取文件夹数据失败%s", err.Error()))
			return
		}
	}
	if *filePath != "" {
		filePaths, err := InitFilePath(*filePath)
		if err != nil {
			cPrint.Error(fmt.Sprintf("获取文件数据失败%s", err.Error()))
			return
		}
		for _, v := range filePaths {
			files = append(files, &request.FileOptions{
				FilePath:   v,
				TargetPath: *newFile,
			})
		}
	}
	request.CompressionFiles(files, 10)
}

// 根据文件夹地址获取到所有的文件地址
func InitDirPath(dirPath string, target string) ([]*request.FileOptions, error) {
	if dirPath == "" {
		return nil, nil
	}
	filePath := make([]*request.FileOptions, 0)
	dirs := strings.Split(dirPath, ",")
	for _, v := range dirs {
		files, err := GetFilesForDir(v, target)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			continue
		}
		filePath = append(filePath, files...)
	}
	return filePath, nil
}

// 根据文件地址获取到所有的文件
func InitFilePath(filePath string) ([]string, error) {
	if filePath == "" {
		return nil, nil
	}
	files := strings.Split(filePath, ",")
	for _, v := range files {
		has, err := FileExist(v)
		if err != nil {
			return nil, fmt.Errorf("%s,err:%s", v, err.Error())
		}
		if !has {
			return nil, fmt.Errorf("%s,err:%s", v, "文件不存在")
		}
		files = append(files, v)
	}
	return files, nil
}

// 判断文件地址
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 判断img是否满足预期
func FileExist(filePath string) (bool, error) {
	fileStat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if fileStat.IsDir() {
		return false, fmt.Errorf("并不是一个文件地址")
	}
	if fileStat.Size() > FileMax {
		return false, fmt.Errorf("文件大小超过5m")
	}
	return true, nil
}

// 获取所有文件信息
func GetFiles(dirPath string) ([]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	filesPath := make([]string, 0)
	for i := 0; i < len(files); i++ {
		file := files[i]
		filePath := path.Join(dirPath, file.Name())
		if file.IsDir() {
			cFiles, err := GetFiles(filePath)
			if err != nil {
				return nil, err
			}
			if len(cFiles) > 0 {
				filesPath = append(filesPath, cFiles...)
			}
		} else {
			if ok, _ := FileExist(filePath); !ok {
				continue
			}
			if FileIsImage(filePath) {
				filesPath = append(filesPath, filePath)
			}
		}
	}
	return filesPath, nil
}

// 获取所有文件信息
func GetFilesForDir(dirPath string, target string) ([]*request.FileOptions, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	if target != "" {
		target = path.Join(target, path.Base(dirPath))
		err = CreateDir(target)
		if err != nil {
			return nil, err
		}
	}
	filesPath := make([]*request.FileOptions, 0)
	for i := 0; i < len(files); i++ {
		file := files[i]
		filePath := path.Join(dirPath, file.Name())
		if file.IsDir() {
			cFiles, err := GetFilesForDir(filePath, target)
			if err != nil {
				return nil, err
			}
			if len(cFiles) > 0 {
				filesPath = append(filesPath, cFiles...)
			}
		} else {
			if ok, _ := FileExist(filePath); !ok {
				continue
			}
			if FileIsImage(filePath) {
				filesPath = append(filesPath, &request.FileOptions{FilePath: filePath, TargetPath: target})
			}
		}
	}
	return filesPath, nil
}

// 判断是否为指定图片格式
func FileIsImage(filepath string) bool {
	ext := path.Ext(filepath)
	for i := 0; i < len(ImgExt); i++ {
		if ImgExt[i] == ext[1:] {
			return true
		}
	}
	return false
}

func CreateDir(dirPath string) error {
	isHas, err := PathExists(dirPath)
	if err != nil {
		return err
	}
	if isHas {
		return nil
	}
	return os.MkdirAll(dirPath, os.ModePerm)
}
