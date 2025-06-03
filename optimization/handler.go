package optimization

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GenerateOptimizationReport 生成完整的優化報告
func (h *Handler) GenerateOptimizationReport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 從請求中獲取命名空間參數
	namespace := ""
	if ns, ok := request.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	report, err := h.service.GenerateOptimizationReport(namespace)
	if err != nil {
		return nil, fmt.Errorf("生成優化報告失敗: %w", err)
	}

	reportJSON, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("序列化優化報告失敗: %w", err)
	}

	return mcp.NewToolResultText(string(reportJSON)), nil
}

// GetOptimizationSummary 取得優化摘要
func (h *Handler) GetOptimizationSummary(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 從請求中獲取命名空間參數
	namespace := ""
	if ns, ok := request.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	// 生成完整報告然後提取摘要
	report, err := h.service.GenerateOptimizationReport(namespace)
	if err != nil {
		return nil, fmt.Errorf("生成優化摘要失敗: %w", err)
	}

	// 創建簡化的摘要回應
	summaryResponse := struct {
		ClusterName string              `json:"clusterName"`
		Namespace   string              `json:"namespace"`
		GeneratedAt string              `json:"generatedAt"`
		Summary     OptimizationSummary `json:"summary"`
		TopIssues   []string            `json:"topIssues"`
	}{
		ClusterName: report.ClusterName,
		Namespace:   report.Namespace,
		GeneratedAt: report.GeneratedAt.Format("2006-01-02 15:04:05"),
		Summary:     report.Summary,
		TopIssues:   h.extractTopIssues(report.Recommendations),
	}

	summaryJSON, err := json.Marshal(summaryResponse)
	if err != nil {
		return nil, fmt.Errorf("序列化優化摘要失敗: %w", err)
	}

	return mcp.NewToolResultText(string(summaryJSON)), nil
}

