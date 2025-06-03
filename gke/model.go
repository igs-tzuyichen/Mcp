package gke

import "time"

// Pod 基本資訊
type Pod struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	NodeName   string            `json:"nodeName"`
	PodIP      string            `json:"podIP"`
	HostIP     string            `json:"hostIP"`
	Labels     map[string]string `json:"labels"`
	CreatedAt  time.Time         `json:"createdAt"`
	Ready      bool              `json:"ready"`
	Containers []Container       `json:"containers"`
}

// 容器資訊
type Container struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	Ready   bool   `json:"ready"`
	Restart int32  `json:"restartCount"`
}

// 資源使用狀況
type ResourceUsage struct {
	PodName    string           `json:"podName"`
	Namespace  string           `json:"namespace"`
	CPU        CPUUsage         `json:"cpu"`
	Memory     MemoryUsage      `json:"memory"`
	Disk       DiskUsage        `json:"disk"`
	Timestamp  time.Time        `json:"timestamp"`
	Containers []ContainerUsage `json:"containers"`
}

// CPU 使用狀況
type CPUUsage struct {
	Current    string  `json:"current"`    // 當前使用量 (例如: "100m")
	Percentage float64 `json:"percentage"` // 使用百分比
	Limit      string  `json:"limit"`      // 限制量
	Request    string  `json:"request"`    // 請求量
}

// 記憶體使用狀況
type MemoryUsage struct {
	Current    string  `json:"current"`    // 當前使用量 (例如: "128Mi")
	Percentage float64 `json:"percentage"` // 使用百分比
	Limit      string  `json:"limit"`      // 限制量
	Request    string  `json:"request"`    // 請求量
}

// 磁碟使用狀況
type DiskUsage struct {
	Used      string            `json:"used"`      // 已使用空間
	Available string            `json:"available"` // 可用空間
	Total     string            `json:"total"`     // 總空間
	Volumes   map[string]Volume `json:"volumes"`   // 各個掛載點的使用狀況
}

// 磁碟卷資訊
type Volume struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	MountPath string `json:"mountPath"`
	Used      string `json:"used"`
	Available string `json:"available"`
	Total     string `json:"total"`
}

// 容器資源使用狀況
type ContainerUsage struct {
	Name   string      `json:"name"`
	CPU    CPUUsage    `json:"cpu"`
	Memory MemoryUsage `json:"memory"`
}

// Pod 詳細資訊 (包含基本資訊和資源使用狀況)
type PodDetails struct {
	Basic  Pod           `json:"basic"`
	Usage  ResourceUsage `json:"usage"`
	Events []Event       `json:"events"`
	Logs   string        `json:"logs"`
}

// Pod 事件
type Event struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// 搜尋條件
type SearchCriteria struct {
	Namespace     string            `json:"namespace"`
	LabelSelector string            `json:"labelSelector"`
	FieldSelector string            `json:"fieldSelector"`
	Status        string            `json:"status"`
	Labels        map[string]string `json:"labels"`
}
