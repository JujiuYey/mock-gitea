package server

import (
	"fmt"
	"math/rand"
	"mockgitea/internal/config"
	"mockgitea/internal/data"
	"mockgitea/internal/models"
	"mockgitea/internal/utils"
	"time"
)

type MockServer struct {
	Users       []models.GiteaUser
	CurrentUser models.GiteaUser
	Repos       []models.MockRepo
	RepoIndex   map[string]*models.MockRepo
	CommitIndex map[string]models.GiteaCommit
	AnchorTime  time.Time
}

func NewMockServer() *MockServer {
	users := data.BuildUsers()
	anchor := time.Now().UTC().Truncate(time.Minute)
	repos, repoIndex, commitIndex := buildRepos(users, anchor)

	return &MockServer{
		Users:       users,
		CurrentUser: users[0],
		Repos:       repos,
		RepoIndex:   repoIndex,
		CommitIndex: commitIndex,
		AnchorTime:  anchor,
	}
}

func (s *MockServer) findRepo(owner, repoName string) (*models.MockRepo, bool) {
	repo, ok := s.RepoIndex[fmt.Sprintf("%s/%s", owner, repoName)]
	return repo, ok
}

func (s *MockServer) syntheticCommit(sha string) models.GiteaCommit {
	rng := rand.New(rand.NewSource(utils.StableSeed("single", sha)))
	userIndex := rng.Intn(len(s.Users))
	author := s.Users[userIndex]
	messagePool := data.BuildCommitMessages()
	message := fmt.Sprintf("%s [mock lookup]", messagePool[rng.Intn(len(messagePool))])
	timestamp := s.AnchorTime.Add(-time.Duration(rng.Intn(config.MaxRecentDays*24)) * time.Hour).Truncate(time.Second)
	additions, deletions := randomStats(rng, message)

	return models.GiteaCommit{
		SHA: sha,
		Commit: models.GiteaCommitInfo{
			Message: message,
			Author: models.GiteaCommitUser{
				Name:  author.FullName,
				Email: author.Email,
				Date:  timestamp,
			},
			Committer: models.GiteaCommitUser{
				Name:  author.FullName,
				Email: author.Email,
				Date:  timestamp.Add(12 * time.Minute),
			},
		},
		Stats: &models.GiteaStats{
			Additions: additions,
			Deletions: deletions,
			Total:     additions + deletions,
		},
		Author: &s.Users[userIndex],
	}
}
