package models

type MockRepo struct {
	Repo       GiteaRepo
	Branches   []MockBranch
	BranchByID map[string]*MockBranch
}

type MockBranch struct {
	Branch  GiteaBranch
	Commits []GiteaCommit
}

type RepoBlueprint struct {
	OwnerLogin  string
	Name        string
	Description string
	Domain      string
}