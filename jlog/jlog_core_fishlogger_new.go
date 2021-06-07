package jlog

import (
	"bufio"
	"fmt"
	//"github.com/fatih/color"
	"github.com/chroblert/jgoutils/jthirdutil/color"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// 设置输出等级
func (fl *FishLogger) setLevel(lv logLevel) {
	if lv < DEBUG || lv > FATAL {
		panic("非法的日志等级")
	}
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.level = lv
}

// 设置最大保存天数
// 小于0不删除
func (fl *FishLogger) setMaxStoreDays(ma int) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.maxStoreDays = ma
}

// 写入文件
func (fl *FishLogger) flush() {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.flushSync()
}

// 设置是否显示调用者的详细信息，所在文件及行号
func (fl *FishLogger) setVerbose(b bool) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.verbose = b
}

// 设置控制台输出
func (fl *FishLogger) setConsole(b bool) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.console = b
}

// 生成日志头信息
func (fl *FishLogger) header(lv logLevel, depth int) *buffer {
	now := time.Now()
	buf := fl.pool.Get().(*buffer)
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	// format yyyymmdd hh:mm:ss.uuuu [DIWEF] file:line] msg
	buf.write4(0, year)
	buf.temp[4] = '/'
	buf.write2(5, int(month))
	buf.temp[7] = '/'
	buf.write2(8, day)
	buf.temp[10] = ' '
	buf.write2(11, hour)
	buf.temp[13] = ':'
	buf.write2(14, minute)
	buf.temp[16] = ':'
	buf.write2(17, second)
	buf.temp[19] = '.'
	buf.write4(20, now.Nanosecond()/1e5)
	buf.temp[24] = ' '
	copy(buf.temp[25:28], lv.Str())
	buf.temp[28] = ' '
	buf.Write(buf.temp[:29])
	// 调用信息
	if fl.verbose {
		_, file, line, ok := runtime.Caller(3 + depth)
		if !ok {
			file = "###"
			line = 1
		} else {
			slash := strings.LastIndex(file, "/")
			if slash >= 0 {
				file = file[slash+1:]
			}
		}
		buf.WriteString(file)
		buf.temp[0] = ':'
		n := buf.writeN(1, line)
		buf.temp[n+1] = ']'
		buf.temp[n+2] = ' '
		buf.Write(buf.temp[:n+3])
	}
	return buf
}

// 换行输出
func (fl *FishLogger) println(lv logLevel, args ...interface{}) {
	if lv < fl.level {
		return
	}
	var buf *buffer
	// 11用来表示Print()
	if lv == 11 {
		buf = &buffer{}
	} else {
		buf = fl.header(lv, 0)
	}
	fmt.Fprintln(buf, args...)
	// 将日志缓存写入到文件中
	fl.write(lv, buf)
}

// 格式输出
func (fl *FishLogger) printf(lv logLevel, format string, args ...interface{}) {
	if lv < fl.level {
		return
	}
	var buf *buffer
	if lv == 11 {
		buf = &buffer{}
		//buf.Write([]byte("\x1b[1K"))
	} else {
		//buf = &buffer{}
		//buf.Write([]byte("\x1b[1K"))
		buf = fl.header(lv, 0)
		//buf.Write(buf2.Bytes())
	}
	//buf := fl.header(lv, 0)
	fmt.Fprintf(buf, format, args...)
	// 210518: 不自动追加\n
	//if buf.Bytes()[buf.Len()-1] != '\n' {
	//	buf.WriteByte('\n')
	//}
	// 210603: 自动追加\x1b[K  清除从光标位置到行尾的所有字符
	//buf.WriteByte('\x1b[K')
	//buf.Write([]byte("\x1b[K"))
	fl.write(lv, buf)
}

