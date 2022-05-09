package jfile

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 判断传入的文件或目录是否存在
func PathExists(path string) (bool, error) {
	path, err := GetAbsPath(path)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 获取当前运行的可执行文件的路径
func GetWorkPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	index := strings.LastIndex(path, string(os.PathSeparator))
	return path[:index], nil
}

// 获取绝对路径
// 若传入的参数是绝对路径，则返回
// 若是相对路径，则将其拼接到当前的工作目录，并返回
func GetAbsPath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		workPath, err := GetWorkPath()
		if err != nil {
			return "", err
		}
		path = filepath.FromSlash(workPath + "/" + path)
	}
	return filepath.Abs(path)
}

// 枚举某个目录下所有的文件
func GetFilenamesByDir(root string) ([]string, error) {
	root, err := GetAbsPath(root)
	if err != nil {
		return nil, err
	}
	var files []string
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}
	var absPath string
	for _, file := range fileInfo {
		absPath, err = GetAbsPath(root + "/" + file.Name())
		if err != nil {
			return nil, err
		}
		files = append(files, filepath.FromSlash(absPath))
	}
	return files, nil
}

// 可以用于处理大文件，按行读取
// filename: 文件名
// pf: 处理每一行的函数
// isContinue: pf函数报错后是否继续处理下一行
func ProcessLine(filename string, pf func(string) error, isContinue bool) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()
	r := bufio.NewReader(f)
	for {
		line, err := readLine(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		// 使用传进来的函数处理line
		err = pf(line)
		if err != nil && !isContinue {
			return err
		}
	}
}

// 解决单行超过4096字节的文本读取问题
func readLine(r *bufio.Reader) (string, error) {
	line2 := []byte{}
	line, isprefix, err := r.ReadLine()
	line2 = append(line2, line...)
	for isprefix && err == nil {
		var bs []byte
		bs, isprefix, err = r.ReadLine()
		line2 = append(line2, bs...)
	}
	return string(line2), err
}

// 判断文件内包含某个字节数组的数量,没有重叠 如：kkkk中包含两个kk
func containsBytesCount(filepa string, cbytes []byte) int {
	f, err := os.Open(filepa)
	if err != nil {
		return 0
	}
	defer f.Close()
	// 每次读500字节
	buf := make([]byte, 50)
	cbytes2 := make([]byte, len(cbytes))
	var seek int64 = 0
	var count = 0
	for {
		rLens, err := f.Read(buf)
		if err != nil {
			break
		}
		//if bytes.Contains(buf,cbytes){
		//	return 1
		//}else{
		// 判断当前读取出来的是否含有第一个字节
		var k = 0
		for ; k < len(buf); k++ {
			//for k,v := range buf{
			if buf[k] == cbytes[0] {
				f.ReadAt(cbytes2, seek+int64(k))
				if bytes.Compare(cbytes, cbytes2) == 0 {
					//jlog.Debug(seek+int64(k))
					k = k + len(cbytes)
					count++
					//return true
				}
			}
		}
		//}
		if rLens < k {
			seek += int64(k)
		} else {
			seek += int64(rLens)
		}
		f.Seek(seek, io.SeekStart)
	}
	return count
}

// 文件复制从src到dst
func FileCopy(src string, dst string) error {
	srcf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcf.Close()
	dstf, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer dstf.Close()
	buf := make([]byte, 500)
	for {
		n, err := srcf.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := dstf.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}
