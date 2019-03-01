package web

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/autograde/aguis/scm"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/sirupsen/logrus"

	webhooks "gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
	"gopkg.in/go-playground/webhooks.v3/gitlab"
)

// GithubHook handles webhook events from GitHub.
func GithubHook(logger logrus.FieldLogger, db database.Database, runner ci.Runner, scriptPath string) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := github.Event(h.Get("X-GitHub-Event"))

		switch event {
		case github.PushEvent:
			p := payload.(github.PushPayload)
			logger.WithField("payload", p).Println("Push event")

			repo, err := db.GetRepository(uint64(p.Repository.ID))
			if err != nil {
				logger.WithError(err).Error("Failed to get repository from database")
				return
			}
			logger.WithField("repo", repo).Info("Found repository, moving on")

			switch {
			case repo.IsTestsRepo():
				// the push event is for the 'tests' repo, which means that we
				// should update the course data (assignments) in the database
				refreshAssignmentsFromTestsRepo(logger, db, repo, uint64(p.Sender.ID))

			case repo.IsStudentRepo():
				// the push event is from a student or group repo; run the tests
				runTests(logger, db, runner, repo, p.Repository.CloneURL, p.HeadCommit.ID, scriptPath)

			default:
				logger.Info("Nothing to do for this push event")
			}

		default:
			logger.WithFields(logrus.Fields{
				"event":   event,
				"payload": payload,
				"header":  h,
			}).Warn("Event not implemented")
		}
	}
}

func refreshAssignmentsFromTestsRepo(logger logrus.FieldLogger, db database.Database, repo *models.Repository, senderID uint64) {
	logger.Info("Refreshing course informaton in database")

	remoteIdentity, err := db.GetRemoteIdentity("github", senderID)
	if err != nil {
		logger.WithError(err).Error("Failed to get sender's remote identity")
		return
	}
	logger.WithField("identity", remoteIdentity).Info("Found sender's remote identity")

	s, err := scm.NewSCMClient("github", remoteIdentity.AccessToken)
	if err != nil {
		logger.WithError(err).Error("Failed to create SCM Client")
		return
	}

	course, err := db.GetCourseByDirectoryID(repo.DirectoryID)
	if err != nil {
		logger.WithError(err).Error("Failed to get course from database")
		return
	}

	assignments, err := FetchAssignments(context.Background(), s, course)
	if err != nil {
		logger.WithError(err).Error("Failed to fetch assignments from 'tests' repository")
	}
	if err = db.UpdateAssignments(assignments); err != nil {
		logger.WithError(err).Error("Failed to update assignments in database")
	}
}

// runTests runs the ci from a RemoteIdentity
func runTests(logger logrus.FieldLogger, db database.Database, runner ci.Runner, repo *models.Repository,
	getURL string, commitHash string, scriptPath string) {

	course, err := db.GetCourseByDirectoryID(repo.DirectoryID)
	if err != nil {
		logger.WithError(err).Error("Failed to get course from database")
		return
	}

	courseCreator, err := db.GetUser(course.CourseCreatorID)
	if err != nil || len(courseCreator.RemoteIdentities) < 1 {
		logger.WithError(err).Error("Failed to fetch course creator")
	}

	selectedAssignment, err := db.GetNextAssignment(course.ID, repo.UserID, repo.GroupID)
	if err != nil {
		logger.WithError(err).Error("Failed to find a next unapproved assignment")
		return
	}
	logger.WithField("Assignment", selectedAssignment).Info("Found assignment")

	testRepos, err := db.GetRepositoriesByCourseIDAndType(course.ID, models.TestsRepo)
	if err != nil || len(testRepos) < 1 {
		logger.WithError(err).Error("Failed to find test repository in database")
		return
	}
	getURLTest := testRepos[0].HTMLURL
	logger.WithField("url", getURL).Info("Code Repository")
	logger.WithField("url", getURLTest).Info("Test repository")

	randomSecret := randomSecret()
	info := ci.AssignmentInfo{
		CreatorAccessToken: courseCreator.RemoteIdentities[0].AccessToken,
		AssignmentName:     selectedAssignment.Name,
		Language:           selectedAssignment.Language,
		GetURL:             getURL,
		TestURL:            getURLTest,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(getURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
		RandomSecret:       randomSecret,
	}
	job, err := ci.ParseScriptTemplate(scriptPath, info)
	if err != nil {
		logger.WithError(err).Error("Failed to parse script template")
		return
	}

	start := time.Now()
	out, err := runner.Run(context.Background(), job)
	if err != nil {
		logger.WithError(err).Error("Docker execution failed")
		return
	}
	execTime := time.Since(start)
	logger.WithField("out", out).WithField("execTime", execTime).Info("Docker execution successful")

	result, err := ci.ExtractResult(out, randomSecret, execTime)
	if err != nil {
		logger.WithError(err).Error("Failed to extract results from log")
		return
	}
	buildInfo, scores, err := result.Marshal()
	if err != nil {
		logger.WithError(err).Error("Failed to marshal build info and scores")
	}
	logger.WithField("result", result).Info("Extracted results")

	err = db.CreateSubmission(&models.Submission{
		AssignmentID: selectedAssignment.ID,
		BuildInfo:    buildInfo,
		CommitHash:   commitHash,
		Score:        result.TotalScore(),
		ScoreObjects: scores,
		UserID:       repo.UserID,
		GroupID:      repo.GroupID,
	})
	if err != nil {
		logger.WithError(err).Error("Failed to add submission to database")
		return
	}
}

func randomSecret() string {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
}

// GitlabHook handles events from Gitlab.
func GitlabHook(logger logrus.FieldLogger) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := gitlab.Event(h.Get("X-Gitlab-Event"))

		switch event {
		case gitlab.PushEvents:
			p := payload.(gitlab.PushEventPayload)
			logger.WithField("payload", p).Println("Push event")
		default:
			logger.WithFields(logrus.Fields{
				"event":   event,
				"payload": payload,
				"header":  h,
			}).Warn("Event not implemented")
		}
	}
}
