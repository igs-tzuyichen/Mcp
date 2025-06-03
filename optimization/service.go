package optimization

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"mcp-gke-monitor/gke"
)

// Logger 接口，用於可選的日誌記錄
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// Service 優化服務
type Service struct {
	gkeService *gke.Service
	mu         sync.RWMutex
	criteria   OptimizationCriteria
	logger     Logger // 可選的 logger
}

// NewService 創建一個新的優化服務
func NewService(gkeService *gke.Service) (*Service, error) {
	return NewServiceWithLogger(gkeService, nil)
}

// NewServiceWithLogger 創建一個帶有 logger 的優化服務
func NewServiceWithLogger(gkeService *gke.Service, logger Logger) (*Service, error) {
	if gkeService == nil {
		return nil, fmt.Errorf("GKE 服務不能為空")
	}

	return &Service{
		gkeService: gkeService,
		criteria: OptimizationCriteria{
			CPUThreshold:    20.0, // CPU 使用率低於 20% 視為過度配置
			MemoryThreshold: 30.0, // 記憶體使用率低於 30% 視為過度配置
			HealthThreshold: 5,    // 重啟次數超過 5 次視為不健康
			IdleThreshold:   5.0,  // 使用率低於 5% 視為閒置
		},
		logger: logger,
	}, nil
}

// GenerateOptimizationReport 生成完整的優化報告
func (s *Service) GenerateOptimizationReport(namespace string) (*OptimizationReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespace == "" {
		namespace = "default"
	}

	if s.logger != nil {
		s.logger.Printf("正在生成 %s 命名空間的優化報告...", namespace)
	}

	// 取得所有 Pod
	pods, err := s.gkeService.GetAllPods(namespace)
	if err != nil {
		return nil, fmt.Errorf("無法取得 Pod 列表: %w", err)
	}

	// 分析所有 Pod
	var podAnalysis []PodOptimization
	var recommendations []Recommendation
	var resourceWaste ResourceWasteAnalysis

	for _, pod := range pods {
		// 分析每個 Pod
		podOpt, err := s.analyzePod(pod)
		if err != nil {
			if s.logger != nil {
				s.logger.Printf("警告: 分析 Pod %s 失敗: %v", pod.Name, err)
			}
			continue
		}
		podAnalysis = append(podAnalysis, *podOpt)

		// 生成建議
		podRecommendations := s.generatePodRecommendations(*podOpt)
		recommendations = append(recommendations, podRecommendations...)
	}

	// 分析資源浪費
	resourceWaste = s.analyzeResourceWaste(podAnalysis)

	// 生成摘要
	summary := s.generateSummary(podAnalysis, resourceWaste)

	report := &OptimizationReport{
		ClusterName:     "GKE-Cluster", // 可以從配置中取得
		Namespace:       namespace,
		GeneratedAt:     time.Now(),
		Summary:         summary,
		Recommendations: recommendations,
		PodAnalysis:     podAnalysis,
		ResourceWaste:   resourceWaste,
	}

	return report, nil
}

// analyzePod 分析單個 Pod
func (s *Service) analyzePod(pod gke.Pod) (*PodOptimization, error) {
	// 取得 Pod 的資源使用狀況
	resourceUsage, err := s.gkeService.GetPodResourceUsage(pod.Name, pod.Namespace)
	if err != nil {
		// 如果無法取得 metrics，創建一個基本的分析
		if s.logger != nil {
			s.logger.Printf("無法取得 Pod %s 的資源使用狀況: %v", pod.Name, err)
		}
		resourceUsage = &gke.ResourceUsage{
			PodName:   pod.Name,
			Namespace: pod.Namespace,
			Timestamp: time.Now(),
		}
	}

	// 分析資源使用
	resourceAnalysis := s.analyzeResourceUsage(*resourceUsage)

	// 分析健康狀態
	healthStatus := s.analyzeHealthStatus(pod)

	// 找出優化問題
	issues := s.identifyOptimizationIssues(resourceAnalysis, healthStatus, pod)

	// 計算優化分數
	optimizationScore := s.calculateOptimizationScore(resourceAnalysis, healthStatus, issues)

	podOpt := &PodOptimization{
		PodName:           pod.Name,
		Namespace:         pod.Namespace,
		Status:            pod.Status,
		OptimizationScore: optimizationScore,
		Issues:            issues,
		ResourceAnalysis:  resourceAnalysis,
		HealthStatus:      healthStatus,
	}

	return podOpt, nil
}

