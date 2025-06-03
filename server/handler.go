package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

type ToolHandler interface {

	// GKE Pod 監控工具
	// 取得所有 Pod 列表
	GetAllPods(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 根據條件搜尋 Pod
	SearchPods(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 取得 Pod 的 CPU 使用狀況
	GetPodCPUUsage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 取得 Pod 的記憶體使用狀況
	GetPodMemoryUsage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 取得 Pod 的磁碟使用狀況
	GetPodDiskUsage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 取得 Pod 的詳細資訊（包含資源使用狀況）
	GetPodDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

type OptimizationHandler interface {

	// GKE 優化建議工具
	// 生成完整的優化報告
	GenerateOptimizationReport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 取得優化摘要
	GetOptimizationSummary(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 取得優化建議
	GetOptimizationRecommendations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 取得資源浪費分析
	GetResourceWasteAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 取得特定 Pod 的優化分析
	GetPodOptimizationAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 取得優化標準
	GetOptimizationCriteria(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 更新優化標準
	UpdateOptimizationCriteria(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
