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

	// API 라우트 설정 (Spring의 @RequestMapping과 유사)
	api := router.PathPrefix("/api").Subrouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","message":"쿠버네티스 관리 애플리케이션이 정상 동작 중입니다"}`))
	}).Methods("GET")

	// 쿠버네티스 관련 API
	api.HandleFunc("/config", kubeController.GetConfig).Methods("GET", "OPTIONS")                       // 현재 config 조회
	api.HandleFunc("/config", kubeController.AddConfig).Methods("POST", "OPTIONS")                      // config 추가
	api.HandleFunc("/contexts", kubeController.GetContexts).Methods("GET", "OPTIONS")                   // context 목록 조회
	api.HandleFunc("/context/use", kubeController.UseContext).Methods("POST", "OPTIONS")                // context 변경
	api.HandleFunc("/context", kubeController.DeleteContext).Methods("DELETE", "OPTIONS")               // context 삭제
	api.HandleFunc("/context/{contextName}", kubeController.GetContextDetail).Methods("GET", "OPTIONS") // context 상세 정보 조회
	api.HandleFunc("/apply", kubeController.ApplyYaml).Methods("POST", "OPTIONS")                       // YAML 적용
	api.HandleFunc("/delete", kubeController.DeleteYaml).Methods("POST", "OPTIONS")                     // YAML 삭제
	api.HandleFunc("/kubectl", terminalController.KubectlTerminal)                                      // WebSocket endpoint

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
	log.Println("✅ CORS 미들웨어 적용 완료 (모든 라우트에 OPTIONS 지원)")

}
