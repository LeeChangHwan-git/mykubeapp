// ===== FILE: mykubeapp/controller/ai_controller.go =====
package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"mykubeapp/model"
	"mykubeapp/service"
)

// AIController - AI 관련 컨트롤러
type AIController struct {
	aiService *service.AIService
}

// NewAIController - AI 컨트롤러 생성자
func NewAIController() *AIController {
	// 환경변수에서 DeepSeek URL 가져오기 (기본값: localhost:11434)
	deepseekURL := os.Getenv("DEEPSEEK_URL")
	if deepseekURL == "" {
		deepseekURL = "http://localhost:11434" // 기본 DeepSeek 로컬 서버 주소
	}

	log.Printf("🤖 DeepSeek 서버 URL: %s", deepseekURL)

	return &AIController{
		aiService: service.NewAIService(deepseekURL),
	}
}

// GenerateYaml - AI로 Kubernetes YAML 생성 (POST /api/ai/generate-yaml)
func (ac *AIController) GenerateYaml(w http.ResponseWriter, r *http.Request) {
	log.Println("🤖 POST /api/ai/generate-yaml - AI YAML 생성 요청")

	var request model.AIYamlRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// 프롬프트 검증
	if strings.TrimSpace(request.Prompt) == "" {
		http.Error(w, "프롬프트는 필수입니다", http.StatusBadRequest)
		return
	}

	response, err := ac.aiService.GenerateKubernetesYaml(request)
	if err != nil {
		http.Error(w, "AI YAML 생성 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// buildTemplatePrompt - 템플릿 타입별 프롬프트 생성
func (ac *AIController) buildTemplatePrompt(request model.AITemplateRequest) string {
	basePrompt := "Create a Kubernetes " + request.TemplateType + " YAML with the following specifications:\n"

	switch strings.ToLower(request.TemplateType) {
	case "deployment":
		return ac.buildDeploymentPrompt(request.Parameters)
	case "service":
		return ac.buildServicePrompt(request.Parameters)
	case "pod":
		return ac.buildPodPrompt(request.Parameters)
	case "configmap":
		return ac.buildConfigMapPrompt(request.Parameters)
	case "secret":
		return ac.buildSecretPrompt(request.Parameters)
	case "ingress":
		return ac.buildIngressPrompt(request.Parameters)
	default:
		return basePrompt + ac.parametersToString(request.Parameters)
	}
}

// buildDeploymentPrompt - Deployment 템플릿 프롬프트
func (ac *AIController) buildDeploymentPrompt(params map[string]interface{}) string {
	prompt := "Create a Kubernetes Deployment YAML with:\n"

	if name, ok := params["name"].(string); ok {
		prompt += "- Name: " + name + "\n"
	}
	if image, ok := params["image"].(string); ok {
		prompt += "- Container image: " + image + "\n"
	}
	if replicas, ok := params["replicas"]; ok {
		prompt += "- Replicas: " + toString(replicas) + "\n"
	}
	if port, ok := params["port"]; ok {
		prompt += "- Container port: " + toString(port) + "\n"
	}
	if labels, ok := params["labels"].(map[string]interface{}); ok {
		prompt += "- Labels: " + mapToString(labels) + "\n"
	}
	if env, ok := params["env"].(map[string]interface{}); ok {
		prompt += "- Environment variables: " + mapToString(env) + "\n"
	}

	return prompt
}

// buildServicePrompt - Service 템플릿 프롬프트
func (ac *AIController) buildServicePrompt(params map[string]interface{}) string {
	prompt := "Create a Kubernetes Service YAML with:\n"

	if name, ok := params["name"].(string); ok {
		prompt += "- Name: " + name + "\n"
	}
	if serviceType, ok := params["type"].(string); ok {
		prompt += "- Type: " + serviceType + "\n"
	}
	if port, ok := params["port"]; ok {
		prompt += "- Port: " + toString(port) + "\n"
	}
	if targetPort, ok := params["targetPort"]; ok {
		prompt += "- Target port: " + toString(targetPort) + "\n"
	}
	if selector, ok := params["selector"].(map[string]interface{}); ok {
		prompt += "- Selector: " + mapToString(selector) + "\n"
	}

	return prompt
}

// buildPodPrompt - Pod 템플릿 프롬프트
func (ac *AIController) buildPodPrompt(params map[string]interface{}) string {
	prompt := "Create a Kubernetes Pod YAML with:\n"

	if name, ok := params["name"].(string); ok {
		prompt += "- Name: " + name + "\n"
	}
	if image, ok := params["image"].(string); ok {
		prompt += "- Container image: " + image + "\n"
	}
	if port, ok := params["port"]; ok {
		prompt += "- Container port: " + toString(port) + "\n"
	}
	if command, ok := params["command"].([]interface{}); ok {
		prompt += "- Command: " + sliceToString(command) + "\n"
	}
	if env, ok := params["env"].(map[string]interface{}); ok {
		prompt += "- Environment variables: " + mapToString(env) + "\n"
	}

	return prompt
}

// buildConfigMapPrompt - ConfigMap 템플릿 프롬프트
func (ac *AIController) buildConfigMapPrompt(params map[string]interface{}) string {
	prompt := "Create a Kubernetes ConfigMap YAML with:\n"

	if name, ok := params["name"].(string); ok {
		prompt += "- Name: " + name + "\n"
	}
	if data, ok := params["data"].(map[string]interface{}); ok {
		prompt += "- Data: " + mapToString(data) + "\n"
	}

	return prompt
}

// buildSecretPrompt - Secret 템플릿 프롬프트
func (ac *AIController) buildSecretPrompt(params map[string]interface{}) string {
	prompt := "Create a Kubernetes Secret YAML with:\n"

	if name, ok := params["name"].(string); ok {
		prompt += "- Name: " + name + "\n"
	}
	if secretType, ok := params["type"].(string); ok {
		prompt += "- Type: " + secretType + "\n"
	}
	if data, ok := params["data"].(map[string]interface{}); ok {
		prompt += "- Data (base64 encoded): " + mapToString(data) + "\n"
	}

	return prompt
}

// buildIngressPrompt - Ingress 템플릿 프롬프트
func (ac *AIController) buildIngressPrompt(params map[string]interface{}) string {
	prompt := "Create a Kubernetes Ingress YAML with:\n"

	if name, ok := params["name"].(string); ok {
		prompt += "- Name: " + name + "\n"
	}
	if host, ok := params["host"].(string); ok {
		prompt += "- Host: " + host + "\n"
	}
	if path, ok := params["path"].(string); ok {
		prompt += "- Path: " + path + "\n"
	}
	if serviceName, ok := params["serviceName"].(string); ok {
		prompt += "- Backend service: " + serviceName + "\n"
	}
	if servicePort, ok := params["servicePort"]; ok {
		prompt += "- Backend service port: " + toString(servicePort) + "\n"
	}

	return prompt
}

// 유틸리티 함수들
func (ac *AIController) parametersToString(params map[string]interface{}) string {
	result := ""
	for key, value := range params {
		result += "- " + key + ": " + toString(value) + "\n"
	}
	return result
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func mapToString(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}

	result := "{"
	first := true
	for key, value := range m {
		if !first {
			result += ", "
		}
		result += key + ": " + toString(value)
		first = false
	}
	result += "}"
	return result
}

func sliceToString(s []interface{}) string {
	if len(s) == 0 {
		return "[]"
	}

	result := "["
	for i, value := range s {
		if i > 0 {
			result += ", "
		}
		result += toString(value)
	}
	result += "]"
	return result
}

// GetAIExamples - AI 사용 예제 반환 (GET /api/ai/examples)
func (ac *AIController) GetAIExamples(w http.ResponseWriter, r *http.Request) {
	log.Println("📚 GET /api/ai/examples - AI 사용 예제 조회")

	examples := map[string]interface{}{
		"success": true,
		"message": "AI 사용 예제 목록",
		"data": map[string]interface{}{
			"yamlGeneration": []map[string]string{
				{
					"title":       "기본 Pod 생성",
					"prompt":      "kubernetes yaml 만들어줘 - 파드이고 이름은 aa 이미지는 bb",
					"description": "간단한 Pod YAML 생성",
				},
				{
					"title":       "Nginx Deployment",
					"prompt":      "Create a deployment with nginx image, 3 replicas, name web-server",
					"description": "Nginx 웹서버 Deployment 생성",
				},
				{
					"title":       "Service 생성",
					"prompt":      "Create a LoadBalancer service for nginx on port 80",
					"description": "LoadBalancer 타입의 Service 생성",
				},
			},
			"generateAndApply": []map[string]string{
				{
					"title":       "Pod 생성 및 적용",
					"prompt":      "kubernetes yaml 만들어서 적용해줘 - nginx 파드, 이름은 my-nginx, 포트는 80",
					"description": "Pod YAML 생성 후 클러스터에 바로 적용",
				},
				{
					"title":       "완전한 앱 배포",
					"prompt":      "Create and deploy a complete Redis setup with persistent volume",
					"description": "Redis 애플리케이션 완전 배포",
				},
			},
			"templates": []map[string]interface{}{
				{
					"type":        "deployment",
					"description": "Deployment 템플릿 기반 생성",
					"parameters": map[string]interface{}{
						"name":     "web-app",
						"image":    "nginx:latest",
						"replicas": 3,
						"port":     80,
					},
				},
				{
					"type":        "service",
					"description": "Service 템플릿 기반 생성",
					"parameters": map[string]interface{}{
						"name":       "web-service",
						"type":       "LoadBalancer",
						"port":       80,
						"targetPort": 80,
					},
				},
			},
			"queries": []map[string]string{
				{
					"question":    "How do I check if my pods are running correctly?",
					"description": "Pod 상태 확인 방법 질문",
				},
				{
					"question":    "What's the difference between ClusterIP and LoadBalancer services?",
					"description": "Service 타입별 차이점 질문",
				},
				{
					"question":    "How to troubleshoot a CrashLoopBackOff pod?",
					"description": "Pod 문제 해결 방법 질문",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(examples)
}

// ValidateYaml - AI가 생성한 YAML 검증 (POST /api/ai/validate)
func (ac *AIController) ValidateYaml(w http.ResponseWriter, r *http.Request) {
	log.Println("✅ POST /api/ai/validate - AI YAML 검증 요청")

	var request struct {
		YamlContent string `json:"yamlContent" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(request.YamlContent) == "" {
		http.Error(w, "YAML 내용은 필수입니다", http.StatusBadRequest)
		return
	}

	// YAML 유효성 검증
	kubeService := service.NewKubeService()
	err := kubeService.ValidateYaml(request.YamlContent)

	var response map[string]interface{}
	if err != nil {
		response = map[string]interface{}{
			"success": false,
			"message": "YAML 검증 실패: " + err.Error(),
			"data": map[string]interface{}{
				"isValid":     false,
				"errorDetail": err.Error(),
				"checkedTime": time.Now().Format("2006-01-02 15:04:05"),
			},
		}
	} else {
		response = map[string]interface{}{
			"success": true,
			"message": "YAML 검증 성공",
			"data": map[string]interface{}{
				"isValid":     true,
				"errorDetail": "",
				"checkedTime": time.Now().Format("2006-01-02 15:04:05"),
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func (ac *AIController) GenerateAndApplyEnhanced(w http.ResponseWriter, r *http.Request) {
	log.Println("🚀 POST /api/ai/generate-apply - AI YAML 생성 및 적용 요청 (Git 지원)")

	var request model.AIApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// 프롬프트 검증
	if strings.TrimSpace(request.Prompt) == "" {
		http.Error(w, "프롬프트는 필수입니다", http.StatusBadRequest)
		return
	}

	// 🆕 Git 관련 키워드 감지
	gitKeywords := []string{"레포지토리", "레포", "repository", "repo", "github", "gitlab", "bitbucket", "git"}
	isGitRelated := false
	lowerPrompt := strings.ToLower(request.Prompt)

	for _, keyword := range gitKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			isGitRelated = true
			break
		}
	}

	// Git 관련 요청이면 Git 컨트롤러로 리다이렉트
	if isGitRelated {
		log.Printf("🔄 Git 관련 요청 감지, Git 처리로 전환: %s", request.Prompt)
		ac.handleGitRelatedPrompt(w, r, request)
		return
	}

	// 기존 로직 유지 (Git이 아닌 일반 AI 처리)
	response, err := ac.aiService.GenerateAndApplyYaml(request)
	if err != nil {
		http.Error(w, "AI YAML 생성 및 적용 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGitRelatedPrompt - Git 관련 프롬프트 처리
func (ac *AIController) handleGitRelatedPrompt(w http.ResponseWriter, r *http.Request, request model.AIApplyRequest) {
	log.Printf("📦 Git 관련 AI 프롬프트 처리: %s", request.Prompt)

	// Git 프롬프트 파싱을 위한 AI 요청
	parseResult, err := ac.parseGitPromptWithAI(request.Prompt)
	if err != nil {
		http.Error(w, "Git 프롬프트 파싱 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 파싱 결과 검증
	if parseResult.RepoURL == "" {
		http.Error(w, "레포지토리 URL을 찾을 수 없습니다. 명확한 레포지토리 주소를 입력해주세요.", http.StatusBadRequest)
		return
	}

	// Git 서비스 생성
	gitService := service.NewGitService()
	defer gitService.CleanupAll() // 함수 종료 시 정리

	// Git 레포지토리 클론
	repoDir, err := gitService.CloneRepository(parseResult.RepoURL, parseResult.Branch)
	if err != nil {
		http.Error(w, "Git 레포지토리 클론 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer gitService.Cleanup(repoDir)

	var yamlFiles []model.GitYamlFile

	// 파일 검색
	if parseResult.Filename != "" {
		// 특정 파일 검색
		yamlFile, err := gitService.GetSpecificYamlFile(repoDir, parseResult.Filename)
		if err != nil {
			http.Error(w, "파일 검색 실패: "+err.Error(), http.StatusNotFound)
			return
		}
		yamlFiles = append(yamlFiles, *yamlFile)
	} else {
		// 모든 YAML 파일 검색
		foundFiles, err := gitService.FindYamlFiles(repoDir)
		if err != nil {
			http.Error(w, "YAML 파일 검색 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}
		yamlFiles = foundFiles
	}

	if len(yamlFiles) == 0 {
		http.Error(w, "레포지토리에서 Kubernetes YAML 파일을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	// YAML 파일들 적용
	applyResult, err := gitService.ApplyYamlFromGit(yamlFiles, parseResult.Namespace, parseResult.DryRun || request.DryRun)
	if err != nil {
		http.Error(w, "YAML 적용 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// AI 분석 추가 (선택적)
	var aiAnalysis *model.AIYamlResponse
	if len(yamlFiles) > 0 {
		aiAnalysis, _ = ac.aiService.GenerateGitYamlWithAI(yamlFiles, "apply")
	}

	// 응답 구성 (Git + AI 결합)
	response := model.AIApplyResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: fmt.Sprintf("Git 레포지토리 YAML 적용 완료 (성공: %d/%d)", applyResult.SuccessFiles, applyResult.TotalFiles),
		},
		Data: model.AIApplyResult{
			GeneratedYaml: ac.formatGitApplyResult(yamlFiles, applyResult, aiAnalysis),
			ApplyResult: model.ApplyYamlResult{
				Output: fmt.Sprintf("Git 레포지토리: %s\n브랜치: %s\n적용된 파일 수: %d\n성공: %d, 실패: %d",
					parseResult.RepoURL, parseResult.Branch, applyResult.TotalFiles, applyResult.SuccessFiles, applyResult.FailedFiles),
				AppliedTime: applyResult.AppliedTime,
				Resources:   applyResult.AllResources,
				DryRun:      applyResult.DryRun,
			},
			Prompt:        request.Prompt,
			GeneratedTime: time.Now().Format("2006-01-02 15:04:05"),
			Source:        "Git Repository + DeepSeek AI",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseGitPromptWithAI - AI를 통한 Git 프롬프트 파싱
func (ac *AIController) parseGitPromptWithAI(prompt string) (*model.GitParseResult, error) {
	systemPrompt := `You are a Git repository parser. Extract information from user prompts about Git repositories and Kubernetes operations.

IMPORTANT: Return ONLY a valid JSON object, no markdown formatting, no code blocks, no explanations.

Extract and return JSON with these fields:
- repoUrl: Full Git repository URL (add https:// if missing, add .git if missing)
- branch: Branch name (default: "main")  
- filename: Specific YAML filename (if mentioned, empty string if not)
- action: "apply" (for 적용/배포/생성) or "show" (for 보기/표시/조회)
- dryRun: true if mentioned (dry-run, 테스트, 시뮬레이션)
- namespace: Kubernetes namespace (if specified, empty string if not)
- confidence: 0.0-1.0 parsing confidence

Example responses:
{"repoUrl": "https://github.com/user/repo.git", "branch": "main", "filename": "app.yaml", "action": "apply", "dryRun": false, "namespace": "", "confidence": 0.9}
{"repoUrl": "https://gitlab.com/org/project.git", "branch": "main", "filename": "", "action": "show", "dryRun": false, "namespace": "", "confidence": 0.8}`

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
		MaxTokens:   300,
		Stream:      false,
	}

	// AI API 호출
	response, err := ac.aiService.CallDeepSeekAPI(aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI API 호출 실패: %v", err)
	}

	log.Printf("🤖 AI 원본 응답: %s", response)

	// AI 응답에서 JSON 추출 및 정제 (개선된 버전)
	cleanedResponse := ac.cleanAIResponseAdvanced(response)
	log.Printf("🔧 정제된 응답: %s", cleanedResponse)

	// JSON 파싱
	var parseResult model.GitParseResult
	if err := json.Unmarshal([]byte(cleanedResponse), &parseResult); err != nil {
		log.Printf("⚠️ JSON 파싱 실패: %v, 정제된 응답: %s", err, cleanedResponse)
		// 파싱 실패 시 폴백 처리
		return ac.fallbackParseGitPrompt(prompt), nil
	}

	// URL 정규화
	if parseResult.RepoURL != "" {
		parseResult.RepoURL = ac.normalizeRepoURL(parseResult.RepoURL)
	}

	// 기본값 설정
	if parseResult.Branch == "" {
		parseResult.Branch = "main"
	}
	if parseResult.Action == "" {
		parseResult.Action = "apply"
	}

	log.Printf("✅ AI 파싱 성공: %+v", parseResult)
	return &parseResult, nil
}

// cleanAIResponseAdvanced - AI 응답에서 JSON 추출 및 정제
func (ac *AIController) cleanAIResponseAdvanced(response string) string {
	log.Printf("🔧 AI 응답 정제 시작: %s", response)

	// 1. 마크다운 코드 블록 제거 (여러 패턴 처리)
	patterns := []string{
		"```json",
		"```JSON",
		"```",
		"`json",
		"`JSON",
		"`",
	}

	for _, pattern := range patterns {
		response = strings.ReplaceAll(response, pattern, "")
	}

	// 2. 앞뒤 공백 및 개행 제거
	response = strings.TrimSpace(response)

	// 3. JSON 객체 추출 (중괄호 기준)
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		log.Printf("⚠️ JSON 객체를 찾을 수 없음: %s", response)
		return ""
	}

	jsonStr := response[startIdx : endIdx+1]

	// 4. 불필요한 문자 정리
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\r", "")
	jsonStr = strings.TrimSpace(jsonStr)

	log.Printf("🔧 AI 응답 정제 완료: %s", jsonStr)
	return jsonStr
}

// cleanAIResponse - AI 응답에서 JSON 추출 및 정제
func (ac *AIController) cleanAIResponse(response string) string {
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
func (ac *AIController) fallbackParseGitPrompt(prompt string) *model.GitParseResult {
	result := &model.GitParseResult{
		Branch:     "main",
		Action:     "apply",
		DryRun:     false,
		Confidence: 0.3,
	}

	lowerPrompt := strings.ToLower(prompt)

	// 간단한 키워드 기반 파싱
	if strings.Contains(lowerPrompt, "보여") || strings.Contains(lowerPrompt, "표시") || strings.Contains(lowerPrompt, "show") {
		result.Action = "show"
	}

	if strings.Contains(lowerPrompt, "dry-run") || strings.Contains(lowerPrompt, "테스트") {
		result.DryRun = true
	}

	// URL 추출 (기본적인 패턴 매칭)
	words := strings.Fields(prompt)
	for _, word := range words {
		if strings.Contains(word, "github.com") || strings.Contains(word, "gitlab.com") || strings.Contains(word, "bitbucket.org") {
			result.RepoURL = ac.normalizeRepoURL(word)
			break
		}
	}

	// 파일명 추출
	for _, word := range words {
		if strings.HasSuffix(word, ".yaml") || strings.HasSuffix(word, ".yml") {
			result.Filename = word
			break
		}
	}

	return result
}

// normalizeRepoURL - 레포지토리 URL 정규화
func (ac *AIController) normalizeRepoURL(repoURL string) string {
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

// formatGitApplyResult - Git 적용 결과 포맷팅
func (ac *AIController) formatGitApplyResult(yamlFiles []model.GitYamlFile, applyResult *model.GitApplyResult, aiAnalysis *model.AIYamlResponse) string {
	var result strings.Builder

	result.WriteString("🔥 Git 레포지토리 YAML 적용 결과\n\n")
	result.WriteString(fmt.Sprintf("📊 요약: 총 %d개 파일, 성공 %d개, 실패 %d개\n\n",
		applyResult.TotalFiles, applyResult.SuccessFiles, applyResult.FailedFiles))

	// 성공한 파일들
	if applyResult.SuccessFiles > 0 {
		result.WriteString("✅ 성공한 파일들:\n")
		for _, fileResult := range applyResult.Results {
			if fileResult.Success {
				result.WriteString(fmt.Sprintf("  - %s (%d개 리소스)\n", fileResult.FilePath, len(fileResult.Resources)))
			}
		}
		result.WriteString("\n")
	}

	// 실패한 파일들
	if applyResult.FailedFiles > 0 {
		result.WriteString("❌ 실패한 파일들:\n")
		for _, fileResult := range applyResult.Results {
			if !fileResult.Success {
				result.WriteString(fmt.Sprintf("  - %s: %s\n", fileResult.FilePath, fileResult.Error))
			}
		}
		result.WriteString("\n")
	}

	// 적용된 리소스 목록
	if len(applyResult.AllResources) > 0 {
		result.WriteString("📦 적용된 리소스들:\n")
		for _, resource := range applyResult.AllResources {
			result.WriteString(fmt.Sprintf("  - %s\n", resource))
		}
		result.WriteString("\n")
	}

	// AI 분석 결과 추가
	if aiAnalysis != nil && aiAnalysis.Data.GeneratedYaml != "" {
		result.WriteString("🤖 AI 분석:\n")
		result.WriteString(aiAnalysis.Data.GeneratedYaml)
		result.WriteString("\n")
	}

	return result.String()
}

// ProcessGitCommand - Git 명령어 직접 처리 (새로운 엔드포인트용)
func (ac *AIController) ProcessGitCommand(w http.ResponseWriter, r *http.Request) {
	log.Println("📦 POST /api/ai/git - AI Git 명령어 처리")

	var request model.AIGitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(request.Prompt) == "" {
		http.Error(w, "프롬프트는 필수입니다", http.StatusBadRequest)
		return
	}

	// Git 관련 키워드 확인
	gitKeywords := []string{"레포지토리", "레포", "repository", "repo", "github", "gitlab", "bitbucket", "git"}
	isGitRelated := false
	lowerPrompt := strings.ToLower(request.Prompt)

	for _, keyword := range gitKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			isGitRelated = true
			break
		}
	}

	if !isGitRelated {
		http.Error(w, "Git 관련 프롬프트가 아닙니다", http.StatusBadRequest)
		return
	}

	// AI를 통한 Git 프롬프트 처리
	gitResponse, err := ac.aiService.ProcessGitPrompt(request.Prompt)
	if err != nil {
		http.Error(w, "Git 프롬프트 처리 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gitResponse)
}

// QueryAI - Kubernetes 관련 AI 질문 (POST /api/ai/query)
func (ac *AIController) QueryAI(w http.ResponseWriter, r *http.Request) {
	log.Println("💬 POST /api/ai/query - AI 질문 요청")

	var request model.AIQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// 질문 검증
	if strings.TrimSpace(request.Question) == "" {
		http.Error(w, "질문은 필수입니다", http.StatusBadRequest)
		return
	}

	response, err := ac.aiService.QueryKubernetesAI(request)
	if err != nil {
		http.Error(w, "AI 질문 처리 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CheckAIHealth - AI 서비스 상태 확인 (GET /api/ai/health)
func (ac *AIController) CheckAIHealth(w http.ResponseWriter, r *http.Request) {
	log.Println("🔍 GET /api/ai/health - AI 서비스 상태 확인")

	startTime := time.Now()

	// DeepSeek 연결 확인
	err := ac.aiService.CheckDeepSeekConnection()
	responseTime := time.Since(startTime)

	var response model.AIHealthResponse
	if err != nil {
		response = model.AIHealthResponse{
			BaseResponse: model.BaseResponse{
				Success: false,
				Message: "AI 서비스 연결 실패: " + err.Error(),
			},
			Data: model.AIHealth{
				DeepSeekURL:     os.Getenv("DEEPSEEK_URL"),
				IsConnected:     false,
				LastChecked:     time.Now().Format("2006-01-02 15:04:05"),
				ResponseTime:    responseTime.String(),
				AvailableModels: []string{},
			},
		}
	} else {
		response = model.AIHealthResponse{
			BaseResponse: model.BaseResponse{
				Success: true,
				Message: "AI 서비스 정상 동작 중",
			},
			Data: model.AIHealth{
				DeepSeekURL:     os.Getenv("DEEPSEEK_URL"),
				IsConnected:     true,
				LastChecked:     time.Now().Format("2006-01-02 15:04:05"),
				ResponseTime:    responseTime.String(),
				AvailableModels: []string{"deepseek-coder"},
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateTemplate - 템플릿 기반 YAML 생성 (POST /api/ai/template)
func (ac *AIController) GenerateTemplate(w http.ResponseWriter, r *http.Request) {
	log.Println("📝 POST /api/ai/template - 템플릿 기반 YAML 생성 요청")

	var request model.AITemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// 템플릿 타입 검증
	if strings.TrimSpace(request.TemplateType) == "" {
		http.Error(w, "템플릿 타입은 필수입니다", http.StatusBadRequest)
		return
	}

	// 템플릿별 프롬프트 생성
	prompt := ac.buildTemplatePrompt(request)

	// AI YAML 생성 요청
	yamlRequest := model.AIYamlRequest{
		Prompt: prompt,
	}

	yamlResponse, err := ac.aiService.GenerateKubernetesYaml(yamlRequest)
	if err != nil {
		http.Error(w, "템플릿 기반 YAML 생성 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 응답 구성
	response := model.AITemplateResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: "템플릿 기반 YAML 생성 완료",
		},
		Data: model.AITemplateResult{
			TemplateType:  request.TemplateType,
			Parameters:    request.Parameters,
			GeneratedYaml: yamlResponse.Data.GeneratedYaml,
			GeneratedTime: yamlResponse.Data.GeneratedTime,
			Source:        yamlResponse.Data.Source,
		},
	}

	// 즉시 적용이 요청된 경우
	if !request.DryRun && request.Parameters["apply"] == true {
		applyRequest := model.ApplyYamlRequest{
			YamlContent: yamlResponse.Data.GeneratedYaml,
			Namespace:   request.Namespace,
			DryRun:      false,
		}

		kubeService := service.NewKubeService()
		applyResult, err := kubeService.ApplyYaml(applyRequest)
		if err != nil {
			log.Printf("⚠️ 템플릿 YAML 적용 실패: %v", err)
		} else {
			response.Data.ApplyResult = applyResult
			response.Message = "템플릿 기반 YAML 생성 및 적용 완료"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
