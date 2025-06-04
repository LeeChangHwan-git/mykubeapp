package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
