package database_test

import (
	"testing"

	pb "github.com/autograde/aguis/ag"
)

func TestGetNextAssignment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	_, err := db.GetNextAssignment(0, 0, 0)
	if err == nil {
		t.Fatal("expected error 'record not found'")
	}

	course := pb.Course{
		Name:           "Distributed Systems",
		Code:           "DAT520",
		Year:           2018,
		Tag:            "Spring",
		Provider:       "fake",
		OrganizationID: 1,
	}

	// create course as teacher
	teacher := createFakeUser(t, db, 10)
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}

	// create and enroll user as student
	user := createFakeUser(t, db, 11)
	if err := db.CreateEnrollment(&pb.Enrollment{CourseID: course.ID, UserID: user.ID}); err != nil {
		t.Fatal(err)
	}
	if err = db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// create group with single student
	group := pb.Group{
		CourseID: course.ID,
		Users: []*pb.User{
			{ID: user.ID},
		},
	}
	if err := db.CreateGroup(&group); err != nil {
		t.Fatal(err)
	}

	_, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if err == nil {
		t.Fatal("expected error 'no assignments found for course 1'")
	}

	// create assignments for course
	assignment1 := pb.Assignment{CourseID: course.ID, Order: 1}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := pb.Assignment{CourseID: course.ID, Order: 2}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := pb.Assignment{CourseID: course.ID, Order: 3, IsGroupLab: true}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}
	assignment4 := pb.Assignment{CourseID: course.ID, Order: 4}
	if err := db.CreateAssignment(&assignment4); err != nil {
		t.Fatal(err)
	}

	_, err = db.GetNextAssignment(course.ID, 0, 0)
	if err == nil {
		t.Fatal("expected error 'record not found'")
	}

	nxtUnapproved, err := db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assignment1.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment1.ID, nxtUnapproved.ID)
	}

	// send new submission for assignment1
	submission1 := pb.Submission{AssignmentID: assignment1.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	// send another submission for assignment1
	submission2 := pb.Submission{AssignmentID: assignment1.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	// send new submission for assignment2
	submission3 := pb.Submission{AssignmentID: assignment2.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}
	// send new submission for assignment3
	submission4 := pb.Submission{AssignmentID: assignment3.ID, GroupID: group.ID}
	if err := db.CreateSubmission(&submission4); err != nil {
		t.Fatal(err)
	}

	// we haven't approved any of the submissions yet; expect same result as above

	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assignment1.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment1.ID, nxtUnapproved.ID)
	}

	// approve submission1
	if err := db.UpdateSubmission(submission1.ID, true); err != nil {
		t.Fatal(err)
	}

	// we have approved the first submission of the first assignment, but since
	// we two submissions for assignment1, this won't change the next to approve.
	// TODO(meling) Is this the desired semantics for this??
	// That is, it seems more reasonable to have a function ApproveAssignment(assignment, user)
	// that finds the latest submission for the user and marks it approved.
	// That is, maybe the UpdateSubmissionById shouldn't be exported.
	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}

	// we approved assignment 1, so the next unapproved should be assignment 2? no idea why it
	if nxtUnapproved.ID != assignment2.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment2.ID, nxtUnapproved.ID)
	}

	// approve submission2
	if err := db.UpdateSubmission(submission2.ID, true); err != nil {
		t.Fatal(err)
	}

	// now the first assignment is approved, moving on to the second
	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assignment2.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment2.ID, nxtUnapproved.ID)
	}

	// approve submission3
	if err := db.UpdateSubmission(submission3.ID, true); err != nil {
		t.Fatal(err)
	}

	// now the second assignment is approved, moving on to the third
	// this fails because the next assignment to approve is a group lab,
	// and we don't provide a group id.
	_, err = db.GetNextAssignment(course.ID, user.ID, 0)
	//TODO(meling) GetNextAssignment semantics has changed; needs to be updated when we understand better what is needed
	// if err == nil {
	// 	t.Fatal("expected error 'record not found'")
	// }

	// moving on to the third assignment, using the group Id this time.
	// fails because user id must be provided.
	_, err = db.GetNextAssignment(course.ID, 0, group.ID)
	//TODO(meling) GetNextAssignment semantics has changed; needs to be updated when we understand better what is needed
	// if err == nil {
	// 	t.Fatal("expected error 'user id must be provided'")
	// }

	// moving on to the third assignment, using both user id and group id this time.
	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assignment3.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment3.ID, nxtUnapproved.ID)
	}

	// approve submission4 for assignment3 (the group lab)
	if err := db.UpdateSubmission(submission4.ID, true); err != nil {
		t.Fatal(err)
	}

	// approving the 4th submission (for assignment3, which is a group lab),
	// should fail because we only provIde user Id, and no group.ID.
	_, err = db.GetNextAssignment(course.ID, user.ID, 0)
	//TODO(meling) GetNextAssignment semantics has changed; needs to be updated when we understand better what is needed
	// if err == nil {
	// 	t.Fatal("expected error 'user id must be provided'")
	// }

	// here it should pass since we also provide the group id.
	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assignment4.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment4.ID, nxtUnapproved.ID)
	}

	// send new submission for assignment4
	submission5 := pb.Submission{AssignmentID: assignment4.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission5); err != nil {
		t.Fatal(err)
	}
	// approve submission5
	if err := db.UpdateSubmission(submission5.ID, true); err != nil {
		t.Fatal(err)
	}
	// all assignments have been approved
	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if nxtUnapproved != nil || err == nil {
		t.Fatal("expected error 'all assignments approved'")
	}
}
