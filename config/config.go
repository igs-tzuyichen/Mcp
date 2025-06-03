package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServerType string

const (
	ServerTypeStdio ServerType = "stdio"
	ServerTypeSSE   ServerType = "sse"
)

// GkeCredentials Google Cloud服务账号凭证配置
type GkeCredentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
	GkeClusterName          string `json:"gke_cluster_name"`
	GkeLocation             string `json:"gke_location"`
}

type GKEConfig struct {
	KubeConfigPath  string `json:"kubeConfigPath"`
	Namespace       string `json:"namespace"`
	ClusterName     string `json:"clusterName"`
	CredentialsFile string `json:"credentialsFile"`
}

type Config struct {
	ServerType ServerType `json:"serverType"`
	SSE        struct {
		BaseURL string      `json:"baseURL"`
		Port    interface{} `json:"port"`
	} `json:"sse"`
	GKE         GKEConfig       `json:"gke"`
	Credentials *GkeCredentials `json:"-"` // 不序列化到JSON
}

func DefaultConfig() Config {
	cfg := Config{
		ServerType: ServerTypeStdio,
	}
	cfg.SSE.BaseURL = "http://127.0.0.1"
	cfg.SSE.Port = 8080
	cfg.GKE.KubeConfigPath = ""                    // 空字串表示使用預設路徑
	cfg.GKE.Namespace = "default"                  // 預設命名空間
	cfg.GKE.ClusterName = ""                       // 空字串表示使用當前上下文
	cfg.GKE.CredentialsFile = "irich-h5-test.json" // 預設凭证文件
	return cfg
}

// LoadGkeCredentials 從文件加載 GKE 凭证
func LoadGkeCredentials(filePath string) (*GkeCredentials, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("讀取 GKE 凭证文件失敗: %w", err)
	}

	var credentials GkeCredentials
	if err := json.Unmarshal(data, &credentials); err != nil {
		return nil, fmt.Errorf("解析 GKE 凭证文件失敗: %w", err)
	}

	return &credentials, nil
}

func LoadFromFile(filePath string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return cfg, fmt.Errorf("讀取配置檔案失敗: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("解析配置檔案失敗: %w", err)
	}

	return cfg, nil
}

func LoadConfig() (Config, error) {
	configPath := "config.json"

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		cfg = DefaultConfig()
	}

	// 加載 GKE 凭证
	if cfg.GKE.CredentialsFile != "" {
		credentials, err := LoadGkeCredentials(cfg.GKE.CredentialsFile)
		if err != nil {
			return cfg, fmt.Errorf("無法載入 GKE 凭证: %w", err)
		}
		cfg.Credentials = credentials
	}

	return cfg, nil
}
