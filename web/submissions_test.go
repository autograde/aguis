package web_test

import (
	"context"
	"reflect"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

func TestSubmissionsAccess(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)

	teacher := createFakeUser(t, db, 2)
	db.UpdateUser(&pb.User{ID: teacher.ID, IsAdmin: true})
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}

	student1 := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student1.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	student2 := createFakeUser(t, db, 4)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student2.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	student3 := createFakeUser(t, db, 5)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student3.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), teacher)

	fakeProvider.CreateOrganization(context.Background(),
		&scm.CreateOrgOptions{Path: "path", Name: "name"},
	)

	users := []*pb.User{student1, student2}
	group_req := &pb.Group{Name: "Test group", CourseID: course.ID, Users: users}

	_, err := ags.CreateGroup(ctx, group_req)
	if err != nil {
		t.Fatal(err)
	}

	// at this stage we have a course teacher, two students enrolled in the course in the same group,
	// and one student and admin not affilated with the course

	if err = db.CreateAssignment(&pb.Assignment{
		CourseID:    course.ID,
		Name:        "lab1",
		Language:    "go",
		Deadline:    "11.11.2022",
		AutoApprove: false,
		Order:       1,
		IsGroupLab:  false,
	}); err != nil {
		t.Fatal(err)
	}

	if err = db.CreateAssignment(&pb.Assignment{
		CourseID:    course.ID,
		Name:        "lab2",
		Language:    "go",
		Deadline:    "11.11.2022",
		AutoApprove: false,
		Order:       2,
		IsGroupLab:  true,
	}); err != nil {
		t.Fatal(err)
	}

	if err = db.CreateSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student1.ID,
	}); err != nil {
		t.Fatal(err)
	}

	if err = db.CreateSubmission(&pb.Submission{
		AssignmentID: 2,
		GroupID:      1,
	}); err != nil {
		t.Fatal(err)
	}

	if err = db.CreateSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student3.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// check that all three submissions have been successfully added to the database
	submission1, err := db.GetSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student1.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	submission2, err := db.GetSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student3.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	submission3, err := db.GetSubmission(&pb.Submission{
		AssignmentID: 2,
		GroupID:      1,
	})
	if err != nil {
		t.Fatal(err)
	}

	allSubmissions := []*pb.Submission{submission1, submission2, submission3}
	latestSubmissions := []*pb.Submission{submission2, submission3}

	// there must be exactly three submissions for given course and assignment in the database
	if len(allSubmissions) != 3 {
		t.Errorf("Expected 3 submissions, got %d: %+v", len(allSubmissions), allSubmissions)
	}

	// teacher must be able to access all of the latest course submissions
	haveSubmissions, err := ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(latestSubmissions, haveSubmissions.GetSubmissions()) {
		t.Errorf("Teacher got submissions %+v, expected submissions: %+v", haveSubmissions.GetSubmissions(), latestSubmissions)
	}

	// admin not enrolled in the course must not be able to access any course submissions
	ctx = withUserContext(context.Background(), admin)
	haveSubmissions, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID})
	if err == nil {
		t.Error("Expected error: user ")
	}
	if len(haveSubmissions.GetSubmissions()) > 0 {
		t.Errorf("Not enrolled admin should not see any submissions, got submisions: %v+ ", haveSubmissions.GetSubmissions())
	}

	// enroll admin as course student
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: admin.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   admin.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	haveSubmissions, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}
	// enrolled as student, admin must be able to access all course submissions
	if !reflect.DeepEqual(latestSubmissions, haveSubmissions.GetSubmissions()) {
		t.Errorf("Admin got submissions %+v, expected submissions: %+v", haveSubmissions.GetSubmissions(), latestSubmissions)
	}

	// the first student must be able to access own submissions as well as submissions made by group he has membership in
	ctx = withUserContext(context.Background(), student1)

	personalSubmission, err := ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, UserID: student1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(personalSubmission.GetSubmissions()) != 1 {
		t.Error("Expected one submission, got ", len(personalSubmission.GetSubmissions()))
	}
	groupSubmission, err := ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, GroupID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(groupSubmission.GetSubmissions()) != 1 {
		t.Error("Expected one submission, got ", len(groupSubmission.GetSubmissions()))
	}

	wantSubmissions := []*pb.Submission{submission1, submission3}
	student1Submissions := []*pb.Submission{personalSubmission.GetSubmissions()[0], groupSubmission.GetSubmissions()[0]}
	if !reflect.DeepEqual(wantSubmissions, student1Submissions) {
		t.Errorf("Student 1 got submissions %+v, expected submissions: %+v", haveSubmissions.GetSubmissions(), wantSubmissions)
	}

	// the second student should not be able to access the submission by student1
	ctx = withUserContext(context.Background(), student2)
	personalSubmission, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, UserID: student1.ID})
	if err == nil {
		t.Error("Expected error: only owner and teachers can get submissions")
	}

	// the second student should no longer be able to access group submissions when removed from the group

	if err = db.UpdateGroup(&pb.Group{
		ID:       1,
		CourseID: course.ID,
		Users:    []*pb.User{student1},
	}); err != nil {
		t.Fatal(err)
	}

	groupSubmission, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, GroupID: 1})
	if err == nil {
		t.Error("Expected error: only owner and teachers can get submissions")
	}

	// the third student (not enrolled in the course) should not be able to access submission even if it belongs to that student
	ctx = withUserContext(context.Background(), student3)
	personalSubmission, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, UserID: student3.ID})
	if err == nil {
		t.Error("Expected error: only owner and teachers can get submissions")
	}
}

