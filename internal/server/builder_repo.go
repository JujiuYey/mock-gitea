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
	"time"
)

func buildRepos(users []models.GiteaUser, anchorTime time.Time) (
	[]models.MockRepo,
	map[string]*models.MockRepo,
	map[string]models.GiteaCommit,
) {

	// 函数内部自己创建 map
	commitIndex := make(map[string]models.GiteaCommit, config.TotalRepos*8*config.CommitsPerBranch)
	repoIndex := make(map[string]*models.MockRepo, config.TotalRepos)

	// 第一步，定义仓库蓝图
	blueprints := data.BuildRepoBluePrint()

	// 第二步：校验数量
	if len(blueprints) != config.TotalRepos {
		log.Fatalf("[mock-gitea] expected %d repos, got %d", config.TotalRepos, len(blueprints))
	}

	// 第三步，构建用户映射
	// 创建一个 map: login -> GiteaUser，方便后续根据 login 查找用户
	// 这里用到 s.Users（在 buildUsers() 中初始化的 10 个用户）
	userByLogin := make(map[string]models.GiteaUser, len(users))
	for _, user := range users {
		userByLogin[user.Login] = user
	}

	// 第四部，创建空切片
	// 创建一个空的 mockRepo 切片，准备存放 30 个仓库
	repos := make([]models.MockRepo, 0, len(blueprints))

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
		createdAt := anchorTime.AddDate(0, 0, -(90 + rng.Intn(240))).Truncate(time.Second)
		updatedAt := anchorTime.Add(-time.Duration(rng.Intn(72)) * time.Hour).Truncate(time.Second)

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
		repo.Branches = buildBranches(users, anchorTime, repo.Repo, blueprint.Domain)

		// 5.6 填充 branchByID 索引，并更新 CommitIndex
		for branchIdx := range repo.Branches {
			branch := &repo.Branches[branchIdx]
			repo.BranchByID[branch.Branch.Name] = branch // 分支名 -> 分支指针

			// 将该分支的所有提交存入 CommitIndex（SHA -> Commit）
			for _, commit := range branch.Commits {
				commitIndex[commit.SHA] = commit
			}
		}

		// 5.7 将构建好的 repo 添加到 repos 切片
		repos = append(repos, repo)
	}

	// ===== 第 6 步：按更新时间排序 =====
	// 最新的排在前面
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Repo.UpdatedAt.After(repos[j].Repo.UpdatedAt)
	})

	// ===== 第 7 步：填充 RepoIndex 索引 =====
	// key 是 "owner/repo" 格式，value 是指向 repo 的指针
	for idx := range repos {
		repo := &repos[idx]
		repoIndex[repo.Repo.FullName] = repo
	}

	// ===== 第 8 步：返回构建好的 repos 切片 =====
	return repos, repoIndex, commitIndex
}
