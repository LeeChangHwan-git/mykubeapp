package controller

import (
	"log"
	"mykubeapp/terminal"
	"net/http"
)

// TerminalController - 터미널 관련 컨트롤러
type TerminalController struct{}

// NewTerminalController - 터미널 컨트롤러 생성자
func NewTerminalController() *TerminalController {
	return &TerminalController{}
}

// KubectlTerminal - kubectl 웹터미널 핸들러
func (tc *TerminalController) KubectlTerminal(w http.ResponseWriter, r *http.Request) {
	log.Println("🖥️  Kubectl 터미널 연결 요청")
	terminal.KubectlTerminalHandler(w, r)
}