// analyzeResourceUsage 分析資源使用狀況
func (s *Service) analyzeResourceUsage(usage gke.ResourceUsage) ResourceAnalysis {
	cpuMetric := s.analyzeResourceMetric(usage.CPU.Current, usage.CPU.Request, usage.CPU.Limit, "CPU")
	memoryMetric := s.analyzeResourceMetric(usage.Memory.Current, usage.Memory.Request, usage.Memory.Limit, "MEMORY")

	// 磁碟分析（簡化版）
	diskMetric := ResourceMetric{
		Current:     usage.Disk.Used,
		Request:     "-",
		Limit:       usage.Disk.Total,
		Utilization: s.calculateDiskUtilization(usage.Disk.Used, usage.Disk.Total),
		Status:      "OPTIMAL",
		Suggestion:  "磁碟使用正常",
	}

	return ResourceAnalysis{
		CPU:    cpuMetric,
		Memory: memoryMetric,
		Disk:   diskMetric,
	}
}

// analyzeResourceMetric 分析單個資源指標
func (s *Service) analyzeResourceMetric(current, request, limit, resourceType string) ResourceMetric {
	metric := ResourceMetric{
		Current: current,
		Request: request,
		Limit:   limit,
	}

	// 計算使用率
	if limit != "" && current != "" {
		utilization := s.calculateUtilization(current, limit)
		metric.Utilization = utilization

		// 判斷狀態和建議
		if utilization < s.criteria.IdleThreshold {
			metric.Status = "IDLE"
			metric.Suggestion = fmt.Sprintf("%s 使用率極低 (%.1f%%)，考慮縮減資源", resourceType, utilization)
		} else if utilization < s.criteria.CPUThreshold && resourceType == "CPU" {
			metric.Status = "OVER_PROVISIONED"
			metric.Suggestion = fmt.Sprintf("CPU 過度配置，使用率僅 %.1f%%，建議減少 CPU 限制", utilization)
		} else if utilization < s.criteria.MemoryThreshold && resourceType == "MEMORY" {
			metric.Status = "OVER_PROVISIONED"
			metric.Suggestion = fmt.Sprintf("記憶體過度配置，使用率僅 %.1f%%，建議減少記憶體限制", utilization)
		} else if utilization > 80 {
			metric.Status = "UNDER_PROVISIONED"
			metric.Suggestion = fmt.Sprintf("%s 使用率過高 (%.1f%%)，建議增加資源限制", resourceType, utilization)
		} else {
			metric.Status = "OPTIMAL"
			metric.Suggestion = fmt.Sprintf("%s 使用率正常 (%.1f%%)", resourceType, utilization)
		}
	} else {
		metric.Status = "UNKNOWN"
		metric.Suggestion = "無法計算使用率，缺少限制或當前使用量資訊"
	}

	return metric
}

// calculateUtilization 計算使用率
func (s *Service) calculateUtilization(current, limit string) float64 {
	currentVal := s.parseResourceValue(current)
	limitVal := s.parseResourceValue(limit)

	if limitVal == 0 {
		return 0
	}

	return (currentVal / limitVal) * 100
}

// calculateDiskUtilization 計算磁碟使用率
func (s *Service) calculateDiskUtilization(used, total string) float64 {
	usedVal := s.parseResourceValue(used)
	totalVal := s.parseResourceValue(total)

	if totalVal == 0 {
		return 0
	}

	return (usedVal / totalVal) * 100
}

