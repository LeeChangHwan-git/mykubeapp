package git

import (
	"context"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

// GitHub 클라이언트 생성
func NewClient(token string) *Client {
	ctx := context.Background()

	var client *github.Client
	if token != "" {
		// 토큰이 있으면 인증된 클라이언트
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		// 토큰이 없으면 공개 API만 사용
		client = github.NewClient(nil)
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

// 사용자 정보 가져오기
func (c *Client) GetUser(username string) (*github.User, error) {
	user, _, err := c.client.Users.Get(c.ctx, username)
	return user, err
}
