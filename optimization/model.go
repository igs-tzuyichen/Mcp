package optimization

import "time"

// OptimizationReport 優化報告
type OptimizationReport struct {
	ClusterName     string                `json:"clusterName"`
	Namespace       string                `json:"namespace"`
	GeneratedAt     time.Time             `json:"generatedAt"`
	Summary         OptimizationSummary   `json:"summary"`
	Recommendations []Recommendation      `json:"recommendations"`
	PodAnalysis     []PodOptimization     `json:"podAnalysis"`
	ResourceWaste   ResourceWasteAnalysis `json:"resourceWaste"`
}

// OptimizationSummary 優化摘要
type OptimizationSummary struct {
	TotalPods               int     `json:"totalPods"`
	PodsNeedingOptimization int     `json:"podsNeedingOptimization"`
	PotentialCPUSavings     string  `json:"potentialCPUSavings"`
	PotentialMemorySavings  string  `json:"potentialMemorySavings"`
	OverallScore            float64 `json:"overallScore"` // 0-100 分
}

// Recommendation 優化建議
type Recommendation struct {
	ID          string             `json:"id"`
	Type        RecommendationType `json:"type"`
	Priority    Priority           `json:"priority"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Impact      string             `json:"impact"`
	Action      string             `json:"action"`
	PodName     string             `json:"podName,omitempty"`
	Namespace   string             `json:"namespace,omitempty"`
}

// RecommendationType 建議類型
type RecommendationType string

const (
	RecommendationCPU      RecommendationType = "CPU"
	RecommendationMemory   RecommendationType = "MEMORY"
	RecommendationReplica  RecommendationType = "REPLICA"
	RecommendationStorage  RecommendationType = "STORAGE"
	RecommendationHealth   RecommendationType = "HEALTH"
	RecommendationSecurity RecommendationType = "SECURITY"
)

// Priority 優先級
type Priority string

const (
	PriorityHigh   Priority = "HIGH"
	PriorityMedium Priority = "MEDIUM"
	PriorityLow    Priority = "LOW"
)

// PodOptimization Pod 優化分析
type PodOptimization struct {
	PodName           string              `json:"podName"`
	Namespace         string              `json:"namespace"`
	Status            string              `json:"status"`
	OptimizationScore float64             `json:"optimizationScore"` // 0-100 分
	Issues            []OptimizationIssue `json:"issues"`
	ResourceAnalysis  ResourceAnalysis    `json:"resourceAnalysis"`
	HealthStatus      HealthStatus        `json:"healthStatus"`
}

// OptimizationIssue 優化問題
type OptimizationIssue struct {
	Type        string   `json:"type"`
	Severity    Priority `json:"severity"`
	Description string   `json:"description"`
	Suggestion  string   `json:"suggestion"`
}

// ResourceAnalysis 資源分析
type ResourceAnalysis struct {
	CPU    ResourceMetric `json:"cpu"`
	Memory ResourceMetric `json:"memory"`
	Disk   ResourceMetric `json:"disk"`
}

// ResourceMetric 資源指標
type ResourceMetric struct {
	Current     string  `json:"current"`
	Request     string  `json:"request"`
	Limit       string  `json:"limit"`
	Utilization float64 `json:"utilization"` // 使用率百分比
	Status      string  `json:"status"`      // "OPTIMAL", "OVER_PROVISIONED", "UNDER_PROVISIONED"
	Suggestion  string  `json:"suggestion"`
}

// HealthStatus 健康狀態
type HealthStatus struct {
	Ready        bool      `json:"ready"`
	RestartCount int32     `json:"restartCount"`
	LastRestart  time.Time `json:"lastRestart,omitempty"`
	HealthScore  float64   `json:"healthScore"` // 0-100 分
	HealthIssues []string  `json:"healthIssues,omitempty"`
}

// ResourceWasteAnalysis 資源浪費分析
type ResourceWasteAnalysis struct {
	OverProvisionedPods []ResourceWaste `json:"overProvisionedPods"`
	UnderUtilizedPods   []ResourceWaste `json:"underUtilizedPods"`
	IdlePods            []string        `json:"idlePods"`
	TotalWastage        WastageStats    `json:"totalWastage"`
}

// ResourceWaste 資源浪費
type ResourceWaste struct {
	PodName         string  `json:"podName"`
	Namespace       string  `json:"namespace"`
	ResourceType    string  `json:"resourceType"` // "CPU", "MEMORY"
	Allocated       string  `json:"allocated"`
	Used            string  `json:"used"`
	WastePercentage float64 `json:"wastePercentage"`
	WasteAmount     string  `json:"wasteAmount"`
}

// WastageStats 浪費統計
type WastageStats struct {
	TotalCPUWaste    string  `json:"totalCPUWaste"`
	TotalMemoryWaste string  `json:"totalMemoryWaste"`
	WastePercentage  float64 `json:"wastePercentage"`
	EstimatedCost    string  `json:"estimatedCost,omitempty"`
}

// OptimizationCriteria 優化標準
type OptimizationCriteria struct {
	CPUThreshold    float64 `json:"cpuThreshold"`    // CPU 使用率閾值
	MemoryThreshold float64 `json:"memoryThreshold"` // 記憶體使用率閾值
	HealthThreshold int32   `json:"healthThreshold"` // 重啟次數閾值
	IdleThreshold   float64 `json:"idleThreshold"`   // 閒置閾值
}
