package server

import (
	"fmt"
	"log"
	"math/rand"
	"mockgitea/internal/config"
	"mockgitea/internal/data"
	"mockgitea/internal/models"
	"mockgitea/internal/utils"
	"sort"
	"strings"
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
	server := &MockServer{
		Users:       data.BuildUsers(),
		RepoIndex:   make(map[string]*models.MockRepo, config.TotalRepos),
		CommitIndex: make(map[string]models.GiteaCommit, config.TotalRepos*8*config.CommitsPerBranch),
		AnchorTime:  time.Now().UTC().Truncate(time.Minute),
	}

	server.CurrentUser = server.Users[0]
	server.Repos = server.buildRepos()

	return server
}

func (s *MockServer) buildRepos() []models.MockRepo {
	// 第一步，定义仓库蓝图
	blueprints := data.BuildRepoBluePrint()

	// 第二步：校验数量
	if len(blueprints) != config.TotalRepos {
		log.Fatalf("[mock-gitea] expected %d Repos, got %d", config.TotalRepos, len(blueprints))
	}

	// 第三步，构建用户映射
	// 创建一个 map: login -> GiteaUser，方便后续根据 login 查找用户
	// 这里用到 s.Users（在 buildUsers() 中初始化的 10 个用户）
	userByLogin := make(map[string]models.GiteaUser, len(s.Users))
	for _, user := range s.Users {
		userByLogin[user.Login] = user
	}

	// 第四部，创建空切片
	// 创建一个空的 mockRepo 切片，准备存放 30 个仓库
	Repos := make([]models.MockRepo, 0, len(blueprints))

	// 第五步，遍历蓝图，创建仓库

	for idx, blueprint := range blueprints {
		// 5.1 根据 OwnerLogin 查找用户作为仓库所有者
		owner, ok := userByLogin[blueprint.OwnerLogin]
		if !ok {
			log.Fatalf("[mock-gitea] unknown repo owner %q", blueprint.OwnerLogin)
		}

		// 5.2 用稳定种子创建随机数生成器（保证每次运行结果一致）
		Reposeed := utils.StableSeed("repo", blueprint.OwnerLogin, blueprint.Name, blueprint.Domain)
		rng := rand.New(rand.NewSource(Reposeed))

		// 5.3 生成创建时间和更新时间（随机偏移）
		// s.AnchorTime 是当前时间，作为基准点
		createdAt := s.AnchorTime.AddDate(0, 0, -(90 + rng.Intn(240))).Truncate(time.Second)
		updatedAt := s.AnchorTime.Add(-time.Duration(rng.Intn(72)) * time.Hour).Truncate(time.Second)

		// 5.4 创建 mockRepo 结构体
		repo := models.MockRepo{
			Repo: models.GiteaRepo{
				ID:              int64(1000 + idx + 1),                                      // 仓库 ID 从 1001 开始
				Owner:           owner,                                                      // 所有者用户
				Name:            blueprint.Name,                                             // 仓库名
				FullName:        fmt.Sprintf("%s/%s", blueprint.OwnerLogin, blueprint.Name), // 完整名称
				Description:     blueprint.Description,                                      // 描述
				DefaultBranch:   "main",                                                     // 默认分支
				StarsCount:      8 + rng.Intn(240),                                          // 随机星标数 8-248
				ForksCount:      1 + rng.Intn(48),                                           // 随机分叉数 1-49
				OpenIssuesCount: rng.Intn(36),                                               // 随机 issue 数 0-36
				CreatedAt:       createdAt,                                                  // 创建时间
				UpdatedAt:       updatedAt,                                                  // 更新时间
			},
			BranchByID: make(map[string]*models.MockBranch, 8), // 预分配 8 个分支的内存
		}

		// 5.5 调用 buildBranches 为仓库创建分支和提交记录
		repo.Branches = s.buildBranches(repo.Repo, blueprint.Domain)

		// 5.6 填充 branchByID 索引，并更新 CommitIndex
		for branchIdx := range repo.Branches {
			branch := &repo.Branches[branchIdx]
			repo.BranchByID[branch.Branch.Name] = branch // 分支名 -> 分支指针

			// 将该分支的所有提交存入 CommitIndex（SHA -> Commit）
			for _, commit := range branch.Commits {
				s.CommitIndex[commit.SHA] = commit
			}
		}

		// 5.7 将构建好的 repo 添加到 Repos 切片
		Repos = append(Repos, repo)
	}

	// ===== 第 6 步：按更新时间排序 =====
	// 最新的排在前面
	sort.Slice(Repos, func(i, j int) bool {
		return Repos[i].Repo.UpdatedAt.After(Repos[j].Repo.UpdatedAt)
	})

	// ===== 第 7 步：填充 RepoIndex 索引 =====
	// key 是 "owner/repo" 格式，value 是指向 repo 的指针
	for idx := range Repos {
		repo := &Repos[idx]
		s.RepoIndex[repo.Repo.FullName] = repo
	}

	// ===== 第 8 步：返回构建好的 Repos 切片 =====
	return Repos
}

