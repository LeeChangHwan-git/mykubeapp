package service

import (
	"fmt"
	"gopkg.in/yaml.v2" // YAML 파싱을 위해 추가 필요
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"mykubeapp/model"
	"mykubeapp/utils"
)

// KubeService - Spring의 @Service와 유사한 역할
type KubeService struct {
	configPath string
}

// NewKubeService - 서비스 생성자
func NewKubeService() *KubeService {
	// 홈 디렉토리의 .kube/config 경로 설정
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("⚠️  홈 디렉토리를 찾을 수 없습니다: %v", err)
		homeDir = "."
	}

	configPath := filepath.Join(homeDir, ".kube", "config")
	log.Printf("🔧 Kube config 경로: %s", configPath)

	return &KubeService{
		configPath: configPath,
	}
}

// GetCurrentConfig - 현재 kube config 파일 내용 반환
func (ks *KubeService) GetCurrentConfig() (string, error) {
	log.Printf("📖 Config 파일 읽기: %s", ks.configPath)

	// 파일 존재 여부 확인
	if !utils.FileExists(ks.configPath) {
		return "", fmt.Errorf("kube config 파일이 존재하지 않습니다: %s", ks.configPath)
	}

	// 파일 내용 읽기
	content, err := utils.ReadFile(ks.configPath)
	if err != nil {
		return "", fmt.Errorf("config 파일 읽기 실패: %v", err)
	}

	log.Printf("✅ Config 파일 읽기 성공 (크기: %d bytes)", len(content))
	return content, nil
}

// AddConfig - kubectl 명령어를 사용하여 새로운 config 추가
func (ks *KubeService) AddConfig(request model.AddConfigRequest) error {
	log.Printf("📝 Config 추가 요청: %s", request.ClusterName)

	// 기존 config 백업
	if utils.FileExists(ks.configPath) {
		if err := utils.BackupFile(ks.configPath); err != nil {
			log.Printf("⚠️  백업 실패 (계속 진행): %v", err)
		}
	}

	// kubectl 명령어를 사용하여 클러스터 추가
	err := ks.addClusterConfig(request)
	if err != nil {
		return fmt.Errorf("클러스터 설정 추가 실패: %v", err)
	}

	// 사용자 자격 증명 추가
	err = ks.addUserConfig(request)
	if err != nil {
		return fmt.Errorf("사용자 설정 추가 실패: %v", err)
	}

	// 컨텍스트 추가
	err = ks.addContextConfig(request)
	if err != nil {
		return fmt.Errorf("컨텍스트 설정 추가 실패: %v", err)
	}

	log.Printf("✅ Config 추가 완료: %s", request.ClusterName)
	return nil
}

// addClusterConfig - 클러스터 설정 추가
func (ks *KubeService) addClusterConfig(request model.AddConfigRequest) error {
	log.Printf("🔧 클러스터 설정 추가: %s", request.ClusterName)

	// kubectl config set-cluster 명령 실행
	args := []string{
		"config", "set-cluster", request.ClusterName,
		"--server=" + request.Server,
	}

	// 인증서 검증 스킵 (개발용)
	args = append(args, "--insecure-skip-tls-verify=true")

	_, err := utils.ExecuteCommand("kubectl", args...)
	if err != nil {
		return fmt.Errorf("클러스터 설정 실패: %v", err)
	}

	log.Printf("✅ 클러스터 설정 완료: %s", request.ClusterName)
	return nil
}

// addUserConfig - 사용자 설정 추가
func (ks *KubeService) addUserConfig(request model.AddConfigRequest) error {
	log.Printf("🔧 사용자 설정 추가: %s", request.User)

	// 토큰이 있으면 토큰 기반 인증 설정
	if request.Token != "" {
		_, err := utils.ExecuteCommand("kubectl", "config", "set-credentials", request.User, "--token="+request.Token)
		if err != nil {
			return fmt.Errorf("토큰 기반 사용자 설정 실패: %v", err)
		}
	} else {
		// 토큰이 없으면 기본 사용자만 생성
		_, err := utils.ExecuteCommand("kubectl", "config", "set-credentials", request.User)
		if err != nil {
			return fmt.Errorf("기본 사용자 설정 실패: %v", err)
		}
	}

	log.Printf("✅ 사용자 설정 완료: %s", request.User)
	return nil
}

