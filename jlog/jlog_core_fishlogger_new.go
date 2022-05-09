package jlog

import (
	"bufio"
	"bytes"
	"fmt"
	"sync"
	//"github.com/fatih/color"
	"github.com/chroblert/jgoutils/jthirdutil/color"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// 定义FishLogger结构体
type FishLogger struct {
	console           bool // 标准输出  默认 false
	verbose           bool // 是否输出行号和文件名 默认 false
	iniCreateNewLog   bool
	maxStoreDays      int           // 最大保留天数
	maxSizePerLogFile int64         // 单个日志最大容量 默认 256MB
	size              int64         // 累计大小 无后缀
	logFullPath       string        // 文件目录 完整路径 logFullPath=logFileName+logFileExt
	logFileName       string        // 文件名
	logFileExt        string        // 文件后缀名 默认 .log
	logCreateDate     string        // 文件创建日期
	logCount          int           // 最大保持日志文件的数量
	flushInterval     time.Duration // 日志写入文件的频率
	bufferSize        int           // 日志缓存大小
	level             logLevel      // 输出的日志等级
	pool              sync.Pool     // Pool
	mu                sync.Mutex    // logger🔒
	writer            *bufio.Writer // 缓存io 缓存到文件
	file              *os.File      // 日志文件
	storeToFile       bool          // 是否将输出内容保存到文件
}

type buffer struct {
	temp [64]byte
	bytes.Buffer
}

// 日志等级
type logLevel int

// 设置输出等级
func (fl *FishLogger) SetLogLevel(lv logLevel) {
	if lv < DEBUG || lv > FATAL {
		panic("非法的日志等级")
	}
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.level = lv
}

// 设置日志文件路径
func (fl *FishLogger) SetLogFullPath(logFullPath string) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.logFullPath = logFullPath
	//日志文件路径设置
	fl.logFileExt = filepath.Ext(fl.logFullPath)                       // .log
	fl.logFileName = strings.TrimSuffix(fl.logFullPath, fl.logFileExt) // logs/app
	if fl.logFileExt == "" {
		fl.logFileExt = ".log"
	}
	os.MkdirAll(filepath.Dir(fl.logFullPath), 0666)
}

// 设置日志文件大小 SetMaxSizePerLogFile
func (fl *FishLogger) SetMaxSizePerLogFile(logfilesize int64) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	//fl.maxStoreDays = ma
	fl.maxSizePerLogFile = logfilesize
}

// iniCreateNewLog
func (fl *FishLogger) IsIniCreateNewLog(iniCreateNewLog bool) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.iniCreateNewLog = iniCreateNewLog
}

// 设置最大保存天数
// 小于0不删除
func (fl *FishLogger) SetMaxStoreDays(ma int) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.maxStoreDays = ma
}

// 设置日志文件最大保存数量
// 小于0不删除
func (fl *FishLogger) SetLogCount(logCount int) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.logCount = logCount
}

// 写入文件
func (fl *FishLogger) Flush() {
	//fl.mu.Lock()
	//defer fl.mu.Unlock()
	// 锁在flushSync函数中加
	fl.flushSync()
}

// 设置是否显示调用者的详细信息，所在文件及行号
func (fl *FishLogger) setVerbose(b bool) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.verbose = b
}

// 设置控制台输出
func (fl *FishLogger) SetUseConsole(b bool) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.console = b
}

// 设置是否保存到文件
func (fl *FishLogger) SetStoreToFile(b bool) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.storeToFile = b
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
	fl.write(lv, buf, true)
}

// 换行输出
// 不带具体日期时间、文件名、行号等信息
func (fl *FishLogger) nprintln(lv logLevel, args ...interface{}) {
	if lv < fl.level {
		return
	}
	var buf *buffer
	buf = &buffer{}
	fmt.Fprintln(buf, args...)
	// 将日志缓存写入到文件中
	fl.write(lv, buf, false)
}

