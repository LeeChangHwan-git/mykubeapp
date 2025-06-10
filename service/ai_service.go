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
			Timeout: 60 * time.Second,
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

	// 현재 클러스터 정보 수집 (컨텍스트 제공)
	contexts, _ := ai.kubeService.GetContexts()
	var currentContext string
	for _, ctx := range contexts {
		if ctx.IsCurrent {
			currentContext = ctx.Name
			break
		}
	}

	// AI 프롬프트 구성
	systemPrompt := `You are a Kubernetes expert assistant. Answer questions about Kubernetes with practical, actionable advice.
Current cluster context: ` + currentContext + `
Provide clear, concise answers with examples when helpful.`

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
		MaxTokens:   1024,
		Stream:      false,
	}

	// AI API 호출
	answer, err := ai.callDeepSeekAPI(aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI API 호출 실패: %v", err)
	}

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