// 写入数据
func (fl *FishLogger) write(lv logLevel, buf *buffer) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	data := buf.Bytes()
	if fl.console {
		//var begColor []byte
		//var endColor []byte
		//var tmpBytes []byte
		switch lv {
		case DEBUG:
			// 黑底蓝字
			//begColor = []byte("\033[1;34;40m")
			//endColor = []byte("\033[0m")
			color.Blue(string(data))
			//color.New(color.FgBlue).Fprintln(os.Stdout, "blue color!")
		case INFO:
			// 黑底白字
			//begColor = []byte("\033[1;37;40m")
			//endColor = []byte("\033[0m")
			color.White(string(data))
		case WARN:
			// 黑底黄字
			//begColor = []byte("\033[1;33;40m")
			//endColor = []byte("\033[0m")
			color.Yellow(string(data))
		case ERROR:
			// 黑底红字
			//begColor = []byte("\033[1;31;40m")
			//endColor = []byte("\033[0m")
			color.Red(string(data))
		case FATAL:
			// 黑底红字，反白显示
			//begColor = []byte("\033[7;31;40m")
			//endColor = []byte("\033[0m")
			color.HiRed(string(data))
		default:
			color.White(string(data))
		}
		//os.Stderr.Write(data)
		//tmpBytes = append(begColor,data...)
		//tmpBytes = append(tmpBytes,endColor...)
		//os.Stdout.Write(tmpBytes)
	}
	// 第一次写入文件
	if fl.file == nil {
		if err := fl.rotate(); err != nil {
			os.Stderr.Write(data)
			fl.exit(err)
		}
	}
	// 自定义的一个loglevel，开头没有日期
	if lv != 11{
		// 按天切割
		if fl.createDate != string(data[0:10]) {
			go fl.delete() // 每天检测一次旧文件
			log.Println("lv:",lv,"rotate测试：",fl.createDate,"string(data[0:10]):",string(data[0:10]),"_")
			if err := fl.rotate(); err != nil {
				fl.exit(err)
			}
		}
	}

	// 按大小切割
	//log.Println("文件最大大小", fl.maxSizePerLogFile)
	if fl.size+int64(len(data)) >= fl.maxSizePerLogFile {
		if err := fl.rotate(); err != nil {
			fl.exit(err)
		}
	}
	n, err := fl.writer.Write(data)
	fl.size += int64(n)
	if err != nil {
		fl.exit(err)
	}
	buf.Reset()
	fl.pool.Put(buf)
}

// 删除旧日志
func (fl *FishLogger) delete() {
	if fl.maxStoreDays < 0 {
		return
	}
	dir := filepath.Dir(fl.fullLogFilePath)
	fakeNow := time.Now().AddDate(0, 0, -fl.maxStoreDays)
	filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "logs: unable to delete old file '%s', error: %v\n", fpath, r)
			}
		}()
		if info == nil {
			return nil
		}
		// 防止误删
		if !info.IsDir() && info.ModTime().Before(fakeNow) && strings.HasSuffix(info.Name(), fl.logFileExt) {
			os.Remove(fpath)
		}
		return nil
	})
}

// 定时写入文件
func (fl *FishLogger) daemon() {
	for range time.NewTicker(flushInterval).C {
		fl.flush()
	}
}

// 不能锁
func (fl *FishLogger) flushSync() {
	if fl.file != nil {
		fl.writer.Flush() // 写入底层数据.写入到内存中
		fl.file.Sync()    // 同步到磁盘.Sync递交文件的当前内容进行稳定的存储。
		// 一般来说，这表示将文件系统的最近写入的数据在内存中的拷贝刷新到硬盘中稳定保存。
	}
}

func (fl *FishLogger) exit(err error) {
	fmt.Fprintf(os.Stderr, "logs: exiting because of error: %s\n", err)
	fl.flushSync()
	os.Exit(0)
}

