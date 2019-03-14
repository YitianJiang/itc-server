package kite

import (
	"context"
	"fmt"
	"os"

	"code.byted.org/gopkg/logs"
)

const (
	DATABUS_RPC_PREFIX = "webarch.rpc."
	DATABUS_APP_PREFIX = "webarch.app."
)

// TraceLogger .
type TraceLogger interface {
	CtxTraceKVs(ctx context.Context, kvs ...interface{})
	CtxErrorKVs(ctx context.Context, kvs ...interface{})
}

var (
	// AccessLogger .
	AccessLogger TraceLogger
	// CallLogger .
	CallLogger    *logs.Logger
	defaultLogger *logs.Logger

	databusRPCChannel string
	databusAPPChannel string
)

// newRPCLogger : a logger made for RPC.
// this logger is no longer a single file logger, we have appended databus & agent provider into it, so I changed the function name.
func newRPCLogger(file string) *logs.Logger {
	logger := logs.NewLogger(1024)
	logger.SetLevel(logs.LevelTrace)
	logger.SetCallDepth(2)

	fileProvider := logs.NewFileProvider(file, logs.DayDur, 0)
	fileProvider.SetLevel(logs.LevelTrace)
	if err := logger.AddProvider(fileProvider); err != nil {
		fmt.Fprintf(os.Stderr, "Add file provider error: %s\n", err)
		return nil
	}

	if DatabusLog && databusRPCChannel != "" {
		databusProvider := logs.NewDatabusProviderWithChannel(DATABUS_RPC_PREFIX+ServiceName, databusRPCChannel) // 此处为RPC log类型
		databusProvider.SetLevel(logs.LevelTrace)
		if err := logger.AddProvider(databusProvider); err != nil {
			fmt.Fprintf(os.Stderr, "Add databus provider error: %s\n", err)
			return nil
		}
	}

	if AgentLog {
		// Create an agent provider only for RPC log.
		agentProvider := logs.NewRPCLogAgentProvider()
		agentProvider.SetLevel(logs.LevelTrace)
		if err := logger.AddProvider(agentProvider); err != nil {
			fmt.Fprintf(os.Stderr, "Add agent provider error: %s\n", err)
			return nil
		}
	}

	logger.StartLogger()
	return logger
}

// initDefaultLog set default logger in logs;
// user just use logs.Error to do application log;
func initDefaultLog(level int) {
	if defaultLogger == nil {
		defaultLogger = logs.NewLogger(1024)
	}
	defaultLogger.SetLevel(level)
	defaultLogger.SetCallDepth(3)
	logs.InitLogger(defaultLogger)
}

func initConsoleProvider() {
	if defaultLogger == nil {
		defaultLogger = logs.NewLogger(1024)
	}
	consoleProvider := logs.NewConsoleProvider()
	consoleProvider.SetLevel(logs.LevelTrace)
	if err := defaultLogger.AddProvider(consoleProvider); err != nil {
		fmt.Fprintf(os.Stderr, "AddProvider consoleProvider error: %s\n", err)
	}
}

func initFileProvider(level int, filename string, dur logs.SegDuration, size int64) {
	if defaultLogger == nil {
		defaultLogger = logs.NewLogger(1024)
	}
	fileProvider := logs.NewFileProvider(filename, dur, size)
	fileProvider.SetLevel(level)
	if err := defaultLogger.AddProvider(fileProvider); err != nil {
		fmt.Fprintf(os.Stderr, "AddProvider fileProvider error: %s\n", err)
	}
}

func initDatabusProvider(level int) {
	if defaultLogger == nil {
		defaultLogger = logs.NewLogger(1024)
	}
	if DatabusLog && databusAPPChannel != "" {
		databusProvider := logs.NewDatabusProviderWithChannel(DATABUS_APP_PREFIX+ServiceName, databusAPPChannel) // 此处为APP log类型
		databusProvider.SetLevel(level)
		if err := defaultLogger.AddProvider(databusProvider); err != nil {
			fmt.Fprintf(os.Stderr, "Add databus provider error: %s\n", err)
		}
	}
}

func initAgentProvider(level int) {
	if defaultLogger == nil {
		defaultLogger = logs.NewLogger(1024)
	}
	if AgentLog {
		agentProvider := logs.NewAgentProvider()
		agentProvider.SetLevel(level)
		if err := defaultLogger.AddProvider(agentProvider); err != nil {
			fmt.Fprintf(os.Stderr, "Add agent provider error: %s\n", err)
		}
	}
}
