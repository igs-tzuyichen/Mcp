# MCP GKE 監控查詢服務

## 簡介
這是一個基於 MCP (Model Context Protocol) 的 Google Kubernetes Engine (GKE) 監控查詢服務，支援透過各種條件查詢 Pod 資訊和資源使用狀況。本服務可作為 LLM/AI 助手的外部工具，讓 AI 可以查詢和監控 GKE 叢集中的 Pod 狀態。

## MCP 協議說明
MCP (Model Context Protocol) 是一套允許 AI 模型與外部工具互動的協議，本服務實現了 MCP 協議，提供以下優勢：
- 讓 AI 助手能夠查詢即時的 GKE 叢集監控數據
- 支援結構化參數和返回值
- 遵循 MCP 標準，易於與各種 AI 平台整合

## 功能特性
- `get_all_pods`: 獲取所有 Pod 列表
- `search_pods`: 根據條件搜尋 Pod（支援透過命名空間、標籤選擇器、欄位選擇器、狀態等搜尋）
- `get_pod_cpu_usage`: 取得 Pod 的 CPU 使用狀況
- `get_pod_memory_usage`: 取得 Pod 的記憶體使用狀況
- `get_pod_disk_usage`: 取得 Pod 的磁碟使用狀況
- `get_pod_details`: 取得 Pod 的詳細資訊（包含資源使用狀況、事件、日誌）

## 資源 (Resources)
本服務除了提供工具外，還提供以下 MCP 資源：

- **GKE 監控查詢指南** (`docs://gke/guide`)
  - 類型：靜態文檔資源
  - 格式：Markdown
  - 內容：完整的使用指南，包含功能說明、使用範例和注意事項
  - 用途：幫助 AI 模型理解如何正確使用本服務的工具

## 專案架構
```
mcp-gke-monitor/
│
├── main.go               # 程式入口點
├── go.mod                # Go 模組定義
├── go.sum                # Go 依賴清單（執行後生成）
├── README.md             # 專案說明文件
├── config.json           # 預設設定檔
│
├── config/               # 配置相關程式碼
│   └── config.go         # 配置載入和管理
│
├── gke/                  # GKE 核心功能
│   ├── handler.go        # GKE MCP 工具處理器
│   ├── model.go          # GKE 數據模型
│   └── service.go        # GKE 業務邏輯
│
├── logger/               # 日誌相關程式碼
│   └── logger.go         # 日誌功能實現
│
├── server/               # MCP 伺服器相關程式碼
│   ├── handler.go        # 伺服器處理器接口
│   └── server.go         # 伺服器建立與設定
│
└── internal/             # 內部資源
    └── docs/             # 文檔資源
        └── guide.md      # 使用指南
```

### 核心模組說明

#### config
負責載入和管理應用程式配置，支援從 JSON 配置檔案讀取設定。主要處理伺服器類型（stdio 或 SSE）、SSE 模式的 URL 與埠號設定，以及 GKE 相關配置（kubeconfig 路徑、命名空間、叢集名稱）。

#### gke
實現 GKE 監控的核心業務邏輯：
- `model.go`: 定義 Pod、資源使用狀況等數據結構
- `service.go`: 實現與 Kubernetes API 的交互，包含 Pod 查詢、資源監控等功能
- `handler.go`: 連接 MCP 工具與 GKE 服務，處理 MCP 請求

#### logger
提供應用程式日誌功能，記錄伺服器啟動、停止和各種操作的日誌。支援與 MCP 伺服器整合的日誌掛鉤機制。

#### server
負責 MCP 伺服器的建立、配置和啟動：
- `server.go`: 實現 MCP 伺服器的建立、工具註冊和資源註冊
- `handler.go`: 定義工具處理器接口

## 前置需求

### 1. 軟體需求
- Go 1.23.2 或更高版本
- 有效的 Kubernetes 叢集存取權限
- 已安裝 kubectl 並配置正確

### 2. GKE 叢集需求
- GKE 叢集已建立並運行
- Metrics Server 已安裝（GKE 預設會安裝）
- 適當的 RBAC 權限以存取 Pod 和 Metrics API

### 3. 權限設定
確保使用的 Service Account 或使用者帳戶具有以下權限：
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: gke-monitor-reader
rules:
- apiGroups: [""]
  resources: ["pods", "events"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get", "list"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods"]
  verbs: ["get", "list"]