// buildBranches 为一个仓库创建多个分支
func (s *MockServer) buildBranches(repo models.GiteaRepo, domain string) []models.MockBranch {

	// ===== 第 1 步：获取该仓库的分支名列表 =====
	// 根据领域和仓库信息生成对应的分支名（如 main, develop, feature/* 等）
	branchNames := buildBranchNames(repo, domain)

	// ===== 第 2 步：创建空切片 =====
	branches := make([]models.MockBranch, 0, len(branchNames))

	// ===== 第 3 步：遍历每个分支名，创建分支和提交 =====
	for _, branchName := range branchNames {

		// 3.1 为这个分支生成提交记录（200 个提交）
		commits := s.buildCommits(repo, branchName)

		// 3.2 创建 mockBranch 结构体
		branches = append(branches, models.MockBranch{
			Branch: models.GiteaBranch{
				Name: branchName,
				Commit: models.GiteaBranchCommit{
					// 分支的最新提交 SHA（取第一个，因为提交是按时间倒序的）
					ID: commits[0].SHA,
				},
			},
			// 这个分支的所有提交
			Commits: commits,
		})
	}

	// ===== 第 4 步：按分支名排序 =====
	// 排序规则：
	//   1. main 分支排第一
	//   2. develop 分支排第二
	//   3. 其他按字母顺序
	sort.Slice(branches, func(i, j int) bool {
		left, right := branches[i].Branch.Name, branches[j].Branch.Name

		// main 永远排最前
		if left == repo.DefaultBranch { // "main"
			return true
		}
		if right == repo.DefaultBranch {
			return false
		}

		// develop 排第二
		if left == "develop" {
			return true
		}
		if right == "develop" {
			return false
		}

		// 其他按字母顺序
		return left < right
	})

	// ===== 第 5 步：返回分支切片 =====
	return branches
}

