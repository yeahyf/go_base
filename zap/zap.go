package zap

//格式化为debug,info,error三种日志信息
//并做了部分固定的设定处理

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var debugLog *zap.Logger
var infoLog *zap.Logger
var errorLog *zap.Logger
var atom zap.AtomicLevel
var logConfigFile string

const (
	Level_Debug = "debug"
	Level_Info  = "info"
	Level_Error = "error"
)

//LogConfig 日志配置结构体
type LogConfig struct {
	Filename   string `json:"logpath"`
	MaxSIze    int    `json:"maxsize"`
	MaxBackups int    `json:"backups"`
	MaxAge     int    `json:"maxage"`
	Compress   bool   `json:"compress"`
	Name       string `json:"name"`
}

//Config 完整配置
type Config struct {
	Logs  []LogConfig `json:"logs"`
	Level string      `json:"level"`
}

//ShortConfig 级别配置
type ShortConfig struct {
	Level string `json:"level"`
}

//SetLogConf 将json数据读取出来并进行初始化
func SetLogConf(configFile *string) {
	logConfigFile = *configFile
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		panic(err)
	}

	fmt.Println("Start set log config ... ")
	logConfig := make([]LogConfig, 3)

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

//SetLevel 动态修改系统的日志级别
func SetLevel(level string) {
	switch level {
	case Level_Debug:
		atom.SetLevel(zapcore.DebugLevel)
	case Level_Info:
		atom.SetLevel(zapcore.InfoLevel)
	case Level_Error:
		atom.SetLevel(zapcore.ErrorLevel)
	}
}

//initConfig 对日志模块进行初始化
func initConfig(config *Config) {
	SetLevel(config.Level)
	for _, logConfig := range config.Logs {
		switch logConfig.Name {
		case Level_Debug:
			debugLog = initLogger(&logConfig)
		case Level_Info:
			infoLog = initLogger(&logConfig)
		case Level_Error:
			errorLog = initLogger(&logConfig)
		}
	}
	go reSetLevel() //定时处理
}

//reSetLevel 读取json中的level配置，并重新设置 实现动态修改log的级别
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

//initLogger 初始化具体的日志对象
func initLogger(logConfig *LogConfig) (logger *zap.Logger) {
	hook := lumberjack.Logger{
		Filename:   logConfig.Filename,   // 日志文件路径
		MaxSize:    logConfig.MaxSIze,    // megabytes  M
		MaxBackups: logConfig.MaxBackups, // 最多保留30个备份
		MaxAge:     logConfig.MaxAge,     // days 天为单位
		Compress:   logConfig.Compress,   // 是否压缩 disabled by default
		LocalTime:  true,                 //使用本地时间，否则使用UTC时间
	}

	writeSyncer := zapcore.AddSync(&hook)
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

	if logConfig.Name == Level_Info {
		logger = zap.New(core)
	} else {
		caller := zap.AddCaller()
		callerSkip := zap.AddCallerSkip(0)
		development := zap.Development()
		logger = zap.New(core, caller, development, callerSkip)
	}
	return
}

//Debug 输出日志
func Debug(msg ...interface{}) {
	info := fmt.Sprint(msg...)
	debugLog.Debug(fmt.Sprintf("%s", info))
}

//Debugf 按照格式输出日志
func Debugf(format string, msg ...interface{}) {
	debugLog.Debug(fmt.Sprintf(format, msg...))
}

//Info 输出日志
func Info(msg ...interface{}) {
	info := fmt.Sprint(msg...)
	infoLog.Info(fmt.Sprintf("%s", info))
}

//Infof 按照格式输出日志
func Infof(format string, msg ...interface{}) {
	infoLog.Info(fmt.Sprintf(format, msg...))
}

//Error 输出日志
func Error(msg ...interface{}) {
	info := fmt.Sprint(msg...)
	errorLog.Error(fmt.Sprintf("%s", info))
}

//Errorf 按照格式输出日志
func Errorf(format string, msg ...interface{}) {
	errorLog.Error(fmt.Sprintf(format, msg...))
}
