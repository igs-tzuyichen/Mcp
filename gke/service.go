package gke

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"

	// Google Cloud 相关导入
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

// Logger 接口，用於可選的日誌記錄
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// Service GKE 服務
type Service struct {
	clientset        *kubernetes.Clientset
	metricsClientset *metricsclientset.Clientset
	mu               sync.RWMutex
	defaultNamespace string
	config           ServiceConfig
	logger           Logger // 可選的 logger
}

// ServiceConfig GKE 服務配置
type ServiceConfig struct {
	UseCredentials   bool
	CredentialsFile  string
	ProjectID        string
	ClusterName      string
	Location         string
	DefaultNamespace string
	Logger           Logger // 可選的 logger
}

// NewService 創建一個新的 GKE 服務
func NewService() (*Service, error) {
	// 為了向後兼容，創建一個默認配置
	config := ServiceConfig{
		UseCredentials:   false,
		DefaultNamespace: "default",
	}
	return NewServiceWithConfig(config)
}

// NewServiceWithConfig 使用配置創建一個新的 GKE 服務
func NewServiceWithConfig(config ServiceConfig) (*Service, error) {
	// 取得 Kubernetes 配置
	kubeConfig, err := getKubeConfigWithCredentials(config)
	if err != nil {
		return nil, fmt.Errorf("無法取得 Kubernetes 配置: %w", err)
	}

	// 建立 Kubernetes 客戶端
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("無法建立 Kubernetes 客戶端: %w", err)
	}

	// 建立 Metrics 客戶端
	metricsClientset, err := metricsclientset.NewForConfig(kubeConfig)
	if err != nil {
		if config.Logger != nil {
			config.Logger.Printf("警告: 無法建立 Metrics 客戶端: %v", err)
		}
		// 繼續執行，但 metrics 功能將不可用
	}

	namespace := config.DefaultNamespace
	if namespace == "" {
		namespace = "default"
	}

	service := &Service{
		clientset:        clientset,
		metricsClientset: metricsClientset,
		defaultNamespace: namespace,
		config:           config,
		logger:           config.Logger,
	}

	// 驗證連接
	if err := service.validateConnection(); err != nil {
		return nil, fmt.Errorf("無法驗證 GKE 連接: %w", err)
	}

	return service, nil
}

// validateConnection 驗證 GKE 連接
func (s *Service) validateConnection() error {
	// 嘗試獲取命名空間列表來驗證連接
	_, err := s.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("連接驗證失敗: %w", err)
	}
	if s.logger != nil {
		s.logger.Println("GKE 連接驗證成功")
	}
	return nil
}

// getKubeConfigWithCredentials 使用凭证取得 Kubernetes 配置
func getKubeConfigWithCredentials(config ServiceConfig) (*rest.Config, error) {
	if config.UseCredentials && config.CredentialsFile != "" {
		return getKubeConfigFromGoogleCredentials(config)
	}
	return getKubeConfig()
}