// parseResourceValue 解析資源值（簡化版）
func (s *Service) parseResourceValue(value string) float64 {
	if value == "" || value == "-" {
		return 0
	}

	// 移除單位並轉換為數值
	value = strings.ToLower(value)

	// CPU 處理（m = millicore）
	if strings.HasSuffix(value, "m") {
		if val, err := strconv.ParseFloat(strings.TrimSuffix(value, "m"), 64); err == nil {
			return val // millicore
		}
	}

	// 記憶體處理
	if strings.HasSuffix(value, "mi") {
		if val, err := strconv.ParseFloat(strings.TrimSuffix(value, "mi"), 64); err == nil {
			return val // MiB
		}
	}
	if strings.HasSuffix(value, "gi") {
		if val, err := strconv.ParseFloat(strings.TrimSuffix(value, "gi"), 64); err == nil {
			return val * 1024 // GiB to MiB
		}
	}

	// 嘗試直接解析數值
	if val, err := strconv.ParseFloat(value, 64); err == nil {
		return val
	}

	return 0
}

// analyzeHealthStatus 分析健康狀態
func (s *Service) analyzeHealthStatus(pod gke.Pod) HealthStatus {
	var totalRestarts int32
	var lastRestart time.Time
	var healthIssues []string

	for _, container := range pod.Containers {
		totalRestarts += container.Restart
		if container.Restart > 0 {
			healthIssues = append(healthIssues, fmt.Sprintf("容器 %s 已重啟 %d 次", container.Name, container.Restart))
		}
		if !container.Ready {
			healthIssues = append(healthIssues, fmt.Sprintf("容器 %s 未就緒", container.Name))
		}
	}

	// 計算健康分數
	healthScore := 100.0
	if totalRestarts > s.criteria.HealthThreshold {
		healthScore -= float64(totalRestarts-s.criteria.HealthThreshold) * 10
	}
	if !pod.Ready {
		healthScore -= 30
	}
	if pod.Status != "Running" {
		healthScore -= 40
	}

	if healthScore < 0 {
		healthScore = 0
	}

	return HealthStatus{
		Ready:        pod.Ready,
		RestartCount: totalRestarts,
		LastRestart:  lastRestart,
		HealthScore:  healthScore,
		HealthIssues: healthIssues,
	}
}

// identifyOptimizationIssues 識別優化問題
func (s *Service) identifyOptimizationIssues(resourceAnalysis ResourceAnalysis, healthStatus HealthStatus, pod gke.Pod) []OptimizationIssue {
	var issues []OptimizationIssue

	// CPU 問題
	if resourceAnalysis.CPU.Status == "OVER_PROVISIONED" {
		issues = append(issues, OptimizationIssue{
			Type:        "CPU_OVER_PROVISIONED",
			Severity:    PriorityMedium,
			Description: "CPU 資源過度配置",
			Suggestion:  resourceAnalysis.CPU.Suggestion,
		})
	} else if resourceAnalysis.CPU.Status == "UNDER_PROVISIONED" {
		issues = append(issues, OptimizationIssue{
			Type:        "CPU_UNDER_PROVISIONED",
			Severity:    PriorityHigh,
			Description: "CPU 資源不足",
			Suggestion:  resourceAnalysis.CPU.Suggestion,
		})
	}

	// 記憶體問題
	if resourceAnalysis.Memory.Status == "OVER_PROVISIONED" {
		issues = append(issues, OptimizationIssue{
			Type:        "MEMORY_OVER_PROVISIONED",
			Severity:    PriorityMedium,
			Description: "記憶體資源過度配置",
			Suggestion:  resourceAnalysis.Memory.Suggestion,
		})
	} else if resourceAnalysis.Memory.Status == "UNDER_PROVISIONED" {
		issues = append(issues, OptimizationIssue{
			Type:        "MEMORY_UNDER_PROVISIONED",
			Severity:    PriorityHigh,
			Description: "記憶體資源不足",
			Suggestion:  resourceAnalysis.Memory.Suggestion,
		})
	}

	// 健康問題
	if healthStatus.RestartCount > s.criteria.HealthThreshold {
		issues = append(issues, OptimizationIssue{
			Type:        "HIGH_RESTART_COUNT",
			Severity:    PriorityHigh,
			Description: fmt.Sprintf("容器重啟次數過多 (%d 次)", healthStatus.RestartCount),
			Suggestion:  "檢查應用程式日誌，修復導致重啟的問題",
		})
	}

	if !pod.Ready {
		issues = append(issues, OptimizationIssue{
			Type:        "POD_NOT_READY",
			Severity:    PriorityHigh,
			Description: "Pod 未就緒",
			Suggestion:  "檢查 Pod 狀態和事件，確保所有容器正常運行",
		})
	}

	return issues
}

