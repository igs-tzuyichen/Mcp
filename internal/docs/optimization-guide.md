# GKE 優化建議服務使用指南

## 📊 **功能概述**

GKE 優化建議服務是 MCP GKE 監控系統的核心功能，專門分析 GKE 叢集中的資源使用狀況，並提供專業的優化建議。

## 🎯 **主要功能**

### 1. **完整優化報告** (`generate_optimization_report`)
生成包含所有 Pod 分析、資源浪費、優化建議的完整報告。

**使用範例**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "generate_optimization_report",
    "arguments": {
      "namespace": "production"
    }
  }
}
```

### 2. **優化摘要** (`get_optimization_summary`)
快速取得關鍵優化指標和主要問題。

**回應格式**:
```json
{
  "clusterName": "GKE-Cluster",
  "namespace": "default",
  "summary": {
    "totalPods": 15,
    "podsNeedingOptimization": 8,
    "potentialCPUSavings": "1200m",
    "potentialMemorySavings": "2.5Gi",
    "overallScore": 65.2
  },
  "topIssues": [
    "CPU: 5 個高優先級問題",
    "MEMORY: 3 個高優先級問題"
  ]
}
```

### 3. **優化建議** (`get_optimization_recommendations`)
取得具體的優化建議，支援篩選。

**參數**:
- `namespace`: 命名空間
- `priority`: 優先級 (HIGH, MEDIUM, LOW)
- `type`: 建議類型 (CPU, MEMORY, HEALTH, STORAGE, REPLICA, SECURITY)

**使用範例**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_optimization_recommendations",
    "arguments": {
      "namespace": "default",
      "priority": "HIGH",
      "type": "CPU"
    }
  }
}
```

### 4. **資源浪費分析** (`get_resource_waste_analysis`)
詳細分析過度配置和閒置資源。

**分析內容**:
- 過度配置的 Pod
- 使用率不足的 Pod
- 完全閒置的 Pod
- 總體浪費統計

### 5. **單 Pod 優化分析** (`get_pod_optimization_analysis`)
針對特定 Pod 的詳細優化分析。

**使用範例**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_pod_optimization_analysis",
    "arguments": {
      "podName": "nginx-deployment-7d5b6c4f8d-abc123",
      "namespace": "default"
    }
  }
}
```

### 6. **優化標準管理**
- `get_optimization_criteria`: 取得當前優化判斷標準
- `update_optimization_criteria`: 更新優化標準

## 🔧 **優化標準說明**

### 預設標準
- **CPU 閾值**: 20% (使用率低於此值視為過度配置)
- **記憶體閾值**: 30% (使用率低於此值視為過度配置)
- **健康閾值**: 5 次 (重啟次數超過此值視為不健康)
- **閒置閾值**: 5% (使用率低於此值視為閒置)

### 調整標準範例
```json
{
  "method": "tools/call",
  "params": {
    "name": "update_optimization_criteria",
    "arguments": {
      "cpuThreshold": 15.0,
      "memoryThreshold": 25.0,
      "healthThreshold": 3,
      "idleThreshold": 3.0
    }
  }
}
```

## 📈 **分析維度**

### 1. **資源分析**
- CPU 使用率和配置狀況
- 記憶體使用率和配置狀況
- 磁碟使用狀況
- 資源請求與限制比較

### 2. **健康分析**
- Pod 就緒狀態
- 容器重啟次數
- 整體健康分數
- 健康問題識別

### 3. **優化分數計算**
- 基準分數: 100 分
- 高優先級問題: -20 分
- 中優先級問題: -10 分
- 低優先級問題: -5 分
- 結合健康分數計算最終評分

## 🎛️ **建議類型分類**

### CPU 優化
- **過度配置**: CPU 使用率過低
- **資源不足**: CPU 使用率過高
- **建議**: 調整 CPU requests 和 limits

### 記憶體優化
- **過度配置**: 記憶體使用率過低
- **資源不足**: 記憶體使用率過高
- **建議**: 調整記憶體 requests 和 limits

### 健康優化
- **重啟問題**: 容器重啟次數過多
- **就緒問題**: Pod 未就緒
- **建議**: 檢查應用程式和健康檢查

### 安全優化
- 安全配置建議
- 權限最小化
- 網路安全

## 📊 **實用場景**

### 1. **日常監控**
```json
{
  "name": "get_optimization_summary",
  "arguments": {"namespace": "production"}
}
```

### 2. **問題排查**
```json
{
  "name": "get_optimization_recommendations",
  "arguments": {
    "namespace": "production",
    "priority": "HIGH"
  }
}
```

### 3. **成本優化**
```json
{
  "name": "get_resource_waste_analysis",
  "arguments": {"namespace": "production"}
}
```

### 4. **單 Pod 診斷**
```json
{
  "name": "get_pod_optimization_analysis",
  "arguments": {
    "podName": "problematic-pod",
    "namespace": "production"
  }
}
```

## 🔄 **最佳實踐**

### 1. **定期檢查**
- 每週生成優化報告
- 監控整體優化分數趨勢
- 關注高優先級建議

### 2. **分階段優化**
- 先處理高優先級問題
- 逐步調整資源配置
- 驗證優化效果

### 3. **標準調整**
- 根據業務需求調整閾值
- 考慮應用特性設定標準
- 定期檢視和更新標準

### 4. **成本控制**
- 重點關注資源浪費分析
- 識別並處理閒置資源
- 優化過度配置的 Pod

## ⚠️ **注意事項**

1. **Metrics 依賴**: 需要 Metrics Server 正常運行
2. **權限要求**: 需要適當的 RBAC 權限
3. **資料準確性**: 建議在穩定運行期間進行分析
4. **漸進式優化**: 避免一次性大幅調整資源

## 🚀 **與原有功能整合**

優化建議服務與原有的 Pod 監控功能完美整合：

1. **數據來源**: 使用相同的 GKE API 和 Metrics API
2. **工具協同**: 可結合 `get_pod_details` 進行深度分析
3. **統一介面**: 通過相同的 MCP 協議提供服務
4. **一致體驗**: 保持相同的使用方式和回應格式 