package scm

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GithubSCM implements the SCM interface.
type GithubSCM struct {
	client *github.Client
}

// NewGithubSCMClient returns a new Github client implementing the SCM interface.
func NewGithubSCMClient(token string) *GithubSCM {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := github.NewClient(oauth2.NewClient(context.Background(), ts))
	return &GithubSCM{
		client: client,
	}
}

// ListDirectories implements the SCM interface.
func (s *GithubSCM) ListDirectories(ctx context.Context) ([]*Directory, error) {
	orgs, _, err := s.client.Organizations.ListOrgMemberships(ctx, nil)
	if err != nil {
		return nil, err
	}

	var directories []*Directory
	for _, org := range orgs {
		directories = append(directories, &Directory{
			ID:     uint64(org.Organization.GetID()),
			Path:   org.Organization.GetLogin(),
			Avatar: org.Organization.GetAvatarURL(),
		})
	}
	return directories, nil
}

// CreateDirectory implements the SCM interface.
func (s *GithubSCM) CreateDirectory(ctx context.Context, opt *CreateDirectoryOptions) (*Directory, error) {
	return nil, ErrNotSupported{
		SCM:    "github",
		Method: "CreateDirectory",
	}
}

// GetDirectory implements the SCM interface.
func (s *GithubSCM) GetDirectory(ctx context.Context, id uint64) (*Directory, error) {
	org, _, err := s.client.Organizations.GetByID(ctx, int(id))
	if err != nil {
		return nil, err
	}

	return &Directory{
		ID:     uint64(org.GetID()),
		Path:   org.GetLogin(),
		Avatar: org.GetAvatarURL(),
	}, nil
}

// CreateRepository implements the SCM interface.
func (s *GithubSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	var repo *github.Repository

	repo, _, err := s.client.Repositories.Create(ctx, opt.Directory.Path, &github.Repository{
		Name:    &opt.Path,
		Private: &opt.Private,
	})
	if err != nil {
		// Could not create a private repo, trying to create public
		err = nil
		repo, _, err = s.client.Repositories.Create(ctx, opt.Directory.Path, &github.Repository{
			Name: &opt.Path,
		})

		if err != nil {
			return nil, err
		}

	}

	return &Repository{
		ID:          uint64(repo.GetID()),
		Path:        repo.GetName(),
		Owner:       repo.Owner.GetLogin(), // TODO: Guard against Owner = nil.
		WebURL:      repo.GetHTMLURL(),
		SSHURL:      repo.GetSSHURL(),
		HTTPURL:     repo.GetCloneURL(),
		DirectoryID: opt.Directory.ID,
	}, nil
}

