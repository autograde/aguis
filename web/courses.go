package web

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/yamlparser"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

// MaxWait is the maximum time a request is allowed to stay open before
// aborting.
const MaxWait = 10 * time.Minute

// NewCourseRequest represents a request for a new course.
type NewCourseRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Year uint   `json:"year"`
	Tag  string `json:"tag"`

	Provider    string `json:"provider"`
	DirectoryID uint64 `json:"directoryid"`
}

func (cr *NewCourseRequest) valid() bool {
	return cr != nil &&
		cr.Name != "" &&
		cr.Code != "" &&
		(cr.Provider == "github" || cr.Provider == "gitlab" || cr.Provider == "fake") &&
		cr.DirectoryID != 0 &&
		cr.Year != 0 &&
		cr.Tag != ""
}

// EnrollUserRequest represent a request for enrolling a user to a course.
type EnrollUserRequest struct {
	Status uint `json:"status"`
}

func (eur *EnrollUserRequest) valid() bool {
	return eur.Status <= models.Teacher
}

// NewGroupRequest represents a new group.
type NewGroupRequest struct {
	Name     string   `json:"name"`
	CourseID uint64   `json:"courseid"`
	UserIDs  []uint64 `json:"userids"`
}

func (grp *NewGroupRequest) valid() bool {
	return grp != nil &&
		grp.Name != "" &&
		len(grp.UserIDs) > 0
}

// UpdateGroupRequest updates group
type UpdateGroupRequest struct {
	Status uint `json:"status"`
}

// ListCourses returns a JSON object containing all the courses in the database.
func ListCourses(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courses, err := db.GetCourses()
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, courses, "\t")
	}
}

// ListCoursesWithEnrollment lists all existing courses with the provided users
// enrollment status.
// If status query param is provided, lists only courses of the student filtered by the query param.
func ListCoursesWithEnrollment(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("uid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		statuses, err := parseEnrollmentStatus(c.QueryParam("status"))
		if err != nil {
			return err
		}

		courses, err := db.GetCoursesByUser(id, statuses...)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, courses, "\t")
	}
}

// ListAssignments lists the assignments for the provided course.
func ListAssignments(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		assignments, err := db.GetAssignmentsByCourse(id)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, assignments, "\t")
	}
}

// BaseHookOptions contains options shared among all webhooks.
type BaseHookOptions struct {
	BaseURL string
	// Secret is used to verify that the event received is legit. GitHub
	// sends back a signature of the payload, while GitLab just sends back
	// the secret. This is all handled by the
	// gopkg.in/go-playground/webhooks.v3 package.
	Secret string
}

