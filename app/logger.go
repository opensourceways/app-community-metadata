package app

import (
	"github.com/gookit/color"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// logger instance
var Logger *zap.Logger

// initLog init log setting
func initLogger() {
	newGenericLogger()
	//newRotatedLogger()

	Logger.Info("logger construction succeeded")
}

func newGenericLogger() {
	var err error
	var cfg zap.Config

	conf := Config.StringMap("log")
	logFile := conf["logFile"]
	errFile := conf["errFile"]

	// replace
	logFile = strings.NewReplacer(
		"{date}", LocTime().Format("20060102"),
	).Replace(logFile)

	errFile = strings.NewReplacer(
		"{date}", LocTime().Format("20060102"),
	).Replace(errFile)

	// create config
	if Debug {
		// cfg = zap.NewDevelopmentConfig()
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		cfg.Development = true
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig = encoderCfg
	} else {
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = []string{logFile}
		cfg.ErrorOutputPaths = []string{errFile}
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig = encoderCfg
	}

	// init some defined fields to log
	cfg.InitialFields = map[string]interface{}{
		"hostname": Hostname,
		// "context": map[string]interface{}{},
	}

	// create logger
	Logger, err = cfg.Build()

	if err != nil {
		panic(err)
	}
}

// see https://github.com/uber-go/zap/blob/master/FAQ.md#does-zap-support-log-rotation
func newRotatedLogger() {
	var cfg zap.Config

	conf := Config.StringMap("log")
	logFile := conf["logFile"]
	errFile := conf["errFile"]

	// replace
	logFile = strings.NewReplacer(
		"{date}", LocTime().Format("20060102"),
		"{hostname}", Hostname,
	).Replace(logFile)

	color.Info.Printf("============ Logger file=%s ============ \n", logFile)

	// create config
	if Debug {
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		cfg.Development = true
		cfg.OutputPaths = []string{"stdout", logFile}
		cfg.ErrorOutputPaths = []string{"stderr", errFile}
	} else {
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stdout", logFile}
		cfg.ErrorOutputPaths = []string{"stdout", errFile}
	}

	// lumberjack.Logger is already safe for concurrent use, so we don't need to lock it.
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename: logFile,
		MaxSize:  10, // megabytes
		// MaxBackups: 3,
		// MaxAge: 28, // days
	})

	core := zapcore.NewCore(
		// zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		w,
		cfg.Level,
	)

	// init some defined fields to log
	Logger = zap.New(core).With(zap.String("hostname", Hostname))
}
