package main

import (
	"github.com/gorilla/mux"
	"log"
	"mykubeapp/controller"
	"net/http"
)

func main() {
	// Spring Boot의 SpringApplication.run() 역할
	log.Println("🚀 쿠버네티스 관리 애플리케이션 시작...")

	// 라우터 생성 (Spring의 @RequestMapping 역할)
	router := mux.NewRouter()

	// CORS 미들웨어 적용
	router.Use(corsMiddleware)

	// API 라우팅 설정
	setupRoutes(router)

	// 서버 시작 (기본적으로 8080 포트)
	port := ":8080"
	log.Printf("🌐 서버가 포트 %s에서 실행 중입니다", port)
	log.Printf("📚 API 문서: http://localhost%s/health", port)

	// HTTP 서버 시작
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal("❌ 서버 시작 실패:", err)
	}
}

// corsMiddleware - CORS 헤더 설정 미들웨어
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS 헤더 설정
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Preflight 요청 처리 (OPTIONS 메서드)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 다음 핸들러 호출
		next.ServeHTTP(w, r)
	})
}

func setupRoutes(router *mux.Router) {
	// 컨트롤러 인스턴스 생성
	kubeController := controller.NewKubeController()
	terminalController := controller.NewTerminalController()
	aiController := controller.NewAIController()
	gitController := controller.NewGitController() // Git 컨트롤러 추가

	// API 라우트 설정 (Spring의 @RequestMapping과 유사)
	api := router.PathPrefix("/api").Subrouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","message":"쿠버네티스 관리 애플리케이션이 정상 동작 중입니다"}`))
	}).Methods("GET")

	// 쿠버네티스 관련 API
	api.HandleFunc("/config", kubeController.GetConfig).Methods("GET", "OPTIONS")
	api.HandleFunc("/config", kubeController.AddConfig).Methods("POST", "OPTIONS")
	api.HandleFunc("/contexts", kubeController.GetContexts).Methods("GET", "OPTIONS")
	api.HandleFunc("/context/use", kubeController.UseContext).Methods("POST", "OPTIONS")
	api.HandleFunc("/context", kubeController.DeleteContext).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/context/{contextName}", kubeController.GetContextDetail).Methods("GET", "OPTIONS")
	api.HandleFunc("/apply", kubeController.ApplyYaml).Methods("POST", "OPTIONS")
	api.HandleFunc("/delete", kubeController.DeleteYaml).Methods("POST", "OPTIONS")
	api.HandleFunc("/kubectl", terminalController.KubectlTerminal)

	// AI 관련 API
	api.HandleFunc("/ai/health", aiController.CheckAIHealth).Methods("GET", "OPTIONS")
	api.HandleFunc("/ai/generate-yaml", aiController.GenerateYaml).Methods("POST", "OPTIONS")
	api.HandleFunc("/ai/generate-apply", aiController.GenerateAndApplyEnhanced).Methods("POST", "OPTIONS") // 🆕 Enhanced 버전 사용
	api.HandleFunc("/ai/query", aiController.QueryAI).Methods("POST", "OPTIONS")
	api.HandleFunc("/ai/template", aiController.GenerateTemplate).Methods("POST", "OPTIONS")
	api.HandleFunc("/ai/validate", aiController.ValidateYaml).Methods("POST", "OPTIONS")
	api.HandleFunc("/ai/examples", aiController.GetAIExamples).Methods("GET", "OPTIONS")
	api.HandleFunc("/ai/git", aiController.ProcessGitCommand).Methods("POST", "OPTIONS") // 🆕 Git 전용 엔드포인트 추가

	// 🆕 Git 관련 API 추가
	api.HandleFunc("/git/yaml", gitController.GetYamlFromGit).Methods("POST", "OPTIONS")    // Git에서 YAML 조회
	api.HandleFunc("/git/apply", gitController.ApplyYamlFromGit).Methods("POST", "OPTIONS") // Git에서 YAML 적용
	api.HandleFunc("/git/ai", gitController.ProcessGitWithAI).Methods("POST", "OPTIONS")    // AI를 통한 Git 연동
	api.HandleFunc("/git/cleanup", gitController.CleanupGitTemp).Methods("GET", "OPTIONS")  // Git 임시 파일 정리

	log.Println("📋 등록된 라우트:")
	log.Println("  GET    /health                    - 헬스 체크")
	log.Println("  GET    /api/config                - 현재 kube config 조회")
	log.Println("  POST   /api/config                - 새로운 config 추가")
	log.Println("  GET    /api/contexts              - context 목록 조회")
	log.Println("  GET    /api/context/{contextName} - context 상세 정보 조회")
	log.Println("  POST   /api/context/use           - context 변경")
	log.Println("  DELETE /api/context               - context 삭제")
	log.Println("  POST   /api/apply                 - YAML 적용")
	log.Println("  POST   /api/delete                - YAML 삭제")
	log.Println("  WS     /api/kubectl               - Kubectl 웹터미널")
	log.Println("")
	log.Println("🤖 AI 관련 라우트:")
	log.Println("  GET    /api/ai/health             - AI 서비스 상태 확인")
	log.Println("  POST   /api/ai/generate-yaml      - AI로 YAML 생성")
	log.Println("  POST   /api/ai/generate-apply     - AI로 YAML 생성 후 적용 (Git 자동감지)")
	log.Println("  POST   /api/ai/query              - AI에게 질문하기")
	log.Println("  POST   /api/ai/template           - 템플릿 기반 YAML 생성")
	log.Println("  POST   /api/ai/validate           - AI YAML 검증")
	log.Println("  POST   /api/ai/git                - AI Git 전용 처리")
	log.Println("  GET    /api/ai/examples           - AI 사용 예제")
	log.Println("")
	log.Println("📦 Git 관련 라우트:")
	log.Println("  POST   /api/git/yaml             - Git 레포지토리 YAML 조회")
	log.Println("  POST   /api/git/apply            - Git 레포지토리 YAML 적용")
	log.Println("  POST   /api/git/ai               - AI를 통한 Git 연동")
	log.Println("  GET    /api/git/cleanup          - Git 임시 파일 정리")
	log.Println("✅ CORS 미들웨어 적용 완료 (모든 라우트에 OPTIONS 지원)")
}