// NewCourse creates a new course and associates it with a directory (organization in github)
// and creates the repositories for the course.
func NewCourse(logger logrus.FieldLogger, db database.Database, bh *BaseHookOptions) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cr NewCourseRequest
		if err := c.Bind(&cr); err != nil {
			return err
		}
		if !cr.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		s, err := getSCM(c, cr.Provider)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()

		directory, err := s.GetDirectory(ctx, cr.DirectoryID)
		if err != nil {
			return err
		}
		repos, err := s.GetRepositories(ctx, directory)
		if err != nil {
			return err
		}
		existing := make(map[string]*scm.Repository)
		for _, repo := range repos {
			existing[repo.Path] = repo
		}

		var paths = []string{InfoRepo, AssignmentRepo, TestsRepo, SolutionsRepo}
		for _, path := range paths {
			var repo *scm.Repository
			var ok bool
			if repo, ok = existing[path]; !ok {
				privRepo := false
				if path == TestsRepo {
					privRepo = true
				}
				var err error
				repo, err = s.CreateRepository(
					ctx,
					&scm.CreateRepositoryOptions{
						Path:      path,
						Directory: directory,
						Private:   privRepo},
				)

				if err != nil {
					logger.WithField("repo", path).WithField("private", privRepo).WithError(err).Warn("Failed to create repository")
					return err
				}
				logger.WithField("repo", repo).Println("Created new repository")
			}

			hooks, err := s.ListHooks(ctx, repo)
			if err != nil {
				logger.WithField("repo", path).WithError(err).Warn("Failed to list hooks for repository")
				return err
			}
			hasAGWebHook := false
			for _, hook := range hooks {
				logger.WithField("url", hook.URL).WithField("id", hook.ID).WithField("name", hook.Name).Println("Hook for repository")
				// TODO this check is specific for the github implementation ; fix this
				if hook.Name == "web" {
					hasAGWebHook = true
					break
				}
			}

			if !hasAGWebHook {
				if err := s.CreateHook(ctx, &scm.CreateHookOptions{
					URL:        GetEventsURL(bh.BaseURL, cr.Provider),
					Secret:     bh.Secret,
					Repository: repo,
				}); err != nil {
					logger.WithField("repo", path).WithError(err).Println("Failed to create webhook for repository")
					return err
				}

				logger.WithField("repo", repo).Println("Created new webhook for repository")
			}

			var repoType models.RepoType
			switch path {
			case InfoRepo:
				repoType = models.CourseInfoRepo
			case AssignmentRepo:
				repoType = models.AssignmentsRepo
			case TestsRepo:
				repoType = models.TestsRepo
			case SolutionsRepo:
				repoType = models.SolutionsRepo
			}

			dbRepo := models.Repository{
				DirectoryID:  directory.ID,
				RepositoryID: repo.ID,
				HTMLURL:      repo.WebURL,
				Type:         repoType,
			}
			if err := db.CreateRepository(&dbRepo); err != nil {
				return err
			}
		}

		user := c.Get("user").(*models.User)
		course := models.Course{
			Name:            cr.Name,
			CourseCreatorID: user.ID,
			Code:            cr.Code,
			Year:            cr.Year,
			Tag:             cr.Tag,
			Provider:        cr.Provider,
			DirectoryID:     directory.ID,
		}
		if err := db.CreateCourse(user.ID, &course); err != nil {
			//TODO(meling) Should we even communicate bad request to the client?
			// We should log errors and debug it on the server side instead.
			// If clients make mistakes, there is nothing it can do with the error message.
			if err == database.ErrCourseExists {
				return c.JSONPretty(http.StatusBadRequest, err.Error(), "\t")
			}
			return err
		}
		return c.JSONPretty(http.StatusCreated, &course, "\t")
	}
}

