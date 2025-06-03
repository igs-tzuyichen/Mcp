package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"mcp-gke-monitor/config"
	"mcp-gke-monitor/logger"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type MCPConfig struct {
	Name    string
	Version string
	Logger  *logger.Logger
}

func NewMCPServer(cfg MCPConfig) *mcpserver.MCPServer {
	loggingHooks := cfg.Logger.ConfigureLoggingHooks()

	s := mcpserver.NewMCPServer(
		cfg.Name,
		cfg.Version,
		mcpserver.WithLogging(),
		mcpserver.WithHooks(loggingHooks),
		mcpserver.WithResourceCapabilities(true, true), // 啟用資源功能
	)

	return s
}

// 註冊所有可用的工具函數
func RegisterTools(s *mcpserver.MCPServer, handler ToolHandler, optimizationHandler OptimizationHandler) []string {
	var registeredTools []string

	// ========== GKE Pod 監控工具 ==========

	// 建立取得所有 Pod 的工具
	getAllPodsTool := mcp.NewTool("get_all_pods",
		mcp.WithDescription("Get all GKE Pod list"),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
	)

	// 建立根據不同條件搜尋 Pod 的工具
	searchPodsTool := mcp.NewTool("search_pods",
		mcp.WithDescription("Search GKE Pods by criteria"),
		mcp.WithString("namespace",
			mcp.Description("Namespace"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Label selector"),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Field selector"),
		),
		mcp.WithString("status",
			mcp.Description("Pod status (Running, Pending, Succeeded, Failed, Unknown)"),
		),
	)

	// 建立取得 Pod CPU 使用狀況的工具
	getPodCPUUsageTool := mcp.NewTool("get_pod_cpu_usage",
		mcp.WithDescription("Get Pod CPU usage"),
		mcp.WithString("podName",
			mcp.Required(),
			mcp.Description("Pod name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
	)

	// 建立取得 Pod 記憶體使用狀況的工具
	getPodMemoryUsageTool := mcp.NewTool("get_pod_memory_usage",
		mcp.WithDescription("Get Pod memory usage"),
		mcp.WithString("podName",
			mcp.Required(),
			mcp.Description("Pod name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
	)

	// 建立取得 Pod 磁碟使用狀況的工具
	getPodDiskUsageTool := mcp.NewTool("get_pod_disk_usage",
		mcp.WithDescription("Get Pod disk usage"),
		mcp.WithString("podName",
			mcp.Required(),
			mcp.Description("Pod name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
	)

	// 建立取得 Pod 詳細資訊的工具
	getPodDetailsTool := mcp.NewTool("get_pod_details",
		mcp.WithDescription("Get Pod detailed information including resource usage"),
		mcp.WithString("podName",
			mcp.Required(),
			mcp.Description("Pod name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
	)

	// ========== GKE 優化建議工具 ==========

	// 建立生成優化報告的工具
	generateOptimizationReportTool := mcp.NewTool("generate_optimization_report",
		mcp.WithDescription("Generate comprehensive GKE optimization report with resource analysis and recommendations"),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
	)

	// 建立取得優化摘要的工具
	getOptimizationSummaryTool := mcp.NewTool("get_optimization_summary",
		mcp.WithDescription("Get GKE optimization summary with key metrics and major issues"),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
	)

	// 建立取得優化建議的工具
	getOptimizationRecommendationsTool := mcp.NewTool("get_optimization_recommendations",
		mcp.WithDescription("Get GKE optimization recommendations filtered by priority and type"),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
		mcp.WithString("priority",
			mcp.Description("Priority filter (HIGH, MEDIUM, LOW)"),
		),
		mcp.WithString("type",
			mcp.Description("Recommendation type filter (CPU, MEMORY, HEALTH, STORAGE, REPLICA, SECURITY)"),
		),
	)

	// 建立取得資源浪費分析的工具
	getResourceWasteAnalysisTool := mcp.NewTool("get_resource_waste_analysis",
		mcp.WithDescription("Analyze GKE resource waste and identify over-provisioned and idle resources"),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
	)

	// 建立取得 Pod 優化分析的工具
	getPodOptimizationAnalysisTool := mcp.NewTool("get_pod_optimization_analysis",
		mcp.WithDescription("Get detailed optimization analysis for specific Pod"),
		mcp.WithString("podName",
			mcp.Required(),
			mcp.Description("Pod name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: default)"),
		),
	)

	// 建立取得優化標準的工具
	getOptimizationCriteriaTool := mcp.NewTool("get_optimization_criteria",
		mcp.WithDescription("Get current optimization criteria"),
	)

	// 建立更新優化標準的工具
	updateOptimizationCriteriaTool := mcp.NewTool("update_optimization_criteria",
		mcp.WithDescription("Update optimization criteria"),
		mcp.WithNumber("cpuThreshold",
			mcp.Description("CPU utilization threshold (default: 20.0)"),
		),
		mcp.WithNumber("memoryThreshold",
			mcp.Description("Memory utilization threshold (default: 30.0)"),
		),
		mcp.WithNumber("healthThreshold",
			mcp.Description("Health threshold (default: 5)"),
		),
		mcp.WithNumber("idleThreshold",
			mcp.Description("Idle threshold (default: 5.0)"),
		),
	)

	// 將所有 GKE Pod 監控工具註冊到伺服器並記錄工具名稱
	s.AddTool(getAllPodsTool, handler.GetAllPods)
	registeredTools = append(registeredTools, "get_all_pods")

	s.AddTool(searchPodsTool, handler.SearchPods)
	registeredTools = append(registeredTools, "search_pods")

	s.AddTool(getPodCPUUsageTool, handler.GetPodCPUUsage)
	registeredTools = append(registeredTools, "get_pod_cpu_usage")

	s.AddTool(getPodMemoryUsageTool, handler.GetPodMemoryUsage)
	registeredTools = append(registeredTools, "get_pod_memory_usage")

	s.AddTool(getPodDiskUsageTool, handler.GetPodDiskUsage)
	registeredTools = append(registeredTools, "get_pod_disk_usage")

	s.AddTool(getPodDetailsTool, handler.GetPodDetails)
	registeredTools = append(registeredTools, "get_pod_details")

	// 將所有 GKE 優化建議工具註冊到伺服器並記錄工具名稱
	s.AddTool(generateOptimizationReportTool, optimizationHandler.GenerateOptimizationReport)
	registeredTools = append(registeredTools, "generate_optimization_report")

	s.AddTool(getOptimizationSummaryTool, optimizationHandler.GetOptimizationSummary)
	registeredTools = append(registeredTools, "get_optimization_summary")

	s.AddTool(getOptimizationRecommendationsTool, optimizationHandler.GetOptimizationRecommendations)
	registeredTools = append(registeredTools, "get_optimization_recommendations")

	s.AddTool(getResourceWasteAnalysisTool, optimizationHandler.GetResourceWasteAnalysis)
	registeredTools = append(registeredTools, "get_resource_waste_analysis")

	s.AddTool(getPodOptimizationAnalysisTool, optimizationHandler.GetPodOptimizationAnalysis)
	registeredTools = append(registeredTools, "get_pod_optimization_analysis")

	s.AddTool(getOptimizationCriteriaTool, optimizationHandler.GetOptimizationCriteria)
	registeredTools = append(registeredTools, "get_optimization_criteria")

	s.AddTool(updateOptimizationCriteriaTool, optimizationHandler.UpdateOptimizationCriteria)
	registeredTools = append(registeredTools, "update_optimization_criteria")

	return registeredTools
}

func readGuideContent() (string, error) {

	// 嘗試從不同路徑讀取指南文件
	possiblePaths := []string{
		filepath.Join("internal", "docs", "guide.md"),
		filepath.Join("..", "internal", "docs", "guide.md"),
	}

	// 如果相對路徑失敗，嘗試使用絕對路徑
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		possiblePaths = append(possiblePaths,
			filepath.Join(execDir, "internal", "docs", "guide.md"),
			filepath.Join(execDir, "..", "internal", "docs", "guide.md"),
		)
	}

	// 嘗試每個可能的路徑
	var lastErr error
	for _, path := range possiblePaths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content), nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("無法讀取指南文件: %v", lastErr)
}

// 註冊所有資源
func RegisterResources(s *mcpserver.MCPServer) {

	// 建立靜態文件資源 - 使用指南
	resource := mcp.NewResource(
		"docs://gke/guide",
		"GKE Monitoring and Query Guide",
		mcp.WithResourceDescription("Service functionality description and usage instructions"),
		mcp.WithMIMEType("text/markdown"),
	)

	s.AddResource(resource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// 讀取指南文件內容
		content, err := readGuideContent()
		if err != nil {
			return nil, err
		}

		// 返回資源內容
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "docs://gke/guide",
				MIMEType: "text/markdown",
				Text:     string(content),
			},
		}, nil
	})
}

// 啟動 Stdio 伺服器
func StartStdioServer(s *mcpserver.MCPServer, logger *logger.Logger) error {
	logger.LogServerStart()

	err := mcpserver.ServeStdio(s)

	logger.LogServerStop()

	if err != nil {
		logger.LogServerError(err)
		return err
	}

	return nil
}

// 啟動 SSE (Server-Sent Events) 伺服器
func StartSSEServer(s *mcpserver.MCPServer, baseURL string, port interface{}, logger *logger.Logger) error {
	portStr := fmt.Sprintf("%v", port)

	// 確保 baseURL 包含埠號
	fullBaseURL := fmt.Sprintf("%s:%s", baseURL, portStr)
	fmt.Printf("sse 伺服器啟動於 %s\n", fullBaseURL)
	logger.LogServerStart()

	// 建立 SSE 伺服器 - 使用包含埠號的完整 URL
	sse := mcpserver.NewSSEServer(s, mcpserver.WithBaseURL(fullBaseURL))

	fmt.Printf("正在啟動 SSE 伺服器於埠號 %s...\n", portStr)

	err := sse.Start(":" + portStr)

	if err != nil {
		errMsg := fmt.Sprintf("伺服器錯誤: %v\n", err)
		fmt.Print(errMsg)
		logger.LogServerError(err)

		fmt.Println("按任意鍵繼續...")
		fmt.Scanln()

		return err
	}

	logger.LogServerStop()
	fmt.Println("sse server stopped")
	return nil
}

// 根據配置啟動適當的伺服器類型
func StartServer(s *mcpserver.MCPServer, appConfig config.Config, logger *logger.Logger) error {

	switch appConfig.ServerType {
	case config.ServerTypeSSE:
		fmt.Println("使用 SSE 模式")
		return StartSSEServer(s, appConfig.SSE.BaseURL, appConfig.SSE.Port, logger)
	case config.ServerTypeStdio:
		// 在 stdio 模式下不輸出，避免干擾 MCP 協議
		logger.Println("使用 Stdio 模式")
		return StartStdioServer(s, logger)
	default:
		// 在非 stdio 模式下才輸出
		if appConfig.ServerType != config.ServerTypeStdio {
			fmt.Printf("未知的伺服器類型 %s, 使用預設的 Stdio 模式\n", appConfig.ServerType)
		}
		logger.Printf("未知的伺服器類型 %s, 使用預設的 Stdio 模式", appConfig.ServerType)
		return StartStdioServer(s, logger)
	}
}
