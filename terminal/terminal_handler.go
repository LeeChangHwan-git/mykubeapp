package terminal

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket 업그레이더 설정
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// TerminalSession - 터미널 세션 정보
type TerminalSession struct {
	ID          string
	Conn        *websocket.Conn
	Cmd         *exec.Cmd
	Stdin       io.WriteCloser
	Stdout      io.ReadCloser
	Stderr      io.ReadCloser
	Mutex       sync.Mutex
	IsClosed    bool
	InputBuffer string // 입력 버퍼 추가
}

// KubectlTerminalHandler - kubectl 전용 웹터미널 핸들러
func KubectlTerminalHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🖥️  새로운 kubectl 터미널 연결 요청")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket 업그레이드 실패: %v", err)
		return
	}
	defer conn.Close()

	session := &TerminalSession{
		ID:          generateSessionID(),
		Conn:        conn,
		InputBuffer: "",
	}

	log.Printf("✅ kubectl 터미널 세션 시작: %s", session.ID)

	// 환영 메시지 전송 (순수 텍스트)
	welcomeMsg := "🚀 Kubectl Terminal Connected!\r\n" +
		"💡 Type kubectl commands directly. Example: kubectl get pods\r\n" +
		"📝 Available commands: kubectl, get, describe, logs, apply, delete, etc.\r\n\r\n" +
		"kubectl> "
	session.SendMessage(welcomeMsg)

	// 메시지 처리 루프
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket 읽기 오류: %v", err)
			}
			break
		}

		// 입력 메시지 처리
		input := string(message)
		log.Printf("🔤 입력 수신: %q", input) // 디버깅용 로그
		session.HandleInput(input)
	}

	log.Printf("🔌 kubectl 터미널 세션 종료: %s", session.ID)
}

// HandleInput - 사용자 입력 처리 (개선됨)
func (s *TerminalSession) HandleInput(input string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.IsClosed {
		return
	}

	// 각 문자 처리
	for _, char := range input {
		switch char {
		case '\r', '\n': // Enter 키
			if s.InputBuffer != "" {
				s.ExecuteCommand(strings.TrimSpace(s.InputBuffer))
				s.InputBuffer = ""
			}
			s.SendMessage("\r\nkubectl> ")

		case '\b', 127: // Backspace 또는 Delete
			if len(s.InputBuffer) > 0 {
				s.InputBuffer = s.InputBuffer[:len(s.InputBuffer)-1]
				s.SendMessage("\b \b") // 백스페이스 효과
			}

		case 3: // Ctrl+C
			s.SendMessage("\r\n^C\r\nkubectl> ")
			s.InputBuffer = ""

		default:
			// 일반 문자
			if char >= 32 && char <= 126 { // 출력 가능한 ASCII 문자
				s.InputBuffer += string(char)
				s.SendMessage(string(char)) // 에코
			}
		}
	}
}

// ExecuteCommand - 명령어 실행 (기존과 동일)
func (s *TerminalSession) ExecuteCommand(command string) {
	log.Printf("🔧 명령어 실행: %s", command)

	if strings.TrimSpace(command) == "" {
		return
	}

	// 특별한 명령어 처리
	switch command {
	case "clear", "cls":
		s.SendMessage("\033[2J\033[H")
		return
	case "exit", "quit":
		s.SendMessage("\r\n👋 Terminal session ended.\r\n")
		s.Close()
		return
	case "help":
		s.ShowHelp()
		return
	}

	// kubectl 명령어가 아닌 경우 자동으로 kubectl 추가
	if !strings.HasPrefix(command, "kubectl") && !isBuiltinCommand(command) {
		command = "kubectl " + command
	}

	s.runCommand(command)
}

// runCommand - 실제 명령어 실행 (기존과 동일)
func (s *TerminalSession) runCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.SendMessage(fmt.Sprintf("\r\n❌ 명령어 실행 실패: %v\r\n", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		s.SendMessage(fmt.Sprintf("\r\n❌ 명령어 실행 실패: %v\r\n", err))
		return
	}

	if err := cmd.Start(); err != nil {
		s.SendMessage(fmt.Sprintf("\r\n❌ 명령어 시작 실패: %v\r\n", err))
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// stdout 읽기
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			s.SendMessage("\r\n" + scanner.Text())
		}
	}()

	// stderr 읽기
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			s.SendMessage("\r\n🔥 " + scanner.Text())
		}
	}()

	// 명령어 완료 대기
	go func() {
		wg.Wait()
		cmd.Wait()
		s.SendMessage("\r\n✅ Command completed\r")
	}()
}

// ShowHelp - 도움말 표시
func (s *TerminalSession) ShowHelp() {
	helpText := `
📚 Available Commands:
  kubectl get pods               - List all pods
  kubectl get services           - List all services
  kubectl get deployments       - List all deployments
  kubectl describe pod <name>    - Describe a pod
  kubectl logs <pod-name>        - Show pod logs
  kubectl apply -f <file>        - Apply configuration
  kubectl delete pod <name>      - Delete a pod
  
🔧 Terminal Commands:
  clear, cls                     - Clear screen
  help                          - Show this help
  exit, quit                    - Close terminal
  
💡 Tips:
  - You can omit 'kubectl' prefix (e.g., just type 'get pods')
  - Press Enter to execute commands
  - Use Ctrl+C to cancel running commands
`
	s.SendMessage(helpText)
}

// SendMessage - 메시지 전송 (순수 텍스트)
func (s *TerminalSession) SendMessage(data string) {
	if s.IsClosed {
		return
	}

	s.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := s.Conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
		log.Printf("❌ 메시지 전송 실패: %v", err)
		s.Close()
	}
}

// Close - 세션 종료
func (s *TerminalSession) Close() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.IsClosed {
		return
	}

	s.IsClosed = true

	if s.Cmd != nil && s.Cmd.Process != nil {
		s.Cmd.Process.Kill()
	}

	if s.Stdin != nil {
		s.Stdin.Close()
	}

	s.Conn.Close()
}

// isBuiltinCommand - 내장 명령어 확인
func isBuiltinCommand(command string) bool {
	builtins := []string{"clear", "cls", "help", "exit", "quit"}
	for _, builtin := range builtins {
		if command == builtin {
			return true
		}
	}
	return false
}

// generateSessionID - 세션 ID 생성
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
