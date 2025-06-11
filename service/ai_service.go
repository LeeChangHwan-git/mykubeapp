package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mykubeapp/utils"
	"net/http"
	"strings"
	"time"

	"mykubeapp/model"
)

// AIService - DeepSeek Coder와 통신하는 서비스
type AIService struct {
	baseURL     string
	httpClient  *http.Client
	kubeService *KubeService
}

// NewAIService - AI 서비스 생성자
func NewAIService(deepseekURL string) *AIService {
	return &AIService{
		baseURL: deepseekURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		kubeService: NewKubeService(),
	}
}

// GenerateKubernetesYaml - AI에게 Kubernetes YAML 생성 요청
func (ai *AIService) GenerateKubernetesYaml(request model.AIYamlRequest) (*model.AIYamlResponse, error) {
	log.Printf("🤖 AI YAML 생성 요청: %s", request.Prompt)

	// AI 프롬프트 구성
	systemPrompt := `You are a Kubernetes expert. Generate valid Kubernetes YAML based on user requirements.
Rules:
1. Always return valid YAML format
2. Use appropriate Kubernetes API versions
3. Include necessary metadata (name, namespace if needed)
4. Add helpful labels and annotations
5. Only return the YAML content, no explanations`

	userPrompt := fmt.Sprintf("Create Kubernetes YAML: %s", request.Prompt)

	// DeepSeek API 요청 구성
	aiRequest := model.DeepSeekRequest{
		Model: "deepseek-coder-v2:16b",
		Messages: []model.DeepSeekMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0.1,
		MaxTokens:   2048,
		Stream:      false,
	}

	// AI API 호출
	yamlContent, err := ai.callDeepSeekAPI(aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI API 호출 실패: %v", err)
	}

	// YAML 내용 정제
	cleanYaml := ai.cleanYamlContent(yamlContent)

	// YAML 유효성 검증
	if err := ai.kubeService.ValidateYaml(cleanYaml); err != nil {
		log.Printf("⚠️ AI가 생성한 YAML이 유효하지 않음: %v", err)
		// 재시도 로직 또는 기본 템플릿 사용 가능
	}

	response := &model.AIYamlResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: "AI YAML 생성 완료",
		},
		Data: model.AIYamlResult{
			GeneratedYaml: cleanYaml,
			Prompt:        request.Prompt,
			GeneratedTime: time.Now().Format("2006-01-02 15:04:05"),
			Source:        "DeepSeek Coder",
		},
	}

	log.Printf("✅ AI YAML 생성 완료")
	return response, nil
}

// GenerateAndApplyYaml - AI로 YAML 생성 후 바로 적용
func (ai *AIService) GenerateAndApplyYaml(request model.AIApplyRequest) (*model.AIApplyResponse, error) {
	log.Printf("🚀 AI YAML 생성 및 적용 요청: %s", request.Prompt)

	// 🆕 삭제 명령어 감지 로직 추가
	deleteKeywords := []string{"삭제", "delete", "제거", "remove", "없애"}
	isDeleteCommand := false
	for _, keyword := range deleteKeywords {
		if strings.Contains(strings.ToLower(request.Prompt), keyword) {
			isDeleteCommand = true
			break
		}
	}

	// 🆕 삭제 명령어라면 별도 처리
	if isDeleteCommand {
		log.Printf("🗑️ 삭제 명령어 감지됨: %s", request.Prompt)
		return ai.HandleDeleteCommand(request)
	}

	// 1단계: AI로 YAML 생성
	yamlRequest := model.AIYamlRequest{
		Prompt: request.Prompt,
	}

	yamlResponse, err := ai.GenerateKubernetesYaml(yamlRequest)
	if err != nil {
		return nil, fmt.Errorf("AI YAML 생성 실패: %v", err)
	}

	// 2단계: 생성된 YAML 적용
	applyRequest := model.ApplyYamlRequest{
		YamlContent: yamlResponse.Data.GeneratedYaml,
		Namespace:   request.Namespace,
		DryRun:      request.DryRun,
	}

	applyResult, err := ai.kubeService.ApplyYaml(applyRequest)
	if err != nil {
		return nil, fmt.Errorf("YAML 적용 실패: %v", err)
	}

	// 응답 구성
	response := &model.AIApplyResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: "AI YAML 생성 및 적용 완료",
		},
		Data: model.AIApplyResult{
			GeneratedYaml: yamlResponse.Data.GeneratedYaml,
			ApplyResult:   *applyResult,
			Prompt:        request.Prompt,
			GeneratedTime: yamlResponse.Data.GeneratedTime,
			Source:        "DeepSeek Coder",
		},
	}

	if request.DryRun {
		log.Printf("✅ AI YAML 생성 및 dry-run 완료")
	} else {
		log.Printf("✅ AI YAML 생성 및 적용 완료 (리소스 수: %d)", len(applyResult.Resources))
	}

	return response, nil
}