// addContextConfig - 컨텍스트 설정 추가
func (ks *KubeService) addContextConfig(request model.AddConfigRequest) error {
	log.Printf("🔧 컨텍스트 설정 추가: %s", request.ContextName)

	_, err := utils.ExecuteCommand("kubectl", "config", "set-context", request.ContextName,
		"--cluster="+request.ClusterName,
		"--user="+request.User)
	if err != nil {
		return fmt.Errorf("컨텍스트 설정 실패: %v", err)
	}

	log.Printf("✅ 컨텍스트 설정 완료: %s", request.ContextName)
	return nil
}

// GetContexts - kubectl config get-contexts 실행하여 context 목록 반환
func (ks *KubeService) GetContexts() ([]model.ContextInfo, error) {
	log.Println("📋 Context 목록 조회 중...")

	// kubectl config get-contexts 명령 실행 (이름만)
	output, err := utils.ExecuteCommand("kubectl", "config", "get-contexts", "--output=name")
	if err != nil {
		return nil, fmt.Errorf("kubectl 명령 실행 실패: %v", err)
	}

	// 현재 context 조회
	currentContext, err := utils.ExecuteCommand("kubectl", "config", "current-context")
	if err != nil {
		log.Printf("⚠️  현재 context 조회 실패: %v", err)
		currentContext = ""
	}
	currentContext = strings.TrimSpace(currentContext)

	// 결과 파싱
	var contexts []model.ContextInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			context := model.ContextInfo{
				Name:      line,
				IsCurrent: line == currentContext,
			}
			contexts = append(contexts, context)
		}
	}

	log.Printf("✅ Context 목록 조회 완료 (총 %d개)", len(contexts))
	return contexts, nil
}

// UseContext - 특정 context 사용 설정
func (ks *KubeService) UseContext(contextName string) error {
	log.Printf("🔄 Context 변경: %s", contextName)

	// kubectl config use-context 명령 실행
	_, err := utils.ExecuteCommand("kubectl", "config", "use-context", contextName)
	if err != nil {
		return fmt.Errorf("context 변경 실패: %v", err)
	}

	log.Printf("✅ Context 변경 완료: %s", contextName)
	return nil
}

// DeleteContext - 특정 context 삭제
func (ks *KubeService) DeleteContext(contextName string) error {
	log.Printf("🗑️ Context 삭제 요청: %s", contextName)

	// 컨텍스트 이름 검증
	if strings.TrimSpace(contextName) == "" {
		return fmt.Errorf("컨텍스트 이름이 비어있습니다")
	}

	// 현재 사용 중인 컨텍스트인지 확인
	currentContext, err := utils.ExecuteCommand("kubectl", "config", "current-context")
	if err == nil {
		currentContext = strings.TrimSpace(currentContext)
		if currentContext == contextName {
			return fmt.Errorf("현재 사용 중인 컨텍스트는 삭제할 수 없습니다: %s", contextName)
		}
	}

	// 컨텍스트 존재 여부 확인
	contexts, err := ks.GetContexts()
	if err != nil {
		return fmt.Errorf("컨텍스트 목록 조회 실패: %v", err)
	}

	contextExists := false
	for _, ctx := range contexts {
		if ctx.Name == contextName {
			contextExists = true
			break
		}
	}

	if !contextExists {
		return fmt.Errorf("존재하지 않는 컨텍스트입니다: %s", contextName)
	}

	// 기존 config 백업
	if utils.FileExists(ks.configPath) {
		if err := utils.BackupFile(ks.configPath); err != nil {
			log.Printf("⚠️  백업 실패 (계속 진행): %v", err)
		}
	}

	// kubectl config delete-context 명령 실행
	_, err = utils.ExecuteCommand("kubectl", "config", "delete-context", contextName)
	if err != nil {
		return fmt.Errorf("컨텍스트 삭제 실패: %v", err)
	}

	log.Printf("✅ Context 삭제 완료: %s", contextName)
	return nil
}