// CreateEnrollment enrolls a user in a course.
func CreateEnrollment(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courseID, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		userID, err := parseUint(c.Param("uid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		var eur EnrollUserRequest
		if err := c.Bind(&eur); err != nil {
			return err
		}
		if !eur.valid() || userID == 0 || courseID == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		enrollment := models.Enrollment{
			UserID:   userID,
			CourseID: courseID,
		}
		if err := db.CreateEnrollment(&enrollment); err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		return c.NoContent(http.StatusCreated)
	}
}

// UpdateEnrollment accepts or rejects a user to enroll in a course.
func UpdateEnrollment(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courseID, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		userID, err := parseUint(c.Param("uid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		var eur EnrollUserRequest
		if err := c.Bind(&eur); err != nil {
			return err
		}
		if !eur.valid() || userID == 0 || courseID == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// check that userID has enrolled in courseID
		if _, err := db.GetEnrollmentByCourseAndUser(courseID, userID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
		user := c.Get("user").(*models.User)
		// TODO: This check should be performed in AccessControl.
		if user.IsAdmin == nil || !*user.IsAdmin {
			// Only admin users are allowed to enroll or reject users to a course.
			// TODO we should also allow users of the 'teachers' team to accept/reject users
			return c.NoContent(http.StatusUnauthorized)
		}

		// TODO If the enrollment is accepted, create repositories and permissions for them with webhooks.
		switch eur.Status {
		case models.Student:
			// update enrollment for student in database
			err = db.EnrollStudent(userID, courseID)
			if err != nil {
				return err
			}
			course, err := db.GetCourse(courseID)
			if err != nil {
				return err
			}
			student, err := db.GetUser(userID)
			if err != nil {
				return err
			}
			remote, err := getRemoteIDFor(student, course.Provider)
			if err != nil {
				return err
			}
			s, err := getSCM(c, course.Provider)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
			defer cancel()

			dir, err := s.GetDirectory(ctx, course.DirectoryID)
			if err != nil {
				return err
			}
			gitUserName, err := s.GetUserNameByID(ctx, remote.ID)
			if err != nil {
				return err
			}

			//TODO(meling) Find the current payment plan and if private is available create team/repo;
			// if not private do not create the team/repo.

			repo, err := s.CreateRepository(ctx, &scm.CreateRepositoryOptions{
				Directory: dir,
				Path:      StudentRepoName(gitUserName),
				Private:   true,
			})
			if err != nil {
				return err
			}
			dbRepo := models.Repository{
				DirectoryID:  course.DirectoryID,
				RepositoryID: repo.ID,
				HTMLURL:      repo.WebURL,
				Type:         models.UserRepo,
				UserID:       userID,
			}
			if err := db.CreateRepository(&dbRepo); err != nil {
				return err
			}
			// Creating team
			team, err := s.CreateTeam(ctx, &scm.CreateTeamOptions{
				Directory: &scm.Directory{Path: dir.Path},
				TeamName:  gitUserName,
				Users:     []string{gitUserName},
			})
			if err != nil {
				return err
			}
			err = s.AddTeamRepo(ctx, &scm.AddTeamRepoOptions{
				TeamID: team.ID,
				Owner:  repo.Owner,
				Repo:   repo.Path,
			})
			if err != nil {
				return err
			}

		case models.Teacher:
			err = db.EnrollTeacher(userID, courseID)
		case models.Rejected:
			err = db.RejectEnrollment(userID, courseID)
		case models.Pending:
			err = db.SetPendingEnrollment(userID, courseID)
		}
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	}
}

// GetCourse find course by id and return JSON object.
func GetCourse(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		course, err := db.GetCourse(id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err

		}

		return c.JSONPretty(http.StatusOK, course, "\t")
	}
}

// RefreshCourse refreshes the information to a course
func RefreshCourse(logger logrus.FieldLogger, db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		course, err := db.GetCourse(cid)
		if err != nil {
			return err
		}
		s, err := getSCM(c, course.Provider)
		if err != nil {
			return err
		}

		user := c.Get("user").(*models.User)

		remoteID, err := getRemoteIDFor(user, course.Provider)
		if err != nil {
			return err
		}

		if user.IsAdmin != nil && *user.IsAdmin {
			// Only admin users should be able to update repos to private, if they are public.
			updateRepoToPrivate(c.Request().Context(), db, s, course.DirectoryID)
		}

		assignments, err := RefreshCourseInformation(c.Request().Context(), logger, db, course, remoteID, s)
		if err != nil {
			return err
		}

		return c.JSONPretty(http.StatusOK, assignments, "\t")
	}
}

var runLock sync.Mutex

// RefreshCourseInformation refreshes the course information on a single course
func RefreshCourseInformation(ctx context.Context, logger logrus.FieldLogger, db database.Database, course *models.Course, remoteID *models.RemoteIdentity, s scm.SCM) ([]yamlparser.NewAssignmentRequest, error) {
	// Have to lock this, so no one tries to refreshes the courses simultanously, since that would result
	// in different routines competing for storage resoruces.
	runLock.Lock()
	defer runLock.Unlock()

	ctxGithub, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	directory, err := s.GetDirectory(ctxGithub, course.DirectoryID)
	if err != nil {
		logger.Error("Problem fetching Directory")
		return nil, err
	}

	path, err := s.CreateCloneURL(ctxGithub, &scm.CreateClonePathOptions{
		Directory:  directory.Path,
		Repository: "tests.git",
		UserToken:  remoteID.AccessToken,
	})
	if err != nil {
		logger.Error("Problem Creating clone URL")
		return nil, err
	}

	runner := ci.Local{}

	// This does not work that well on Windows because the path should be
	// /mnt/c/Users/{user}/AppData/Local/Temp
	// cloneDirectory := filepath.Join(os.TempDir(), "agclonepath")
	cloneDirectory := "agclonepath"

	// Clone all tests from tests repositry
	job := &ci.Job{
		Commands: []string{
			"mkdir " + cloneDirectory,
			"cd " + cloneDirectory,
			"git clone " + path,
		},
	}
	logger.WithField("job", job).Info("Running Job")
	_, err = runner.Run(ctx, job)
	if err != nil {
		logger.Error("Problem Running CI runner")
		runner.Run(ctx, &ci.Job{
			Commands: []string{
				"yes | rm -r " + cloneDirectory,
			},
		})
		return nil, err
	}

	// Parse assignments in the test directory
	assignments, err := yamlparser.Parse("agclonepath/tests")
	if err != nil {
		logger.Error("Problem getting assignments")
		return nil, err
	}

	// Cleanup downloaded
	runner.Run(ctx, &ci.Job{
		Commands: []string{
			"yes | rm -r " + cloneDirectory,
		},
	})

	for _, v := range assignments {
		assignment, err := createAssignment(&v, course)
		if err != nil {
			logger.Error("Problem createing assignment")
			return nil, err
		}
		if err := db.CreateAssignment(assignment); err != nil {
			logger.Error("Problem adding assignment to DB")
			return nil, err
		}
	}
	return assignments, nil
}

func getRemoteIDFor(user *models.User, provider string) (*models.RemoteIdentity, error) {
	var remoteID *models.RemoteIdentity
	for _, v := range user.RemoteIdentities {
		if v.Provider == provider {
			remoteID = v
			break
		}
	}
	if remoteID == nil {
		return nil, echo.ErrNotFound
	}
	return remoteID, nil
}

func createAssignment(request *yamlparser.NewAssignmentRequest, course *models.Course) (*models.Assignment, error) {
	date, err := time.Parse("02-01-2006 15:04", request.Deadline)
	if err != nil {
		return nil, err
	}

	return &models.Assignment{
		AutoApprove: request.AutoApprove,
		CourseID:    course.ID,
		Deadline:    date,
		Language:    request.Language,
		Name:        request.Name,
		Order:       request.AssignmentID,
		IsGroupLab:  request.IsGroupLab,
	}, nil
}

// GetSubmission returns a single submission for a assignment and a user
func GetSubmission(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		assignmentID, err := parseUint(c.Param("aid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		user := c.Get("user").(*models.User)

		submission, err := db.GetSubmissionForUser(assignmentID, user.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		return c.JSONPretty(http.StatusOK, submission, "\t")
	}
}

// ListSubmissions returns all the latests submissions for a user to a course
func ListSubmissions(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courseID, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// Check if a user is provided, else used logged in user
		userID, err := parseUint(c.Param("uid"))
		if err != nil {
			userID = c.Get("user").(*models.User).ID
		}

		submission, err := db.GetSubmissions(courseID, userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		return c.JSONPretty(http.StatusOK, submission, "\t")
	}
}

// UpdateCourse updates an existing course
func UpdateCourse(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		if _, err := db.GetCourse(id); err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		// TODO: Might be better to define a Validate method on models.Course and bind to that.
		var cr NewCourseRequest
		if err := c.Bind(&cr); err != nil {
			return err
		}
		if !cr.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		s, err := getSCM(c, cr.Provider)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()

		// Check that the directory exists.
		_, err = s.GetDirectory(ctx, cr.DirectoryID)
		if err != nil {
			return err
		}

		if err := db.UpdateCourse(&models.Course{
			ID:          id,
			Name:        cr.Name,
			Code:        cr.Code,
			Year:        cr.Year,
			Tag:         cr.Tag,
			Provider:    cr.Provider,
			DirectoryID: cr.DirectoryID,
		}); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)

	}
}

// GetEnrollmentsByCourse get all enrollments for a course.
func GetEnrollmentsByCourse(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		statuses, err := parseEnrollmentStatus(c.QueryParam("status"))
		if err != nil {
			return err
		}

		enrollments, err := db.GetEnrollmentsByCourse(id, statuses...)
		if err != nil {
			return err
		}

		for _, enrollment := range enrollments {
			enrollment.User, err = db.GetUser(enrollment.UserID)
			if err != nil {
				return err
			}
		}

		return c.JSONPretty(http.StatusOK, enrollments, "\t")
	}
}