// 格式输出
// 不带具体日期时间、文件名、行号等信息
func (fl *FishLogger) nprintf(lv logLevel, format string, args ...interface{}) {
	if lv < fl.level {
		return
	}
	var buf *buffer
	buf = &buffer{}
	fmt.Fprintf(buf, format, args...)
	fl.write(lv, buf, false)
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
	//buf := fl.header(Lv, 0)
	fmt.Fprintf(buf, format, args...)
	// 210518: 不自动追加\n
	//if buf.Bytes()[buf.Len()-1] != '\n' {
	//	buf.WriteByte('\n')
	//}
	// 210603: 自动追加\x1b[K  清除从光标位置到行尾的所有字符
	//buf.WriteByte('\x1b[K')
	//buf.Write([]byte("\x1b[K"))
	fl.write(lv, buf, true)
}

// 写入数据
// isverbose: buf中是否有带有具体日期时间及文件名行号这些信息
func (fl *FishLogger) write(lv logLevel, buf *buffer, isverbose bool) {
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
	if !fl.storeToFile {
		return
	}
	// 第一次写入文件
	if fl.file == nil {
		if err := fl.rotate(); err != nil {
			os.Stderr.Write(data)
			fl.exit(err)
		}
	}

	// 按天切割
	if fl.logCreateDate != time.Now().Format("2006/01/02") {
		go fl.delete() // 每天检测一次旧文件
		//log.Println("Lv:",Lv,"rotate测试：",fl.logCreateDate,"string(data[0:10]):",string(data[0:10]),"_")
		if err := fl.rotate(); err != nil {
			fl.exit(err)
		}
	}

	// 按大小切割
	//log.Println("文件最大大小", fl.MaxSizePerLogFile)
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
	dir := filepath.Dir(fl.logFullPath)
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
		if !info.IsDir() && info.ModTime().Before(fakeNow) && strings.HasSuffix(info.Name(), fl.logFileExt) && strings.HasPrefix(info.Name(), fl.logFileName+".") {
			return os.Remove(fpath)

		}
		return nil
	})
}

// 定时写入文件，监测到Ctrl+C时写入文件
func (fl *FishLogger) daemon(stopChannel chan os.Signal) {
	tickTimer := time.NewTicker(fl.flushInterval)
	for {
		select {
		case <-tickTimer.C:
			fl.Flush()
		case <-stopChannel:
			//fmt.Println("监测到信号")
			fl.Flush()
			// 220111 bugfix
			os.Exit(-1)
			//fmt.Println("结束")
		}
	}

	//for range time.NewTicker(FlushInterval).C {
	//	fl.Flush()
	//}
}