// QueryKubernetesAI - Kubernetes 관련 질문을 AI에게 물어보기
func (ai *AIService) QueryKubernetesAI(request model.AIQueryRequest) (*model.AIQueryResponse, error) {
	log.Printf("💬 AI 쿠버네티스 질문: %s", request.Question)
	// 현재 클러스터 정보 수집 (타임아웃 방지를 위해 간소화)
	var currentContext string

	// 컨텍스트 조회를 고루틴으로 처리하여 타임아웃 방지
	contextChan := make(chan string, 1)
	go func() {
		contexts, err := ai.kubeService.GetContexts()
		if err != nil {
			log.Printf("⚠️ 컨텍스트 조회 실패 (무시하고 계속): %v", err)
			contextChan <- "unknown"
			return
		}

		for _, ctx := range contexts {
			if ctx.IsCurrent {
				contextChan <- ctx.Name
				return
			}
		}
		contextChan <- "default"
	}()

	// 3초 내에 컨텍스트 조회 완료되지 않으면 기본값 사용
	select {
	case currentContext = <-contextChan:
		log.Printf("✅ 현재 컨텍스트: %s", currentContext)
	case <-time.After(3 * time.Second):
		currentContext = "unknown"
		log.Printf("⚠️ 컨텍스트 조회 타임아웃, 기본값 사용")
	}

	// AI 프롬프트 구성 (더 간결하게)
	systemPrompt := `You are a Kubernetes expert assistant. Answer questions about Kubernetes clearly and concisely.
Current cluster context: ` + currentContext + `
Provide practical, actionable advice with examples when helpful.`

	aiRequest := model.DeepSeekRequest{
		Model: "deepseek-coder-v2:16b",
		Messages: []model.DeepSeekMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: request.Question,
			},
		},
		Temperature: 0.3,
		MaxTokens:   800, // 1024 → 800으로 줄여서 응답 속도 향상
		Stream:      false,
	}

	// AI API 호출
	log.Printf("🌐 AI API 질문 요청 시작...")
	answer, err := ai.callDeepSeekAPI(aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI API 호출 실패: %v", err)
	}
	log.Printf("✅ AI API 질문 응답 완료")

	response := &model.AIQueryResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: "AI 질문 응답 완료",
		},
		Data: model.AIQueryResult{
			Question:     request.Question,
			Answer:       answer,
			Context:      currentContext,
			AnsweredTime: time.Now().Format("2006-01-02 15:04:05"),
			Source:       "DeepSeek Coder",
		},
	}

	log.Printf("✅ AI 질문 응답 완료")
	return response, nil
}

// callDeepSeekAPI - DeepSeek API 실제 호출
func (ai *AIService) callDeepSeekAPI(request model.DeepSeekRequest) (string, error) {
	// JSON 요청 생성
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("JSON 인코딩 실패: %v", err)
	}

	// HTTP 요청 생성
	url := fmt.Sprintf("%s/v1/chat/completions", ai.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	// 헤더 설정
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// API 호출
	log.Printf("🌐 DeepSeek API 호출: %s", url)
	resp, err := ai.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API 호출 실패: %v", err)
	}
	defer resp.Body.Close()

	// 응답 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("응답 읽기 실패: %v", err)
	}

	// HTTP 상태 확인
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 오류 (상태: %d): %s", resp.StatusCode, string(body))
	}

	// 응답 파싱
	var apiResponse model.DeepSeekResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", fmt.Errorf("응답 파싱 실패: %v", err)
	}

	// 응답 내용 추출
	if len(apiResponse.Choices) == 0 {
		return "", fmt.Errorf("API 응답에 내용이 없습니다")
	}

	content := apiResponse.Choices[0].Message.Content
	log.Printf("✅ DeepSeek API 응답 수신 (길이: %d)", len(content))

	return content, nil
}

