package git

import "github.com/google/go-github/v57/github"

// 사용자의 레포지토리 목록 가져오기
func (c *Client) GetUserRepos(username string) ([]*github.Repository, error) {
	opt := &github.RepositoryListOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 10},
	}

	repos, _, err := c.client.Repositories.List(c.ctx, username, opt)
	return repos, err
}

// 레포지토리 정보 가져오기
func (c *Client) GetRepo(owner, repo string) (*github.Repository, error) {
	repository, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
	return repository, err
}
