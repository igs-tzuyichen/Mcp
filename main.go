package main

import (
	"fmt"
	"log"

	"mcp-gke-monitor/config"
	"mcp-gke-monitor/gke"
	"mcp-gke-monitor/logger"
	"mcp-gke-monitor/optimization"
	"mcp-gke-monitor/server"
)

func main() {
	//-----------------------------------------------------------------
	// 組態
	//-----------------------------------------------------------------
	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("載入配置失敗: %v", err)
	}

	// 檢查是否為 stdio 模式，如果是則不輸出到 stdout
	isStdioMode := appConfig.ServerType == config.ServerTypeStdio

	if !isStdioMode {
		fmt.Println("正在啟動 MCP GKE 監控查詢服務...")
		fmt.Printf("伺服器類型: %s\n", appConfig.ServerType)
	}

	//-----------------------------------------------------------------
	// 日誌
	//-----------------------------------------------------------------
	appLogger, err := logger.New("mcp_log.txt")
	if err != nil {
		log.Fatalf("初始化日誌系統失敗: %v", err)
	}
	defer appLogger.Close()

	// 記錄到日誌文件
	appLogger.Println("正在啟動 MCP GKE 監控查詢服務...")
	appLogger.Printf("伺服器類型: %s", appConfig.ServerType)

	// 檢查是否成功讀取 GKE 凭证
	if appConfig.Credentials == nil {
		if !isStdioMode {
			fmt.Println("警告: 未載入 GKE 凭证，將使用預設 kubeconfig")
		}
		appLogger.Println("警告: 未載入 GKE 凭证，將使用預設 kubeconfig")
	} else {
		if !isStdioMode {
			fmt.Printf("已載入 GKE 凭证，項目ID: %s\n", appConfig.Credentials.ProjectID)
		}
		appLogger.Printf("已載入 GKE 凭证，項目ID: %s", appConfig.Credentials.ProjectID)
	}

	//-----------------------------------------------------------------
	// GKE 服務
	//-----------------------------------------------------------------
	var gkeService *gke.Service

	if appConfig.Credentials != nil {
		// 使用 Google Cloud 凭证創建 GKE 服務
		gkeConfig := gke.ServiceConfig{
			UseCredentials:   true,
			CredentialsFile:  appConfig.GKE.CredentialsFile,
			ProjectID:        appConfig.Credentials.ProjectID,
			ClusterName:      appConfig.Credentials.GkeClusterName,
			Location:         appConfig.Credentials.GkeLocation,
			DefaultNamespace: appConfig.GKE.Namespace,
			Logger:           appLogger,
		}

		gkeService, err = gke.NewServiceWithConfig(gkeConfig)
		if err != nil {
			log.Fatalf("使用 Google Cloud 凭证初始化 GKE 服務失敗: %v", err)
		}
		successMsg := fmt.Sprintf("成功使用 Google Cloud 凭证連接到 GKE 集群: %s", appConfig.Credentials.GkeClusterName)
		if !isStdioMode {
			fmt.Println(successMsg)
		}
		appLogger.Println(successMsg)
	} else {
		// 使用傳統的 kubeconfig 方式
		defaultConfig := gke.ServiceConfig{
			Logger: appLogger,
		}
		gkeService, err = gke.NewServiceWithConfig(defaultConfig)
		if err != nil {
			log.Fatalf("初始化 GKE 服務失敗: %v", err)
		}
		if !isStdioMode {
			fmt.Println("使用傳統 kubeconfig 連接到 GKE")
		}
		appLogger.Println("使用傳統 kubeconfig 連接到 GKE")
	}

	gkeHandler := gke.NewHandler(gkeService)

	//-----------------------------------------------------------------
	// 優化服務
	//-----------------------------------------------------------------
	optimizationService, err := optimization.NewServiceWithLogger(gkeService, appLogger)
	if err != nil {
		log.Fatalf("初始化優化服務失敗: %v", err)
	}

	optimizationHandler := optimization.NewHandler(optimizationService)

	//-----------------------------------------------------------------
	// MCP 伺服器
	//-----------------------------------------------------------------

	// 初始化 MCP 伺服器
	mcpServer := server.NewMCPServer(server.MCPConfig{
		Name:    "mcp-gke-monitor",
		Version: "0.0.1",
		Logger:  appLogger,
	})

	// 註冊工具
	registeredTools := server.RegisterTools(mcpServer, gkeHandler, optimizationHandler)

	// 註冊資源
	server.RegisterResources(mcpServer)

	if !isStdioMode {
		fmt.Println("MCP 伺服器初始化完成")
		// 顯示已註冊的工具列表
		fmt.Printf("已註冊 %d 個工具:\n", len(registeredTools))
		for i, toolName := range registeredTools {
			fmt.Printf("  %d. %s\n", i+1, toolName)
		}
	}

	// 記錄到日誌文件
	appLogger.Println("MCP 伺服器初始化完成")
	appLogger.Printf("已註冊 %d 個工具", len(registeredTools))
	for i, toolName := range registeredTools {
		appLogger.Printf("  %d. %s", i+1, toolName)
	}

	// 啟動伺服器 (根據組態決定啟動模式)
	if err := server.StartServer(mcpServer, appConfig, appLogger); err != nil {
		log.Fatalf("伺服器錯誤: %v", err)
	}
}
