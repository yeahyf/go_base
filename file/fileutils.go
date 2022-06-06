package file

import (
	"bufio"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/yeahyf/go_base/log"
	"github.com/yeahyf/go_base/utils"
)

//ExistFile 判断文件是否存在或者不是文件,返回true表示文件存在，false标识不存在
func ExistFile(filePath string) bool {
	fInfo, err := os.Stat(filePath)
	return err == nil && !fInfo.IsDir()
}

//ExistsPath 判断目录是否存在,返回true表示存在,false表示不存在
func ExistsPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//CopyFile 拷贝文件  要拷贝的文件路径 拷贝到哪里
func CopyFile(source, dest string) (int64, error) {
	if source == "" || dest == "" {
		return 0, fmt.Errorf("%s or %s is illegle", source, dest)
	}
	var sourceInfo os.FileInfo
	var err error
	sourceInfo, err = os.Stat(source)
	if err != nil {
		return 0, fmt.Errorf("stat %s error", source)
	}
	if !sourceInfo.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is Not a regular file", source)
	}

	//打开文件资源
	var sourceFile *os.File
	sourceFile, err = os.Open(source)
	//养成好习惯。操作文件时候记得添加 defer 关闭文件资源代码
	if err != nil {
		return 0, fmt.Errorf("couldn't open file %s", source)
	}
	defer utils.CloseAction(sourceFile)

	destFileName := filepath.Base(dest)
	index := strings.LastIndex(dest, destFileName)
	destPath := dest[0:index]

	if result, _ := ExistsPath(destPath); !result {
		err := os.MkdirAll(destPath, os.ModePerm)
		if err != nil {
			log.Errorf("couldn't create dirs", err)
		}
	}
	var destFile *os.File
	destFile, err = os.Create(dest)
	if err != nil {
		return 0, fmt.Errorf("couldn't create file %s", dest)
	}
	defer utils.CloseAction(destFile)

	//进行数据拷贝
	return io.Copy(destFile, sourceFile)
}

//ReadLine 逐行读取文本文件进行处理
func ReadLine(fileName string, handler func(*string)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer utils.CloseAction(f)

	buffer := bufio.NewReader(f)
	for {
		line, err := buffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				//可能最后一行不是空行
				line = strings.TrimSpace(line)
				if len(line) > 0 {
					handler(&line)
				}
				return nil
			}
			return err
		}
		line = strings.TrimSpace(line)
		handler(&line)
	}
}

//Compress 压缩文件
//原始地址，目标地址
func Compress(srcFile, destFile *string) error {
	//创建目标文件
	var newFile *os.File
	var err error
	newFile, err = os.Create(*destFile)
	if err != nil {
		return err
	}
	defer utils.CloseAction(newFile)

	var oldFile *os.File
	oldFile, err = os.Open(*srcFile)
	if err != nil {
		return err
	}
	defer utils.CloseAction(oldFile)

	zw := gzip.NewWriter(newFile)
	var fileStat os.FileInfo
	fileStat, err = oldFile.Stat()
	if err != nil {
		return err
	}

	zw.Name = fileStat.Name()
	zw.ModTime = fileStat.ModTime()
	_, err = io.Copy(zw, oldFile)
	if err != nil {
		return err
	}

	err = zw.Flush()
	if err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return nil
}

//SHA1File 计算文件的sha1值
func SHA1File(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer utils.CloseAction(file)

	m := sha1.New()
	_, err = io.Copy(m, file)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(m.Sum(nil))
}

//FtpFile 使用ftp进行文件传输，注意需要对ftp服务器进行适当的配置
func FtpFile(destHost, sourceFilePath, destPath, destFileName, user, passwd string) error {
	//建立连接
	c, err := ftp.Dial(destHost, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	//退出
	defer func() {
		err = c.Quit()
		if err != nil {
			log.Error("ftp quit error!", err)
		}
	}()

	//登录
	err = c.Login(user, passwd)
	if err != nil {
		return err
	}

	//修改路径，需要修改selinux的配置才可以
	err = c.ChangeDir(destPath)
	if err != nil {
		return err
	}

	file, err := os.Open(sourceFilePath)
	if err != nil {
		return err
	}

	r := bufio.NewReader(file)

	//开始传递文件
	err = c.Stor(destFileName, r)
	if err != nil {
		return err
	}
	return nil
}
