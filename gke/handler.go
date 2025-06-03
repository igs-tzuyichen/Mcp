package gke

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

// GetAllPods 取得所有 Pod
func (h *Handler) GetAllPods(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 從請求中獲取命名空間參數
	namespace := ""
	if ns, ok := request.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	pods, err := h.service.GetAllPods(namespace)
	if err != nil {
		return nil, fmt.Errorf("取得 Pod 列表失敗: %w", err)
	}

	podsJSON, err := json.Marshal(pods)
	if err != nil {
		return nil, fmt.Errorf("序列化 Pod 資料失敗: %w", err)
	}

	return mcp.NewToolResultText(string(podsJSON)), nil
}

// SearchPods 根據條件搜尋 Pod
func (h *Handler) SearchPods(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	criteria := SearchCriteria{}

	// 從請求中獲取搜尋參數
	if namespace, ok := request.Params.Arguments["namespace"].(string); ok && namespace != "" {
		criteria.Namespace = namespace
	}

	if labelSelector, ok := request.Params.Arguments["labelSelector"].(string); ok && labelSelector != "" {
		criteria.LabelSelector = labelSelector
	}

	if fieldSelector, ok := request.Params.Arguments["fieldSelector"].(string); ok && fieldSelector != "" {
		criteria.FieldSelector = fieldSelector
	}

	if status, ok := request.Params.Arguments["status"].(string); ok && status != "" {
		criteria.Status = status
	}

	pods, err := h.service.SearchPods(criteria)
	if err != nil {
		return nil, fmt.Errorf("搜尋 Pod 失敗: %w", err)
	}

	podsJSON, err := json.Marshal(pods)
	if err != nil {
		return nil, fmt.Errorf("序列化 Pod 資料失敗: %w", err)
	}

	return mcp.NewToolResultText(string(podsJSON)), nil
}

// GetPodCPUUsage 取得 Pod 的 CPU 使用狀況
func (h *Handler) GetPodCPUUsage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	usage, err := h.service.GetPodResourceUsage(podName, namespace)
	if err != nil {
		return nil, fmt.Errorf("取得 Pod 資源使用狀況失敗: %w", err)
	}

	// 只返回 CPU 相關資訊
	cpuInfo := struct {
		PodName    string           `json:"podName"`
		Namespace  string           `json:"namespace"`
		CPU        CPUUsage         `json:"cpu"`
		Timestamp  string           `json:"timestamp"`
		Containers []ContainerUsage `json:"containers"`
	}{
		PodName:    usage.PodName,
		Namespace:  usage.Namespace,
		CPU:        usage.CPU,
		Timestamp:  usage.Timestamp.Format("2006-01-02 15:04:05"),
		Containers: usage.Containers,
	}

	cpuJSON, err := json.Marshal(cpuInfo)
	if err != nil {
		return nil, fmt.Errorf("序列化 CPU 使用資料失敗: %w", err)
	}

	return mcp.NewToolResultText(string(cpuJSON)), nil
}

// GetPodMemoryUsage 取得 Pod 的記憶體使用狀況
func (h *Handler) GetPodMemoryUsage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	usage, err := h.service.GetPodResourceUsage(podName, namespace)
	if err != nil {
		return nil, fmt.Errorf("取得 Pod 資源使用狀況失敗: %w", err)
	}

	// 只返回記憶體相關資訊
	memoryInfo := struct {
		PodName    string           `json:"podName"`
		Namespace  string           `json:"namespace"`
		Memory     MemoryUsage      `json:"memory"`
		Timestamp  string           `json:"timestamp"`
		Containers []ContainerUsage `json:"containers"`
	}{
		PodName:    usage.PodName,
		Namespace:  usage.Namespace,
		Memory:     usage.Memory,
		Timestamp:  usage.Timestamp.Format("2006-01-02 15:04:05"),
		Containers: usage.Containers,
	}

	memoryJSON, err := json.Marshal(memoryInfo)
	if err != nil {
		return nil, fmt.Errorf("序列化記憶體使用資料失敗: %w", err)
	}

	return mcp.NewToolResultText(string(memoryJSON)), nil
}

// GetPodDiskUsage 取得 Pod 的磁碟使用狀況
func (h *Handler) GetPodDiskUsage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	usage, err := h.service.GetPodResourceUsage(podName, namespace)
	if err != nil {
		return nil, fmt.Errorf("取得 Pod 資源使用狀況失敗: %w", err)
	}

	// 只返回磁碟相關資訊
	diskInfo := struct {
		PodName   string    `json:"podName"`
		Namespace string    `json:"namespace"`
		Disk      DiskUsage `json:"disk"`
		Timestamp string    `json:"timestamp"`
	}{
		PodName:   usage.PodName,
		Namespace: usage.Namespace,
		Disk:      usage.Disk,
		Timestamp: usage.Timestamp.Format("2006-01-02 15:04:05"),
	}

	diskJSON, err := json.Marshal(diskInfo)
	if err != nil {
		return nil, fmt.Errorf("序列化磁碟使用資料失敗: %w", err)
	}

	return mcp.NewToolResultText(string(diskJSON)), nil
}

// GetPodDetails 取得 Pod 的詳細資訊
func (h *Handler) GetPodDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	details, err := h.service.GetPodDetails(podName, namespace)
	if err != nil {
		return nil, fmt.Errorf("取得 Pod 詳細資訊失敗: %w", err)
	}

	// 格式化時間戳
	formattedDetails := struct {
		Basic     Pod           `json:"basic"`
		Usage     ResourceUsage `json:"usage"`
		Events    []Event       `json:"events"`
		Logs      string        `json:"logs"`
		Timestamp string        `json:"timestamp"`
	}{
		Basic:     details.Basic,
		Usage:     details.Usage,
		Events:    details.Events,
		Logs:      details.Logs,
		Timestamp: details.Usage.Timestamp.Format("2006-01-02 15:04:05"),
	}

	detailsJSON, err := json.Marshal(formattedDetails)
	if err != nil {
		return nil, fmt.Errorf("序列化 Pod 詳細資訊失敗: %w", err)
	}

	return mcp.NewToolResultText(string(detailsJSON)), nil
}
