package server

import (
	"fmt"
	"mockgitea/internal/config"
	"mockgitea/internal/models"
	"mockgitea/internal/utils"
	"net/http"
	"strings"
)

func (s *MockServer) HandleUser(w http.ResponseWriter, r *http.Request) {
	if !utils.EnsureGET(w, r) {
		return
	}
	utils.WriteJSON(w, http.StatusOK, s.CurrentUser)
}

func (s *MockServer) HandleReposearch(w http.ResponseWriter, r *http.Request) {
	if !utils.EnsureGET(w, r) {
		return
	}

	page := utils.ParsePositiveInt(r.URL.Query().Get("page"), config.DefaultPage)
	limit := utils.ParsePositiveInt(r.URL.Query().Get("limit"), config.DefaultLimit)
	ReposPage := utils.Paginate(s.Repos, page, limit)

	type ReposearchResponse struct {
		Data []models.GiteaRepo `json:"data"`
		OK   bool               `json:"ok"`
	}

	data := make([]models.GiteaRepo, 0, len(ReposPage))
	for _, repo := range ReposPage {
		data = append(data, repo.Repo)
	}

	utils.WriteJSON(w, http.StatusOK, ReposearchResponse{
		Data: data,
		OK:   true,
	})
}

func (s *MockServer) HandleRepoRoutes(w http.ResponseWriter, r *http.Request) {
	if !utils.EnsureGET(w, r) {
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/repos/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")

	switch {
	case len(parts) == 3 && parts[2] == "branches":
		s.HandleBranches(w, r, parts[0], parts[1])
	case len(parts) == 3 && parts[2] == "commits":
		s.handleCommits(w, r, parts[0], parts[1])
	case len(parts) == 5 && parts[2] == "git" && parts[3] == "commits":
		s.HandleSingleCommit(w, r, parts[0], parts[1], parts[4])
	default:
		s.HandleNotFound(w, r)
	}
}

func (s *MockServer) HandleBranches(w http.ResponseWriter, r *http.Request, owner, repoName string) {
	repo, ok := s.findRepo(owner, repoName)
	if !ok {
		utils.WriteJSON(w, http.StatusNotFound, map[string]string{"message": "Repository not found"})
		return
	}

	page := utils.ParsePositiveInt(r.URL.Query().Get("page"), config.DefaultPage)
	limit := utils.ParsePositiveInt(r.URL.Query().Get("limit"), config.DefaultLimit)
	branchesPage := utils.Paginate(repo.Branches, page, limit)

	data := make([]models.GiteaBranch, 0, len(branchesPage))
	for _, branch := range branchesPage {
		data = append(data, branch.Branch)
	}
	utils.WriteJSON(w, http.StatusOK, data)
}

func (s *MockServer) handleCommits(w http.ResponseWriter, r *http.Request, owner, repoName string) {
	repo, ok := s.findRepo(owner, repoName)
	if !ok {
		utils.WriteJSON(w, http.StatusNotFound, map[string]string{"message": "Repository not found"})
		return
	}

	sha := r.URL.Query().Get("sha")
	if sha == "" {
		sha = repo.Repo.DefaultBranch
	}

	branch, ok := repo.BranchByID[sha]
	if !ok {
		utils.WriteJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("branch %q not found", sha)})
		return
	}

	page := utils.ParsePositiveInt(r.URL.Query().Get("page"), config.DefaultPage)
	limit := utils.ParsePositiveInt(r.URL.Query().Get("limit"), config.DefaultLimit)
	utils.WriteJSON(w, http.StatusOK, utils.Paginate(branch.Commits, page, limit))
}

func (s *MockServer) HandleSingleCommit(w http.ResponseWriter, r *http.Request, owner, repoName, sha string) {
	if _, ok := s.findRepo(owner, repoName); !ok {
		utils.WriteJSON(w, http.StatusNotFound, map[string]string{"message": "Repository not found"})
		return
	}

	if commit, ok := s.CommitIndex[sha]; ok {
		utils.WriteJSON(w, http.StatusOK, commit)
		return
	}

	utils.WriteJSON(w, http.StatusOK, s.syntheticCommit(sha))
}

func (s *MockServer) HandleNotFound(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
}