// GetOptimizationRecommendations 取得優化建議
func (h *Handler) GetOptimizationRecommendations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 從請求中獲取參數
	namespace := ""
	if ns, ok := request.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	priority := ""
	if p, ok := request.Params.Arguments["priority"].(string); ok {
		priority = p
	}

	recommendationType := ""
	if rt, ok := request.Params.Arguments["type"].(string); ok {
		recommendationType = rt
	}

	// 生成完整報告
	report, err := h.service.GenerateOptimizationReport(namespace)
	if err != nil {
		return nil, fmt.Errorf("取得優化建議失敗: %w", err)
	}

	// 過濾建議
	filteredRecommendations := h.filterRecommendations(report.Recommendations, priority, recommendationType)

	// 創建回應
	response := struct {
		ClusterName     string           `json:"clusterName"`
		Namespace       string           `json:"namespace"`
		GeneratedAt     string           `json:"generatedAt"`
		TotalCount      int              `json:"totalCount"`
		FilteredCount   int              `json:"filteredCount"`
		Recommendations []Recommendation `json:"recommendations"`
	}{
		ClusterName:     report.ClusterName,
		Namespace:       report.Namespace,
		GeneratedAt:     report.GeneratedAt.Format("2006-01-02 15:04:05"),
		TotalCount:      len(report.Recommendations),
		FilteredCount:   len(filteredRecommendations),
		Recommendations: filteredRecommendations,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("序列化優化建議失敗: %w", err)
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// GetResourceWasteAnalysis 取得資源浪費分析
func (h *Handler) GetResourceWasteAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 從請求中獲取命名空間參數
	namespace := ""
	if ns, ok := request.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	// 生成完整報告
	report, err := h.service.GenerateOptimizationReport(namespace)
	if err != nil {
		return nil, fmt.Errorf("取得資源浪費分析失敗: %w", err)
	}

	// 創建詳細的浪費分析回應
	response := struct {
		ClusterName   string                `json:"clusterName"`
		Namespace     string                `json:"namespace"`
		GeneratedAt   string                `json:"generatedAt"`
		ResourceWaste ResourceWasteAnalysis `json:"resourceWaste"`
		Insights      []string              `json:"insights"`
	}{
		ClusterName:   report.ClusterName,
		Namespace:     report.Namespace,
		GeneratedAt:   report.GeneratedAt.Format("2006-01-02 15:04:05"),
		ResourceWaste: report.ResourceWaste,
		Insights:      h.generateWasteInsights(report.ResourceWaste),
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("序列化資源浪費分析失敗: %w", err)
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// GetPodOptimizationAnalysis 取得特定 Pod 的優化分析
func (h *Handler) GetPodOptimizationAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Pod 名稱是必要參數
	podName, ok := request.Params.Arguments["podName"].(string)
	if !ok || podName == "" {
		return nil, errors.New("必須提供有效的 Pod 名稱")
	}

	// 命名空間是可選參數
	namespace := ""
	if ns, ok := request.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	// 生成完整報告
	report, err := h.service.GenerateOptimizationReport(namespace)
	if err != nil {
		return nil, fmt.Errorf("取得 Pod 優化分析失敗: %w", err)
	}

	// 找到指定的 Pod 分析
	var podAnalysis *PodOptimization
	for _, analysis := range report.PodAnalysis {
		if analysis.PodName == podName {
			podAnalysis = &analysis
			break
		}
	}

	if podAnalysis == nil {
		return nil, fmt.Errorf("找不到 Pod %s 的分析資料", podName)
	}

	// 找到相關的建議
	var relatedRecommendations []Recommendation
	for _, rec := range report.Recommendations {
		if rec.PodName == podName {
			relatedRecommendations = append(relatedRecommendations, rec)
		}
	}

	// 創建詳細的 Pod 分析回應
	response := struct {
		PodName                 string           `json:"podName"`
		Namespace               string           `json:"namespace"`
		GeneratedAt             string           `json:"generatedAt"`
		PodAnalysis             PodOptimization  `json:"podAnalysis"`
		RelatedRecommendations  []Recommendation `json:"relatedRecommendations"`
		OptimizationSuggestions []string         `json:"optimizationSuggestions"`
	}{
		PodName:                 podName,
		Namespace:               namespace,
		GeneratedAt:             report.GeneratedAt.Format("2006-01-02 15:04:05"),
		PodAnalysis:             *podAnalysis,
		RelatedRecommendations:  relatedRecommendations,
		OptimizationSuggestions: h.generatePodOptimizationSuggestions(*podAnalysis),
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("序列化 Pod 優化分析失敗: %w", err)
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// GetOptimizationCriteria 取得優化標準
func (h *Handler) GetOptimizationCriteria(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	criteria := h.service.GetOptimizationCriteria()

	response := struct {
		Criteria    OptimizationCriteria `json:"criteria"`
		Description map[string]string    `json:"description"`
	}{
		Criteria: criteria,
		Description: map[string]string{
			"cpuThreshold":    "CPU 使用率低於此值視為過度配置",
			"memoryThreshold": "記憶體使用率低於此值視為過度配置",
			"healthThreshold": "重啟次數超過此值視為不健康",
			"idleThreshold":   "使用率低於此值視為閒置",
		},
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("序列化優化標準失敗: %w", err)
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// UpdateOptimizationCriteria 更新優化標準
func (h *Handler) UpdateOptimizationCriteria(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 解析新的標準
	var newCriteria OptimizationCriteria

	if cpuThreshold, ok := request.Params.Arguments["cpuThreshold"].(float64); ok {
		newCriteria.CPUThreshold = cpuThreshold
	} else {
		newCriteria.CPUThreshold = h.service.GetOptimizationCriteria().CPUThreshold
	}

	if memoryThreshold, ok := request.Params.Arguments["memoryThreshold"].(float64); ok {
		newCriteria.MemoryThreshold = memoryThreshold
	} else {
		newCriteria.MemoryThreshold = h.service.GetOptimizationCriteria().MemoryThreshold
	}

	if healthThreshold, ok := request.Params.Arguments["healthThreshold"].(float64); ok {
		newCriteria.HealthThreshold = int32(healthThreshold)
	} else {
		newCriteria.HealthThreshold = h.service.GetOptimizationCriteria().HealthThreshold
	}

	if idleThreshold, ok := request.Params.Arguments["idleThreshold"].(float64); ok {
		newCriteria.IdleThreshold = idleThreshold
	} else {
		newCriteria.IdleThreshold = h.service.GetOptimizationCriteria().IdleThreshold
	}

	// 更新標準
	h.service.UpdateOptimizationCriteria(newCriteria)

	response := struct {
		Message     string               `json:"message"`
		UpdatedAt   string               `json:"updatedAt"`
		NewCriteria OptimizationCriteria `json:"newCriteria"`
	}{
		Message:     "優化標準已成功更新",
		UpdatedAt:   fmt.Sprintf("%v", request.Params.Arguments),
		NewCriteria: newCriteria,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("序列化更新結果失敗: %w", err)
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// 輔助函數

// extractTopIssues 提取主要問題
func (h *Handler) extractTopIssues(recommendations []Recommendation) []string {
	var topIssues []string
	issueCount := make(map[string]int)

	// 統計問題類型
	for _, rec := range recommendations {
		if rec.Priority == PriorityHigh {
			issueCount[string(rec.Type)]++
		}
	}

	// 提取前 5 個最常見的問題
	for issueType, count := range issueCount {
		if len(topIssues) < 5 {
			topIssues = append(topIssues, fmt.Sprintf("%s: %d 個高優先級問題", issueType, count))
		}
	}

	if len(topIssues) == 0 {
		topIssues = append(topIssues, "目前沒有發現高優先級問題")
	}

	return topIssues
}

// filterRecommendations 過濾建議
func (h *Handler) filterRecommendations(recommendations []Recommendation, priority, recommendationType string) []Recommendation {
	var filtered []Recommendation

	for _, rec := range recommendations {
		// 優先級過濾
		if priority != "" && string(rec.Priority) != priority {
			continue
		}

		// 類型過濾
		if recommendationType != "" && string(rec.Type) != recommendationType {
			continue
		}

		filtered = append(filtered, rec)
	}

	return filtered
}

// generateWasteInsights 生成浪費洞察
func (h *Handler) generateWasteInsights(waste ResourceWasteAnalysis) []string {
	var insights []string

	if len(waste.OverProvisionedPods) > 0 {
		insights = append(insights, fmt.Sprintf("發現 %d 個過度配置的 Pod", len(waste.OverProvisionedPods)))
	}

	if len(waste.IdlePods) > 0 {
		insights = append(insights, fmt.Sprintf("發現 %d 個閒置 Pod，建議考慮縮減或刪除", len(waste.IdlePods)))
	}

	if waste.TotalWastage.WastePercentage > 20 {
		insights = append(insights, fmt.Sprintf("整體資源浪費率達 %.1f%%，建議立即優化", waste.TotalWastage.WastePercentage))
	} else if waste.TotalWastage.WastePercentage > 10 {
		insights = append(insights, fmt.Sprintf("整體資源浪費率為 %.1f%%，有優化空間", waste.TotalWastage.WastePercentage))
	} else {
		insights = append(insights, "資源使用效率良好")
	}

	if len(insights) == 0 {
		insights = append(insights, "未發現明顯的資源浪費問題")
	}

	return insights
}

// generatePodOptimizationSuggestions 生成 Pod 優化建議
func (h *Handler) generatePodOptimizationSuggestions(podAnalysis PodOptimization) []string {
	var suggestions []string

	if podAnalysis.OptimizationScore < 50 {
		suggestions = append(suggestions, "該 Pod 需要重點優化，建議檢查所有資源配置")
	} else if podAnalysis.OptimizationScore < 70 {
		suggestions = append(suggestions, "該 Pod 有改善空間，建議檢查主要問題")
	} else {
		suggestions = append(suggestions, "該 Pod 運行狀況良好")
	}

	for _, issue := range podAnalysis.Issues {
		if issue.Severity == PriorityHigh {
			suggestions = append(suggestions, fmt.Sprintf("高優先級: %s", issue.Suggestion))
		}
	}

	if len(suggestions) == 1 {
		suggestions = append(suggestions, "持續監控資源使用狀況")
	}

	return suggestions
}
