package buildpack

import (
	"errors"
	"fmt"
)

type RepoMan struct {
	repos map[string]*RepositoryConfig
}

func (r *RepoMan) Print(){
	for _, repo := range r.repos{
		fmt.Println(repo.Id, repo.Publisher, repo.StableChannel, repo.UnstableChannel)
	}
}

func (r *RepoMan) UpdateUserName(id, username string) {
	repo, ok := r.repos[id]
	if !ok {
		return
	}
	repo.StableChannel.Username = username
	repo.UnstableChannel.Username = username
}

func (r *RepoMan) UpdatePassword(id, password string) {
	repo, ok := r.repos[id]
	if !ok {
		return
	}
	repo.StableChannel.Password = password
	repo.UnstableChannel.Password = password
}

func (r *RepoMan) FindRepo(id string) (RepositoryConfig, error) {
	repo, ok := r.repos[id]
	if !ok {
		return RepositoryConfig{}, errors.New("not found repository by id " + id)
	}
	return *repo, nil
}

func (r *RepoMan) FindChannelById(release bool, id string) (ChannelConfig, error) {
	repo, err := r.FindRepo(id)
	if err != nil {
		return ChannelConfig{}, err
	}
	if release {
		return repo.StableChannel, nil
	}

	return repo.UnstableChannel, nil
}

func (r *RepoMan) FindChannelByAddress(release bool, address string) (ChannelConfig, error) {
	for _, repo := range r.repos {
		if release {
			if repo.StableChannel.Address == address {
				return repo.StableChannel, nil
			}
		} else {
			if repo.UnstableChannel.Address == address {
				return repo.UnstableChannel, nil
			}
		}
	}
	return ChannelConfig{}, errors.New("can not find repo by address " + address)
}

func (r *RepoMan) FindChannelByIdAndAddress(id, address string) (ChannelConfig, error) {
	repo, err := r.FindRepo(id)
	if err != nil {
		return ChannelConfig{}, err
	}

	if repo.StableChannel.Address == address {
		return repo.StableChannel, nil
	}

	if repo.UnstableChannel.Address == address {
		return repo.UnstableChannel, nil
	}

	return ChannelConfig{}, errors.New("can not find repo by id " + id + " and address " + address)
}
