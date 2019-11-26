package scm

import (
	"context"
	"errors"
	"strconv"

	pb "github.com/autograde/aguis/ag"
)

// FakeSCM implements the SCM interface.
type FakeSCM struct {
	Repositories  map[uint64]*Repository
	Organizations map[uint64]*pb.Organization
	Hooks         map[uint64]int
}

// NewFakeSCMClient returns a new Fake client implementing the SCM interface.
func NewFakeSCMClient() *FakeSCM {
	return &FakeSCM{
		Repositories:  make(map[uint64]*Repository),
		Organizations: make(map[uint64]*pb.Organization),
		Hooks:         make(map[uint64]int),
	}
}

// ListOrganizations implements the SCM interface.
func (s *FakeSCM) ListOrganizations(ctx context.Context) ([]*pb.Organization, error) {
	var orgs []*pb.Organization
	for _, org := range s.Organizations {
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// CreateOrganization implements the SCM interface.
func (s *FakeSCM) CreateOrganization(ctx context.Context, opt *CreateOrgOptions) (*pb.Organization, error) {
	id := len(s.Organizations) + 1
	org := &pb.Organization{
		ID:     uint64(id),
		Path:   opt.Path,
		Avatar: "https://avatars3.githubusercontent.com/u/1000" + strconv.Itoa(id) + "?v=3",
	}
	s.Organizations[org.ID] = org
	return org, nil
}

// UpdateOrganization implements the SCM interface.
func (s *FakeSCM) UpdateOrganization(ctx context.Context, opt *CreateOrgOptions) error {
	// TODO no implementation provided yet
	return nil
}

// GetOrganization implements the SCM interface.
func (s *FakeSCM) GetOrganization(ctx context.Context, opt *GetOrgOptions) (*pb.Organization, error) {
	org, ok := s.Organizations[opt.ID]
	if !ok {
		return nil, errors.New("directory not found")
	}
	return org, nil
}

// CreateRepository implements the SCM interface.
func (s *FakeSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	id := len(s.Repositories) + 1
	repo := &Repository{
		ID:      uint64(id),
		Path:    opt.Path,
		WebURL:  "https://example.com/" + opt.Organization.Path + "/" + opt.Path,
		SSHURL:  "git@example.com:" + opt.Organization.Path + "/" + opt.Path,
		HTTPURL: "https://example.com/" + opt.Organization.Path + "/" + opt.Path + ".git",
		OrgID:   opt.Organization.ID,
	}
	s.Repositories[repo.ID] = repo
	return repo, nil
}

// GetRepository implements the SCM interface.
func (s *FakeSCM) GetRepository(cts context.Context, opt *RepositoryOptions) (*Repository, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// GetRepositories implements the SCM interface.
func (s *FakeSCM) GetRepositories(ctx context.Context, org *pb.Organization) ([]*Repository, error) {
	var repos []*Repository
	for _, repo := range s.Repositories {
		if repo.OrgID == org.ID {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

// DeleteRepository implements the SCM interface.
func (s *FakeSCM) DeleteRepository(ctx context.Context, opt *RepositoryOptions) error {
	if _, ok := s.Repositories[opt.ID]; !ok {
		return errors.New("repository not found")
	}
	delete(s.Repositories, opt.ID)
	return nil
}

// UpdateRepoAccess implements the SCM interface.
func (s *FakeSCM) UpdateRepoAccess(ctx context.Context, repo *Repository, user, permission string) error {
	// TODO no implementation provided yet
	return nil
}

// RepositoryIsEmpty implements the SCM interface
func (s *FakeSCM) RepositoryIsEmpty(ctx context.Context, opt *RepositoryOptions) bool {
	// TODO no implementation provided yet
	return false
}

// ListHooks implements the SCM interface.
func (s *FakeSCM) ListHooks(ctx context.Context, repo *Repository, org string) ([]*Hook, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// CreateHook implements the SCM interface.
func (s *FakeSCM) CreateHook(ctx context.Context, opt *CreateHookOptions) error {
	if _, ok := s.Repositories[opt.Repository.ID]; !ok {
		return errors.New("repository not found")
	}
	s.Hooks[opt.Repository.ID]++
	return nil
}

// CreateOrgHook implements the scm interface
func (s *FakeSCM) CreateOrgHook(ctx context.Context, opt *OrgHookOptions) error {
	// TODO no implementation provided yet
	return nil
}

// CreateTeam implements the SCM interface.
func (s *FakeSCM) CreateTeam(ctx context.Context, opt *TeamOptions) (*Team, error) {
	// TODO no implementation provided yet
	return &Team{ID: 1, Name: "", URL: ""}, nil
}

// DeleteTeam implements the SCM interface.
func (s *FakeSCM) DeleteTeam(ctx context.Context, opt *TeamOptions) error {
	// TODO no implementation provided yet
	return nil
}

// GetTeam implements the SCM interface
func (s *FakeSCM) GetTeam(ctx context.Context, opt *TeamOptions) (*Team, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// GetTeams implements the SCM interface
func (s *FakeSCM) GetTeams(ctx context.Context, org *pb.Organization) ([]*Team, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// AddTeamMember implements the scm interface
func (s *FakeSCM) AddTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	// TODO no implementation provided yet
	return nil
}

// RemoveTeamMember implements the scm interface
func (s *FakeSCM) RemoveTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	// TODO no implementation provided yet
	return nil
}

// UpdateTeamMembers implements the SCM interface.
func (s *FakeSCM) UpdateTeamMembers(ctx context.Context, opt *TeamOptions) error {
	// TODO no implementation provided yet
	return nil
}

// CreateCloneURL implements the SCM interface.
func (s *FakeSCM) CreateCloneURL(opt *CreateClonePathOptions) string {
	return ""
}

// AddTeamRepo implements the SCM interface.
func (s *FakeSCM) AddTeamRepo(ctx context.Context, opt *AddTeamRepoOptions) error {
	return nil
}

// GetUserName implements the SCM interface.
func (s *FakeSCM) GetUserName(ctx context.Context) (string, error) {
	return "", nil
}

// GetUserNameByID implements the SCM interface.
func (s *FakeSCM) GetUserNameByID(ctx context.Context, remoteID uint64) (string, error) {
	return "", nil
}

// UpdateOrgMembership implements the SCM interface
func (s *FakeSCM) UpdateOrgMembership(ctx context.Context, opt *OrgMembershipOptions) error {
	// TODO no implementation provided yet
	return nil
}

// RevokeOrgMembership implements the SCM interface
func (s *FakeSCM) RevokeOrgMembership(ctx context.Context, opt *OrgMembershipOptions) error {
	// TODO no implementation provided yet
	return nil
}

// GetUserScopes implements the SCM interface
func (s *FakeSCM) GetUserScopes(ctx context.Context) *Authorization {
	// TODO no implementation provided yet
	return nil
}