// getKubeConfigFromGoogleCredentials 從 Google Cloud 凭证建立 Kubernetes 配置
func getKubeConfigFromGoogleCredentials(config ServiceConfig) (*rest.Config, error) {
	// 讀取凭证文件
	credentialsBytes, err := os.ReadFile(config.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("無法讀取凭证文件: %w", err)
	}

	// 解析凭证
	var credentials map[string]interface{}
	if err := json.Unmarshal(credentialsBytes, &credentials); err != nil {
		return nil, fmt.Errorf("無法解析凭证文件: %w", err)
	}

	// 建立 Google 凭证
	googleCredentials, err := google.CredentialsFromJSON(context.Background(), credentialsBytes, container.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("無法建立 Google 凭证: %w", err)
	}

	// 建立 Container 服務客戶端
	containerService, err := container.NewService(context.Background(), option.WithCredentials(googleCredentials))
	if err != nil {
		return nil, fmt.Errorf("無法建立 Container 服務: %w", err)
	}

	// 取得集群資訊
	clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", config.ProjectID, config.Location, config.ClusterName)
	cluster, err := containerService.Projects.Locations.Clusters.Get(clusterPath).Do()
	if err != nil {
		return nil, fmt.Errorf("無法取得集群資訊: %w", err)
	}

	// 解碼 CA 証書 (base64 解碼)
	caCertData, err := base64.StdEncoding.DecodeString(cluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return nil, fmt.Errorf("無法解碼 CA 證書: %w", err)
	}

	// 建立 Kubernetes REST 配置
	kubeConfig := &rest.Config{
		Host: fmt.Sprintf("https://%s", cluster.Endpoint),
		TLSClientConfig: rest.TLSClientConfig{
			CAData: caCertData,
		},
	}

	// 設定 Google 認證
	tokenSource := googleCredentials.TokenSource
	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("無法取得認證令牌: %w", err)
	}

	kubeConfig.BearerToken = token.AccessToken

	// 設定令牌刷新
	kubeConfig.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return &tokenRefreshTransport{
			base:        rt,
			tokenSource: tokenSource,
		}
	})

	if config.Logger != nil {
		config.Logger.Printf("使用 Google Cloud 凭证成功建立 GKE 連接")
		config.Logger.Printf("集群端點: %s", cluster.Endpoint)
		config.Logger.Printf("集群狀態: %s", cluster.Status)
	}

	return kubeConfig, nil
}

// tokenRefreshTransport 自動刷新令牌的傳輸層
type tokenRefreshTransport struct {
	base        http.RoundTripper
	tokenSource oauth2.TokenSource
}

func (t *tokenRefreshTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("無法刷新令牌: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	return t.base.RoundTrip(req)
}

// getKubeConfig 取得 Kubernetes 配置 (原有的方法，用於向後兼容)
func getKubeConfig() (*rest.Config, error) {
	// 嘗試使用 in-cluster 配置
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// 如果不在叢集內，使用 kubeconfig 檔案
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("無法載入 kubeconfig: %w", err)
	}

	return config, nil
}

// GetAllPods 取得所有 Pod
func (s *Service) GetAllPods(namespace string) ([]Pod, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespace == "" {
		namespace = s.defaultNamespace
	}

	pods, err := s.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("無法取得 Pod 列表: %w", err)
	}

	var result []Pod
	for _, pod := range pods.Items {
		result = append(result, s.convertPod(&pod))
	}

	return result, nil
}

// SearchPods 根據條件搜尋 Pod
func (s *Service) SearchPods(criteria SearchCriteria) ([]Pod, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	namespace := criteria.Namespace
	if namespace == "" {
		namespace = s.defaultNamespace
	}

	listOptions := metav1.ListOptions{}

	// 設定標籤選擇器
	if criteria.LabelSelector != "" {
		listOptions.LabelSelector = criteria.LabelSelector
	}

	// 設定欄位選擇器
	if criteria.FieldSelector != "" {
		listOptions.FieldSelector = criteria.FieldSelector
	}

	pods, err := s.clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("無法搜尋 Pod: %w", err)
	}

	var result []Pod
	for _, pod := range pods.Items {
		convertedPod := s.convertPod(&pod)

		// 額外過濾條件
		if criteria.Status != "" && convertedPod.Status != criteria.Status {
			continue
		}

		result = append(result, convertedPod)
	}

	return result, nil
}

