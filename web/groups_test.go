package web_test

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/grpcservice"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc/metadata"
)

func TestNewGroup(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpcservice.NewAutograderService(db, testscms, web.BaseHookOptions{})

	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(admin.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	// Prepare provider
	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(cont,
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	testscms["token"] = fakeProvider

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{ID: user.ID})
	group_req := &pb.Group{Name: "Hein's Group", CourseID: course.ID, Users: users}

	respGroup, err := test_ag.CreateGroup(cont, group_req)
	if err != nil {
		t.Fatal(err)
	}

	group, err := test_ag.GetGroup(cont, &pb.RecordRequest{ID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}

	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	//group.Enrollments = nil
	if !reflect.DeepEqual(&respGroup, &group) {
		t.Errorf("have response group %+v, while database has %+v", respGroup, group)
	}
}

func TestNewGroupTeacherCreator(t *testing.T) {
	const route = "/courses/:cid/groups"

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollTeacher(teacher.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpcservice.NewAutograderService(db, testscms, web.BaseHookOptions{})

	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(teacher.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(cont,
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	testscms["token"] = fakeProvider

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{ID: user.ID})
	group_req := &pb.Group{Name: "Hein's Group", CourseID: course.ID, Users: users}

	respGroup, err := test_ag.CreateGroup(cont, group_req)
	if err != nil {
		t.Fatal(err)
	}

	group, err := test_ag.GetGroup(cont, &pb.RecordRequest{ID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}

	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	//group.Enrollments = nil
	if !cmp.Equal(respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", respGroup, group)
	}
}

func TestNewGroupStudentCreateGroupWithTeacher(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollTeacher(teacher.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpcservice.NewAutograderService(db, testscms, web.BaseHookOptions{})

	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(user.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(cont,
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	testscms["token"] = fakeProvider

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{ID: user.ID})
	users = append(users, &pb.User{ID: teacher.ID})
	group_req := &pb.Group{Name: "Hein's Group", CourseID: course.ID, Users: users}

	_, err = test_ag.CreateGroup(cont, group_req)
	if err == nil {
		t.Error("Student trying to enroll teacher should not be possible!")
	}
}
func TestStudentCreateNewGroupTeacherUpdateGroup(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	testscms := make(map[string]scm.SCM)
	test_ag := grpcservice.NewAutograderService(db, testscms, web.BaseHookOptions{})
	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(context.Background(),
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	testscms["token"] = fakeProvider

	admin := createFakeUser(t, db, 1)
	course := pb.Course{Provider: "fake", DirectoryID: 1}
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollTeacher(teacher.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	user1 := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user1.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	user2 := createFakeUser(t, db, 4)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user2.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	user3 := createFakeUser(t, db, 5)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user3.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user3.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// group with two students
	users := make([]*pb.User, 0)
	users = append(users, user1)
	users = append(users, user2)
	newGroupReq := &pb.Group{Name: "Hein's two member Group", TeamID: 1, CourseID: course.ID, Users: users}

	// set ID of user3 to context, user3 is not member of group (should fail)
	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(user3.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	if _, err = test_ag.CreateGroup(cont, newGroupReq); err == nil {
		t.Error("expected error 'student must be member of new group'")
	}

	// set ID of user1, which is group member
	meta = metadata.New(map[string]string{"user": strconv.Itoa(int(user1.ID))})
	cont = metadata.NewIncomingContext(context.Background(), meta)

	respGroup, err := test_ag.CreateGroup(cont, newGroupReq)
	if err != nil {
		t.Fatal(err)
	}

	group, err := test_ag.GetGroup(cont, &pb.RecordRequest{ID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}

	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	//group.Enrollments = nil
	if !reflect.DeepEqual(respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", respGroup, group)
	}

	// ******************* Admin/Teacher UpdateGroup *******************

	// group with three students

	users1 := make([]*pb.User, 0)
	users1 = append(users1, user1)
	users1 = append(users1, user2)
	users1 = append(users1, user3)

	updateGroupReq := &pb.Group{ID: group.ID, Name: "Hein's three member Group", CourseID: course.ID, Users: users1}

	// set admin ID in context
	meta = metadata.New(map[string]string{"user": strconv.Itoa(int(admin.ID))})
	cont = metadata.NewIncomingContext(context.Background(), meta)

	_, err = test_ag.UpdateGroup(cont, updateGroupReq) //test_ag.UpdateGroup(cont, updateGroupReq)
	if err != nil {
		t.Error(err)
	}

	// check that the group have changed group membership
	haveGroup, err := db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}
	userIDs := make([]uint64, 0)
	for _, usr := range haveGroup.Users {
		userIDs = append(userIDs, usr.ID)
	}

	grpUsers, err := db.GetUsers(userIDs...)
	if err != nil {
		t.Fatal(err)
	}

	wantGroup := group
	wantGroup.Name = updateGroupReq.Name
	wantGroup.Users = grpUsers
	// UpdateGroup will autoApprove group on update
	wantGroup.Status = pb.Group_APPROVED
	haveGroup.Enrollments = nil
	wantGroup.Enrollments = nil
	if !cmp.Equal(haveGroup, wantGroup) {
		t.Errorf("have group %+v", haveGroup)
		t.Errorf("want group %+v", wantGroup)
	}

	// ******************* Teacher Only UpdateGroup *******************

	// change group to only one student
	users2 := make([]*pb.User, 0)
	users2 = append(users2, user1)
	updateGroupReq1 := &pb.Group{ID: group.ID, Name: "Hein's single member Group", CourseID: course.ID, Users: users2}

	// set teacher ID in context
	meta = metadata.New(map[string]string{"user": strconv.Itoa(int(teacher.ID))})
	cont = metadata.NewIncomingContext(context.Background(), meta)

	_, err = test_ag.UpdateGroup(cont, updateGroupReq1)
	if err != nil {
		t.Error(err)
	}
	// check that the group have changed group membership
	haveGroup, err = db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}
	userIDs = make([]uint64, 0)
	for _, usr := range updateGroupReq1.Users {
		userIDs = append(userIDs, usr.ID)
	}

	grpUsers, err = db.GetUsers(userIDs...)
	if err != nil {
		t.Fatal(err)
	}

	if len(haveGroup.Users) != 1 {
		t.Fatal("expected only single member group")
	}
	wantGroup = updateGroupReq
	wantGroup.Name = updateGroupReq1.Name
	wantGroup.Users = grpUsers
	// UpdateGroup will autoApprove group on update
	wantGroup.Status = pb.Group_APPROVED
	wantGroup.TeamID = 1
	haveGroup.Enrollments = nil
	wantGroup.Enrollments = nil
	if !reflect.DeepEqual(wantGroup, haveGroup) {
		t.Errorf("have group %+v", haveGroup)
		t.Errorf("want group %+v", wantGroup)
	}
}

func TestDeleteGroup(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	testCourse := pb.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
	}
	admin := createFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.ID, &testCourse); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: testCourse.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, testCourse.ID); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{Name: "Test Delete Group", CourseID: testCourse.ID, Users: []*pb.User{user}}

	testscms := make(map[string]scm.SCM)
	test_ag := grpcservice.NewAutograderService(db, testscms, web.BaseHookOptions{})
	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(user.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	respGroup, err := test_ag.CreateGroup(cont, group)
	if err != nil {
		t.Fatal(err)
	}

	_, err = test_ag.DeleteGroup(cont, respGroup)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetGroup(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	testCourse := pb.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
	}
	admin := createFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.ID, &testCourse); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: testCourse.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, testCourse.ID); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpcservice.NewAutograderService(db, testscms, web.BaseHookOptions{})
	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(user.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	group := &pb.Group{Name: "Test Group", CourseID: testCourse.ID, Users: []*pb.User{user}}
	respGroup, err := test_ag.CreateGroup(cont, group)
	if err != nil {
		t.Fatal(err)
	}

	gotGroup, err := test_ag.GetGroup(cont, &pb.RecordRequest{ID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotGroup, respGroup) {
		t.Errorf("have response group %+v, while database has %+v", &gotGroup, &respGroup)
	}

}

func TestPatchGroupStatus(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	course := pb.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
		ID:          1,
	}

	admin := createFakeUser(t, db, 1)
	err := db.CreateCourse(admin.ID, &course)
	if err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpcservice.NewAutograderService(db, testscms, web.BaseHookOptions{})
	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(admin.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	f := scm.NewFakeSCMClient()
	if _, err := f.CreateDirectory(context.Background(), &scm.CreateDirectoryOptions{
		Name: course.Code,
		Path: course.Code,
	}); err != nil {
		t.Fatal(err)
	}
	testscms["token"] = f

	user1 := createFakeUser(t, db, 2)
	user2 := createFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user1.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user1.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user2.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user2.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{
		ID:       1,
		Name:     "Test Group",
		CourseID: course.ID,
		Users:    []*pb.User{user1, user2},
		TeamID:   1,
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}
	// get the group as stored in db with enrollments
	prePatchGroup, err := db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}

	prePatchGroup.Status = pb.Group_APPROVED
	_, err = test_ag.UpdateGroup(cont, prePatchGroup)
	if err != nil {
		t.Error(err)
	}

	// check that the group didn't change
	haveGroup, err := db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(prePatchGroup, haveGroup) {
		t.Errorf("have group %+v want %+v", haveGroup, prePatchGroup)
	}
}

func TestGetGroupByUserAndCourse(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	course := pb.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
		ID:          1,
	}

	admin := createFakeUser(t, db, 1)
	err := db.CreateCourse(admin.ID, &course)
	if err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpcservice.NewAutograderService(db, testscms, web.BaseHookOptions{})
	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(admin.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	user1 := createFakeUser(t, db, 2)
	user2 := createFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user1.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user1.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user2.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user2.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{
		ID:       1,
		CourseID: course.ID,
		Users:    []*pb.User{user1, user2},
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}

	respGroup, err := test_ag.GetGroupByUserAndCourse(cont, &pb.ActionRequest{UserID: user1.ID, CourseID: course.ID})
	if err != nil {
		t.Error(err)
	}

	dbGroup, err := test_ag.GetGroup(cont, &pb.RecordRequest{ID: group.ID})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(respGroup, dbGroup) {
		t.Errorf("have response group %+v, while database has %+v", respGroup, dbGroup)
	}
}

func TestDeleteApprovedGroup(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	course := allCourses[0]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpcservice.NewAutograderService(db, testscms, web.BaseHookOptions{})
	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(admin.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	f := scm.NewFakeSCMClient()
	if _, err := f.CreateDirectory(context.Background(), &scm.CreateDirectoryOptions{
		Name: course.Code,
		Path: course.Code,
	}); err != nil {
		t.Fatal(err)
	}
	testscms["token"] = f

	user1 := createFakeUser(t, db, 2)
	user2 := createFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user1.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user2.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollTeacher(admin.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{
		ID:       1,
		CourseID: course.ID,
		Name:     "Test Group",
		Users:    []*pb.User{user1, user2},
	}
	createdGroup, err := test_ag.CreateGroup(cont, group)
	if err != nil {
		t.Fatal(err)
	}

	// first approve the group
	createdGroup.Status = pb.Group_APPROVED
	if _, err = test_ag.UpdateGroup(cont, createdGroup); err != nil {
		t.Fatal(err)
	}

	// then get user enrollments with group ID
	enr1, err := db.GetEnrollmentByCourseAndUser(course.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}
	enr2, err := db.GetEnrollmentByCourseAndUser(course.ID, user2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// reject the group
	createdGroup.Status = pb.Group_REJECTED
	if _, err = test_ag.UpdateGroup(cont, createdGroup); err != nil {
		t.Fatal(err)
	}

	// get updated enrollments of group members
	newEnr1, err := db.GetEnrollmentByCourseAndUser(course.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}
	newEnr2, err := db.GetEnrollmentByCourseAndUser(course.ID, user2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// now nullify manually group ID for original enrollments
	enr1.GroupID = 0
	enr2.GroupID = 0

	// then check that new enrollments have group IDs nullified automatically
	if !reflect.DeepEqual(enr1, newEnr1) {
		t.Errorf("want enrollment %+v, while database has %+v", enr1, newEnr1)
	}
	if !reflect.DeepEqual(enr2, newEnr2) {
		t.Errorf("want enrollment %+v, while database has %+v", enr2, newEnr2)
	}
}