// rotate
// 切割文件
// 如果是第一次写入日志，
//       -> 判断是否存在app.log文件；若存在，则重命名
//		 -> 创建日志文件app.log
//		 -> 判断当前日志文件数量是否小于规定个数；若大于则删除
// 如果不是第一次写入日志，
//       -> 判断当前日志文件的大小是否小于规定大小；若大于，则切割，
func (fl *FishLogger) rotate() error {
	now := time.Now()
	// 分割文件
	// 若日志文件已打开，则将缓存写入内存，再刷入磁盘
	if fl.file != nil {
		// 写入内存
		fl.writer.Flush()
		// 写入磁盘
		fl.file.Sync()
		// 关闭文件
		err := fl.file.Close()
		if err != nil {
			log.Println("fl.file", err)
		}
		// 对日志文件进行重命名
		fileBackupName := filepath.Join(fl.logFileName + now.Format(".2006-01-02_150405") + fl.logFileExt)
		err = os.Rename(fl.fullLogFilePath, fileBackupName)
		if err != nil {
			log.Println("rename", err)
		}
		// 创建新日志文件app.log
		newLogFile, err := os.OpenFile(fl.fullLogFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		fl.file = newLogFile
		fl.size = 0
		// 日志缓存
		fl.writer = bufio.NewWriterSize(fl.file, bufferSize)
	}else if fl.file == nil{
		// 对于第一次写入文件
		// 判断是否存在app.log日志文件，若存在则重命名
		_, err := os.Stat(fl.fullLogFilePath)
		if err == nil {
			// 获取当前日志文件的创建日期
			// 对日志文件进行重命名
			fileBackupName := filepath.Join(fl.logFileName + now.Format(".2006-01-02_150405") + fl.logFileExt)
			err = os.Rename(fl.fullLogFilePath, fileBackupName)
			if err != nil {
				log.Println("rename", err)
			}
		}
		// 创建新日志文件app.log
		newLogFile, err := os.OpenFile(fl.fullLogFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		fl.file = newLogFile
		fl.size = 0
		// 日志缓存
		fl.writer = bufio.NewWriterSize(fl.file, bufferSize)
	}
	fileInfo, err := os.Stat(fl.fullLogFilePath)
	fl.createDate = now.Format("2006/01/02")
	if err == nil {
		// 获取当前日志文件的大小
		fl.size = fileInfo.Size()
		// 获取当前日志文件的创建日期
		fl.createDate = fileInfo.ModTime().Format("2006/01/02")
	}
	//fl.writer = bufio.NewWriterSize(fl.file, bufferSize)
	// 日志文件的个数不能超过logCount个，若超过，则刪除最先创建的日志文件
	pattern := fl.logFileName + ".*" + fl.logFileExt
	for files, _ := filepath.Glob(pattern); len(files) > logCount; files, _ = filepath.Glob(pattern) {
		// 删除log文件
		os.Remove(files[0])
		if fl.level == DEBUG {
			tmpBuffer := fl.header(DEBUG, 0)
			fmt.Fprintf(tmpBuffer, "删除旧日志文件")
			fmt.Fprintf(tmpBuffer, files[0])
			//fmt.Fprintf(tmpBuffer,"\033[0m")
			fmt.Fprintf(tmpBuffer, "\n")
			// 黑底蓝色
			//fmt.Fprintf(os.Stdout,"\033[1;34;40m"+string(tmpBuffer.Bytes())+"\033[0m")
			color.Blue(string(tmpBuffer.Bytes()))
			fl.writer.Write(tmpBuffer.Bytes())
		}
	}
	return nil
}

// -------- 实例 自定义

func (fl *FishLogger) debug(args ...interface{}) {
	fl.println(DEBUG, args...)
}

func (fl *FishLogger) debugf(format string, args ...interface{}) {
	fl.printf(DEBUG, format, args...)
}
func (fl *FishLogger) info(args ...interface{}) {
	fl.println(INFO, args...)
}

func (fl *FishLogger) infof(format string, args ...interface{}) {
	fl.printf(INFO, format, args...)
}

func (fl *FishLogger) warn(args ...interface{}) {
	fl.println(WARN, args...)
}

func (fl *FishLogger) warnf(format string, args ...interface{}) {
	fl.printf(WARN, format, args...)
}

func (fl *FishLogger) error(args ...interface{}) {
	fl.println(ERROR, args...)
}

func (fl *FishLogger) errorf(format string, args ...interface{}) {
	fl.printf(ERROR, format, args...)
}

func (fl *FishLogger) fatal(args ...interface{}) {
	fl.println(FATAL, args...)
	os.Exit(0)
}

func (fl *FishLogger) fatalf(format string, args ...interface{}) {
	fl.printf(FATAL, format, args...)
	os.Exit(0)
}
