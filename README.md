# 🚀 쿠버네티스 관리 Go 애플리케이션

Spring Boot 스타일로 구현된 쿠버네티스 config 관리 백엔드 애플리케이션입니다.

## AI 연계(DeepSeek-Coder)
# 1. Ollama 컨테이너 실행
docker run -d -p 11434:11434 --name ollama ollama/ollama:latest

# 2. 균형잡힌 성능의 모델 다운로드 (권장)
docker exec -it ollama ollama pull deepseek-coder-v2:16b

# 3. 환경변수 설정
export DEEPSEEK_URL=http://localhost:11434