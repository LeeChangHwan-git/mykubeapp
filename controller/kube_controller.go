package controller

import (
	"encoding/json"
	"log"
	"net/http"

	"mykubeapp/model"
	"mykubeapp/service"
)

// KubeController - Spring의 @RestController와 유사한 역할
type KubeController struct {
	kubeService *service.KubeService
}

// NewKubeController - 컨트롤러 생성자 (Spring의 @Autowired 역할)
func NewKubeController() *KubeController {
	return &KubeController{
		kubeService: service.NewKubeService(),
	}
}

// GetConfig - 현재 kube config 내용 반환 (GET /api/config)
func (kc *KubeController) GetConfig(w http.ResponseWriter, r *http.Request) {
	log.Println("📋 GET /api/config - kube config 조회 요청")

	configContent, err := kc.kubeService.GetCurrentConfig()
	if err != nil {
		http.Error(w, "Config 파일을 읽을 수 없습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.ConfigResponse{}
	response.Success = true
	response.Message = "Config 조회 성공"
	response.Data = configContent

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddConfig - 새로운 config 추가 (POST /api/config)
func (kc *KubeController) AddConfig(w http.ResponseWriter, r *http.Request) {
	log.Println("📝 POST /api/config - config 추가 요청")

	var request model.AddConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	err := kc.kubeService.AddConfig(request)
	if err != nil {
		http.Error(w, "Config 추가 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.BaseResponse{
		Success: true,
		Message: "Config가 성공적으로 추가되었습니다",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetContexts - kubectl config get-contexts 결과 반환 (GET /api/contexts)
func (kc *KubeController) GetContexts(w http.ResponseWriter, r *http.Request) {
	log.Println("📋 GET /api/contexts - context 목록 조회 요청")

	contexts, err := kc.kubeService.GetContexts()
	if err != nil {
		http.Error(w, "Context 목록을 가져올 수 없습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.ContextsResponse{}
	response.Success = true
	response.Message = "Context 목록 조회 성공"
	response.Data = contexts

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UseContext - 특정 context 사용 설정 (POST /api/context/use)
func (kc *KubeController) UseContext(w http.ResponseWriter, r *http.Request) {
	log.Println("🔄 POST /api/context/use - context 변경 요청")

	var request model.UseContextRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	err := kc.kubeService.UseContext(request.ContextName)
	if err != nil {
		http.Error(w, "Context 변경 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.BaseResponse{
		Success: true,
		Message: "Context가 성공적으로 변경되었습니다: " + request.ContextName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
