package jfile

import (
	"github.com/chroblert/jgoutils/jlog"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func PathExists(path string) (bool, error) {
	path = getAbsPath(path)
	jlog.Debug(path)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 获取编译后的可执行路径
func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}

// 获取绝对路径
func getAbsPath(path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.FromSlash(GetAppPath() + "/" + path)
	}
	return path
}

// 枚举某个目录下所有的文件
func GetFilenamesByDir(root string) ([]string, error) {
	root = getAbsPath(root)
	var files []string
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, filepath.FromSlash(getAbsPath(root+"/"+file.Name())))
	}
	return files, nil
}