func TestApproveSubmission(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)

	course := allCourses[0]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	student := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	lab := &pb.Assignment{
		CourseID: course.ID,
		Name:     "test lab",
		Language: "go",
		Order:    1,
	}
	if err = db.CreateAssignment(lab); err != nil {
		t.Fatal(err)
	}

	wantSubmission := &pb.Submission{
		AssignmentID: lab.ID,
		UserID:       student.ID,
		Score:        17,
	}
	if err = db.CreateSubmission(wantSubmission); err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), admin)

	fakeProvider.CreateOrganization(context.Background(),
		&scm.CreateOrgOptions{Path: "path", Name: "name"},
	)

	if _, err = ags.UpdateSubmission(ctx, &pb.UpdateSubmissionRequest{
		SubmissionID: wantSubmission.ID,
		CourseID:     course.ID,
		Approve:      true,
	}); err != nil {
		t.Fatal(err)
	}

	updatedSubmission, err := db.GetSubmission(&pb.Submission{ID: wantSubmission.ID})
	wantSubmission.Approved = true

	if !reflect.DeepEqual(wantSubmission.GetApproved(), updatedSubmission.GetApproved()) {
		t.Errorf("Expected submission approval to be %+v, got: %+v", wantSubmission.GetApproved(), updatedSubmission.GetApproved())
	}

	if _, err = ags.UpdateSubmission(ctx, &pb.UpdateSubmissionRequest{
		SubmissionID: wantSubmission.ID,
		CourseID:     course.ID,
		Approve:      false,
	}); err != nil {
		t.Fatal(err)
	}

	updatedSubmission, err = db.GetSubmission(&pb.Submission{ID: wantSubmission.ID})
	wantSubmission.Approved = false

	if !reflect.DeepEqual(wantSubmission.GetApproved(), updatedSubmission.GetApproved()) {
		t.Errorf("Expected submission approval to be %+v, got: %+v", wantSubmission.GetApproved(), updatedSubmission.GetApproved())
	}
}

