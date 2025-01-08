package fileopt

import (
	"os"
	"path/filepath"
	"strings"
)

func GetContents(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func FileGetContents(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func PutContents[DType string | []byte](filename string, data DType, isAppend ...bool) error {
	return FilePutContents(filename, data, isAppend...)
}

// FilePutContents file_put_contents
func FilePutContents[DType string | []byte](filename string, data DType, isAppend ...bool) error {
	if dir := filepath.Dir(filename); dir != "" && dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	byteDate := []byte(data)
	needAppend := len(isAppend) > 0 && isAppend[0] == true
	// write to file
	if !needAppend {
		return os.WriteFile(filename, byteDate, 0644)
	}
	// append to file
	fl, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	_, err = fl.Write(byteDate)
	if err1 := fl.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func IsExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err == nil {
		return true
	}
	if os.IsExist(err) {
		return true
	}
	return false
}

func IsExistOrCreate[T string | []byte](path string, init ...T) error {
	if IsExist(path) {
		return nil
	}
	var initData []byte
	if len(init) == 1 {
		initData = []byte(init[0])
	}
	return FilePutContents(path, initData)
}

func DirExistOrCreate(dirPath string) error {
	if IsExist(dirPath) {
		return nil
	} else {
		return os.MkdirAll(dirPath, 0755)
	}
}

func AbsPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") || path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}
		path = filepath.Join(homeDir, path[2:])
	}
	return path, nil
}

func Filename(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
