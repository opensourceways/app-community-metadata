/*
Copyright 2021 The Opensourceways Group.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
		encoderCfg.TimeKey = "ts"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig = encoderCfg
	} else {
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stdout", logFile}
		cfg.ErrorOutputPaths = []string{"stderr", errFile}
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "ts"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig = encoderCfg
	}

	// init some defined fields to log
	cfg.InitialFields = map[string]interface{}{
		//TODO: Add useful field here.
	}

	// create logger
	Logger, err = cfg.Build()

	if err != nil {
		panic(err)
	}
}