// cleanYamlContent - AI가 생성한 YAML 내용 정제
func (ai *AIService) cleanYamlContent(content string) string {
	// 코드 블록 마커 제거
	content = strings.ReplaceAll(content, "```yaml", "")
	content = strings.ReplaceAll(content, "```yml", "")
	content = strings.ReplaceAll(content, "```", "")

	// 앞뒤 공백 제거
	content = strings.TrimSpace(content)

	// 시작 부분의 설명 제거 (YAML이 아닌 내용)
	lines := strings.Split(content, "\n")
	var yamlLines []string
	yamlStarted := false

	for _, line := range lines {
		// YAML이 시작되는 지점 찾기 (apiVersion 또는 kind로 시작)
		if !yamlStarted && (strings.HasPrefix(strings.TrimSpace(line), "apiVersion:") ||
			strings.HasPrefix(strings.TrimSpace(line), "kind:")) {
			yamlStarted = true
		}

		if yamlStarted {
			yamlLines = append(yamlLines, line)
		}
	}

	return strings.Join(yamlLines, "\n")
}

// CheckDeepSeekConnection - DeepSeek 연결 상태 확인
func (ai *AIService) CheckDeepSeekConnection() error {
	url := fmt.Sprintf("%s/v1/models", ai.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("연결 테스트 요청 생성 실패: %v", err)
	}

	resp, err := ai.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DeepSeek 서버 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DeepSeek 서버 응답 오류: %d", resp.StatusCode)
	}

	log.Println("✅ DeepSeek 연결 확인 완료")
	return nil
}

// 🆕 HandleDeleteCommand - 삭제 명령어 처리 (새로 추가된 함수)
func (ai *AIService) HandleDeleteCommand(request model.AIApplyRequest) (*model.AIApplyResponse, error) {
	log.Printf("🗑️ AI 삭제 명령어 처리 시작: %s", request.Prompt)

	// AI에게 삭제할 리소스 파악 요청
	systemPrompt := `You are a Kubernetes expert. The user wants to DELETE resources.
Parse the user's delete request and identify the exact resources to delete.

Rules:
1. Return ONLY resource names in format: "resourceType/resourceName"
2. Multiple resources should be separated by newlines
3. Examples:
   - "nginx-service 서비스 삭제" → "service/nginx-service"
   - "nginx-deployment 삭제" → "deployment/nginx-deployment"
   - "nginx-service 서비스 삭제, nginx-deployment 삭제" → "service/nginx-service\ndeployment/nginx-deployment"
4. Do NOT generate YAML, only return resource identifiers to delete`

	aiRequest := model.DeepSeekRequest{
		Model: "deepseek-coder-v2:16b",
		Messages: []model.DeepSeekMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: "Parse this delete request: " + request.Prompt,
			},
		},
		Temperature: 0.1,
		MaxTokens:   512,
		Stream:      false,
	}

	// AI API 호출
	resourceList, err := ai.callDeepSeekAPI(aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI API 호출 실패: %v", err)
	}

	log.Printf("🔍 AI가 파악한 삭제 대상: %s", resourceList)

	// 리소스 목록 파싱 및 삭제 실행
	resources := strings.Split(strings.TrimSpace(resourceList), "\n")
	var deleteResults []string
	var successResources []string

	for _, resource := range resources {
		resource = strings.TrimSpace(resource)
		if resource == "" {
			continue
		}

		log.Printf("🗑️ 삭제 시도: %s", resource)

		// kubectl delete 명령 구성
		cmd := []string{"delete", resource}

		if request.Namespace != "" && request.Namespace != "default" {
			cmd = append(cmd, "-n", request.Namespace)
		}

		if request.DryRun {
			cmd = append(cmd, "--dry-run=client")
		}

		// kubectl 명령 실행
		result, err := utils.ExecuteCommand("kubectl", cmd...)
		if err != nil {
			deleteResults = append(deleteResults, fmt.Sprintf("❌ %s: %v", resource, err))
			log.Printf("❌ 삭제 실패 %s: %v", resource, err)
		} else {
			deleteResults = append(deleteResults, fmt.Sprintf("✅ %s: %s", resource, strings.TrimSpace(result)))
			successResources = append(successResources, resource)
			log.Printf("✅ 삭제 성공 %s: %s", resource, result)
		}
	}

	// 응답 구성
	response := &model.AIApplyResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: "AI 리소스 삭제 처리 완료",
		},
		Data: model.AIApplyResult{
			GeneratedYaml: "# 삭제 명령어 실행 결과\n" + strings.Join(deleteResults, "\n"),
			ApplyResult: model.ApplyYamlResult{
				Output:      strings.Join(deleteResults, "\n"),
				AppliedTime: time.Now().Format("2006-01-02 15:04:05"),
				Resources:   successResources,
				DryRun:      request.DryRun,
			},
			Prompt:        request.Prompt,
			GeneratedTime: time.Now().Format("2006-01-02 15:04:05"),
			Source:        "DeepSeek Coder (Delete Mode)",
		},
	}

	log.Printf("✅ AI 삭제 명령어 처리 완료 (성공: %d개)", len(successResources))
	return response, nil
}

