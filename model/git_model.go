package model

// GitYamlRequest - Git 레포지토리에서 YAML 가져오기 요청
type GitYamlRequest struct {
	RepoURL  string `json:"repoUrl" binding:"required"` // Git 레포지토리 URL
	Branch   string `json:"branch"`                     // 브랜치 (선택사항, 기본값: main/master)
	Filename string `json:"filename"`                   // 특정 파일명 (선택사항)
}

// GitApplyRequest - Git 레포지토리에서 YAML 가져와서 적용 요청
type GitApplyRequest struct {
	RepoURL   string `json:"repoUrl" binding:"required"` // Git 레포지토리 URL
	Branch    string `json:"branch"`                     // 브랜치 (선택사항)
	Filename  string `json:"filename"`                   // 특정 파일명 (선택사항, 없으면 모든 YAML)
	Namespace string `json:"namespace"`                  // 네임스페이스 (선택사항)
	DryRun    bool   `json:"dryRun"`                     // dry-run 모드 (선택사항)
}

// GitYamlResponse - Git YAML 조회 응답
type GitYamlResponse struct {
	BaseResponse             // 익명 임베딩
	Data         GitYamlData `json:"data"`
}

// GitYamlData - Git YAML 조회 결과
type GitYamlData struct {
	RepoURL     string        `json:"repoUrl"`     // 레포지토리 URL
	Branch      string        `json:"branch"`      // 사용된 브랜치
	YamlFiles   []GitYamlFile `json:"yamlFiles"`   // 발견된 YAML 파일들
	TotalFiles  int           `json:"totalFiles"`  // 총 파일 수
	RetrievedAt string        `json:"retrievedAt"` // 조회 시간
}

// GitYamlFile - Git에서 가져온 YAML 파일 정보
type GitYamlFile struct {
	Path         string `json:"path"`         // 상대 경로
	FullPath     string `json:"fullPath"`     // 전체 경로 (서버 내부용)
	Content      string `json:"content"`      // 파일 내용
	Size         int64  `json:"size"`         // 파일 크기 (bytes)
	IsKubernetes bool   `json:"isKubernetes"` // Kubernetes YAML인지 여부
}

// GitApplyResponse - Git YAML 적용 응답
type GitApplyResponse struct {
	BaseResponse              // 익명 임베딩
	Data         GitApplyData `json:"data"`
}

// GitApplyData - Git YAML 적용 결과
type GitApplyData struct {
	RepoURL     string         `json:"repoUrl"`     // 레포지토리 URL
	Branch      string         `json:"branch"`      // 사용된 브랜치
	ApplyResult GitApplyResult `json:"applyResult"` // 적용 결과
	RetrievedAt string         `json:"retrievedAt"` // 조회 시간
}

// GitApplyResult - Git에서 가져온 YAML들의 적용 결과
type GitApplyResult struct {
	TotalFiles   int                  `json:"totalFiles"`   // 총 파일 수
	SuccessFiles int                  `json:"successFiles"` // 성공한 파일 수
	FailedFiles  int                  `json:"failedFiles"`  // 실패한 파일 수
	AppliedTime  string               `json:"appliedTime"`  // 적용 시간
	Results      []GitFileApplyResult `json:"results"`      // 각 파일별 적용 결과
	AllResources []string             `json:"allResources"` // 모든 적용된 리소스 목록
	DryRun       bool                 `json:"dryRun"`       // dry-run 여부
}

// GitFileApplyResult - 개별 파일 적용 결과
type GitFileApplyResult struct {
	FilePath  string   `json:"filePath"`  // 파일 경로
	Success   bool     `json:"success"`   // 성공 여부
	Output    string   `json:"output"`    // kubectl 출력
	Resources []string `json:"resources"` // 적용된 리소스 목록
	Error     string   `json:"error"`     // 에러 메시지 (실패시)
}

// AIGitRequest - AI를 통한 Git 연동 요청
type AIGitRequest struct {
	Prompt string `json:"prompt" binding:"required"` // AI 프롬프트 (예: "xx레포지토리에서 aa.yaml 적용시켜줘")
}

// AIGitResponse - AI를 통한 Git 연동 응답
type AIGitResponse struct {
	BaseResponse           // 익명 임베딩
	Data         AIGitData `json:"data"`
}

// AIGitData - AI Git 연동 결과
type AIGitData struct {
	ParsedRequest   GitParseResult `json:"parsedRequest"`   // AI가 파싱한 요청 내용
	RepoURL         string         `json:"repoUrl"`         // 파싱된 레포지토리 URL
	Branch          string         `json:"branch"`          // 파싱된 브랜치
	Filename        string         `json:"filename"`        // 파싱된 파일명
	Action          string         `json:"action"`          // 수행할 액션 (apply, show, list)
	ExecutionResult interface{}    `json:"executionResult"` // 실행 결과 (GitYamlData 또는 GitApplyData)
	ProcessedTime   string         `json:"processedTime"`   // 처리 시간
}

// GitParseResult - AI가 파싱한 Git 요청 결과
type GitParseResult struct {
	RepoURL      string  `json:"repoUrl"`      // 추출된 레포지토리 URL
	Branch       string  `json:"branch"`       // 추출된 브랜치
	Filename     string  `json:"filename"`     // 추출된 파일명
	Action       string  `json:"action"`       // 수행할 액션
	DryRun       bool    `json:"dryRun"`       // dry-run 여부
	Namespace    string  `json:"namespace"`    // 네임스페이스
	Confidence   float64 `json:"confidence"`   // 파싱 신뢰도 (0.0-1.0)
	ErrorMessage string  `json:"errorMessage"` // 파싱 오류 메시지
}
