package core

import "C"
import (
	"context"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/utils"
	"log"
	"os"
	"time"
)

func signature() *object.Signature {
	return &object.Signature{
		Name: "system",
		When: time.Now(),
	}
}

type GitOption struct {
	WorkDir       string
	Branch        string
	RemoteAddress string
	config.GitCredential
}

type GitClient struct {
	GitOption
	ReferenceName plumbing.ReferenceName
	Repo          *git.Repository
}

func (c *GitClient) auth() (transport.AuthMethod, error) {
	return authWithCred(c.GitCredential)
}

func (c *GitClient) CloneIntoMemory(ctx context.Context) error {
	auth, err := c.auth()
	if err != nil {
		return err
	}
	c.ReferenceName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", c.Branch))
	log.Printf("cloning branch %s into memory from %s", c.Branch, c.GitOption.RemoteAddress)
	c.Repo, err = git.CloneContext(ctx, memory.NewStorage(), memfs.New(), &git.CloneOptions{
		ReferenceName: c.ReferenceName,
		URL:           c.GitOption.RemoteAddress,
		Auth:          auth,
		SingleBranch:  true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *GitClient) WriteSingleFile(data []byte, file, commitMsg string) error {
	wt, err := c.Repo.Worktree()
	if err != nil {
		return err
	}
	f, err := wt.Filesystem.OpenFile(file, os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	_, err = wt.Add(file)
	if err != nil {
		return err
	}
	_, err = wt.Commit(commitMsg, &git.CommitOptions{
		All:       true,
		Author:    signature(),
		Committer: signature(),
	})
	if err != nil {
		return err
	}
	obj, err := c.Repo.CommitObjects()
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("commit object is null")
	}
	return nil
}

func (c *GitClient) Push(ctx context.Context) error {
	auth, err := c.auth()
	if err != nil {
		return err
	}
	err = c.Repo.PushContext(ctx, &git.PushOptions{
		RefSpecs: []gitconfig.RefSpec{
			gitconfig.RefSpec(c.ReferenceName + ":" + c.ReferenceName),
		},
		Force:    true,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		return err
	}
	return nil
}

func tagExists(tag string, r *git.Repository) (bool, error) {
	tagFoundErr := "tag was found"
	tags, err := r.TagObjects()
	if err != nil {
		return false, err
	}
	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name == tag {
			return fmt.Errorf(tagFoundErr)
		}
		return nil
	})
	return err != nil, nil
}

func (c *GitClient) Tag(ctx context.Context, version string) error {
	exist, err := tagExists(version, c.Repo)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("tag was found")
	}

	h, err := c.Repo.Head()
	if err != nil {
		return err
	}
	_, err = c.Repo.CreateTag(version, h.Hash(), &git.CreateTagOptions{
		Tagger:  signature(),
		Message: "v" + version,
	})

	if err != nil {
		return err
	}
	auth, err := c.auth()
	if err != nil {
		return err
	}

	_ = c.Repo.DeleteRemote("update-code")
	remote, err := c.Repo.CreateRemote(&gitconfig.RemoteConfig{
		Name: "update-code",
		URLs: []string{c.RemoteAddress},
	})
	if err != nil {
		return fmt.Errorf("can not create anonymouse remote %v", err)
	}
	po := &git.PushOptions{
		RemoteName: remote.Config().Name,
		Progress:   os.Stdout,
		RefSpecs: []gitconfig.RefSpec{
			gitconfig.RefSpec("refs/tags/*:refs/tags/*"),
		},
		Auth: auth,
	}
	return c.Repo.Push(po)
}

func authWithCred(cred config.GitCredential) (transport.AuthMethod, error) {
	switch cred.Type {
	case config.CredentialToken:
		return &http.BasicAuth{
			Username: "token",
			Password: utils.ReadEnvVariableIfHas(cred.AccessToken),
		}, nil
	case config.CredentialAccount:
		return &http.BasicAuth{
			Username: utils.ReadEnvVariableIfHas(cred.Username),
			Password: utils.ReadEnvVariableIfHas(cred.Password),
		}, nil
	}
	return nil, fmt.Errorf("can not recognize credential type")
}

func PullLatestCode(ctx context.Context, c GitClient) error {
	repo, err := git.PlainOpen(c.WorkDir)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	_ = repo.DeleteRemote("update-code")
	remote, err := repo.CreateRemote(&gitconfig.RemoteConfig{
		Name: "update-code",
		URLs: []string{c.RemoteAddress},
	})
	if err != nil {
		return fmt.Errorf("can not create anonymouse remote %v", err)
	}
	auth, err := authWithCred(c.GitCredential)
	if err != nil {
		return err
	}
	err = wt.PullContext(ctx, &git.PullOptions{
		Force:         true,
		RemoteName:    remote.Config().Name,
		Auth:          auth,
		Progress:      os.Stdout,
		ReferenceName: c.ReferenceName,
	})
	if err != nil {
		return err
	}
	ref, err := repo.Head()
	if err != nil {
		return err
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}
	log.Printf("latest commit hash: %s", commit.Hash)
	return nil
}

func (c *GitClient) CreateNewBranch(branchName string) error {
	newBranchName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName))

	wt, err := c.Repo.Worktree()
	if err != nil {
		return err
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: newBranchName,
		Force:  true,
		Create: true,
		Keep:   true,
	})
	if err != nil {
		return err
	}
	auth, err := c.auth()
	if err != nil {
		return err
	}
	err = c.Repo.Push(&git.PushOptions{
		RefSpecs: []gitconfig.RefSpec{
			gitconfig.RefSpec(newBranchName + ":" + newBranchName),
		},
		Progress: os.Stdout,
		Auth:     auth,
		Force:    true,
	})
	if err != nil {
		return err
	}

	oldBranchName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", c.Branch))
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: oldBranchName,
		Force:  true,
		Create: false,
		Keep:   false,
	})
	if err != nil {
		return err
	}

	return nil
}
