# GKE 監控查詢服務使用指南

## 簡介
本服務提供 Google Kubernetes Engine (GKE) 的 Pod 監控和查詢功能，支援透過 MCP (Model Context Protocol) 協議與 AI 助手整合。

## 主要功能

### 1. Pod 列表查詢
**工具名稱**: `get_all_pods`

**功能描述**: 取得指定命名空間中的所有 Pod 列表

**參數**:
- `namespace` (可選): 命名空間名稱，預設為 "default"

**使用範例**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_all_pods",
    "arguments": {
      "namespace": "production"
    }
  }
}
```

### 2. Pod 搜尋
**工具名稱**: `search_pods`

**功能描述**: 根據多種條件搜尋 Pod

**參數**:
- `namespace` (可選): 命名空間名稱
- `labelSelector` (可選): 標籤選擇器 (例如: "app=nginx")
- `fieldSelector` (可選): 欄位選擇器 (例如: "status.phase=Running")
- `status` (可選): Pod 狀態 (Running, Pending, Succeeded, Failed, Unknown)

**使用範例**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "search_pods",
    "arguments": {
      "namespace": "default",
      "labelSelector": "app=nginx",
      "status": "Running"
    }
  }
}
```

### 3. Pod CPU 使用狀況
**工具名稱**: `get_pod_cpu_usage`

**功能描述**: 查詢特定 Pod 的 CPU 使用狀況

**參數**:
- `podName` (必要): Pod 名稱
- `namespace` (可選): 命名空間名稱，預設為 "default"

**使用範例**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_pod_cpu_usage",
    "arguments": {
      "podName": "nginx-deployment-7d5b6c4f8d-abc123",
      "namespace": "default"
    }
  }
}
```

### 4. Pod 記憶體使用狀況
**工具名稱**: `get_pod_memory_usage`

**功能描述**: 查詢特定 Pod 的記憶體使用狀況

**參數**:
- `podName` (必要): Pod 名稱
- `namespace` (可選): 命名空間名稱，預設為 "default"

**使用範例**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_pod_memory_usage",
    "arguments": {
      "podName": "nginx-deployment-7d5b6c4f8d-abc123",
      "namespace": "default"
    }
  }
}
```

### 5. Pod 磁碟使用狀況
**工具名稱**: `get_pod_disk_usage`

**功能描述**: 查詢特定 Pod 的磁碟使用狀況

**參數**:
- `podName` (必要): Pod 名稱
- `namespace` (可選): 命名空間名稱，預設為 "default"

**使用範例**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_pod_disk_usage",
    "arguments": {
      "podName": "nginx-deployment-7d5b6c4f8d-abc123",
      "namespace": "default"
    }
  }
}
```

### 6. Pod 詳細資訊
**工具名稱**: `get_pod_details`

**功能描述**: 取得 Pod 的完整詳細資訊，包含基本資訊、資源使用狀況、事件和日誌

**參數**:
- `podName` (必要): Pod 名稱
- `namespace` (可選): 命名空間名稱，預設為 "default"

**使用範例**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_pod_details",
    "arguments": {
      "podName": "nginx-deployment-7d5b6c4f8d-abc123",
      "namespace": "default"
    }
  }
}
```

## 回應格式

### Pod 基本資訊
```json
{
  "name": "nginx-deployment-7d5b6c4f8d-abc123",
  "namespace": "default",
  "status": "Running",
  "nodeName": "gke-cluster-1-default-pool-12345678-abcd",
  "podIP": "10.244.1.5",
  "hostIP": "10.128.0.2",
  "labels": {
    "app": "nginx",
    "version": "1.0"
  },
  "createdAt": "2024-01-15T10:30:00Z",
  "ready": true,
  "containers": [
    {
      "name": "nginx",
      "image": "nginx:1.21",
      "status": "Running",
      "ready": true,
      "restartCount": 0
    }
  ]
}
```

### 資源使用狀況
```json
{
  "podName": "nginx-deployment-7d5b6c4f8d-abc123",
  "namespace": "default",
  "cpu": {
    "current": "50m",
    "percentage": 25.0,
    "limit": "200m",
    "request": "100m"
  },
  "memory": {
    "current": "128Mi",
    "percentage": 25.6,
    "limit": "500Mi",
    "request": "256Mi"
  },
  "disk": {
    "used": "500Mi",
    "available": "1.5Gi",
    "total": "2Gi",
    "volumes": {
      "data-volume": {
        "name": "data-volume",
        "type": "PVC",
        "mountPath": "/data",
        "used": "100Mi",
        "available": "900Mi",
        "total": "1Gi"
      }
    }
  },
  "timestamp": "2024-01-15 10:35:00"
}
```

## 使用注意事項

### 1. 權限要求
- 服務需要適當的 Kubernetes RBAC 權限來存取 Pod 資訊
- 需要 Metrics Server 安裝和運行以取得資源使用資料

### 2. 配置需求
- 確保 kubeconfig 檔案配置正確
- 設定適當的命名空間存取權限

### 3. 效能考量
- 大量 Pod 查詢可能需要較長時間
- 建議使用適當的過濾條件來限制結果數量

### 4. 錯誤處理
- 當 Metrics API 不可用時，資源使用資料將無法取得
- 請檢查 GKE 叢集的 Metrics Server 狀態

## 常見使用案例

### 1. 監控特定應用的 Pod 狀態
```json
{
  "name": "search_pods",
  "arguments": {
    "labelSelector": "app=my-application",
    "status": "Running"
  }
}
```

### 2. 檢查高 CPU 使用的 Pod
先取得所有 Pod，然後逐一檢查 CPU 使用狀況：
```json
{
  "name": "get_pod_cpu_usage",
  "arguments": {
    "podName": "suspicious-pod-name"
  }
}
```

### 3. 診斷 Pod 問題
取得完整的 Pod 詳細資訊，包含事件和日誌：
```json
{
  "name": "get_pod_details",
  "arguments": {
    "podName": "problematic-pod-name"
  }
}
```

## 支援的 Kubernetes 版本
- Kubernetes 1.20+
- GKE Standard 和 Autopilot 模式

## 相依套件
- `k8s.io/client-go`: Kubernetes 客戶端函式庫
- `k8s.io/metrics`: Metrics API 客戶端
- `github.com/mark3labs/mcp-go`: MCP 協議實現 