// GetRepositories implements the SCM interface.
func (s *GithubSCM) GetRepositories(ctx context.Context, directory *Directory) ([]*Repository, error) {
	var path string
	if directory.Path != "" {
		path = directory.Path
	} else {
		directory, err := s.GetDirectory(ctx, directory.ID)
		if err != nil {
			return nil, err
		}
		path = directory.Path
	}

	repos, _, err := s.client.Repositories.ListByOrg(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var repositories []*Repository
	for _, repo := range repos {
		repositories = append(repositories, &Repository{
			ID:          uint64(repo.GetID()),
			Path:        repo.GetName(),
			Owner:       repo.Owner.GetLogin(), // TODO: Guard against Owner = nil.
			WebURL:      repo.GetHTMLURL(),
			SSHURL:      repo.GetSSHURL(),
			HTTPURL:     repo.GetCloneURL(),
			DirectoryID: directory.ID,
		})
	}

	return repositories, nil
}

// DeleteRepository implements the SCM interface.
func (s *GithubSCM) DeleteRepository(ctx context.Context, id uint64) error {
	repo, _, err := s.client.Repositories.GetByID(ctx, int(id))
	if err != nil {
		return err
	}
	if _, err := s.client.Repositories.Delete(ctx, repo.Owner.GetLogin(), repo.GetName()); err != nil {
		return err
	}
	return nil
}

// ListHooks implements the SCM interface.
func (s *GithubSCM) ListHooks(ctx context.Context, repo *Repository) ([]*Hook, error) {
	githubHooks, _, err := s.client.Repositories.ListHooks(ctx, repo.Owner, repo.Path, nil)
	var hooks []*Hook
	for _, hook := range githubHooks {
		hooks = append(hooks, &Hook{
			ID:   uint64(hook.GetID()),
			Name: hook.GetName(),
			URL:  hook.GetURL(),
		})
	}
	return hooks, err
}

// CreateHook implements the SCM interface.
func (s *GithubSCM) CreateHook(ctx context.Context, opt *CreateHookOptions) (err error) {
	name := "web"
	_, _, err = s.client.Repositories.CreateHook(ctx, opt.Repository.Owner, opt.Repository.Path, &github.Hook{
		Name: &name,
		Config: map[string]interface{}{
			"url":          opt.URL,
			"secret":       opt.Secret,
			"content_type": "json",
			"insecure_ssl": "0",
		},
	})
	return
}

// CreateTeam implements the SCM interface.
func (s *GithubSCM) CreateTeam(ctx context.Context, opt *CreateTeamOptions) (*Team, error) {
	t, _, err := s.client.Organizations.CreateTeam(ctx, opt.Directory.Path, &github.Team{
		Name: &opt.TeamName,
	})
	if err != nil {
		return nil, err
	}
	for _, user := range opt.Users {
		_, _, err = s.client.Organizations.AddTeamMembership(ctx, t.GetID(), user, nil)
		if err != nil {
			return nil, err
		}
	}
	return &Team{
		ID:   uint64(t.GetID()),
		Name: t.GetName(),
		URL:  t.GetURL(),
	}, nil
}

// CreateCloneURL implements the SCM interface.
func (s *GithubSCM) CreateCloneURL(ctx context.Context, opt *CreateClonePathOptions) (string, error) {
	return "https://" + opt.UserToken + "@github.com/" + opt.Directory + "/" + opt.Repository, nil
}

// AddTeamRepo implements the SCM interface.
func (s *GithubSCM) AddTeamRepo(ctx context.Context, opt *AddTeamRepoOptions) error {
	_, err := s.client.Organizations.AddTeamRepo(ctx, int(opt.TeamID), opt.Owner, opt.Repo, &github.OrganizationAddTeamRepoOptions{
		Permission: "push", // This make sure that users can pull and push
	})
	if err != nil {
		return err
	}
	return nil
}

// GetUserName implements the SCM interface.
func (s *GithubSCM) GetUserName(ctx context.Context) (string, error) {
	user, _, err := s.client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

// GetUserNameByID implements the SCM interface.
func (s *GithubSCM) GetUserNameByID(ctx context.Context, remoteID uint64) (string, error) {
	user, _, err := s.client.Users.GetByID(ctx, int(remoteID))
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

// GetPaymentPlan implements the SCM interface.
func (s *GithubSCM) GetPaymentPlan(ctx context.Context, orgID uint64) (*PaymentPlan, error) {
	org, _, err := s.client.Organizations.GetByID(ctx, int(orgID))
	if err != nil {
		return nil, err
	}
	plan := &PaymentPlan{
		Name:         org.Plan.GetName(),
		PrivateRepos: uint64(org.Plan.GetPrivateRepos()),
	}
	return plan, nil
}

// UpdateRepository implements the SCM interface
func (s *GithubSCM) UpdateRepository(ctx context.Context, repo *Repository) error {
	// TODO - make this more flexible rather than only making stuff private.
	gitRepo, _, err := s.client.Repositories.GetByID(ctx, int(repo.ID))
	if err != nil {
		return err
	}

	*gitRepo.Private = true
	_, _, err = s.client.Repositories.Edit(ctx, gitRepo.Owner.GetLogin(), gitRepo.GetName(), gitRepo)
	if err != nil {
		return err
	}

	return nil
}