func TestGetCourseLabSubmissions(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)

	course1 := allCourses[2]
	course2 := allCourses[3]
	if err := db.CreateCourse(admin.ID, course1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateCourse(admin.ID, course2); err != nil {
		t.Fatal(err)
	}

	student := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student.ID, CourseID: course1.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student.ID, CourseID: course2.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student.ID,
		CourseID: course1.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student.ID,
		CourseID: course2.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	// make labs with similar lab names for both courses
	lab1c1 := &pb.Assignment{
		CourseID: course1.ID,
		Name:     "lab 1",
		Language: "go",
		Deadline: "2020-02-23T18:00:00",
		Order:    1,
	}

	lab2c1 := &pb.Assignment{
		CourseID: course1.ID,
		Name:     "lab 2",
		Language: "go",
		Deadline: "2020-03-23T18:00:00",
		Order:    2,
	}
	lab1c2 := &pb.Assignment{
		CourseID: course2.ID,
		Name:     "lab 1",
		Language: "go",
		Deadline: "2020-04-23T18:00:00",
		Order:    1,
	}
	lab2c2 := &pb.Assignment{
		CourseID: course2.ID,
		Name:     "lab 2",
		Language: "go",
		Deadline: "2020-05-23T18:00:00",
		Order:    2,
	}
	if err := db.CreateAssignment(lab1c1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateAssignment(lab2c1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateAssignment(lab1c2); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateAssignment(lab2c2); err != nil {
		t.Fatal(err)
	}

	sub1 := &pb.Submission{
		UserID:       student.ID,
		AssignmentID: lab1c1.ID,
		Score:        44,
	}
	sub2 := &pb.Submission{
		UserID:       student.ID,
		AssignmentID: lab2c2.ID,
		Score:        66,
	}
	if err := db.CreateSubmission(sub1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSubmission(sub2); err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), admin)

	fakeProvider.CreateOrganization(context.Background(),
		&scm.CreateOrgOptions{Path: "path", Name: "name"},
	)

	// check that all assignments were saved for the correct courses
	wantAssignments1 := []*pb.Assignment{lab1c1, lab2c1}
	wantAssignments2 := []*pb.Assignment{lab1c2, lab2c2}

	haveAssignments1, err := ags.GetAssignments(ctx, &pb.CourseRequest{CourseID: course1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantAssignments1, haveAssignments1.GetAssignments()) {
		t.Errorf("Expected assigments for course 1: %+v, got %+v", wantAssignments1, haveAssignments1.GetAssignments())
	}
	haveAssignments2, err := ags.GetAssignments(ctx, &pb.CourseRequest{CourseID: course2.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantAssignments2, haveAssignments2.GetAssignments()) {
		t.Errorf("Expected assigments for course 2: %+v, got %+v", wantAssignments1, haveAssignments1.GetAssignments())
	}

	// check that all submissions were saved for the correct labs
	labsForCourse1, err := ags.GetSubmissionsByCourse(ctx, &pb.LabRequest{CourseID: course1.ID})
	if err != nil {
		t.Fatal(err)
	}
	for _, labLink := range labsForCourse1.GetLabs() {
		if labLink.GetEnrollment().GetUserID() == student.ID {
			labs := labLink.GetSubmissions()
			if len(labs) != 1 {
				t.Errorf("Expected 1 submission for course 1, got %d", len(labs))
			}
			if !reflect.DeepEqual(sub1, labs[0]) {
				t.Errorf("Want submission %+v, got %+v", sub1, labs[0])
			}
		}
	}

	labsForCourse2, err := ags.GetSubmissionsByCourse(ctx, &pb.LabRequest{CourseID: course2.ID})
	if err != nil {
		t.Fatal(err)
	}
	for _, labLink := range labsForCourse2.GetLabs() {
		if labLink.GetEnrollment().GetUserID() == student.ID {
			labs := labLink.GetSubmissions()
			if len(labs) != 1 {
				t.Errorf("Expected 1 submission for course 1, got %d", len(labs))
			}
			if !reflect.DeepEqual(sub2, labs[0]) {
				t.Errorf("Want submission %+v, got %+v", sub1, labs[0])
			}
		}
	}
	// check that no submissions will be returned for a wrong course ID
	if _, err = ags.GetSubmissionsByCourse(ctx, &pb.LabRequest{CourseID: 234}); err == nil {
		t.Error("Expected 'no submissions found'")
	}

	// check that method fails with empty context
	if _, err = ags.GetSubmissionsByCourse(context.Background(), &pb.LabRequest{CourseID: course1.ID}); err == nil {
		t.Error("Expected 'authorization failed. please try to logout and sign in again'")
	}

	// check that method fails for non-teacher user
	ctx = withUserContext(ctx, student)
	if _, err = ags.GetSubmissionsByCourse(context.Background(), &pb.LabRequest{CourseID: course1.ID}); err == nil {
		t.Error("Expected 'only teachers can get all lab submissions'")
	}

}
