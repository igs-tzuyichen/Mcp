package logger

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Logger struct {
	*log.Logger
	file *os.File
}

func New(filePath string) (*Logger, error) {

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger := &Logger{
		Logger: log.New(file, "", log.LstdFlags),
		file:   file,
	}

	return logger, nil
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) LogServerStart() {
	l.Println("MCP 服務已啟動，時間:", time.Now().Format("2006-01-02 15:04:05"))
}

func (l *Logger) LogServerStop() {
	l.Println("MCP 服務已停止，時間:", time.Now().Format("2006-01-02 15:04:05"))
}

func (l *Logger) LogServerError(err error) {
	l.Printf("伺服器錯誤: %v\n", err)
}

// 設定 req/res 的 logging Hooks
func (l *Logger) ConfigureLoggingHooks() *server.Hooks {
	hooks := &server.Hooks{}

	// 請求
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		reqJSON, _ := json.MarshalIndent(message, "", "  ")
		l.Printf("收到請求 [%s] ID:%v\n請求內容: %s\n", method, id, string(reqJSON))
	})

	// 成功的回應
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		resJSON, _ := json.MarshalIndent(result, "", "  ")
		l.Printf("回應請求 [%s] ID:%v\n回應內容: %s\n", method, id, string(resJSON))
	})

	// 錯誤的回應
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		l.Printf("請求錯誤 [%s] ID:%v\n錯誤訊息: %v\n", method, id, err)
	})

	return hooks
}