// NewGroup creates a new group under a course
func NewGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		if _, err := db.GetCourse(cid); err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}

		var grp NewGroupRequest
		if err := c.Bind(&grp); err != nil {
			return err
		}
		if !grp.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// Don't add remote identities here since these users are returned to the client.
		users, err := db.GetUsers(false, grp.UserIDs...)
		if err != nil {
			return err
		}
		// check if provided user ids are valid
		if len(users) != len(grp.UserIDs) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// signed in student user must be member of the group
		signedInUser := c.Get("user").(*models.User)
		signedInUserEnrollment, err := db.GetEnrollmentByCourseAndUser(cid, signedInUser.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Not able to retreive enrollment for signed in user")
		}
		signedInUserInGroup := false

		// only enrolled users can join a group
		// prevent group override if a student is already in a group in this course
		for _, user := range users {
			enrollment, err := db.GetEnrollmentByCourseAndUser(cid, user.ID)
			switch {
			case err == gorm.ErrRecordNotFound:
				return echo.NewHTTPError(http.StatusNotFound, "user not enrolled in course")
			case err != nil:
				return err
			case enrollment.GroupID > 0:
				return echo.NewHTTPError(http.StatusBadRequest, "user already enrolled in another group")
			case enrollment.Status < models.Student:
				return echo.NewHTTPError(http.StatusBadRequest, "user not yet accepted for this course")
			case enrollment.Status == models.Teacher && signedInUserEnrollment.Status != models.Teacher:
				return echo.NewHTTPError(http.StatusBadRequest, "A teacher has to create this group")
			case signedInUser.ID == user.ID && enrollment.Status == models.Student:
				signedInUserInGroup = true
			}
		}

		// If user is a teacher it should be allowed to proceed and create a group with only the "enrolled" persons.
		if signedInUserEnrollment.Status == models.Teacher {
			signedInUserInGroup = true
		}

		if !signedInUserInGroup {
			return echo.NewHTTPError(http.StatusBadRequest, "student must be member of new group")
		}

		group := models.Group{
			Name:     grp.Name,
			CourseID: cid,
			Users:    users,
		}
		// CreateGroup creates a new group and update group_id in enrollment table
		if err := db.CreateGroup(&group); err != nil {
			if err == database.ErrDuplicateGroup {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return err
		}

		return c.JSONPretty(http.StatusCreated, &group, "\t")
	}
}