// CallDeepSeekAPI - 외부에서 호출 가능한 DeepSeek API 메서드 (Git Controller에서 사용)
func (ai *AIService) CallDeepSeekAPI(request model.DeepSeekRequest) (string, error) {
	return ai.callDeepSeekAPI(request)
}

// ProcessGitPrompt - Git 관련 프롬프트 처리 (개선된 버전)
func (ai *AIService) ProcessGitPrompt(prompt string) (*model.AIGitResponse, error) {
	log.Printf("🤖 Git 프롬프트 처리: %s", prompt)

	// Git 관련 키워드 감지
	gitKeywords := []string{"레포지토리", "레포", "repository", "repo", "github", "gitlab", "bitbucket", "git"}
	isGitRelated := false

	lowerPrompt := strings.ToLower(prompt)
	for _, keyword := range gitKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			isGitRelated = true
			break
		}
	}

	if !isGitRelated {
		return nil, fmt.Errorf("Git 관련 프롬프트가 아닙니다")
	}

	// Git 프롬프트 파싱을 위한 AI 요청
	systemPrompt := `You are a Git repository parser for Kubernetes operations. 
Parse user requests and extract Git repository information.

IMPORTANT: Return ONLY a valid JSON object, no markdown formatting, no code blocks, no explanations.

Required JSON format:
{
  "repoUrl": "https://github.com/user/repo.git",
  "branch": "main",
  "filename": "deployment.yaml",
  "action": "apply",
  "dryRun": false,
  "namespace": "",
  "confidence": 0.95
}

Rules:
1. repoUrl: Add https:// if missing, add .git if missing
2. branch: Default "main" if not specified
3. filename: Specific file name if mentioned, empty string if not
4. action: "apply" for 적용/배포/생성, "show" for 보기/표시/조회
5. dryRun: true if dry-run/테스트/시뮬레이션 mentioned
6. namespace: Kubernetes namespace if specified
7. confidence: 0.0-1.0 based on parsing certainty`

	aiRequest := model.DeepSeekRequest{
		Model: "deepseek-coder-v2:16b",
		Messages: []model.DeepSeekMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.1,
		MaxTokens:   200,
		Stream:      false,
	}

	// AI API 호출
	response, err := ai.callDeepSeekAPI(aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI API 호출 실패: %v", err)
	}

	log.Printf("🤖 AI 원본 응답: %s", response)

	// AI 응답 정제
	cleanedResponse := ai.cleanAIResponse(response)

	// JSON 파싱
	var parseResult model.GitParseResult
	if err := json.Unmarshal([]byte(cleanedResponse), &parseResult); err != nil {
		// JSON 파싱 실패 시 기본값으로 처리
		log.Printf("⚠️ JSON 파싱 실패, 기본 파싱 사용: %v", err)
		parseResult = ai.fallbackParseGitPrompt(prompt)
	}

	// URL 정규화
	if parseResult.RepoURL != "" {
		parseResult.RepoURL = ai.normalizeRepoURL(parseResult.RepoURL)
	}

	// 응답 구성
	aiGitResponse := &model.AIGitResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: "Git 프롬프트 파싱 완료",
		},
		Data: model.AIGitData{
			ParsedRequest: parseResult,
			RepoURL:       parseResult.RepoURL,
			Branch:        parseResult.Branch,
			Filename:      parseResult.Filename,
			Action:        parseResult.Action,
			ProcessedTime: time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	return aiGitResponse, nil
}

// normalizeRepoURL - 레포지토리 URL 정규화
func (ai *AIService) normalizeRepoURL(repoURL string) string {
	// https:// 접두사 추가
	if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}

	// .git 접미사 추가
	if !strings.HasSuffix(repoURL, ".git") {
		repoURL = repoURL + ".git"
	}

	return repoURL
}

// cleanAIResponse - AI 응답에서 JSON 추출 및 정제
func (ai *AIService) cleanAIResponse(response string) string {
	// 마크다운 코드 블록 제거
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")

	// 앞뒤 공백 제거
	response = strings.TrimSpace(response)

	// JSON 시작/끝 찾기
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		response = response[startIdx : endIdx+1]
	}

	log.Printf("🔧 AI 응답 정제 결과: %s", response)
	return response
}

