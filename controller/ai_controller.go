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

// GenerateAndApply - AI로 YAML 생성 후 바로 적용 (POST /api/ai/generate-apply)
func (ac *AIController) GenerateAndApply(w http.ResponseWriter, r *http.Request) {
	log.Println("🚀 POST /api/ai/generate-apply - AI YAML 생성 및 적용 요청")

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

	response, err := ac.aiService.GenerateAndApplyYaml(request)
	if err != nil {
		http.Error(w, "AI YAML 생성 및 적용 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