```

## 安裝與設定

### 1. 克隆專案
```bash
git clone <repository-url>
cd mcp-gke-monitor
```

### 2. 安裝相依套件
```bash
go mod download
```

### 3. 配置設定
編輯 `config.json` 檔案以設定您的環境：

```json
{
  "serverType": "stdio",
  "sse": {
    "baseURL": "http://127.0.0.1",
    "port": 8080
  },
  "gke": {
    "kubeConfigPath": "",
    "namespace": "default",
    "clusterName": ""
  }
}
```

**配置說明**：
- `kubeConfigPath`: kubeconfig 檔案路徑，空字串表示使用預設路徑 (~/.kube/config)
- `namespace`: 預設命名空間
- `clusterName`: 叢集名稱，空字串表示使用當前上下文

### 4. 編譯程式
```bash
go build -o mcp-gke-monitor
```

## 使用方式

### 1. 啟動服務

#### stdio 模式（預設）
```bash
./mcp-gke-monitor
```

#### SSE 模式
修改 `config.json` 中的 `serverType` 為 `"sse"`，然後啟動：
```bash
./mcp-gke-monitor
```

### 2. 驗證連接
服務啟動後，您應該看到類似以下的輸出：
```
正在啟動 MCP GKE 監控查詢服務...
伺服器類型: stdio
MCP 伺服器初始化完成
使用 Stdio 模式
stdio 伺服器啟動
```

## MCP 工具使用說明

### 取得所有 Pod
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_all_pods",
    "arguments": {
      "namespace": "default"
    }
  }
}
```

### 搜尋特定標籤的 Pod
```json
{
  "method": "tools/call",
  "params": {
    "name": "search_pods",
    "arguments": {
      "namespace": "production",
      "labelSelector": "app=nginx",
      "status": "Running"
    }
  }
}
```

### 查詢 Pod CPU 使用狀況
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

### 查詢 Pod 詳細資訊
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

## 啟動模式配置

### stdio 模式配置
```json
{
  "serverType": "stdio",
  "gke": {
    "kubeConfigPath": "",
    "namespace": "default",
    "clusterName": ""
  }
}
```

### SSE 模式配置
```json
{
  "serverType": "sse",
  "sse": {
    "baseURL": "http://127.0.0.1",
    "port": 8080
  },
  "gke": {
    "kubeConfigPath": "",
    "namespace": "default",
    "clusterName": ""
  }
}
```

## 整合至 AI 平台

### MCP Host 配置範例
```json
{
  "mcpServers": {
    "mcp-gke-monitor-stdio": {
      "command": "cmd",
      "args": [
        "/c", 
        "C:\\path\\to\\mcp-gke-monitor\\mcp-gke-monitor.exe"
      ],
      "env": {}
    },
    "mcp-gke-monitor-sse": {
      "url": "http://127.0.0.1:8080/sse",
      "args": [],
      "env": {}
    } 
  }
}
```

## 疑難排解

### 1. 無法連接到 Kubernetes 叢集
**問題**: 服務無法連接到 GKE 叢集
**解決方案**:
- 確認 kubectl 命令能正常運作
- 檢查 kubeconfig 檔案路徑是否正確
- 驗證叢集憑證是否過期

### 2. Metrics API 不可用
**問題**: 無法取得資源使用狀況
**解決方案**:
- 確認 GKE 叢集的 Metrics Server 正在運行
- 檢查是否有權限存取 metrics.k8s.io API

### 3. 權限不足
**問題**: 無法存取 Pod 或命名空間
**解決方案**:
- 檢查 RBAC 權限設定
- 確認 Service Account 具有適當的角色綁定

### 4. 服務無法啟動
**問題**: 程式啟動失敗
**解決方案**:
- 檢查 config.json 格式是否正確
- 確認指定埠號未被佔用（SSE模式）
- 查看錯誤日誌以獲取詳細資訊

## 開發與貢獻

### 建構開發環境
```bash
# 克隆專案
git clone <repository-url>
cd mcp-gke-monitor

# 安裝相依套件
go mod download

# 執行測試
go test ./...

# 執行開發版本
go run main.go
```

### 程式碼結構
- 遵循 Go 語言慣例
- 使用 MCP 協議標準
- 支援並發安全的操作
- 包含完整的錯誤處理

## 版本資訊
- 當前版本: 0.0.1
- Go 版本: 1.23.2
- MCP-Go 版本: 0.20.1
- Kubernetes Client 版本: 0.31.1

## 授權條款
本專案採用 [授權條款] 授權。

## 支援與回饋
如有問題或建議，請透過以下方式聯繫：
- 建立 Issue 回報問題
- 提交 Pull Request 貢獻程式碼
- 聯繫專案維護者 