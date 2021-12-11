package jlog

//// logger
//type FishLogger struct {
//	console           bool          // 标准输出  默认 false
//	verbose           bool          // 是否输出行号和文件名 默认 false
//	MaxStoreDays      int           // 最大保留天数
//	MaxSizePerLogFile int64         // 单个日志最大容量 默认 256MB
//	size              int64         // 累计大小 无后缀
//	LogFullPath   string        // 文件目录 完整路径 LogFullPath=logFileName+logFileExt
//	logFileName       string        // 文件名
//	logFileExt        string        // 文件后缀名 默认 .log
//	logCreateDate        string        // 文件创建日期
//	level             logLevel      // 输出的日志等级
//	pool              sync.Pool     // Pool
//	mu                sync.Mutex    // logger🔒
//	writer            *bufio.Writer // 缓存io 缓存到文件
//	file              *os.File      // 日志文件
//}
//
//type buffer struct {
//	temp [64]byte
//	bytes.Buffer
//}
//
//// 日志等级
//type logLevel int
