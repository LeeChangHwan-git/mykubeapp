package model

// AIYamlRequest - AI YAML 생성 요청
type AIYamlRequest struct {
	Prompt string `json:"prompt" binding:"required"` // AI에게 보낼 프롬프트
}

// AIYamlResponse - AI YAML 생성 응답
type AIYamlResponse struct {
	BaseResponse              // 익명 임베딩
	Data         AIYamlResult `json:"data"`
}

// AIYamlResult - AI YAML 생성 결과
type AIYamlResult struct {
	GeneratedYaml string `json:"generatedYaml"` // 생성된 YAML 내용
	Prompt        string `json:"prompt"`        // 원본 프롬프트
	GeneratedTime string `json:"generatedTime"` // 생성 시간
	Source        string `json:"source"`        // AI 모델 소스
}

// AIApplyRequest - AI YAML 생성 및 적용 요청
type AIApplyRequest struct {
	Prompt    string `json:"prompt" binding:"required"` // AI에게 보낼 프롬프트
	Namespace string `json:"namespace"`                 // 네임스페이스 (선택사항)
	DryRun    bool   `json:"dryRun"`                    // dry-run 모드 (선택사항)
}

// AIApplyResponse - AI YAML 생성 및 적용 응답
type AIApplyResponse struct {
	BaseResponse               // 익명 임베딩
	Data         AIApplyResult `json:"data"`
}

// AIApplyResult - AI YAML 생성 및 적용 결과
type AIApplyResult struct {
	GeneratedYaml string          `json:"generatedYaml"` // 생성된 YAML 내용
	ApplyResult   ApplyYamlResult `json:"applyResult"`   // 적용 결과
	Prompt        string          `json:"prompt"`        // 원본 프롬프트
	GeneratedTime string          `json:"generatedTime"` // 생성 시간
	Source        string          `json:"source"`        // AI 모델 소스
}

// AIQueryRequest - AI 질문 요청
type AIQueryRequest struct {
	Question string `json:"question" binding:"required"` // AI에게 할 질문
}

// AIQueryResponse - AI 질문 응답
type AIQueryResponse struct {
	BaseResponse               // 익명 임베딩
	Data         AIQueryResult `json:"data"`
}

// AIQueryResult - AI 질문 결과
type AIQueryResult struct {
	Question     string `json:"question"`     // 원본 질문
	Answer       string `json:"answer"`       // AI 답변
	Context      string `json:"context"`      // 현재 클러스터 컨텍스트
	AnsweredTime string `json:"answeredTime"` // 답변 시간
	Source       string `json:"source"`       // AI 모델 소스
}

// DeepSeek API 관련 구조체들

// DeepSeekRequest - DeepSeek API 요청
type DeepSeekRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Stream      bool              `json:"stream"`
}

// DeepSeekMessage - DeepSeek 메시지
type DeepSeekMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// DeepSeekResponse - DeepSeek API 응답
type DeepSeekResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []DeepSeekChoice `json:"choices"`
	Usage   DeepSeekUsage    `json:"usage"`
}

// DeepSeekChoice - DeepSeek 선택지
type DeepSeekChoice struct {
	Index        int             `json:"index"`
	Message      DeepSeekMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// DeepSeekUsage - DeepSeek 사용량 정보
type DeepSeekUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIHealthResponse - AI 서비스 상태 응답
type AIHealthResponse struct {
	BaseResponse          // 익명 임베딩
	Data         AIHealth `json:"data"`
}

// AIHealth - AI 서비스 상태
type AIHealth struct {
	DeepSeekURL     string   `json:"deepSeekUrl"`     // DeepSeek 서버 URL
	IsConnected     bool     `json:"isConnected"`     // 연결 상태
	LastChecked     string   `json:"lastChecked"`     // 마지막 확인 시간
	ResponseTime    string   `json:"responseTime"`    // 응답 시간
	AvailableModels []string `json:"availableModels"` // 사용 가능한 모델 목록
}

// AITemplateRequest - AI 템플릿 기반 생성 요청
type AITemplateRequest struct {
	TemplateType string                 `json:"templateType" binding:"required"` // "deployment", "service", "pod", "configmap" 등
	Parameters   map[string]interface{} `json:"parameters"`                      // 템플릿 파라미터
	Namespace    string                 `json:"namespace"`                       // 네임스페이스 (선택사항)
	DryRun       bool                   `json:"dryRun"`                          // dry-run 모드 (선택사항)
}

// AITemplateResponse - AI 템플릿 기반 생성 응답
type AITemplateResponse struct {
	BaseResponse                  // 익명 임베딩
	Data         AITemplateResult `json:"data"`
}

// AITemplateResult - AI 템플릿 기반 생성 결과
type AITemplateResult struct {
	TemplateType  string                 `json:"templateType"`          // 사용된 템플릿 타입
	Parameters    map[string]interface{} `json:"parameters"`            // 사용된 파라미터
	GeneratedYaml string                 `json:"generatedYaml"`         // 생성된 YAML
	ApplyResult   *ApplyYamlResult       `json:"applyResult,omitempty"` // 적용 결과 (적용한 경우)
	GeneratedTime string                 `json:"generatedTime"`         // 생성 시간
	Source        string                 `json:"source"`                // AI 모델 소스
}
