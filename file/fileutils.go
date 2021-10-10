///部分文件处理工具
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

//逐行读取文本文件进行处理
func ReadLine(fileName string, handler func(*string)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	buffer := bufio.NewReader(f)

	for {
		line, err := buffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				//可能最后一行不是空行
				line = strings.TrimSpace(line)
				if len(line) > 0{
					handler(&line)
				}
				return nil
			}
			return err
		}
		line = strings.TrimSpace(line)
		handler(&line)
	}
	return nil
}

//压缩文件
//原始地址，目标地址
func Compress(srcFile, destFile *string) error {
	//创建目标文件
	newfile, err := os.Create(*destFile)
	if err != nil {
		return err
	}
	defer newfile.Close()

	oldFile, err := os.Open(*srcFile)
	if err != nil {
		return err
	}
	defer oldFile.Close()

	zw := gzip.NewWriter(newfile)
	filestat, err := oldFile.Stat()
	if err != nil {
		return err
	}

	zw.Name = filestat.Name()
	zw.ModTime = filestat.ModTime()
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

//计算文件的sha1值
func SHA1File(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	m := sha1.New()
	_, err = io.Copy(m, file)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(m.Sum(nil))
}

//使用ftp进行文件传输，注意需要对ftp服务器进行适当的配置
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
