package model

// BaseResponse - 기본 응답 구조체 (Spring의 ResponseEntity와 유사)
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ConfigResponse - Config 조회 응답
type ConfigResponse struct {
	BaseResponse        // 익명 임베딩
	Data         string `json:"data"`
}

// ContextsResponse - Context 목록 응답
type ContextsResponse struct {
	BaseResponse               // 익명 임베딩
	Data         []ContextInfo `json:"data"`
}

// AddConfigRequest - Config 추가 요청 DTO
type AddConfigRequest struct {
	ClusterName string `json:"clusterName" binding:"required"` // 클러스터 이름
	Server      string `json:"server" binding:"required"`      // API 서버 주소
	ContextName string `json:"contextName" binding:"required"` // Context 이름
	User        string `json:"user" binding:"required"`        // 사용자 이름
	Token       string `json:"token"`                          // 인증 토큰 (선택사항)
	CertData    string `json:"certData"`                       // 인증서 데이터 (선택사항)
}

// UseContextRequest - Context 변경 요청 DTO
type UseContextRequest struct {
	ContextName string `json:"contextName" binding:"required"` // 사용할 Context 이름
}

// ContextInfo - Context 정보
type ContextInfo struct {
	Name      string `json:"name"`      // Context 이름
	IsCurrent bool   `json:"isCurrent"` // 현재 사용 중인지 여부
}

// KubeConfig - Kubernetes Config 구조체 (참고용)
type KubeConfig struct {
	APIVersion     string                 `yaml:"apiVersion"`
	Kind           string                 `yaml:"kind"`
	Clusters       []ClusterConfig        `yaml:"clusters"`
	Contexts       []ContextConfig        `yaml:"contexts"`
	Users          []UserConfig           `yaml:"users"`
	CurrentContext string                 `yaml:"current-context"`
	Preferences    map[string]interface{} `yaml:"preferences,omitempty"`
}

// ClusterConfig - 클러스터 설정
type ClusterConfig struct {
	Name    string            `yaml:"name"`
	Cluster ClusterConfigData `yaml:"cluster"`
}

// ClusterConfigData - 클러스터 설정 데이터
type ClusterConfigData struct {
	Server                   string `yaml:"server"`
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
	InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify,omitempty"`
}

// ContextConfig - Context 설정
type ContextConfig struct {
	Name    string            `yaml:"name"`
	Context ContextConfigData `yaml:"context"`
}

// ContextConfigData - Context 설정 데이터
type ContextConfigData struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}

// UserConfig - 사용자 설정
type UserConfig struct {
	Name string         `yaml:"name"`
	User UserConfigData `yaml:"user"`
}

// UserConfigData - 사용자 설정 데이터
type UserConfigData struct {
	Token                 string `yaml:"token,omitempty"`
	ClientCertificateData string `yaml:"client-certificate-data,omitempty"`
	ClientKeyData         string `yaml:"client-key-data,omitempty"`
}
