package jfile

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func PathExists(path string) (bool, error) {
	path = GetAbsPath(path)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 获取当前运行的可执行文件的路径
func GetWorkPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	return path[:index]
}

// 获取绝对路径
func GetAbsPath(path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.FromSlash(GetWorkPath() + "/" + path)
	}
	return path
}

// 枚举某个目录下所有的文件
func GetFilenamesByDir(root string) ([]string, error) {
	root = GetAbsPath(root)
	var files []string
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, filepath.FromSlash(GetAbsPath(root+"/"+file.Name())))
	}
	return files, nil
}
