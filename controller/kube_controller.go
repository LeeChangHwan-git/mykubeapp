package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"

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

// DeleteContext - 특정 context 삭제 (DELETE /api/context)
func (kc *KubeController) DeleteContext(w http.ResponseWriter, r *http.Request) {
	log.Println("🗑️ DELETE /api/context - context 삭제 요청")

	var request model.DeleteContextRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// 빈 컨텍스트 이름 검증
	if strings.TrimSpace(request.ContextName) == "" {
		http.Error(w, "컨텍스트 이름은 필수입니다", http.StatusBadRequest)
		return
	}

	err := kc.kubeService.DeleteContext(request.ContextName)
	if err != nil {
		http.Error(w, "Context 삭제 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.BaseResponse{
		Success: true,
		Message: "Context가 성공적으로 삭제되었습니다: " + request.ContextName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetContextDetail - 특정 context의 상세 정보 조회 (GET /api/context/{contextName})
func (kc *KubeController) GetContextDetail(w http.ResponseWriter, r *http.Request) {
	log.Println("📋 GET /api/context/{contextName} - context 상세 정보 조회 요청")

	// URL에서 contextName 파라미터 추출
	vars := mux.Vars(r)
	contextName := vars["contextName"]

	if strings.TrimSpace(contextName) == "" {
		http.Error(w, "컨텍스트 이름은 필수입니다", http.StatusBadRequest)
		return
	}

	contextDetail, err := kc.kubeService.GetContextDetail(contextName)
	if err != nil {
		http.Error(w, "Context 상세 정보 조회 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.ContextDetailResponse{}
	response.Success = true
	response.Message = "Context 상세 정보 조회 성공"
	response.Data = *contextDetail

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ApplyYaml - YAML 내용을 kubectl apply로 적용 (POST /api/apply)
func (kc *KubeController) ApplyYaml(w http.ResponseWriter, r *http.Request) {
	log.Println("🚀 POST /api/apply - YAML 적용 요청")

	var request model.ApplyYamlRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// YAML 내용 검증
	if strings.TrimSpace(request.YamlContent) == "" {
		http.Error(w, "YAML 내용은 필수입니다", http.StatusBadRequest)
		return
	}

	result, err := kc.kubeService.ApplyYaml(request)
	if err != nil {
		http.Error(w, "YAML 적용 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.ApplyYamlResponse{}
	response.Success = true
	if request.DryRun {
		response.Message = fmt.Sprintf("YAML dry-run 실행 완료: %s", result.Output)
	} else {
		response.Message = "YAML 적용 완료"
	}
	response.Data = *result

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteYaml - YAML 내용을 kubectl delete로 삭제 (POST /api/delete)
func (kc *KubeController) DeleteYaml(w http.ResponseWriter, r *http.Request) {
	log.Println("🗑️ POST /api/delete - YAML 삭제 요청")

	var request model.DeleteYamlRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// YAML 내용 검증
	if strings.TrimSpace(request.YamlContent) == "" {
		http.Error(w, "YAML 내용은 필수입니다", http.StatusBadRequest)
		return
	}

	result, err := kc.kubeService.DeleteYaml(request)
	if err != nil {
		http.Error(w, "YAML 삭제 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.ApplyYamlResponse{}
	response.Success = true
	response.Message = "YAML 삭제 완료"
	response.Data = *result

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
