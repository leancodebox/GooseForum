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
	_, err := os.Stat(path)
	return err == nil
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

func AbsPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") || p == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return p, err
		}
		p = filepath.Join(homeDir, p[2:])
	}
	return p, nil
}

func Filename(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
