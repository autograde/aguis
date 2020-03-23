package database

import (
	pb "github.com/autograde/aguis/ag"
	"github.com/jinzhu/gorm"
)

// CreateSubmission creates a new submission record or updates the most
// recent submission, as defined by the provided submissionQuery.
// The submissionQuery must always specify the assignment, and may specify the ID of
// either an individual student or a group, but not both.
func (db *GormDB) CreateSubmission(submission *pb.Submission) error {
	// Primary key must be greater than 0.
	if submission.AssignmentID < 1 {
		return gorm.ErrRecordNotFound
	}

	// Either user or group id must be set, but not both.
	var m *gorm.DB
	switch {
	case submission.UserID > 0 && submission.GroupID > 0:
		return gorm.ErrRecordNotFound
	case submission.UserID > 0:
		m = db.conn.First(&pb.User{ID: submission.UserID})
	case submission.GroupID > 0:
		m = db.conn.First(&pb.Group{ID: submission.GroupID})
	default:
		return gorm.ErrRecordNotFound
	}

	// Check that user/group with given ID exists.
	var group uint64
	if err := m.Count(&group).Error; err != nil {
		return err
	}

	// Checks that the assignment exists.
	var assignment uint64
	if err := db.conn.Model(&pb.Assignment{}).Where(&pb.Assignment{
		ID: submission.AssignmentID,
	}).Count(&assignment).Error; err != nil {
		return err
	}

	if assignment+group != 2 {
		return gorm.ErrRecordNotFound
	}

	// Make a new submission struct for the database query to check
	// whether a submission record for the given lab and user/group
	// already exists. We cannot reuse the incoming submission
	// because the query would attempt to match all the test result
	// fields as well.
	query := &pb.Submission{
		AssignmentID: submission.GetAssignmentID(),
		UserID:       submission.GetUserID(),
		GroupID:      submission.GetGroupID(),
	}

	// We want the last record as there can be multiple submissions
	// for the same student/group and lab in the database.
	if err := db.conn.Last(query, query).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// If a submission for the given assignment and student/group already exists, update it.
	// Otherwise create a new submission record
	var labSubmission pb.Submission
	err := db.conn.Where(query).Assign(submission).FirstOrCreate(&labSubmission).Error
	submission.ID = labSubmission.GetID()
	return err
}

// GetSubmission fetches a submission record.
func (db *GormDB) GetSubmission(query *pb.Submission) (*pb.Submission, error) {
	var submission pb.Submission
	if err := db.conn.Where(query).Last(&submission).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetSubmissions returns all submissions for the active assignment for the given course.
func (db *GormDB) GetSubmissions(courseID uint64, query *pb.Submission) ([]*pb.Submission, error) {
	var course pb.Course
	if err := db.conn.Preload("Assignments").First(&course, courseID).Error; err != nil {
		return nil, err
	}

	// note that, this creates a query with possibly both user and group;
	// it will only limit the number of submissions returned if both are supplied.
	q := &pb.Submission{UserID: query.GetUserID(), GroupID: query.GetGroupID()}
	var latestSubs []*pb.Submission
	for _, a := range course.Assignments {
		q.AssignmentID = a.GetID()
		temp, err := db.GetSubmission(q)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return nil, err
		}
		latestSubs = append(latestSubs, temp)
	}
	return latestSubs, nil
}

// GetCourseSubmissions returns all individual lab submissions or group submissions for the course ID
// depending on the provided groupLabs boolean.
func (db *GormDB) GetCourseSubmissions(courseID uint64, groupLabs bool) ([]pb.Submission, error) {
	m := db.conn

	// fetch the course entry with all associated assignments and active enrollments
	course, err := db.GetCourse(courseID, true)
	if err != nil {
		return nil, err
	}

	// get IDs of all individual labs or group labs for the course
	courseAssignmentIDs := make([]uint64, 0)
	for _, a := range course.Assignments {
		if a.IsGroupLab == groupLabs {
			// collect either group labs or non-group labs
			// but not both in the same collection.
			courseAssignmentIDs = append(courseAssignmentIDs, a.GetID())
		}
	}

	var allLabs []pb.Submission
	if err := m.Where("assignment_id IN (?)", courseAssignmentIDs).Find(&allLabs).Error; err != nil {
		return nil, err
	}

	return allLabs, nil
}

// UpdateSubmission updates submission with the given approved status.
func (db *GormDB) UpdateSubmission(sid uint64, approved bool) error {
	return db.conn.
		Model(&pb.Submission{}).
		Where(&pb.Submission{ID: sid}).
		Update("approved", approved).Error
}