// UpdateGroup update a group
// TODO: Remove this function? similar exists in web/groups.go
func UpdateGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		course, err := db.GetCourse(cid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}

		gid, err := parseUint(c.Param("gid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		// we don't remote identities here; we only do this to check that the group exists.
		oldgrp, err := db.GetGroup(false, gid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "group not found")
			}
			return err
		}

		user := c.Get("user").(*models.User)
		enrollment, err := db.GetEnrollmentByCourseAndUser(cid, user.ID)
		if err != nil {
			return err
		}
		if enrollment.Status != models.Teacher {
			return echo.NewHTTPError(http.StatusForbidden, "only teacher can update a group")
		}

		var grp NewGroupRequest
		if err := c.Bind(&grp); err != nil {
			return err
		}
		if !grp.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		// we need the remote identities of users of the group to find their scm user names
		users, err := db.GetUsers(true, grp.UserIDs...)
		if err != nil {
			return err
		}
		// check if provided user ids are valid
		if len(users) != len(grp.UserIDs) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// only enrolled user i.e accepted to the course can join a group
		// prevent group override if a student is already in a group in this course
		for _, user := range users {
			enrollment, err := db.GetEnrollmentByCourseAndUser(cid, user.ID)
			switch {
			case err == gorm.ErrRecordNotFound:
				return echo.NewHTTPError(http.StatusNotFound, "user is not enrolled to this course")
			case err != nil:
				return err
			case enrollment.GroupID > 0 && enrollment.GroupID != oldgrp.ID:
				return echo.NewHTTPError(http.StatusBadRequest, "user is already in another group")
			case enrollment.Status < models.Student:
				return echo.NewHTTPError(http.StatusBadRequest, "user is not yet accepted to this course")
			}
		}

		if err := db.UpdateGroup(&models.Group{
			ID:       oldgrp.ID,
			Name:     grp.Name,
			CourseID: cid,
			Users:    users,
		}); err != nil {
			if err == database.ErrDuplicateGroup {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return err
		}

		s, err := getSCM(c, course.Provider)
		if err != nil {
			return err
		}

		var userRemoteIdentity []*models.RemoteIdentity
		// TODO move this into the for loop above, modify db.GetUsers() to also retreive RemoteIdentity
		// so we can remove individual GetUser calls
		for _, user := range users {
			remoteIdentityUser, _ := db.GetUser(user.ID)
			if err != nil {
				return err
			}
			// TODO, figure out which remote identity to be used!
			if len(remoteIdentityUser.RemoteIdentities) > 0 {
				userRemoteIdentity = append(userRemoteIdentity, remoteIdentityUser.RemoteIdentities[0])
			}
		}

		// TODO move this functionality down into SCM?
		// Note: This Requires alot of calls to git.
		// Figure out all group members git-username
		var gitUserNames []string
		for _, identity := range userRemoteIdentity {
			gitName, err := s.GetUserNameByID(c.Request().Context(), identity.RemoteID)
			if err != nil {
				return err
			}
			gitUserNames = append(gitUserNames, gitName)
		}

		// Create and add repo to autograder group
		dir, err := s.GetDirectory(c.Request().Context(), course.DirectoryID)
		if err != nil {
			return err
		}
		repo, err := s.CreateRepository(c.Request().Context(), &scm.CreateRepositoryOptions{
			Directory: dir,
			Path:      grp.Name,
			Private:   true,
		})
		if err != nil {
			return err
		}
		// Add repo to DB
		dbRepo := models.Repository{
			DirectoryID:  course.DirectoryID,
			RepositoryID: repo.ID,
			Type:         models.UserRepo,
			UserID:       grp.UserIDs[0], // Should this be groupID ????
		}
		if err := db.CreateRepository(&dbRepo); err != nil {
			return err
		}
		// Create git-team
		team, err := s.CreateTeam(c.Request().Context(), &scm.CreateTeamOptions{
			Directory: &scm.Directory{Path: dir.Path},
			TeamName:  grp.Name,
			Users:     gitUserNames,
		})
		if err != nil {
			return err
		}
		// Adding Repo to git-team
		if err = s.AddTeamRepo(c.Request().Context(), &scm.AddTeamRepoOptions{
			TeamID: team.ID,
			Owner:  repo.Owner,
			Repo:   repo.Path,
		}); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}

