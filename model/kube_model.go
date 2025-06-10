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

// DeleteContextRequest - Context 삭제 요청 DTO
type DeleteContextRequest struct {
	ContextName string `json:"contextName" binding:"required"` // 삭제할 Context 이름
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

// ContextDetailResponse - Context 상세 정보 응답
type ContextDetailResponse struct {
	BaseResponse               // 익명 임베딩
	Data         ContextDetail `json:"data"`
}

// ContextDetail - Context 상세 정보
type ContextDetail struct {
	Name      string        `json:"name"`      // 컨텍스트 이름
	IsCurrent bool          `json:"isCurrent"` // 현재 사용 중인지 여부
	Cluster   ClusterDetail `json:"cluster"`   // 클러스터 정보
	User      UserDetail    `json:"user"`      // 사용자 정보
	Namespace string        `json:"namespace"` // 네임스페이스 (선택사항)
}

// ClusterDetail - 클러스터 상세 정보 (토큰 제외)
type ClusterDetail struct {
	Name                    string `json:"name"`                    // 클러스터 이름
	Server                  string `json:"server"`                  // API 서버 주소
	InsecureSkipTLSVerify   bool   `json:"insecureSkipTLSVerify"`   // TLS 검증 스킵 여부
	HasCertificateAuthority bool   `json:"hasCertificateAuthority"` // CA 인증서 존재 여부
}

// UserDetail - 사용자 상세 정보 (토큰 제외)
type UserDetail struct {
	Name                 string `json:"name"`                 // 사용자 이름
	HasToken             bool   `json:"hasToken"`             // 토큰 존재 여부 (값은 제외)
	HasClientCertificate bool   `json:"hasClientCertificate"` // 클라이언트 인증서 존재 여부
	HasClientKey         bool   `json:"hasClientKey"`         // 클라이언트 키 존재 여부
	AuthenticationMethod string `json:"authenticationMethod"` // 인증 방식 요약
}

// ApplyYamlRequest - YAML 적용 요청 DTO
type ApplyYamlRequest struct {
	YamlContent string `json:"yamlContent" binding:"required"` // YAML 내용
	Namespace   string `json:"namespace"`                      // 네임스페이스 (선택사항)
	DryRun      bool   `json:"dryRun"`                         // dry-run 모드 (선택사항)
}

// ApplyYamlResponse - YAML 적용 응답
type ApplyYamlResponse struct {
	BaseResponse                 // 익명 임베딩
	Data         ApplyYamlResult `json:"data"`
}

// ApplyYamlResult - YAML 적용 결과
type ApplyYamlResult struct {
	Output      string   `json:"output"`      // kubectl 명령 출력
	AppliedTime string   `json:"appliedTime"` // 적용 시간
	Resources   []string `json:"resources"`   // 적용된 리소스 목록
	DryRun      bool     `json:"dryRun"`      // dry-run 여부
}

// DeleteYamlRequest - YAML 삭제 요청 DTO
type DeleteYamlRequest struct {
	YamlContent string `json:"yamlContent" binding:"required"` // YAML 내용
	Namespace   string `json:"namespace"`                      // 네임스페이스 (선택사항)
}
