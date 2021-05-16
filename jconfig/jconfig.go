package jconfig

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

// requests配置
type requestsConfig struct {
	Proxy       string
	CAPath      string
	IsRedirect  bool
	IsVerifySSL bool
	Timeout     int
}

// async配置
// TaskMaxLimit 使用最大允许协程数量
type asyncConfig struct {
	TaskMaxLimit int
}

//lv 日志等级
//IsConsole bool // 是否在控制台输出
//MaxStoreDays int // 最大保存天数
//MaxSize int // 单个日志文件最大大小，单位B
//LogCount int // 保存日志的个数
//BufferSize int // 缓存的字节大小，单位B
//FlushInterval int // 日志写入文件的间隔，单位time.Second
type logConfig struct {
	LV            int
	IsConsole     bool
	MaxStoreDays  int
	MaxSize       int64
	LogCount      int
	BufferSize    int
	FlushInterval int
	LogFileName   string
}

// 配置
type config struct {
	RequestsConfig *requestsConfig
	AsyncConfig    *asyncConfig
	LogConfig      *logConfig
}

var Conf config

// 从json文件中读取配置
func init() {
	// 配置文件所在的路径
	viper.AddConfigPath("conf")
	// 配置文件的名称
	viper.SetConfigName("config")
	// 配置文件的类型
	viper.SetConfigType("json")
	// 读取配置文件到viper中
	if err := viper.ReadInConfig(); err != nil {
		log.Println("viper 读取配置文件失败", err)
		log.Println("使用内置的配置")
		setDefaultConfig()
		return
	}
	// 将读取的配置信息保存至全局变量Conf
	if err := viper.Unmarshal(&Conf); err != nil {
		log.Println("viper 反序列化配置文件失败", err)
		log.Println("使用内置的配置")
		setDefaultConfig()
		return
	}
	// 监控配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println("配置文件被修改")
		if err := viper.Unmarshal(&Conf); err != nil {
			log.Println("viper 反序列化配置文件失败")
			return
		}
	})
	//log.Println(Conf)
	//log.Println("LogConfig:",Conf.LogConfig)
	//log.Println("AsyncConfig:",Conf.AsyncConfig)
	//log.Println("RequestsConfig:",Conf.RequestsConfig)
}

// 设置默认的配置
func setDefaultConfig() {
	Conf.RequestsConfig = &requestsConfig{
		Proxy:       "",
		CAPath:      "conf/cas",
		IsRedirect:  false,
		IsVerifySSL: false,
		Timeout:     15,
	}
	Conf.AsyncConfig = &asyncConfig{
		TaskMaxLimit: 10000,
	}
	Conf.LogConfig = &logConfig{
		LV:            0,
		IsConsole:     true,
		MaxStoreDays:  5,
		MaxSize:       1024 * 1024 * 256,
		LogCount:      5,
		BufferSize:    1024 * 256,
		FlushInterval: 5,
		LogFileName:   "logs/app.log",
	}
}
