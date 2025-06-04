package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileExists - 파일 존재 여부 확인
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// ReadFile - 파일 내용 읽기
func ReadFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile - 파일에 내용 쓰기
func WriteFile(filename, content string) error {
	// 디렉토리가 없으면 생성
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %v", err)
	}

	// 파일 쓰기
	return ioutil.WriteFile(filename, []byte(content), 0644)
}

// ExecuteCommand - 외부 명령어 실행
func ExecuteCommand(name string, args ...string) (string, error) {
	log.Printf("🔧 명령어 실행: %s %s", name, strings.Join(args, " "))

	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("❌ 명령어 실행 실패: %v", err)
		log.Printf("📄 출력: %s", string(output))
		return "", fmt.Errorf("명령어 실행 실패: %v, 출력: %s", err, string(output))
	}

	result := string(output)
	log.Printf("✅ 명령어 실행 성공")
	log.Printf("📄 출력: %s", result)

	return result, nil
}

// IsKubectlAvailable - kubectl 명령어 사용 가능 여부 확인
func IsKubectlAvailable() bool {
	_, err := exec.LookPath("kubectl")
	return err == nil
}

// GetHomeDir - 홈 디렉토리 경로 반환
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// GetKubeConfigPath - kube config 파일 경로 반환
func GetKubeConfigPath() (string, error) {
	// 환경변수 KUBECONFIG 확인
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig, nil
	}

	// 기본 경로 ($HOME/.kube/config)
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".kube", "config"), nil
}

// BackupFile - 파일 백업
func BackupFile(filename string) error {
	if !FileExists(filename) {
		return fmt.Errorf("백업할 파일이 존재하지 않습니다: %s", filename)
	}

	backupPath := filename + ".backup"
	content, err := ReadFile(filename)
	if err != nil {
		return fmt.Errorf("원본 파일 읽기 실패: %v", err)
	}

	err = WriteFile(backupPath, content)
	if err != nil {
		return fmt.Errorf("백업 파일 생성 실패: %v", err)
	}

	log.Printf("✅ 파일 백업 완료: %s -> %s", filename, backupPath)
	return nil
}
