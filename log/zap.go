package log

//格式化为debug,info,warn,error四种日志信息
//并做了部分固定的设定处理

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var debugLog *zap.Logger
var infoLog *zap.Logger
var errorLog *zap.Logger
var warnLog *zap.Logger
var atom zap.AtomicLevel
var logConfigFile string

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelError = "error"
	LevelWarn  = "warn"
)

// LevelConfig 日志配置结构体
type LevelConfig struct {
	Filename   string `json:"logpath"`
	MaxSize    int    `json:"maxsize"`
	MaxBackups int    `json:"backups"`
	MaxAge     int    `json:"maxage"`
	Compress   bool   `json:"compress"`
	Name       string `json:"name"`
	//0默认为lumberjack，1 小时，2 分钟,3 天
	Type int `json:"type"`
	//滚动时间，type=1 单位:小时，type=2 单位:分钟, type=3 单位:天
	Rotation int `json:"rotation"`
}

// Config 完整配置
type Config struct {
	Logs  []LevelConfig `json:"logs"`
	Level string        `json:"level"`
}

// ShortConfig 级别配置
type ShortConfig struct {
	Level string `json:"level"`
}

// SetLogConf 将json数据读取出来并进行初始化
func SetLogConf(configFile *string) {
	logConfigFile = *configFile
	data, err := os.ReadFile(logConfigFile)
	if err != nil {
		panic(err)
	}

	//fmt.Println("Start set log config ... ")
	logConfig := make([]LevelConfig, 3)

	config := Config{
		Logs: logConfig,
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	} else {
		atom = zap.NewAtomicLevel()
		initConfig(&config)
	}
}

// SetLevel 动态修改系统的日志级别
func SetLevel(level string) {
	switch level {
	case LevelDebug:
		atom.SetLevel(zapcore.DebugLevel)
	case LevelInfo:
		atom.SetLevel(zapcore.InfoLevel)
	case LevelError:
		atom.SetLevel(zapcore.ErrorLevel)
	case LevelWarn:
		atom.SetLevel(zapcore.WarnLevel)
	}
}

func IsDebug() bool {
	return atom.Enabled(zapcore.DebugLevel)
}

// initConfig 对日志模块进行初始化
func initConfig(config *Config) {
	SetLevel(config.Level)
	for _, logConfig := range config.Logs {
		switch logConfig.Name {
		case LevelDebug:
			debugLog = initLogger(&logConfig)
		case LevelInfo:
			infoLog = initLogger(&logConfig)
		case LevelError:
			errorLog = initLogger(&logConfig)
		case LevelWarn:
			warnLog = initLogger(&logConfig)
		}
	}
	go reSetLevel() //定时处理
}

// reSetLevel 读取json中的level配置，并重新设置 实现动态修改log的级别
func reSetLevel() {
	for {
		time.Sleep(time.Second * 30)
		//Debug("start reload log configf file ...")
		data, err := ioutil.ReadFile(logConfigFile)
		if err != nil {
			Error("Read log config file Error", err)
			continue
		}
		shortConfig := ShortConfig{}
		err = json.Unmarshal(data, &shortConfig)
		if err != nil {
			Error("Unmarshal log config file Error", err)
			continue
		} else {
			SetLevel(shortConfig.Level)
		}
	}
}

// initLogger 初始化具体的日志对象
func initLogger(logConfig *LevelConfig) (logger *zap.Logger) {
	var hook io.Writer
	logPath := path.Dir(logConfig.Filename)
	//强制建立目录，如果目录存在，建立失败，不影响
	_ = os.MkdirAll(logPath, os.ModePerm)
	if !strings.HasSuffix(logConfig.Filename, ".log") {
		panic("Filename must end as .log!")
	}
	if logConfig.Type == 0 {
		lumberLog := lumberjack.Logger{
			Filename:   logConfig.Filename,   // 日志文件路径
			MaxSize:    logConfig.MaxSize,    // megabytes  M
			MaxBackups: logConfig.MaxBackups, // 最多保留30个备份
			MaxAge:     logConfig.MaxAge,     // days 天为单位
			Compress:   logConfig.Compress,   // 是否压缩 disabled by default
			LocalTime:  true,                 //使用本地时间，否则使用UTC时间
		}
		hook = &lumberLog
	} else {
		var err error
		hook, err = initRotateLog(logConfig)
		if err != nil {
			fmt.Println("Init Log Error", err)
			panic(err)
		}
	}
	writeSyncer := zapcore.AddSync(hook)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey: "time",
		// LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "linenum",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",

		LineEnding: zapcore.DefaultLineEnding,
		// EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.StringDurationEncoder, //
		//	EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:   zapcore.FullNameEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		writeSyncer,
		atom,
	)

	if logConfig.Name == LevelInfo {
		logger = zap.New(core)
	} else {
		caller := zap.AddCaller()
		callerSkip := zap.AddCallerSkip(1) //让日志中能够打印出业务代码的行数
		development := zap.Development()
		logger = zap.New(core, caller, development, callerSkip)
	}
	return
}

func initRotateLog(logConfig *LevelConfig) (io.Writer, error) {
	var patternTemp string
	var rotationTime time.Duration
	switch logConfig.Type {
	case 1:
		{
			patternTemp = "_%Y%m%d%H.log"
			rotationTime = time.Hour * time.Duration(logConfig.Rotation)
		}
	case 2:
		{
			patternTemp = "_%Y%m%d%H%M.log"
			rotationTime = time.Minute * time.Duration(logConfig.Rotation)
		}
	case 3:
		{
			patternTemp = "_%Y%m%d.log"
			rotationTime = time.Hour * 24 * time.Duration(logConfig.Rotation)
		}
	}
	pattern := strings.ReplaceAll(logConfig.Filename, ".log", patternTemp)
	hook, err := rotatelogs.New(
		pattern, rotatelogs.WithLinkName(logConfig.Filename),
		rotatelogs.WithMaxAge(time.Hour*24*time.Duration(logConfig.MaxAge)), // 保存天数
		rotatelogs.WithRotationTime(rotationTime),                           //切割频率 小时
	)
	return hook, err
}

// Debug 输出日志
func Debug(msg ...interface{}) {
	info := fmt.Sprint(msg...)
	debugLog.Debug(fmt.Sprintf("%s", info))
}

// Debugf 按照格式输出日志
func Debugf(format string, msg ...interface{}) {
	debugLog.Debug(fmt.Sprintf(format, msg...))
}

// Info 输出日志
func Info(msg ...interface{}) {
	info := fmt.Sprint(msg...)
	infoLog.Info(fmt.Sprintf("%s", info))
}

// Infof 按照格式输出日志
func Infof(format string, msg ...interface{}) {
	infoLog.Info(fmt.Sprintf(format, msg...))
}

// Error 输出日志
func Error(msg ...interface{}) {
	info := fmt.Sprint(msg...)
	errorLog.Error(fmt.Sprintf("%s", info))
}

// Errorf 按照格式输出日志
func Errorf(format string, msg ...interface{}) {
	errorLog.Error(fmt.Sprintf(format, msg...))
}

func Warn(msg ...interface{}) {
	info := fmt.Sprint(msg...)
	warnLog.Warn(fmt.Sprintf("%s", info))
}

func Warnf(format string, msg ...interface{}) {
	warnLog.Warn(fmt.Sprintf(format, msg...))
}