// GetContextDetail - 특정 context의 상세 정보 조회
func (ks *KubeService) GetContextDetail(contextName string) (*model.ContextDetail, error) {
	log.Printf("📋 Context 상세 정보 조회: %s", contextName)

	// 컨텍스트 이름 검증
	if strings.TrimSpace(contextName) == "" {
		return nil, fmt.Errorf("컨텍스트 이름이 비어있습니다")
	}

	// kube config 파일 읽기
	configContent, err := ks.GetCurrentConfig()
	if err != nil {
		return nil, fmt.Errorf("config 파일 읽기 실패: %v", err)
	}

	// YAML 파싱
	var kubeConfig model.KubeConfig
	if err := yaml.Unmarshal([]byte(configContent), &kubeConfig); err != nil {
		return nil, fmt.Errorf("config 파싱 실패: %v", err)
	}

	// 현재 컨텍스트 확인
	currentContext := strings.TrimSpace(kubeConfig.CurrentContext)

	// 요청한 컨텍스트 찾기
	var targetContext *model.ContextConfig
	for _, ctx := range kubeConfig.Contexts {
		if ctx.Name == contextName {
			targetContext = &ctx
			break
		}
	}

	if targetContext == nil {
		return nil, fmt.Errorf("컨텍스트를 찾을 수 없습니다: %s", contextName)
	}

	// 클러스터 정보 찾기
	var clusterDetail model.ClusterDetail
	for _, cluster := range kubeConfig.Clusters {
		if cluster.Name == targetContext.Context.Cluster {
			clusterDetail = model.ClusterDetail{
				Name:                    cluster.Name,
				Server:                  cluster.Cluster.Server,
				InsecureSkipTLSVerify:   cluster.Cluster.InsecureSkipTLSVerify,
				HasCertificateAuthority: cluster.Cluster.CertificateAuthorityData != "",
			}
			break
		}
	}

	// 사용자 정보 찾기
	var userDetail model.UserDetail
	for _, user := range kubeConfig.Users {
		if user.Name == targetContext.Context.User {
			authMethod := ks.determineAuthMethod(user.User)
			userDetail = model.UserDetail{
				Name:                 user.Name,
				HasToken:             user.User.Token != "",
				HasClientCertificate: user.User.ClientCertificateData != "",
				HasClientKey:         user.User.ClientKeyData != "",
				AuthenticationMethod: authMethod,
			}
			break
		}
	}

	// 컨텍스트 상세 정보 구성
	contextDetail := &model.ContextDetail{
		Name:      contextName,
		IsCurrent: contextName == currentContext,
		Cluster:   clusterDetail,
		User:      userDetail,
		Namespace: targetContext.Context.Namespace,
	}

	log.Printf("✅ Context 상세 정보 조회 완료: %s", contextName)
	return contextDetail, nil
}

// determineAuthMethod - 인증 방식 결정
func (ks *KubeService) determineAuthMethod(user model.UserConfigData) string {
	if user.Token != "" {
		return "Token"
	}
	if user.ClientCertificateData != "" && user.ClientKeyData != "" {
		return "Client Certificate"
	}
	if user.ClientCertificateData != "" {
		return "Certificate Only"
	}
	return "None"
}