// fallbackParseGitPrompt - AI 파싱 실패 시 폴백 파싱
func (ai *AIService) fallbackParseGitPrompt(prompt string) model.GitParseResult {
	log.Println("🔄 폴백 Git 프롬프트 파싱 사용")

	result := model.GitParseResult{
		Branch:     "main",
		DryRun:     false,
		Confidence: 0.5,
	}

	lowerPrompt := strings.ToLower(prompt)

	// 액션 감지
	applyKeywords := []string{"적용", "배포", "생성", "apply", "deploy", "create"}
	showKeywords := []string{"보여", "표시", "조회", "show", "display", "list"}

	for _, keyword := range applyKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			result.Action = "apply"
			break
		}
	}

	if result.Action == "" {
		for _, keyword := range showKeywords {
			if strings.Contains(lowerPrompt, keyword) {
				result.Action = "show"
				break
			}
		}
	}

	// 기본값
	if result.Action == "" {
		result.Action = "show"
	}

	// DryRun 감지
	dryRunKeywords := []string{"dry-run", "dryrun", "테스트", "시뮬레이션", "test"}
	for _, keyword := range dryRunKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			result.DryRun = true
			break
		}
	}

	// 간단한 URL 추출 (개선 필요)
	words := strings.Fields(prompt)
	for _, word := range words {
		if strings.Contains(word, "github.com") || strings.Contains(word, "gitlab.com") || strings.Contains(word, "bitbucket.org") {
			if !strings.HasPrefix(word, "http") {
				word = "https://" + word
			}
			if !strings.HasSuffix(word, ".git") {
				word = word + ".git"
			}
			result.RepoURL = word
			break
		}
	}

	// 파일명 추출 (.yaml, .yml 파일)
	for _, word := range words {
		if strings.HasSuffix(word, ".yaml") || strings.HasSuffix(word, ".yml") {
			result.Filename = word
			break
		}
	}

	return result
}

// GenerateGitYamlWithAI - AI로 Git에서 가져온 YAML 분석 및 설명
func (ai *AIService) GenerateGitYamlWithAI(yamlFiles []model.GitYamlFile, action string) (*model.AIYamlResponse, error) {
	log.Printf("🤖 Git YAML AI 분석: %d개 파일, 액션: %s", len(yamlFiles), action)

	if len(yamlFiles) == 0 {
		return nil, fmt.Errorf("분석할 YAML 파일이 없습니다")
	}

	// YAML 파일들 요약
	var yamlSummary strings.Builder
	yamlSummary.WriteString("발견된 Kubernetes YAML 파일들:\n")

	for i, file := range yamlFiles {
		if i >= 5 { // 최대 5개 파일만 요약
			yamlSummary.WriteString(fmt.Sprintf("... 그 외 %d개 파일\n", len(yamlFiles)-5))
			break
		}
		yamlSummary.WriteString(fmt.Sprintf("- %s (%d bytes)\n", file.Path, file.Size))

		// 첫 번째 파일의 내용 일부 포함
		if i == 0 && len(file.Content) > 0 {
			lines := strings.Split(file.Content, "\n")
			yamlSummary.WriteString("  내용 미리보기:\n")
			for j, line := range lines {
				if j >= 10 { // 최대 10줄만
					yamlSummary.WriteString("  ...\n")
					break
				}
				yamlSummary.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
	}

	// AI 프롬프트 구성
	var systemPrompt string
	if action == "apply" {
		systemPrompt = `You are a Kubernetes expert. Analyze the provided YAML files and provide:
1. Summary of what will be created/applied
2. Potential issues or warnings
3. Recommended namespace if not specified
4. Dependencies between resources
5. Estimated resource requirements

Be concise but thorough in your analysis.`
	} else {
		systemPrompt = `You are a Kubernetes expert. Analyze the provided YAML files and provide:
1. Overview of the Kubernetes resources
2. Architecture explanation
3. Purpose and functionality of each component
4. Best practices assessment
5. Suggestions for improvement

Be educational and helpful in your explanation.`
	}

	aiRequest := model.DeepSeekRequest{
		Model: "deepseek-coder-v2:16b",
		Messages: []model.DeepSeekMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: yamlSummary.String(),
			},
		},
		Temperature: 0.3,
		MaxTokens:   1000,
		Stream:      false,
	}

	// AI API 호출
	analysis, err := ai.callDeepSeekAPI(aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI 분석 실패: %v", err)
	}

	// 응답 구성
	response := &model.AIYamlResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: "Git YAML AI 분석 완료",
		},
		Data: model.AIYamlResult{
			GeneratedYaml: analysis, // 분석 결과를 GeneratedYaml 필드에 저장
			Prompt:        fmt.Sprintf("Git 레포지토리 YAML 분석 (%d개 파일)", len(yamlFiles)),
			GeneratedTime: time.Now().Format("2006-01-02 15:04:05"),
			Source:        "DeepSeek Coder (Git Analysis)",
		},
	}

	return response, nil
}
