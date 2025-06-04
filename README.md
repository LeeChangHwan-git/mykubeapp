# 🚀 쿠버네티스 관리 Go 애플리케이션

Spring Boot 스타일로 구현된 쿠버네티스 config 관리 백엔드 애플리케이션입니다.

## 📁 프로젝트 구조

```
mykubeapp/
├── main.go                    # 애플리케이션 진입점
├── go.mod                     # Go 모듈 파일
├── controller/
│   └── kube_controller.go     # REST API 컨트롤러
├── service/
│   └── kube_service.go        # 비즈니스 로직
├── model/
│   └── kube_model.go          # 데이터 모델/DTO
└── utils/
    └── kube_utils.go          # 유틸리티 함수
```

## 🛠️ 설치 및 실행

### 1. 프로젝트 초기화
```bash
# 프로젝트 폴더 생성
mkdir mykubeapp
cd mykubeapp

# Go 모듈 초기화
go mod init mykubeapp

# 의존성 설치
go mod tidy
```

### 2. 애플리케이션 실행
```bash
go run main.go
```

### 3. 빌드
```bash
# 실행 파일 생성
go build -o mykubeapp main.go

# 실행
./mykubeapp
```

## 📚 API 엔드포인트

| Method | URL | 설명 |
|--------|-----|------|
| GET | `/health` | 헬스 체크 |
| GET | `/api/config` | 현재 kube config 조회 |
| POST | `/api/config` | 새로운 config 추가 |
| GET | `/api/contexts` | context 목록 조회 |
| POST | `/api/context/use` | context 변경 |

## 📝 API 사용 예제

### 1. 헬스 체크
```bash
curl http://localhost:8080/health
```

### 2. 현재 config 조회
```bash
curl http://localhost:8080/api/config
```

### 3. 새로운 config 추가
```bash
curl -X POST http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{
    "clusterName": "my-cluster",
    "server": "https://kubernetes.default.svc",
    "contextName": "my-context",
    "user": "my-user",
    "token": "your-token-here"
  }'
```

### 4. context 목록 조회
```bash
curl http://localhost:8080/api/contexts
```

### 5. context 변경
```bash
curl -X POST http://localhost:8080/api/context/use \
  -H "Content-Type: application/json" \
  -d '{
    "contextName": "my-context"
  }'
```

## 🔧 주요 기능

- **Config 관리**: `~/.kube/config` 파일 읽기/쓰기
- **Context 관리**: kubectl 명령어를 통한 context 조회 및 변경
- **REST API**: Spring Boot 스타일의 컨트롤러 패턴
- **구조화된 코드**: Controller → Service → Utils 계층 분리

## 📋 사전 요구사항

- Go 1.21 이상
- kubectl 명령어 설치
- 기존 Kubernetes config 파일 (`~/.kube/config`)

## 🚨 주의사항

1. **kubectl 필수**: `kubectl` 명령어가 시스템에 설치되어 있어야 합니다.
2. **권한 확인**: config 파일 읽기/쓰기 권한이 필요합니다.
3. **백업 권장**: 중요한 config 파일은 미리 백업하세요.

## 🔄 다음 단계

1. **YAML 파싱**: 현재는 단순 문자열 조작이므로 `gopkg.in/yaml.v2` 사용하여 정확한 YAML 파싱 구현
2. **에러 처리**: 더 세밀한 에러 처리 및 검증 로직 추가
3. **테스트**: 단위 테스트 및 통합 테스트 작성
4. **프론트엔드**: React 또는 Vue.js로 웹 인터페이스 구현
5. **도커화**: Dockerfile 및 docker-compose.yml 추가

## 🐛 문제 해결

### kubectl 명령어 오류
```bash
# kubectl 설치 확인
kubectl version --client

# PATH 확인
echo $PATH
```

### 권한 오류
```bash
# config 파일 권한 확인
ls -la ~/.kube/config

# 권한 수정 (필요시)
chmod 600 ~/.kube/config
```