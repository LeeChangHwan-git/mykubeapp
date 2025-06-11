package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"mykubeapp/model"
	"mykubeapp/utils"
)

// GitService - Git 관련 서비스
type GitService struct {
	tempDir     string
	kubeService *KubeService
}

// NewGitService - Git 서비스 생성자
func NewGitService() *GitService {
	tempDir := filepath.Join(os.TempDir(), "kubectl-git-repos")
	os.MkdirAll(tempDir, 0755)

	return &GitService{
		tempDir:     tempDir,
		kubeService: NewKubeService(),
	}
}

// CloneRepository - Git 레포지토리 클론
func (gs *GitService) CloneRepository(repoURL, branch string) (string, error) {
	log.Printf("📦 Git 레포지토리 클론 시작: %s (branch: %s)", repoURL, branch)

	// 레포지토리 이름 추출
	repoName := gs.extractRepoName(repoURL)
	if repoName == "" {
		return "", fmt.Errorf("잘못된 레포지토리 URL: %s", repoURL)
	}

	// 클론 대상 디렉토리
	cloneDir := filepath.Join(gs.tempDir, fmt.Sprintf("%s_%d", repoName, time.Now().Unix()))

	// 기존 디렉토리가 있으면 삭제
	if utils.FileExists(cloneDir) {
		os.RemoveAll(cloneDir)
	}

	// git clone 명령어 구성
	args := []string{"clone"}

	// 브랜치 지정
	if branch != "" && branch != "main" && branch != "master" {
		args = append(args, "-b", branch)
	}

	// 얕은 클론으로 속도 향상
	args = append(args, "--depth", "1", repoURL, cloneDir)

	// git clone 실행
	_, err := utils.ExecuteCommand("git", args...)
	if err != nil {
		return "", fmt.Errorf("Git 클론 실패: %v", err)
	}

	log.Printf("✅ Git 레포지토리 클론 완료: %s", cloneDir)
	return cloneDir, nil
}

// FindYamlFiles - 디렉토리에서 YAML 파일 찾기
func (gs *GitService) FindYamlFiles(repoDir string) ([]model.GitYamlFile, error) {
	log.Printf("🔍 YAML 파일 검색: %s", repoDir)

	var yamlFiles []model.GitYamlFile

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// .git 디렉토리 스킵
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// YAML 파일 확인
		if !info.IsDir() && gs.isYamlFile(info.Name()) {
			relativePath, _ := filepath.Rel(repoDir, path)

			// 파일 내용 읽기
			content, err := ioutil.ReadFile(path)
			if err != nil {
				log.Printf("⚠️ 파일 읽기 실패 (스킵): %s - %v", path, err)
				return nil
			}

			// Kubernetes YAML인지 확인
			if gs.isKubernetesYaml(string(content)) {
				yamlFile := model.GitYamlFile{
					Path:         relativePath,
					FullPath:     path,
					Content:      string(content),
					Size:         info.Size(),
					IsKubernetes: true,
				}
				yamlFiles = append(yamlFiles, yamlFile)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("디렉토리 탐색 실패: %v", err)
	}

	log.Printf("✅ YAML 파일 검색 완료: %d개 발견", len(yamlFiles))
	return yamlFiles, nil
}

// GetSpecificYamlFile - 특정 YAML 파일 가져오기
func (gs *GitService) GetSpecificYamlFile(repoDir, filename string) (*model.GitYamlFile, error) {
	log.Printf("📄 특정 YAML 파일 검색: %s", filename)

	var foundFile *model.GitYamlFile

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// .git 디렉토리 스킵
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// 파일 이름 매칭 (대소문자 무시)
		if !info.IsDir() && strings.EqualFold(info.Name(), filename) {
			relativePath, _ := filepath.Rel(repoDir, path)

			content, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("파일 읽기 실패: %v", err)
			}

			foundFile = &model.GitYamlFile{
				Path:         relativePath,
				FullPath:     path,
				Content:      string(content),
				Size:         info.Size(),
				IsKubernetes: gs.isKubernetesYaml(string(content)),
			}
			return fmt.Errorf("found") // 찾았으므로 탐색 중단
		}

		return nil
	})

	if err != nil && err.Error() != "found" {
		return nil, fmt.Errorf("파일 검색 실패: %v", err)
	}

	if foundFile == nil {
		return nil, fmt.Errorf("파일을 찾을 수 없습니다: %s", filename)
	}

	log.Printf("✅ 파일 검색 완료: %s", foundFile.Path)
	return foundFile, nil
}