// buildBranchNames 为一个仓库生成分支名列表
// 参数：
//   - repo: 仓库信息（获取 FullName 用于生成稳定种子）
//   - domain: 领域类型，决定 feature 分支的类型
//
// 返回：分支名切片
func buildBranchNames(repo models.GiteaRepo, domain string) []string {

	// ===== 第 1 步：创建随机数生成器 =====
	// 用仓库全名和领域作为种子，保证每次运行分支名一致
	rng := rand.New(rand.NewSource(utils.StableSeed("branches", repo.FullName, domain)))

	// ===== 第 2 步：定义各类分支名 =====

	// 基础分支：每个仓库都有
	base := []string{"main", "develop"}

	// 发布分支：随机选 1 个
	releases := []string{"release/v1.0", "release/v1.1", "release/v1.2", "release/v2.0"}

	// 功能分支：根据领域不同，选用不同的 feature 分支
	featuresByDomain := map[string][]string{
		"frontend": {"feature/ui-refresh", "feature/dashboard-filter", "feature/export-center", "feature/theme-token"},
		"backend":  {"feature/query-cache", "feature/batch-sync", "feature/audit-log", "feature/webhook-retry"},
		"mobile":   {"feature/offline-cache", "feature/push-center", "feature/profile-redesign", "feature/perf-monitor"},
		"infra":    {"feature/canary-release", "feature/pod-budget", "feature/secret-rotation", "feature/cluster-cost"},
		"tooling":  {"feature/schema-lint", "feature/generator-v2", "feature/template-market", "feature/dev-cli"},
		"data":     {"feature/metric-drilldown", "feature/etl-backfill", "feature/cohort-export", "feature/data-guard"},
		"ai":       {"feature/agent-memory", "feature/prompt-playground", "feature/rag-cache", "feature/model-router"},
	}

	// 热修复分支：全部可选
	hotfixes := []string{"hotfix/login-timeout", "hotfix/cron-retry", "hotfix/date-range", "hotfix/cache-ttl"}

	// 杂项分支：全部可选
	choreBranches := []string{"chore/upgrade-deps", "perf/query-plan", "test/e2e-stabilize"}

	// ===== 第 3 步：构建分支名列表 =====

	// 3.1 首先添加基础分支（main, develop）
	names := append([]string{}, base...)

	// 3.2 随机添加 1 个发布分支
	names = append(names, releases[rng.Intn(len(releases))])

	// 3.3 获取当前领域的功能分支池
	featurePool := append([]string{}, featuresByDomain[domain]...)

	// 3.4 随机打乱功能分支的顺序
	rng.Shuffle(len(featurePool), func(i, j int) {
		featurePool[i], featurePool[j] = featurePool[j], featurePool[i]
	})

	// 3.5 随机决定分支总数（5~8 个）
	targetCount := 5 + rng.Intn(4) // 5, 6, 7, 或 8

	// 3.6 依次添加功能分支，直到达到目标数量
	for _, feature := range featurePool {
		if len(names) >= targetCount {
			break
		}
		names = append(names, feature)
	}

	// 3.7 添加热修复分支（如果还有空位）
	for _, branch := range hotfixes {
		if len(names) >= targetCount {
			break
		}
		names = append(names, branch)
	}

	// 3.8 添加杂项分支（如果还有空位）
	for _, branch := range choreBranches {
		if len(names) >= targetCount {
			break
		}
		names = append(names, branch)
	}

	// ===== 第 4 步：返回分支名列表 =====
	return names
}

// buildCommits 为一个分支生成提交记录
// 参数：
//   - repo: 仓库信息
//   - branchName: 分支名
//
// 返回：包含 200 个提交的切片
func (s *MockServer) buildCommits(repo models.GiteaRepo, branchName string) []models.GiteaCommit {

	// ===== 第 1 步：创建随机数生成器 =====
	// 用仓库名和分支名作为种子，保证每次运行提交记录一致
	rng := rand.New(rand.NewSource(utils.StableSeed("commits", repo.FullName, branchName)))

	// ===== 第 2 步：获取预定义的提交消息模板 =====
	messages := data.BuildCommitMessages()

	// ===== 第 3 步：构建加权日期偏移数组 =====
	// 用于模拟真实的工作日分布（工作日权重高，周末权重低）
	weightedDays := buildWeightedDayOffsets(s.AnchorTime)

	// ===== 第 4 步：创建空提交切片 =====
	// 每个分支固定生成 commitsPerBranch (200) 个提交
	commits := make([]models.GiteaCommit, 0, config.CommitsPerBranch)

	// ===== 第 5 步：循环生成 200 个提交 =====
	for idx := 0; idx < config.CommitsPerBranch; idx++ {

		// 5.1 随机选择作者（从 10 个用户中选）
		authorIndex := rng.Intn(len(s.Users))

		// 5.2 随机选择提交者
		// 82% 的情况：提交者 = 作者
		// 18% 的情况：提交者 = 随机用户（模拟代码审查场景）
		committerIndex := authorIndex
		if rng.Intn(100) < 18 {
			committerIndex = rng.Intn(len(s.Users))
		}

		// 5.3 随机选择提交日期（基于加权分布）
		// 权重：周一/周五 = 2, 周二三四 = 3, 周末 = 1
		dayOffset := weightedDays[rng.Intn(len(weightedDays))]
		timestamp := s.randomTimestampForOffset(rng, dayOffset)

		// 5.4 随机选择提交消息
		// 从预定义的消息池中循环选取
		message := messages[(idx+rng.Intn(len(messages)))%len(messages)]

		// 5.5 根据分支类型调整提交消息
		// 如果是 hotfix 分支但消息不是 fix 开头，强制改成 fix
		if strings.HasPrefix(branchName, "hotfix/") && !strings.HasPrefix(message, "fix") {
			message = "fix(hotfix): stabilize production incident handling"
		}

		// 如果是 release 分支但消息是 feat 开头，改成 chore
		if strings.HasPrefix(branchName, "release/") && strings.HasPrefix(message, "feat") {
			message = "chore(release): prepare release candidate and bump version metadata"
		}

		// 5.6 根据消息类型生成代码变更统计
		// 不同类型的提交，代码行数不同（如 docs 改动少，refactor 改动大）
		additions, deletions := randomStats(rng, message)

		// 5.7 生成 SHA（提交 ID）
		sha := utils.ShortSHA(repo.FullName, branchName, idx, timestamp.UnixNano(), message)

		// 5.8 获取作者和提交者用户信息
		author := s.Users[authorIndex]
		committer := s.Users[committerIndex]

		// 5.9 组装提交结构体
		commit := models.GiteaCommit{
			SHA: sha,
			Commit: models.GiteaCommitInfo{
				Message: message,
				Author: models.GiteaCommitUser{
					Name:  author.FullName,
					Email: author.Email,
					Date:  timestamp,
				},
				Committer: models.GiteaCommitUser{
					Name:  committer.FullName,
					Email: committer.Email,
					// 提交时间比作者时间晚 0~45 分钟（模拟审查延迟）
					Date: timestamp.Add(time.Duration(rng.Intn(45)) * time.Minute),
				},
			},
			Stats: &models.GiteaStats{
				Additions: additions,             // 新增行数
				Deletions: deletions,             // 删除行数
				Total:     additions + deletions, // 总变更行数
			},
			Author: &s.Users[authorIndex],
		}

		// 5.10 添加到提交切片
		commits = append(commits, commit)
	}

	// ===== 第 6 步：按提交时间倒序排序 =====
	// 最新的提交排在前面
	sort.Slice(commits, func(i, j int) bool {
		left := commits[i].Commit.Author.Date
		right := commits[j].Commit.Author.Date

		// 如果时间相同，按 SHA 字典序（保证顺序一致）
		if left.Equal(right) {
			return commits[i].SHA > commits[j].SHA
		}

		// 否则按时间倒序（新的在前）
		return left.After(right)
	})

	// ===== 第 7 步：返回提交切片 =====
	return commits
}

