package util

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"io/ioutil"
	"encoding/json"
)

// 获取当前时间戳
func GetTime() int64 {
	return time.Now().Unix()
}

// 获取当前格式化时间
func GetTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 格式化运行时间
func FormateRunTime(second int64) time.Duration {
	return time.Duration(second * 1000 * 1000 * 1000)
}

// 格式化unixtime
func FormatUnixTime(unixTime int64) string {
	return time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
}

// 检查文件夹是否存在
func CheckPathExist(pathName string) bool {
	if fi, err := os.Stat(pathName); err == nil {
		return fi.IsDir()
	}

	return false
}

// 检查文件是否存在
func CheckFileExist(filename string) bool {
	if fi, err := os.Stat(filename); err == nil {
		return !fi.IsDir()
	}

	return false
}

// 遍历文件夹，读取所有文件
func ReadDir(pathname string) ([]string, error) {
	files := make([]string, 0)

	if !CheckPathExist(pathname) {
		return files, errors.New("path is not exists:" + pathname)
	}

	filepath.Walk(pathname,
		func(path string, f os.FileInfo, err error) error {
			if f == nil || f.IsDir() {
				return err
			}
			files = append(files, path)
			return nil
		})

	return files, nil
}

// 获取当前路径
func GetDirPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)

	return filepath.Dir(path)
}

// 获取im的根目录
func GetImPath() string {
	return filepath.Dir(GetDirPath())
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// 截取字符串函数
func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}

// 获取token值
func GetToken(encryptKey string, userId int, time int64) string {
	tokenStr := fmt.Sprintf("%s#%d#%d", encryptKey, userId, time)
	md5Str := GetMd5String(tokenStr)
	return Substr(md5Str, 0, 4)
}

// 读取JSON配置
func LoadJsonConfig(filename string) (map[string]interface{}, error) {
	configs := make(map[string]interface{})

	if !CheckFileExist(filename) {
		return configs, errors.New("config file not exists:" + filename)
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return configs, errors.New("ReadFile:" + err.Error())
	}

	if err := json.Unmarshal(bytes, &configs); err != nil {
		return configs, errors.New("JsonDecode:" + err.Error())
	}

	return configs, nil
}