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

func setupRoutes(router *mux.Router) {
	// 컨트롤러 인스턴스 생성
	kubeController := controller.NewKubeController()

	// API 라우트 설정 (Spring의 @RequestMapping과 유사)
	api := router.PathPrefix("/api").Subrouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","message":"쿠버네티스 관리 애플리케이션이 정상 동작 중입니다"}`))
	}).Methods("GET")

	// 쿠버네티스 관련 API
	api.HandleFunc("/config", kubeController.GetConfig).Methods("GET")        // 현재 config 조회
	api.HandleFunc("/config", kubeController.AddConfig).Methods("POST")       // config 추가
	api.HandleFunc("/contexts", kubeController.GetContexts).Methods("GET")    // context 목록 조회
	api.HandleFunc("/context/use", kubeController.UseContext).Methods("POST") // context 변경

	log.Println("📋 등록된 라우트:")
	log.Println("  GET  /health         - 헬스 체크")
	log.Println("  GET  /api/config     - 현재 kube config 조회")
	log.Println("  POST /api/config     - 새로운 config 추가")
	log.Println("  GET  /api/contexts   - context 목록 조회")
	log.Println("  POST /api/context/use - context 변경")
}
