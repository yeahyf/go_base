package file

import "os"

//判断文件是否存在或者不是文件
//返回true标识文件存在，false标识不存在
func FileExist(filePath string) bool {
	fInfo, err := os.Stat(filePath)
	return err == nil && !fInfo.IsDir()
}
