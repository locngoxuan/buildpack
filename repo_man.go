package buildpack

import (
	"errors"
	"fmt"
)

type RepoManager struct {
	Repos map[string]Repository
}

type Repository struct {
	Name     string
	Stable   *RepoChannel
	Unstable *RepoChannel
}

type RepoChannel struct {
	Address  string
	Username string
	Password string
}

func CreateRepoManager() RepoManager {
	return RepoManager{}
}

func CreateRepository() (r Repository, err error) {
	return
}

func (rm RepoManager) PickOne(name string) (r Repository, err error) {
	r, ok := rm.Repos[name]
	if !ok {
		err = errors.New(fmt.Sprintf("repo %s may be not registered", r))
	}
	return
}
