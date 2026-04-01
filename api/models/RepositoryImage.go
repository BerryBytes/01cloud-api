package models

import (
	"errors"
)

type RepositoryImage struct {
	Name          string
	Repository    string
	Tag           string
	CommitMessage *Commit
}

type GitRepository struct {
	GitUrl string
	Branch string
}

func (repo RepositoryImage) Validate(input *Environment) error {
	if repo.Repository == "" {
		return errors.New("required image repository")
	}
	if repo.Tag == "" {
		return errors.New("required image tag")
	}
	if repo.Repository != input.Application.ImageUrl {
		return errors.New("invalid image repository url")
	}
	if repo.Tag == input.ImageTag {
		return errors.New("tags is already deployed")
	}
	return nil
}

func (repo GitRepository) Validate(input *Environment) error {
	if repo.GitUrl == "" {
		return errors.New("required git repository")
	}
	if repo.Branch == "" {
		return errors.New("required image tag")
	}
	if repo.GitUrl != input.Application.GitUrl {
		return errors.New("invalid git repository url")
	}
	if repo.Branch == input.GitBranch {
		return errors.New("branch is already deployed")
	}
	return nil
}