func buildWeightedDayOffsets(anchor time.Time) []int {
	offsets := make([]int, 0, config.MaxRecentDays*3)
	dayStart := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, time.UTC)
	for offset := 0; offset < config.MaxRecentDays; offset++ {
		day := dayStart.AddDate(0, 0, -offset)
		weight := 3
		switch day.Weekday() {
		case time.Saturday:
			weight = 1
		case time.Sunday:
			weight = 1
		case time.Monday, time.Friday:
			weight = 2
		}
		for repeat := 0; repeat < weight; repeat++ {
			offsets = append(offsets, offset)
		}
	}
	return offsets
}

func randomStats(rng *rand.Rand, message string) (int, int) {
	switch {
	case strings.HasPrefix(message, "docs"):
		return 4 + rng.Intn(40), 0 + rng.Intn(8)
	case strings.HasPrefix(message, "test"):
		return 10 + rng.Intn(80), 2 + rng.Intn(24)
	case strings.HasPrefix(message, "refactor"):
		return 30 + rng.Intn(160), 20 + rng.Intn(120)
	case strings.HasPrefix(message, "perf"):
		return 18 + rng.Intn(90), 10 + rng.Intn(40)
	case strings.HasPrefix(message, "ci"), strings.HasPrefix(message, "build"), strings.HasPrefix(message, "chore"):
		return 6 + rng.Intn(60), 2 + rng.Intn(20)
	case strings.HasPrefix(message, "fix"):
		return 8 + rng.Intn(70), 3 + rng.Intn(30)
	default:
		return 20 + rng.Intn(160), 6 + rng.Intn(70)
	}
}

func (s *MockServer) randomTimestampForOffset(rng *rand.Rand, dayOffset int) time.Time {
	dayStart := time.Date(s.AnchorTime.Year(), s.AnchorTime.Month(), s.AnchorTime.Day(), 0, 0, 0, 0, time.UTC).
		AddDate(0, 0, -dayOffset)

	hour := 9 + rng.Intn(11)
	minute := rng.Intn(60)
	second := rng.Intn(60)
	return dayStart.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute + time.Duration(second)*time.Second)
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