// ApplyYaml - YAML 내용을 kubectl apply로 적용
func (ks *KubeService) ApplyYaml(request model.ApplyYamlRequest) (*model.ApplyYamlResult, error) {
	log.Printf("🚀 YAML 적용 시작 (DryRun: %t)", request.DryRun)

	// 임시 파일 생성
	tempFile, err := ks.createTempYamlFile(request.YamlContent)
	if err != nil {
		return nil, fmt.Errorf("임시 파일 생성 실패: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tempFile) // 함수 종료 시 임시 파일 삭제

	// kubectl apply 명령어 구성
	args := []string{"apply", "-f", tempFile}

	// 네임스페이스 지정
	if request.Namespace != "" {
		args = append(args, "-n", request.Namespace)
	}

	// dry-run 모드
	if request.DryRun {
		args = append(args, "--dry-run=client")
	}

	// 상세 출력
	args = append(args, "-v=0")

	// kubectl 명령 실행
	output, err := utils.ExecuteCommand("kubectl", args...)
	if err != nil {
		return nil, fmt.Errorf("kubectl apply 실패: %v", err)
	}

	// 적용된 리소스 목록 추출
	resources := ks.extractResourcesFromOutput(output)

	result := &model.ApplyYamlResult{
		Output:      output,
		AppliedTime: time.Now().Format("2006-01-02 15:04:05"),
		Resources:   resources,
		DryRun:      request.DryRun,
	}

	if request.DryRun {
		log.Printf("✅ YAML dry-run 완료")
	} else {
		log.Printf("✅ YAML 적용 완료 (리소스 수: %d)", len(resources))
	}

	return result, nil
}

// DeleteYaml - YAML 내용을 kubectl delete로 삭제
func (ks *KubeService) DeleteYaml(request model.DeleteYamlRequest) (*model.ApplyYamlResult, error) {
	log.Printf("🗑️ YAML 삭제 시작")

	// 임시 파일 생성
	tempFile, err := ks.createTempYamlFile(request.YamlContent)
	if err != nil {
		return nil, fmt.Errorf("임시 파일 생성 실패: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tempFile) // 함수 종료 시 임시 파일 삭제

	// kubectl delete 명령어 구성
	args := []string{"delete", "-f", tempFile}

	// 네임스페이스 지정
	if request.Namespace != "" {
		args = append(args, "-n", request.Namespace)
	}

	// 리소스가 없어도 에러 무시
	args = append(args, "--ignore-not-found=true")

	// kubectl 명령 실행
	output, err := utils.ExecuteCommand("kubectl", args...)
	if err != nil {
		return nil, fmt.Errorf("kubectl delete 실패: %v", err)
	}

	// 삭제된 리소스 목록 추출
	resources := ks.extractResourcesFromOutput(output)

	result := &model.ApplyYamlResult{
		Output:      output,
		AppliedTime: time.Now().Format("2006-01-02 15:04:05"),
		Resources:   resources,
		DryRun:      false,
	}

	log.Printf("✅ YAML 삭제 완료 (리소스 수: %d)", len(resources))
	return result, nil
}

// createTempYamlFile - 임시 YAML 파일 생성
func (ks *KubeService) createTempYamlFile(yamlContent string) (string, error) {
	// 임시 디렉토리에 파일 생성
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("kubectl-apply-%d.yaml", time.Now().UnixNano()))

	// YAML 내용을 파일에 쓰기
	err := os.WriteFile(tempFile, []byte(yamlContent), 0644)
	if err != nil {
		return "", fmt.Errorf("임시 파일 쓰기 실패: %v", err)
	}

	log.Printf("📝 임시 YAML 파일 생성: %s", tempFile)
	return tempFile, nil
}

// extractResourcesFromOutput - kubectl 출력에서 리소스 목록 추출
func (ks *KubeService) extractResourcesFromOutput(output string) []string {
	var resources []string

	// kubectl 출력에서 "리소스타입/이름 action" 패턴 찾기
	// 예: "deployment.apps/my-app created", "service/my-service configured"
	re := regexp.MustCompile(`([a-zA-Z0-9.\-/]+)\s+(created|configured|unchanged|deleted)`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			resources = append(resources, match[1])
		}
	}

	// 중복 제거
	seen := make(map[string]bool)
	var uniqueResources []string
	for _, resource := range resources {
		if !seen[resource] {
			seen[resource] = true
			uniqueResources = append(uniqueResources, resource)
		}
	}

	return uniqueResources
}

// ValidateYaml - YAML 구문 검증 (선택적으로 사용 가능)
func (ks *KubeService) ValidateYaml(yamlContent string) error {
	// 기본적인 YAML 구문 검증
	var temp interface{}
	err := yaml.Unmarshal([]byte(yamlContent), &temp)
	if err != nil {
		return fmt.Errorf("잘못된 YAML 형식: %v", err)
	}
	return nil
}