// ApplyYamlFromGit - Git에서 가져온 YAML 적용
func (gs *GitService) ApplyYamlFromGit(yamlFiles []model.GitYamlFile, namespace string, dryRun bool) (*model.GitApplyResult, error) {
	log.Printf("🚀 Git YAML 적용 시작 (파일 수: %d, DryRun: %t)", len(yamlFiles), dryRun)

	var results []model.GitFileApplyResult
	var allResources []string
	successCount := 0

	for _, yamlFile := range yamlFiles {
		log.Printf("📝 적용 중: %s", yamlFile.Path)

		// YAML 적용 요청 생성
		applyRequest := model.ApplyYamlRequest{
			YamlContent: yamlFile.Content,
			Namespace:   namespace,
			DryRun:      dryRun,
		}

		// YAML 적용
		applyResult, err := gs.kubeService.ApplyYaml(applyRequest)

		fileResult := model.GitFileApplyResult{
			FilePath: yamlFile.Path,
			Success:  err == nil,
		}

		if err != nil {
			fileResult.Error = err.Error()
			log.Printf("❌ 적용 실패 %s: %v", yamlFile.Path, err)
		} else {
			fileResult.Output = applyResult.Output
			fileResult.Resources = applyResult.Resources
			allResources = append(allResources, applyResult.Resources...)
			successCount++
			log.Printf("✅ 적용 성공 %s: %d개 리소스", yamlFile.Path, len(applyResult.Resources))
		}

		results = append(results, fileResult)
	}

	result := &model.GitApplyResult{
		TotalFiles:   len(yamlFiles),
		SuccessFiles: successCount,
		FailedFiles:  len(yamlFiles) - successCount,
		AppliedTime:  time.Now().Format("2006-01-02 15:04:05"),
		Results:      results,
		AllResources: gs.removeDuplicates(allResources),
		DryRun:       dryRun,
	}

	log.Printf("✅ Git YAML 적용 완료 (성공: %d/%d)", successCount, len(yamlFiles))
	return result, nil
}

// Cleanup - 임시 파일 정리
func (gs *GitService) Cleanup(repoDir string) error {
	if repoDir != "" && utils.FileExists(repoDir) {
		err := os.RemoveAll(repoDir)
		if err != nil {
			log.Printf("⚠️ 임시 디렉토리 삭제 실패: %v", err)
			return err
		}
		log.Printf("🧹 임시 디렉토리 삭제 완료: %s", repoDir)
	}
	return nil
}

// CleanupAll - 모든 임시 파일 정리
func (gs *GitService) CleanupAll() error {
	if utils.FileExists(gs.tempDir) {
		err := os.RemoveAll(gs.tempDir)
		if err != nil {
			return fmt.Errorf("전체 임시 디렉토리 삭제 실패: %v", err)
		}
		log.Printf("🧹 전체 임시 디렉토리 삭제 완료: %s", gs.tempDir)
	}
	return nil
}

// 유틸리티 메서드들

// extractRepoName - URL에서 레포지토리 이름 추출
func (gs *GitService) extractRepoName(repoURL string) string {
	// GitHub URL 패턴 매칭
	patterns := []string{
		`github\.com[/:]([^/]+)/([^/]+?)(\.git)?/?$`,
		`gitlab\.com[/:]([^/]+)/([^/]+?)(\.git)?/?$`,
		`bitbucket\.org[/:]([^/]+)/([^/]+?)(\.git)?/?$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(repoURL)
		if len(matches) >= 3 {
			return fmt.Sprintf("%s-%s", matches[1], matches[2])
		}
	}

	// 일반적인 패턴으로 추출
	if strings.Contains(repoURL, "/") {
		parts := strings.Split(strings.TrimRight(repoURL, "/"), "/")
		if len(parts) >= 2 {
			repoName := parts[len(parts)-1]
			repoName = strings.TrimSuffix(repoName, ".git")
			return repoName
		}
	}

	return ""
}

// isYamlFile - YAML 파일인지 확인
func (gs *GitService) isYamlFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".yaml" || ext == ".yml"
}

// isKubernetesYaml - Kubernetes YAML인지 확인
func (gs *GitService) isKubernetesYaml(content string) bool {
	// Kubernetes YAML의 기본 필드들 확인
	requiredFields := []string{
		"apiVersion:",
		"kind:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(content, field) {
			return false
		}
	}

	// Kubernetes 리소스 타입 확인
	k8sKinds := []string{
		"Pod", "Service", "Deployment", "ConfigMap", "Secret",
		"Ingress", "PersistentVolume", "PersistentVolumeClaim",
		"ServiceAccount", "Role", "RoleBinding", "ClusterRole",
		"ClusterRoleBinding", "Namespace", "DaemonSet", "StatefulSet",
		"Job", "CronJob", "HorizontalPodAutoscaler", "NetworkPolicy",
	}

	for _, kind := range k8sKinds {
		if strings.Contains(content, "kind: "+kind) {
			return true
		}
	}

	return false
}

// removeDuplicates - 중복 제거
func (gs *GitService) removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