// calculateOptimizationScore 計算優化分數
func (s *Service) calculateOptimizationScore(resourceAnalysis ResourceAnalysis, healthStatus HealthStatus, issues []OptimizationIssue) float64 {
	score := 100.0

	// 根據問題減分
	for _, issue := range issues {
		switch issue.Severity {
		case PriorityHigh:
			score -= 20
		case PriorityMedium:
			score -= 10
		case PriorityLow:
			score -= 5
		}
	}

	// 根據健康分數調整
	score = (score + healthStatus.HealthScore) / 2

	if score < 0 {
		score = 0
	}

	return score
}

// generatePodRecommendations 為 Pod 生成建議
func (s *Service) generatePodRecommendations(podOpt PodOptimization) []Recommendation {
	var recommendations []Recommendation
	idCounter := 1

	for _, issue := range podOpt.Issues {
		rec := Recommendation{
			ID:          fmt.Sprintf("REC-%s-%d", podOpt.PodName, idCounter),
			Type:        s.mapIssueTypeToRecommendationType(issue.Type),
			Priority:    issue.Severity,
			Title:       issue.Description,
			Description: issue.Suggestion,
			PodName:     podOpt.PodName,
			Namespace:   podOpt.Namespace,
		}

		// 設定影響和行動
		switch issue.Type {
		case "CPU_OVER_PROVISIONED":
			rec.Impact = "減少 CPU 成本，提高資源利用率"
			rec.Action = "調整 CPU requests 和 limits"
		case "MEMORY_OVER_PROVISIONED":
			rec.Impact = "減少記憶體成本，提高資源利用率"
			rec.Action = "調整記憶體 requests 和 limits"
		case "HIGH_RESTART_COUNT":
			rec.Impact = "提高應用程式穩定性和可用性"
			rec.Action = "檢查應用程式日誌並修復問題"
		case "POD_NOT_READY":
			rec.Impact = "確保服務正常運行"
			rec.Action = "檢查 Pod 狀態和健康檢查"
		}

		recommendations = append(recommendations, rec)
		idCounter++
	}

	return recommendations
}

// mapIssueTypeToRecommendationType 將問題類型映射到建議類型
func (s *Service) mapIssueTypeToRecommendationType(issueType string) RecommendationType {
	switch {
	case strings.Contains(issueType, "CPU"):
		return RecommendationCPU
	case strings.Contains(issueType, "MEMORY"):
		return RecommendationMemory
	case strings.Contains(issueType, "RESTART") || strings.Contains(issueType, "READY"):
		return RecommendationHealth
	default:
		return RecommendationHealth
	}
}