// GetGroups returns all groups under a course
func GetGroups(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		if _, err := db.GetCourse(cid); err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}
		groups, err := db.GetGroupsByCourse(cid)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, groups, "\t")
	}
}

// UpdateSubmission updates a submission
func UpdateSubmission(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		sid, err := parseUint(c.Param("sid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		err = db.UpdateSubmissionByID(sid, true)
		if err != nil {
			return err
		}

		return nil
	}
}

// ListGroupSubmissions fetches all submissions from specific group
func ListGroupSubmissions(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		gid, err := parseUint(c.Param("gid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		submission, err := db.GetGroupSubmissions(cid, gid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}

		return c.JSONPretty(http.StatusOK, submission, "\t")

	}
}

// GetCourseInformationURL returns the course information html as string
func GetCourseInformationURL(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse courseID")
		}
		courseInfoRepo, err := db.GetRepositoriesByCourseIDAndType(cid, models.CourseInfoRepo)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not retrieve any course information repos")
		}
		// There should only exist 1 course info pr course.
		if len(courseInfoRepo) > 1 && len(courseInfoRepo) == 0 {
			return echo.NewHTTPError(http.StatusInternalServerError, "Too many course information repositories exists for this course")
		}

		// Have to be in string array to be able to jsonify so frontend recognize it.
		// See public/src/HttpHelper.ts -> send()
		var courseInfoURL []string
		courseInfoURL = append(courseInfoURL, courseInfoRepo[0].HTMLURL)
		return c.JSONPretty(http.StatusOK, &courseInfoURL, "\t")
	}
}