// 写入到文件
func (fl *FishLogger) flushSync() {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	//fmt.Println("写入文件")
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
			//log.Println("fl.file", err)
			return err
		}
		// 对日志文件进行重命名
		fileBackupName := filepath.Join(fl.logFileName + now.Format(".2006-01-02_150405") + fl.logFileExt)
		err = os.Rename(fl.logFullPath, fileBackupName)
		if err != nil {
			//log.Println("rename", err)
			return err
		}
		// 创建新日志文件app.log
		newLogFile, err := os.OpenFile(fl.logFullPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		fl.file = newLogFile
		fl.size = 0
		// 日志缓存
		fl.writer = bufio.NewWriterSize(fl.file, fl.bufferSize)
	} else if fl.file == nil {
		// TODO 判断每次运行是否重命名原有日志文件
		if fl.iniCreateNewLog {
			// 对于第一次写入文件
			// 判断是否存在app.log日志文件，若存在则重命名
			_, err := os.Stat(fl.logFullPath)
			if err == nil {
				// 获取当前日志文件的创建日期
				// 对日志文件进行重命名
				fileBackupName := filepath.Join(fl.logFileName + now.Format(".2006-01-02_150405") + fl.logFileExt)
				err = os.Rename(fl.logFullPath, fileBackupName)
				if err != nil {
					//log.Println("rename", err)
					return err
				}
			}
		}
		// 创建或打开日志文件app.log
		newLogFile, err := os.OpenFile(fl.logFullPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		fl.file = newLogFile
		fl.size = 0
		// 日志缓存
		fl.writer = bufio.NewWriterSize(fl.file, fl.bufferSize)
	}
	fileInfo, err := os.Stat(fl.logFullPath)
	fl.logCreateDate = now.Format("2006/01/02")
	if err == nil {
		// 获取当前日志文件的大小
		fl.size = fileInfo.Size()
		// 获取当前日志文件的创建日期
		fl.logCreateDate = fileInfo.ModTime().Format("2006/01/02")
	}
	//fl.writer = bufio.NewWriterSize(fl.file, BufferSize)
	// 日志文件的个数不能超过logCount个，若超过，则刪除最先创建的日志文件
	if fl.logCount > 0 {
		pattern := fl.logFileName + ".*" + fl.logFileExt
		for files, _ := filepath.Glob(pattern); len(files) > fl.logCount; files, _ = filepath.Glob(pattern) {
			// 删除log文件
			os.Remove(files[0])
			if fl.level == -1 {
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
	}
	return nil
}

// -------- 实例 自定义

//func (fl *FishLogger) debug(args ...interface{}) {
//	fl.println(DEBUG, args...)
//}
//
//func (fl *FishLogger) debugf(format string, args ...interface{}) {
//	fl.printf(DEBUG, format, args...)
//}
//func (fl *FishLogger) info(args ...interface{}) {
//	fl.println(INFO, args...)
//}
//
//func (fl *FishLogger) infof(format string, args ...interface{}) {
//	fl.printf(INFO, format, args...)
//}
//
//func (fl *FishLogger) warn(args ...interface{}) {
//	fl.println(WARN, args...)
//}
//
//func (fl *FishLogger) warnf(format string, args ...interface{}) {
//	fl.printf(WARN, format, args...)
//}
//
//func (fl *FishLogger) error(args ...interface{}) {
//	fl.println(ERROR, args...)
//}
//
//func (fl *FishLogger) errorf(format string, args ...interface{}) {
//	fl.printf(ERROR, format, args...)
//}
//
//func (fl *FishLogger) fatal(args ...interface{}) {
//	fl.println(FATAL, args...)
//	os.Exit(0)
//}
//
//func (fl *FishLogger) fatalf(format string, args ...interface{}) {
//	fl.printf(FATAL, format, args...)
//	os.Exit(0)
//}

// 用户自行实例化
func (fl *FishLogger) Debug(args ...interface{}) {
	fl.println(DEBUG, args...)
}

func (fl *FishLogger) Debugf(format string, args ...interface{}) {
	fl.printf(DEBUG, format, args...)
}
func (fl *FishLogger) Info(args ...interface{}) {
	fl.println(INFO, args...)
}

func (fl *FishLogger) Infof(format string, args ...interface{}) {
	fl.printf(INFO, format, args...)
}

func (fl *FishLogger) Warn(args ...interface{}) {
	fl.println(WARN, args...)
}

func (fl *FishLogger) Warnf(format string, args ...interface{}) {
	fl.printf(WARN, format, args...)
}

func (fl *FishLogger) Error(args ...interface{}) {
	fl.println(ERROR, args...)
}

func (fl *FishLogger) Errorf(format string, args ...interface{}) {
	fl.printf(ERROR, format, args...)
}

func (fl *FishLogger) Fatal(args ...interface{}) {
	fl.println(FATAL, args...)
	os.Exit(0)
}

func (fl *FishLogger) Fatalf(format string, args ...interface{}) {
	fl.printf(FATAL, format, args...)
	os.Exit(0)
}

// Nxxxx

func (fl *FishLogger) NDebug(args ...interface{}) {
	fl.nprintln(DEBUG, args...)
}

func (fl *FishLogger) NDebugf(format string, args ...interface{}) {
	fl.nprintf(DEBUG, format, args...)
}
func (fl *FishLogger) NInfo(args ...interface{}) {
	fl.nprintln(INFO, args...)
}

func (fl *FishLogger) NInfof(format string, args ...interface{}) {
	fl.nprintf(INFO, format, args...)
}

func (fl *FishLogger) NWarn(args ...interface{}) {
	fl.nprintln(WARN, args...)
}

func (fl *FishLogger) NWarnf(format string, args ...interface{}) {
	fl.nprintf(WARN, format, args...)
}

func (fl *FishLogger) NError(args ...interface{}) {
	fl.nprintln(ERROR, args...)
}

func (fl *FishLogger) NErrorf(format string, args ...interface{}) {
	fl.nprintf(ERROR, format, args...)
}

func (fl *FishLogger) NFatal(args ...interface{}) {
	fl.nprintln(FATAL, args...)
	os.Exit(0)
}

func (fl *FishLogger) NFatalf(format string, args ...interface{}) {
	fl.nprintf(FATAL, format, args...)
	os.Exit(0)
}