// GetPodResourceUsage 取得 Pod 的資源使用狀況
func (s *Service) GetPodResourceUsage(podName, namespace string) (*ResourceUsage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespace == "" {
		namespace = s.defaultNamespace
	}

	if s.metricsClientset == nil {
		return nil, fmt.Errorf("Metrics API 不可用")
	}

	// 取得 Pod metrics
	podMetrics, err := s.metricsClientset.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("無法取得 Pod metrics: %w", err)
	}

	// 取得 Pod 資訊以獲取資源限制和請求
	pod, err := s.clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("無法取得 Pod 資訊: %w", err)
	}

	usage := &ResourceUsage{
		PodName:   podName,
		Namespace: namespace,
		Timestamp: time.Now(),
	}

	// 計算總的 CPU 和記憶體使用量
	totalCPU := int64(0)
	totalMemory := int64(0)
	var containerUsages []ContainerUsage

	for _, container := range podMetrics.Containers {
		cpu := container.Usage.Cpu().MilliValue()
		memory := container.Usage.Memory().Value()

		totalCPU += cpu
		totalMemory += memory

		// 找到對應的容器規格
		var containerSpec *corev1.Container
		for _, spec := range pod.Spec.Containers {
			if spec.Name == container.Name {
				containerSpec = &spec
				break
			}
		}

		containerUsage := ContainerUsage{
			Name: container.Name,
			CPU: CPUUsage{
				Current: fmt.Sprintf("%dm", cpu),
			},
			Memory: MemoryUsage{
				Current: fmt.Sprintf("%dMi", memory/(1024*1024)),
			},
		}

		if containerSpec != nil {
			// CPU 限制和請求
			if cpuLimit := containerSpec.Resources.Limits.Cpu(); cpuLimit != nil {
				containerUsage.CPU.Limit = cpuLimit.String()
				if cpuLimit.MilliValue() > 0 {
					containerUsage.CPU.Percentage = float64(cpu) / float64(cpuLimit.MilliValue()) * 100
				}
			}
			if cpuRequest := containerSpec.Resources.Requests.Cpu(); cpuRequest != nil {
				containerUsage.CPU.Request = cpuRequest.String()
			}

			// 記憶體限制和請求
			if memLimit := containerSpec.Resources.Limits.Memory(); memLimit != nil {
				containerUsage.Memory.Limit = memLimit.String()
				if memLimit.Value() > 0 {
					containerUsage.Memory.Percentage = float64(memory) / float64(memLimit.Value()) * 100
				}
			}
			if memRequest := containerSpec.Resources.Requests.Memory(); memRequest != nil {
				containerUsage.Memory.Request = memRequest.String()
			}
		}

		containerUsages = append(containerUsages, containerUsage)
	}

	// 設定總體使用量
	usage.CPU = CPUUsage{
		Current: fmt.Sprintf("%dm", totalCPU),
	}
	usage.Memory = MemoryUsage{
		Current: fmt.Sprintf("%dMi", totalMemory/(1024*1024)),
	}
	usage.Containers = containerUsages

	// 取得磁碟使用狀況 (模擬資料，實際需要額外的監控工具)
	usage.Disk = s.getMockDiskUsage(pod)

	return usage, nil
}

// GetPodDetails 取得 Pod 的詳細資訊
func (s *Service) GetPodDetails(podName, namespace string) (*PodDetails, error) {
	if namespace == "" {
		namespace = s.defaultNamespace
	}

	// 取得基本資訊
	pod, err := s.clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("無法取得 Pod 資訊: %w", err)
	}

	// 取得資源使用狀況
	usage, err := s.GetPodResourceUsage(podName, namespace)
	if err != nil {
		if s.logger != nil {
			s.logger.Printf("警告: 無法取得資源使用狀況: %v", err)
		}
		// 建立一個空的使用狀況
		usage = &ResourceUsage{
			PodName:   podName,
			Namespace: namespace,
			Timestamp: time.Now(),
		}
	}

	// 取得事件
	events, err := s.getPodEvents(podName, namespace)
	if err != nil {
		if s.logger != nil {
			s.logger.Printf("警告: 無法取得 Pod 事件: %v", err)
		}
		events = []Event{}
	}

	// 取得日誌 (最新 100 行)
	logs, err := s.getPodLogs(podName, namespace, 100)
	if err != nil {
		if s.logger != nil {
			s.logger.Printf("警告: 無法取得 Pod 日誌: %v", err)
		}
		logs = "無法取得日誌"
	}

	details := &PodDetails{
		Basic:  s.convertPod(pod),
		Usage:  *usage,
		Events: events,
		Logs:   logs,
	}

	return details, nil
}

