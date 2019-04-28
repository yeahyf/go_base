package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//判断文件是否存在或者不是文件
//返回true标识文件存在，false标识不存在
func FileExist(filePath string) bool {
	fInfo, err := os.Stat(filePath)
	return err == nil && !fInfo.IsDir()
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//拷贝文件  要拷贝的文件路径 拷贝到哪里
func CopyFile(source, dest string) (int64, error) {
	if source == "" || dest == "" {
		return 0, fmt.Errorf("%s or %s is illegle", source, dest)
	}

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return 0, fmt.Errorf("Stat %s error!", source)
	}
	if !sourceInfo.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is Not a regular file", source)
	}

	//打开文件资源
	sourceFile, err := os.Open(source)
	//养成好习惯。操作文件时候记得添加 defer 关闭文件资源代码
	if err != nil {
		return 0, fmt.Errorf("open %s error!", source)
	}
	defer sourceFile.Close()

	destFileName := filepath.Base(dest)
	index := strings.LastIndex(dest, destFileName)
	destPath := dest[0:index]

	if result, _ := PathExists(destPath); !result {
		os.MkdirAll(destPath, os.ModePerm)
	}

	destFile, err := os.Create(dest)
	if err != nil {
		return 0, fmt.Errorf("Create %s error!", dest)
	}
	defer destFile.Close()

	//进行数据拷贝
	return io.Copy(destFile, sourceFile)
}
