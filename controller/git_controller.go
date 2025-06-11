package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"mykubeapp/model"
	"mykubeapp/service"
)

// GitController - Git 관련 컨트롤러
type GitController struct {
	gitService *service.GitService
	aiService  *service.AIService
}

// NewGitController - Git 컨트롤러 생성자
func NewGitController() *GitController {
	return &GitController{
		gitService: service.NewGitService(),
		aiService:  service.NewAIService("http://localhost:11434"), // DeepSeek URL
	}
}

// GetYamlFromGit - Git 레포지토리에서 YAML 파일들 가져오기 (GET /api/git/yaml)
func (gc *GitController) GetYamlFromGit(w http.ResponseWriter, r *http.Request) {
	log.Println("📦 GET /api/git/yaml - Git 레포지토리 YAML 조회 요청")

	var request model.GitYamlRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// 레포지토리 URL 검증
	if strings.TrimSpace(request.RepoURL) == "" {
		http.Error(w, "레포지토리 URL은 필수입니다", http.StatusBadRequest)
		return
	}

	// 브랜치 기본값 설정
	if request.Branch == "" {
		request.Branch = "main"
	}

	// Git 레포지토리 클론
	repoDir, err := gc.gitService.CloneRepository(request.RepoURL, request.Branch)
	if err != nil {
		http.Error(w, "Git 레포지토리 클론 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer gc.gitService.Cleanup(repoDir) // 함수 종료 시 정리

	var yamlFiles []model.GitYamlFile

	if request.Filename != "" {
		// 특정 파일 검색
		yamlFile, err := gc.gitService.GetSpecificYamlFile(repoDir, request.Filename)
		if err != nil {
			http.Error(w, "파일 검색 실패: "+err.Error(), http.StatusNotFound)
			return
		}
		yamlFiles = append(yamlFiles, *yamlFile)
	} else {
		// 모든 YAML 파일 검색
		foundFiles, err := gc.gitService.FindYamlFiles(repoDir)
		if err != nil {
			http.Error(w, "YAML 파일 검색 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}
		yamlFiles = foundFiles
	}

	// 응답 구성
	response := model.GitYamlResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: "Git 레포지토리 YAML 조회 완료",
		},
		Data: model.GitYamlData{
			RepoURL:     request.RepoURL,
			Branch:      request.Branch,
			YamlFiles:   yamlFiles,
			TotalFiles:  len(yamlFiles),
			RetrievedAt: time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ApplyYamlFromGit - Git 레포지토리에서 YAML 가져와서 적용 (POST /api/git/apply)
func (gc *GitController) ApplyYamlFromGit(w http.ResponseWriter, r *http.Request) {
	log.Println("🚀 POST /api/git/apply - Git 레포지토리 YAML 적용 요청")

	var request model.GitApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// 레포지토리 URL 검증
	if strings.TrimSpace(request.RepoURL) == "" {
		http.Error(w, "레포지토리 URL은 필수입니다", http.StatusBadRequest)
		return
	}

	// 브랜치 기본값 설정
	if request.Branch == "" {
		request.Branch = "main"
	}

	// Git 레포지토리 클론
	repoDir, err := gc.gitService.CloneRepository(request.RepoURL, request.Branch)
	if err != nil {
		http.Error(w, "Git 레포지토리 클론 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer gc.gitService.Cleanup(repoDir) // 함수 종료 시 정리

	var yamlFiles []model.GitYamlFile

	if request.Filename != "" {
		// 특정 파일 적용
		yamlFile, err := gc.gitService.GetSpecificYamlFile(repoDir, request.Filename)
		if err != nil {
			http.Error(w, "파일 검색 실패: "+err.Error(), http.StatusNotFound)
			return
		}
		yamlFiles = append(yamlFiles, *yamlFile)
	} else {
		// 모든 YAML 파일 적용
		foundFiles, err := gc.gitService.FindYamlFiles(repoDir)
		if err != nil {
			http.Error(w, "YAML 파일 검색 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}
		yamlFiles = foundFiles
	}

	// YAML 파일들 적용
	applyResult, err := gc.gitService.ApplyYamlFromGit(yamlFiles, request.Namespace, request.DryRun)
	if err != nil {
		http.Error(w, "YAML 적용 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 응답 구성
	response := model.GitApplyResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: "Git 레포지토리 YAML 적용 완료",
		},
		Data: model.GitApplyData{
			RepoURL:     request.RepoURL,
			Branch:      request.Branch,
			ApplyResult: *applyResult,
			RetrievedAt: time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ProcessGitWithAI - AI를 통한 Git 연동 처리 (POST /api/git/ai)
func (gc *GitController) ProcessGitWithAI(w http.ResponseWriter, r *http.Request) {
	log.Println("🤖 POST /api/git/ai - AI Git 연동 요청")

	var request model.AIGitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// 프롬프트 검증
	if strings.TrimSpace(request.Prompt) == "" {
		http.Error(w, "프롬프트는 필수입니다", http.StatusBadRequest)
		return
	}

	// AI를 통해 프롬프트 파싱
	parseResult, err := gc.parseGitPromptWithAI(request.Prompt)
	if err != nil {
		http.Error(w, "AI 프롬프트 파싱 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 파싱 결과 검증
	if parseResult.RepoURL == "" {
		http.Error(w, "레포지토리 URL을 찾을 수 없습니다", http.StatusBadRequest)
		return
	}

	var executionResult interface{}
	var message string

	// 액션에 따른 처리
	switch parseResult.Action {
	case "show", "list", "display":
		// YAML 내용 조회
		yamlRequest := model.GitYamlRequest{
			RepoURL:  parseResult.RepoURL,
			Branch:   parseResult.Branch,
			Filename: parseResult.Filename,
		}

		yamlData, err := gc.executeYamlRetrieval(yamlRequest)
		if err != nil {
			http.Error(w, "YAML 조회 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}
		executionResult = yamlData
		message = "Git 레포지토리 YAML 조회 완료"

	case "apply", "deploy", "create":
		// YAML 적용
		applyRequest := model.GitApplyRequest{
			RepoURL:   parseResult.RepoURL,
			Branch:    parseResult.Branch,
			Filename:  parseResult.Filename,
			Namespace: parseResult.Namespace,
			DryRun:    parseResult.DryRun,
		}

		applyData, err := gc.executeYamlApplication(applyRequest)
		if err != nil {
			http.Error(w, "YAML 적용 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}
		executionResult = applyData
		if parseResult.DryRun {
			message = "Git 레포지토리 YAML dry-run 완료"
		} else {
			message = "Git 레포지토리 YAML 적용 완료"
		}

	default:
		http.Error(w, "지원하지 않는 액션입니다: "+parseResult.Action, http.StatusBadRequest)
		return
	}

	// 응답 구성
	response := model.AIGitResponse{
		BaseResponse: model.BaseResponse{
			Success: true,
			Message: message,
		},
		Data: model.AIGitData{
			ParsedRequest:   *parseResult,
			RepoURL:         parseResult.RepoURL,
			Branch:          parseResult.Branch,
			Filename:        parseResult.Filename,
			Action:          parseResult.Action,
			ExecutionResult: executionResult,
			ProcessedTime:   time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseGitPromptWithAI - AI를 통해 Git 프롬프트 파싱
func (gc *GitController) parseGitPromptWithAI(prompt string) (*model.GitParseResult, error) {
	log.Printf("🤖 AI Git 프롬프트 파싱: %s", prompt)

	// AI 시스템 프롬프트 구성
	systemPrompt := `You are a Git repository parser for Kubernetes operations. Parse user requests about Git repositories and YAML files.

Extract the following information from the user prompt:
1. Repository URL (GitHub, GitLab, Bitbucket, etc.)
2. Branch name (if specified, default: main)
3. Filename (if specific file mentioned)
4. Action (apply/deploy/create or show/list/display)
5. DryRun (if mentioned: dry-run, test, 시뮬레이션)
6. Namespace (if specified)

Return ONLY a JSON object with this structure:
{
  "repoUrl": "extracted repository URL",
  "branch": "branch name or main",
  "filename": "specific filename or empty",
  "action": "apply or show",
  "dryRun": boolean,
  "namespace": "namespace or empty",
  "confidence": 0.95,
  "errorMessage": "error if parsing failed"
}

Examples:
- "github.com/myorg/k8s-manifests 레포에서 deployment.yaml 적용해줘" → {"repoUrl": "https://github.com/myorg/k8s-manifests", "filename": "deployment.yaml", "action": "apply", ...}
- "https://github.com/example/repo의 yaml 파일들 모두 보여줘" → {"repoUrl": "https://github.com/example/repo", "action": "show", ...}
- "xx레포의 yaml 모두 dry-run으로 적용" → {"action": "apply", "dryRun": true, ...}`

	userPrompt := fmt.Sprintf("Parse this Git request: %s", prompt)

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
		MaxTokens:   512,
		Stream:      false,
	}

	// AI API 호출
	response, err := gc.aiService.CallDeepSeekAPI(aiRequest)
	if err != nil {
		return nil, fmt.Errorf("AI API 호출 실패: %v", err)
	}

	fmt.Println("===========================")
	fmt.Println(response)

	// AI 응답 정제 (마크다운 코드 블록 제거)
	cleanedResponse := gc.cleanAIResponseAdvanced(response)
	fmt.Println("===== 정제된 응답 =====")
	fmt.Println(cleanedResponse)

	// JSON 응답 파싱
	var parseResult model.GitParseResult
	if err := json.Unmarshal([]byte(cleanedResponse), &parseResult); err != nil {
		log.Printf("⚠️ JSON 파싱 실패: %v, 원본 응답: %s", err, response)
		log.Printf("⚠️ 정제된 응답: %s", cleanedResponse)
		// 파싱 실패 시 폴백 처리
		return gc.fallbackParseGitPrompt(prompt), nil
	}

	// 레포지토리 URL 정규화
	if parseResult.RepoURL != "" {
		parseResult.RepoURL = gc.normalizeRepoURL(parseResult.RepoURL)
	}

	// 기본값 설정
	if parseResult.Branch == "" {
		parseResult.Branch = "main"
	}

	log.Printf("✅ AI 파싱 완료: %+v", parseResult)
	return &parseResult, nil
}

// cleanAIResponseAdvanced - AI 응답에서 JSON 추출 및 정제 (개선된 버전)
func (gc *GitController) cleanAIResponseAdvanced(response string) string {
	log.Printf("🔧 AI 응답 정제 시작")

	// 1. 다양한 마크다운 패턴 제거
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

	// 3. JSON 객체 추출 (첫 번째 { 부터 마지막 } 까지)
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		log.Printf("⚠️ JSON 객체를 찾을 수 없음, 원본 응답: %s", response)
		// 기본 JSON 반환
		return `{"repoUrl": "", "branch": "main", "filename": "", "action": "show", "dryRun": false, "namespace": "", "confidence": 0.3, "errorMessage": "JSON parsing failed"}`
	}

	jsonStr := response[startIdx : endIdx+1]

	// 4. 추가 정제
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\r", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\t", "")
	jsonStr = strings.TrimSpace(jsonStr)

	log.Printf("🔧 AI 응답 정제 완료: %s", jsonStr)
	return jsonStr
}

// fallbackParseGitPrompt - AI 파싱 실패 시 폴백 파싱 (개선된 버전)
func (gc *GitController) fallbackParseGitPrompt(prompt string) *model.GitParseResult {
	log.Println("🔄 폴백 Git 프롬프트 파싱 사용")

	result := &model.GitParseResult{
		Branch:     "main",
		DryRun:     false,
		Confidence: 0.5,
		Action:     "show", // 기본값을 show로 설정
	}

	lowerPrompt := strings.ToLower(prompt)

	// 액션 감지
	applyKeywords := []string{"적용", "배포", "생성", "apply", "deploy", "create"}
	for _, keyword := range applyKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			result.Action = "apply"
			break
		}
	}

	// DryRun 감지
	dryRunKeywords := []string{"dry-run", "dryrun", "테스트", "시뮬레이션", "test"}
	for _, keyword := range dryRunKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			result.DryRun = true
			break
		}
	}

	// URL 추출 (더 정교하게)
	words := strings.Fields(prompt)
	for _, word := range words {
		// GitHub, GitLab, Bitbucket URL 패턴 감지
		if strings.Contains(word, "github.com") || strings.Contains(word, "gitlab.com") || strings.Contains(word, "bitbucket.org") {
			// URL 정규화
			url := word
			if !strings.HasPrefix(url, "http") {
				url = "https://" + url
			}
			if !strings.HasSuffix(url, ".git") {
				url = url + ".git"
			}
			result.RepoURL = url
			result.Confidence = 0.7 // URL을 찾았으므로 신뢰도 증가
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

	// 브랜치 감지
	branchKeywords := []string{"branch", "브랜치"}
	for i, word := range words {
		for _, keyword := range branchKeywords {
			if strings.Contains(strings.ToLower(word), keyword) && i+1 < len(words) {
				result.Branch = words[i+1]
				break
			}
		}
	}

	log.Printf("🔄 폴백 파싱 결과: %+v", result)
	return result
}

// cleanAIResponse - AI 응답에서 JSON 추출 및 정제
func (gc *GitController) cleanAIResponse(response string) string {
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

	// 추가 정제
	lines := strings.Split(response, "\n")
	var jsonLines []string
	jsonStarted := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// JSON 시작 감지
		if !jsonStarted && (strings.HasPrefix(trimmed, "{") || strings.Contains(trimmed, "{")) {
			jsonStarted = true
		}

		// JSON 부분만 추가
		if jsonStarted {
			jsonLines = append(jsonLines, line)

			// JSON 끝 감지
			if strings.Contains(trimmed, "}") && strings.Count(strings.Join(jsonLines, ""), "{") <= strings.Count(strings.Join(jsonLines, ""), "}") {
				break
			}
		}
	}

	result := strings.Join(jsonLines, "\n")
	log.Printf("🔧 AI 응답 정제 결과: %s", result)
	return result
}

// executeYamlRetrieval - YAML 조회 실행
func (gc *GitController) executeYamlRetrieval(request model.GitYamlRequest) (*model.GitYamlData, error) {
	// Git 레포지토리 클론
	repoDir, err := gc.gitService.CloneRepository(request.RepoURL, request.Branch)
	if err != nil {
		return nil, fmt.Errorf("Git 레포지토리 클론 실패: %v", err)
	}
	defer gc.gitService.Cleanup(repoDir)

	var yamlFiles []model.GitYamlFile

	if request.Filename != "" {
		// 특정 파일 검색
		yamlFile, err := gc.gitService.GetSpecificYamlFile(repoDir, request.Filename)
		if err != nil {
			return nil, fmt.Errorf("파일 검색 실패: %v", err)
		}
		yamlFiles = append(yamlFiles, *yamlFile)
	} else {
		// 모든 YAML 파일 검색
		foundFiles, err := gc.gitService.FindYamlFiles(repoDir)
		if err != nil {
			return nil, fmt.Errorf("YAML 파일 검색 실패: %v", err)
		}
		yamlFiles = foundFiles
	}

	return &model.GitYamlData{
		RepoURL:     request.RepoURL,
		Branch:      request.Branch,
		YamlFiles:   yamlFiles,
		TotalFiles:  len(yamlFiles),
		RetrievedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// executeYamlApplication - YAML 적용 실행
func (gc *GitController) executeYamlApplication(request model.GitApplyRequest) (*model.GitApplyData, error) {
	// Git 레포지토리 클론
	repoDir, err := gc.gitService.CloneRepository(request.RepoURL, request.Branch)
	if err != nil {
		return nil, fmt.Errorf("Git 레포지토리 클론 실패: %v", err)
	}
	defer gc.gitService.Cleanup(repoDir)

	var yamlFiles []model.GitYamlFile

	if request.Filename != "" {
		// 특정 파일 적용
		yamlFile, err := gc.gitService.GetSpecificYamlFile(repoDir, request.Filename)
		if err != nil {
			return nil, fmt.Errorf("파일 검색 실패: %v", err)
		}
		yamlFiles = append(yamlFiles, *yamlFile)
	} else {
		// 모든 YAML 파일 적용
		foundFiles, err := gc.gitService.FindYamlFiles(repoDir)
		if err != nil {
			return nil, fmt.Errorf("YAML 파일 검색 실패: %v", err)
		}
		yamlFiles = foundFiles
	}

	// YAML 파일들 적용
	applyResult, err := gc.gitService.ApplyYamlFromGit(yamlFiles, request.Namespace, request.DryRun)
	if err != nil {
		return nil, fmt.Errorf("YAML 적용 실패: %v", err)
	}

	return &model.GitApplyData{
		RepoURL:     request.RepoURL,
		Branch:      request.Branch,
		ApplyResult: *applyResult,
		RetrievedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// normalizeRepoURL - 레포지토리 URL 정규화
func (gc *GitController) normalizeRepoURL(repoURL string) string {
	// https:// 접두사가 없으면 추가
	if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") {
		if strings.Contains(repoURL, "github.com") || strings.Contains(repoURL, "gitlab.com") || strings.Contains(repoURL, "bitbucket.org") {
			repoURL = "https://" + repoURL
		}
	}

	// .git 접미사가 없으면 추가
	if !strings.HasSuffix(repoURL, ".git") {
		repoURL = repoURL + ".git"
	}

	return repoURL
}

// CleanupGitTemp - Git 임시 파일 정리 (GET /api/git/cleanup)
func (gc *GitController) CleanupGitTemp(w http.ResponseWriter, r *http.Request) {
	log.Println("🧹 GET /api/git/cleanup - Git 임시 파일 정리 요청")

	err := gc.gitService.CleanupAll()
	if err != nil {
		http.Error(w, "임시 파일 정리 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.BaseResponse{
		Success: true,
		Message: "Git 임시 파일 정리 완료",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