// convertPod 轉換 Kubernetes Pod 為內部 Pod 結構
func (s *Service) convertPod(pod *corev1.Pod) Pod {
	var containers []Container
	ready := true

	for _, container := range pod.Spec.Containers {
		containerStatus := s.getContainerStatus(pod, container.Name)
		containerReady := containerStatus != nil && containerStatus.Ready
		if !containerReady {
			ready = false
		}

		containers = append(containers, Container{
			Name:    container.Name,
			Image:   container.Image,
			Status:  s.getContainerStatusString(containerStatus),
			Ready:   containerReady,
			Restart: s.getContainerRestartCount(containerStatus),
		})
	}

	return Pod{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Status:     string(pod.Status.Phase),
		NodeName:   pod.Spec.NodeName,
		PodIP:      pod.Status.PodIP,
		HostIP:     pod.Status.HostIP,
		Labels:     pod.Labels,
		CreatedAt:  pod.CreationTimestamp.Time,
		Ready:      ready,
		Containers: containers,
	}
}

// getContainerStatus 取得容器狀態
func (s *Service) getContainerStatus(pod *corev1.Pod, containerName string) *corev1.ContainerStatus {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
			return &status
		}
	}
	return nil
}

// getContainerStatusString 取得容器狀態字串
func (s *Service) getContainerStatusString(status *corev1.ContainerStatus) string {
	if status == nil {
		return "Unknown"
	}
	if status.State.Running != nil {
		return "Running"
	}
	if status.State.Waiting != nil {
		return "Waiting"
	}
	if status.State.Terminated != nil {
		return "Terminated"
	}
	return "Unknown"
}

// getContainerRestartCount 取得容器重啟次數
func (s *Service) getContainerRestartCount(status *corev1.ContainerStatus) int32 {
	if status == nil {
		return 0
	}
	return status.RestartCount
}

// getPodEvents 取得 Pod 事件
func (s *Service) getPodEvents(podName, namespace string) ([]Event, error) {
	fieldSelector := fields.OneTermEqualSelector("involvedObject.name", podName).String()
	events, err := s.clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	var result []Event
	for _, event := range events.Items {
		result = append(result, Event{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Timestamp: event.FirstTimestamp.Time,
			Source:    event.Source.Component,
		})
	}

	return result, nil
}

// getPodLogs 取得 Pod 日誌
func (s *Service) getPodLogs(podName, namespace string, tailLines int) (string, error) {
	tailLines64 := int64(tailLines)
	req := s.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: &tailLines64,
	})

	logs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf := make([]byte, 1024*1024) // 1MB buffer
	n, err := logs.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return "", err
	}

	return string(buf[:n]), nil
}

// getMockDiskUsage 取得模擬的磁碟使用狀況 (實際需要額外的監控工具)
func (s *Service) getMockDiskUsage(pod *corev1.Pod) DiskUsage {
	volumes := make(map[string]Volume)

	// 模擬一些基本的磁碟使用資訊
	for _, volume := range pod.Spec.Volumes {
		volumes[volume.Name] = Volume{
			Name:      volume.Name,
			Type:      s.getVolumeType(&volume),
			MountPath: "/data", // 模擬掛載路徑
			Used:      "100Mi",
			Available: "900Mi",
			Total:     "1Gi",
		}
	}

	return DiskUsage{
		Used:      "500Mi",
		Available: "1.5Gi",
		Total:     "2Gi",
		Volumes:   volumes,
	}
}

// getVolumeType 取得卷類型
func (s *Service) getVolumeType(volume *corev1.Volume) string {
	switch {
	case volume.EmptyDir != nil:
		return "EmptyDir"
	case volume.PersistentVolumeClaim != nil:
		return "PVC"
	case volume.ConfigMap != nil:
		return "ConfigMap"
	case volume.Secret != nil:
		return "Secret"
	default:
		return "Unknown"
	}
}