// GetRepositoryURL returns the course information html as string
func GetRepositoryURL(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse courseID")
		}

		// parseUint does not allow 0 values, since model.RepoType can be 0 we convert it ourselves.
		repoType, err := strconv.ParseUint(c.QueryParam("type"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse repoType")
		}
		identifiedRepoType, err := models.IdentifyRepoTypeFromFrontEnd(repoType)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse Repository type")
		}

		var repos []*models.Repository
		if identifiedRepoType == models.UserRepo {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(http.StatusBadRequest, "user not registered")
			}
			userRepo, err := db.GetRepoByCourseIDUserIDandType(cid, user.ID, models.UserRepo)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "GetRepoByCourseIDUserIDandType: Could not retrieve any UserRepo")
			}
			repos = append(repos, userRepo)
		} else {
			repos, err = db.GetRepositoriesByCourseIDAndType(cid, identifiedRepoType)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "GetRepositoriesByCourseIDAndType: Could not retrieve any repos")
			}
		}

		// There should only exist one of each specific repo pr course.
		// AssignmentRepo, CourseInfoRepo, SolutionRepo, TestRepo
		// There can exist many UserRepo, but only one pr user
		if len(repos) > 1 && len(repos) == 0 {
			return echo.NewHTTPError(http.StatusInternalServerError, "Too many course information repositories exists for this course")
		}

		// Have to be in string array to be able to jsonify so frontend recognize it.
		// See public/src/HttpHelper.ts -> send()
		var repoURL []string
		repoURL = append(repoURL, repos[0].HTMLURL)
		return c.JSONPretty(http.StatusOK, &repoURL, "\t")
	}
}

func updateRepoToPrivate(ctx context.Context, db database.Database, s scm.SCM, directoryID uint64) {
	repositories, err := db.GetRepositoriesByDirectory(directoryID)
	if err != nil {
		return
	}

	payment, _ := s.GetPaymentPlan(ctx, directoryID)
	// If privaterepos is bigger than 0, we know that the org/team is paid for.
	if payment.PrivateRepos > 0 {
		for _, repo := range repositories {
			if repo.Type != models.AssignmentsRepo &&
				repo.Type != models.CourseInfoRepo &&
				repo.Type != models.SolutionsRepo {

				scmRepo := &scm.Repository{
					DirectoryID: repo.DirectoryID,
					ID:          repo.RepositoryID,
				}
				err := s.UpdateRepository(ctx, scmRepo)
				if err != nil {
					return
				}
			}
		}
	}
}