// analyzeResourceWaste 分析資源浪費
func (s *Service) analyzeResourceWaste(podAnalyses []PodOptimization) ResourceWasteAnalysis {
	var overProvisionedPods []ResourceWaste
	var underUtilizedPods []ResourceWaste
	var idlePods []string

	totalCPUWaste := 0.0
	totalMemoryWaste := 0.0

	for _, podAnalysis := range podAnalyses {
		// 檢查過度配置
		if podAnalysis.ResourceAnalysis.CPU.Status == "OVER_PROVISIONED" {
			wastePercentage := 100 - podAnalysis.ResourceAnalysis.CPU.Utilization
			overProvisionedPods = append(overProvisionedPods, ResourceWaste{
				PodName:         podAnalysis.PodName,
				Namespace:       podAnalysis.Namespace,
				ResourceType:    "CPU",
				Allocated:       podAnalysis.ResourceAnalysis.CPU.Limit,
				Used:            podAnalysis.ResourceAnalysis.CPU.Current,
				WastePercentage: wastePercentage,
				WasteAmount:     fmt.Sprintf("%.1f%%", wastePercentage),
			})
			totalCPUWaste += wastePercentage
		}

		if podAnalysis.ResourceAnalysis.Memory.Status == "OVER_PROVISIONED" {
			wastePercentage := 100 - podAnalysis.ResourceAnalysis.Memory.Utilization
			overProvisionedPods = append(overProvisionedPods, ResourceWaste{
				PodName:         podAnalysis.PodName,
				Namespace:       podAnalysis.Namespace,
				ResourceType:    "MEMORY",
				Allocated:       podAnalysis.ResourceAnalysis.Memory.Limit,
				Used:            podAnalysis.ResourceAnalysis.Memory.Current,
				WastePercentage: wastePercentage,
				WasteAmount:     fmt.Sprintf("%.1f%%", wastePercentage),
			})
			totalMemoryWaste += wastePercentage
		}

		// 檢查閒置 Pod
		if podAnalysis.ResourceAnalysis.CPU.Utilization < s.criteria.IdleThreshold &&
			podAnalysis.ResourceAnalysis.Memory.Utilization < s.criteria.IdleThreshold {
			idlePods = append(idlePods, podAnalysis.PodName)
		}
	}

	// 計算總體浪費
	avgWastePercentage := 0.0
	if len(overProvisionedPods) > 0 {
		avgWastePercentage = (totalCPUWaste + totalMemoryWaste) / float64(len(overProvisionedPods)*2)
	}

	wastageStats := WastageStats{
		TotalCPUWaste:    fmt.Sprintf("%.1f%%", totalCPUWaste),
		TotalMemoryWaste: fmt.Sprintf("%.1f%%", totalMemoryWaste),
		WastePercentage:  avgWastePercentage,
		EstimatedCost:    "需要更多成本資訊來計算",
	}

	return ResourceWasteAnalysis{
		OverProvisionedPods: overProvisionedPods,
		UnderUtilizedPods:   underUtilizedPods,
		IdlePods:            idlePods,
		TotalWastage:        wastageStats,
	}
}

// generateSummary 生成摘要
func (s *Service) generateSummary(podAnalyses []PodOptimization, resourceWaste ResourceWasteAnalysis) OptimizationSummary {
	totalPods := len(podAnalyses)
	podsNeedingOptimization := 0

	for _, podAnalysis := range podAnalyses {
		if len(podAnalysis.Issues) > 0 {
			podsNeedingOptimization++
		}
	}

	// 計算總體分數
	totalScore := 0.0
	for _, podAnalysis := range podAnalyses {
		totalScore += podAnalysis.OptimizationScore
	}
	overallScore := 0.0
	if totalPods > 0 {
		overallScore = totalScore / float64(totalPods)
	}

	return OptimizationSummary{
		TotalPods:               totalPods,
		PodsNeedingOptimization: podsNeedingOptimization,
		PotentialCPUSavings:     resourceWaste.TotalWastage.TotalCPUWaste,
		PotentialMemorySavings:  resourceWaste.TotalWastage.TotalMemoryWaste,
		OverallScore:            overallScore,
	}
}

// GetOptimizationCriteria 取得優化標準
func (s *Service) GetOptimizationCriteria() OptimizationCriteria {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.criteria
}

// UpdateOptimizationCriteria 更新優化標準
func (s *Service) UpdateOptimizationCriteria(criteria OptimizationCriteria) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.criteria = criteria
}
