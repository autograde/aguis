import { IMap, MapHelper } from "../map";
import {
    IAssignment,
    ICourseLinkAssignment,
    ICourseUserLink,
    IError,
    IGroupCourse,
    INewGroup,
    isCourse,
    IStatusCode,
    IStudentSubmission,
    ISubmission,
    IUserCourse,
    IUserRelation,

} from "../models";

import { UserManager } from "../managers";
import { ILogger } from "./LogManager";
import { Course, Enrollment, User, Directory, Group, Void } from "../../proto/ag_pb";

export interface ICourseProvider {
    getCourses(): Promise<Course[]>;
    getAssignments(courseID: number): Promise<IMap<IAssignment>>;
    // getCoursesStudent(): Promise<ICourseUserLink[]>;
    getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<ICourseEnrollment[]>;
    getUsersForCourse(course: Course, state?: Enrollment.UserStatus[]): Promise<IUserEnrollment[]>;

    addUserToCourse(user: User, course: Course): Promise<boolean>;
    changeUserState(link: ICourseUserLink, state: Enrollment.UserStatus): Promise<boolean>;

    createNewCourse(courseData: Course): Promise<Course | IError>;
    getCourse(id: number): Promise<Course | null>;
    updateCourse(courseID: number, courseData: Course): Promise<Void | IError>;

    getCourseGroups(courseID: number): Promise<Group[]>;
    updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<boolean>;
    createGroup(groupData: INewGroup, courseId: number): Promise<Group | IError>;
    getGroup(groupID: number): Promise<Group | null>;
    deleteGroup(groupID: number): Promise<boolean>;
    getGroupByUserAndCourse(userid: number, courseid: number): Promise<Group | null>;
    updateGroup(groupData: Group): Promise<IStatusCode | IError>;
    // deleteCourse(id: number): Promise<boolean>;

    getAllLabInfos(courseID: number, userId: number): Promise<IMap<ISubmission>>;
    getAllGroupLabInfos(courseID: number, groupID: number): Promise<IMap<ISubmission>>;
    getDirectories(provider: string): Promise<Directory[]>;
    getProviders(): Promise<string[]>;
    refreshCoursesFor(courseID: number): Promise<any>;
    approveSubmission(submissionID: number): Promise<void>;

    getCourseInformationURL(cid: number): Promise<string>;
    getRepositoryURL(cid: number, type: number): Promise<string>;
}

export function isUserEnrollment(enroll: IEnrollment): enroll is ICourseEnrollment {
    if ((enroll as any).course) {
        return true;
    }
    return false;
}

export function isCourseEnrollment(enroll: IEnrollment): enroll is IUserEnrollment {
    if ((enroll as any).user) {
        return true;
    }
    return false;
}

export interface ICourseEnrollment extends IEnrollment {
    course: Course;
}

export interface IUserEnrollment extends IEnrollment {
    user: User;
    status: Enrollment.UserStatus;
}

export interface IEnrollment {
    userid: number;
    courseid: number;
    status?: Enrollment.UserStatus;

    course?: Course;
    user?: User;
}

export class CourseManager {
    private courseProvider: ICourseProvider;

    constructor(courseProvider: ICourseProvider, logger: ILogger) {
        this.courseProvider = courseProvider;
    }

    /**
     * Adds a user to a course
     * @param user The user to be added to a course
     * @param course The course the user should be added to
     * @returns True if succeeded and false otherwise
     */
    public async addUserToCourse(user: User, course: Course): Promise<boolean> {
        return this.courseProvider.addUserToCourse(user, course);
    }

    /**
     * Get a course from and id
     * @param ID The id of the course
     */
    public async getCourse(ID: number): Promise<Course | null> {
        // const a = (await this.courseProvider.getCourses())[id];
        // if (a) {
        //     return a;
        // }
        // return null;
        return await this.courseProvider.getCourse(ID);
    }

    /**
     * Get all the courses available at the server
     */
    public async getCourses(): Promise<Course[]> {
        // return MapHelper.toArray(await this.courseProvider.getCourses());
        return await this.courseProvider.getCourses();
    }

    public async getCoursesWithState(user: User): Promise<IUserCourse[]> {
        const userCourses = await this.courseProvider.getCoursesFor(user);
        const newMap = userCourses.map<IUserCourse>((ele) => {
            return {
                assignments: [],
                course: ele.course,
                link: ele.status !== undefined ?
                    { courseId: ele.courseid, userid: ele.userid, state: ele.status } : undefined,
            };
        });
        return newMap;
    }

    /**
     * returns all the courses
     * if user is enrolled to a course, enrolled field will have a non-negative value
     * else enrolled field will have -1
     * @param user
     * @returns {Promise<ICourseWithEnrollStatus[]>}
     */
    /*public async getCoursesWithEnrollStatus(user: IUser): Promise<ICourseWithEnrollStatus[]> {
        const userCourses = await this.courseProvider.getCoursesWithEnrollStatus(user);
        return userCourses;
    }*/

    /**
     * Get all courses related to a user
     * @param user The user to get courses to
     * @param state Optional. The state the relations should be in, all if not present
     */
    public async getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<Course[]> {
        return (await this.courseProvider.getCoursesFor(user, state)).map((ele) => ele.course);
    }

    /**
     * Retrives one assignment from a single course
     * @param course The course the assignment is in
     * @param assignmentID The id to the assignment
     */
    public async getAssignment(course: Course, assignmentID: number): Promise<IAssignment | null> {
        const assignments = await this.courseProvider.getAssignments(course.getId());
        const assign = assignments[assignmentID];
        if (assign) {
            return assign;
        }
        return null;
    }

    /**
     * Get all assignments in a single course
     * @param courseID The course id or ICourse to retrive assignments from
     */
    public async getAssignments(courseID: number | Course): Promise<IAssignment[]> {
        if (isCourse(courseID)) {
            courseID = courseID.getId();
        }
        return MapHelper.toArray(await this.courseProvider.getAssignments(courseID));
    }

    /**
     * Change the userstate for a relation between a course and a user
     * @param link The link to change state of
     * @param state The new state of the relation
     */
    public async changeUserState(link: ICourseUserLink, state: Enrollment.UserStatus): Promise<boolean> {
        return this.courseProvider.changeUserState(link, state);
    }

    /**
     * Creates a new course in the backend
     * @param courseData The course information to create a course from
     */
    public async createNewCourse(courseData: Course): Promise<Course | IError> {
        return this.courseProvider.createNewCourse(courseData);
    }

    /**
     * Updates a course with new information
     * @param courseData The new information for the course
     */
    public async updateCourse(courseID: number, courseData: Course): Promise<Void | IError> {
        return await this.courseProvider.updateCourse(courseID, courseData);
    }

    /**
     * Load an IUserCourse object for a single user and a single course
     * @param student The student the information should be retrived from
     * @param course The course the data should be loaded for
     */
    public async getStudentCourse(student: User, course: Course): Promise<IUserCourse | null> {
        const courses = await this.courseProvider.getCoursesFor(student);
        for (const crs of courses) {
            if (crs.courseid === course.getId()) {
                const userCourse: IUserCourse = {
                    link: crs.status !== undefined ?
                        { userid: student.getId(), courseId: course.getId(), state: crs.status } : undefined,
                    assignments: [],
                    course,
                };
                await this.fillLinks(student, userCourse);
                return userCourse;
            }
        }
        return null;
    }

    /**
     * Load an IUserCourse object for a single user and a single course
     * @param student The student the information should be retrived from
     * @param course The course the data should be loaded for
     */
    public async getStudentCourseForTeacher(student: IUserRelation, course: Course, assignments: IAssignment[]): Promise<IUserCourse | null> {
        const userCourse: IUserCourse = {
            link: { userid: student.user.getId(), courseId: course.getId(), state: student.link.state },
            assignments: [],
            course,
        };
        await this.fillLinks(student.user, userCourse, assignments);
        return userCourse;
    }

    /**
     * Loads a single IStudentSubmission for a student and an assignment.
     * This will contains information about an assignment and the lates
     * sumbission information related to that assignment.
     * @param student The student the information should be retrived from
     * @param assignment The assignment the data should be loaded for
     */
    public async getUserSubmittions(student: User, assignment: IAssignment): Promise<IStudentSubmission> {
        const labsInfo = MapHelper.find(await this.courseProvider.getAllLabInfos(assignment.courseid, student.getId()),
            (ele) => ele.userid === student.getId() && ele.assignmentid === assignment.id);
        if (labsInfo) {
            return {
                assignment,
                latest: labsInfo,
            };
        }
        return {
            assignment,
            latest: undefined,
        };
    }

    /**
     * Retrives all course relations, and courses related to a
     * a single student
     * @param student The student to load the information for
     */
    public async getStudentCourses(student: User, state?: Enrollment.UserStatus[]): Promise<ICourseLinkAssignment[]> {
        const links: IUserCourse[] = [];
        const userCourses = await this.courseProvider.getCoursesFor(student, state);
        for (const course of userCourses) {
            links.push({
                assignments: [],
                course: course.course,
                link: course.status !== undefined ?
                    { courseId: course.courseid, userid: student.getId(), state: course.status } : undefined,
            });
        }

        for (const link of links) {
            await this.fillLinks(student, link);
        }
        return links;
    }

    /**
     * Retrives all users related to a single course
     * @param course The course to retrive userinformation to
     * @param userMan Usermanager to be able to get user information
     * @param state Optional. The state of the user to course relation
     */
    public async getUsersForCourse(
        course: Course,
        userMan: UserManager,
        state?: Enrollment.UserStatus[]): Promise<IUserRelation[]> {

        return (await this.courseProvider.getUsersForCourse(course, state)).map<IUserRelation>((user) => {
            return {
                link: { courseId: course.getId(), userid: user.userid, state: user.status },
                user: user.user,
            };
        });
    }

    public async createGroup(groupData: INewGroup, courseID: number): Promise<Group | IError> {
        return await this.courseProvider.createGroup(groupData, courseID);
    }

    public async updateGroup(groupData: Group): Promise<IStatusCode | IError> {
        return await this.courseProvider.updateGroup(groupData);
    }

    /**
     * getCourseGroup returns all the groups under a course
     * @param courseID course id of a course
     */
    public async getCourseGroups(courseID: number): Promise<Group[]> {
        return await this.courseProvider.getCourseGroups(courseID);
    }

    /**
     * Load an IGroupCourse object for a single group and a single course
     * @param group The group the information should be retrived from
     * @param course The course the data should be loaded for
     */
    public async getGroupCourse(group: Group, course: Course): Promise<IGroupCourse | null> {
        // Fetching group enrollment status
        if (group.getCourseid() === course.getId()) {
            const groupCourse: IGroupCourse = {
                link: { groupid: group.getId(), courseId: course.getId(), state: group.getStatus()},
                assignments: [],
                course,
            };
            await this.fillLinksGroup(group, groupCourse);
            return groupCourse;
        }
        return null;
    }

    public async getGroupCourseForTeacher(group: Group, course: Course, assignments: IAssignment[]): Promise<IGroupCourse | null> {
        // Fetching group enrollment status
        if (group.getCourseid() === course.getId()) {
            const groupCourse: IGroupCourse = {
                link: { groupid: group.getId(), courseId: course.getId(), state: group.getStatus()},
                assignments: [],
                course,
            };
            await this.fillLinksGroup(group, groupCourse, assignments);
            return groupCourse;
        }
        return null;
    }

    public async getGroupByUserAndCourse(userID: number, courseID: number): Promise<Group | null> {
        return await this.courseProvider.getGroupByUserAndCourse(userID, courseID);
    }

    public async updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<boolean> {
        return await this.courseProvider.updateGroupStatus(groupID, status);
    }

    public async getGroup(groupID: number): Promise<Group | null> {
        return await this.courseProvider.getGroup(groupID);
    }

    public async deleteGroup(groupID: number): Promise<boolean> {
        return await this.courseProvider.deleteGroup(groupID);
    }

    public async refreshCoursesFor(courseID: number): Promise<any> {
        return await this.courseProvider.refreshCoursesFor(courseID);
    }

    /**
     * Get all available directories or organisations for a single provider
     * @param provider The provider to load information from, for instance github og gitlab
     */
    public async getDirectories(provider: string): Promise<Directory[]> {
        return await this.courseProvider.getDirectories(provider);
    }

    public async getProviders(): Promise<string[]> {
        return await this.courseProvider.getProviders();
    }

    public async getCourseInformationURL(cid: number): Promise<string> {
        return await this.courseProvider.getCourseInformationURL(cid);
    }

    public async getRepositoryURL(cid: number, type: number): Promise<string> {
        return this.courseProvider.getRepositoryURL(cid, type);
    }

    public async approveSubmission(submissionID: number): Promise<void> {
        return await this.courseProvider.approveSubmission(submissionID);
    }

    /**
     * Add IStudentSubmissions to an IUserCourse
     * @param student The student
     * @param studentCourse The student course
     */
    private async fillLinks(student: User, studentCourse: IUserCourse, assignments?: IAssignment[]): Promise<void> {
        if (!studentCourse.link) {
            return;
        }
        if (!assignments) {
            assignments = await this.getAssignments(studentCourse.course.getId());
        }
        if (assignments.length > 0) {
            const submissions = MapHelper.toArray(
                await this.courseProvider.getAllLabInfos(studentCourse.course.getId(), student.getId()));

            for (const a of assignments) {
                const submission = submissions.find((sub) => sub.assignmentid === a.id);
                studentCourse.assignments.push({ assignment: a, latest: submission });
            }
        }
    }

    /**
     * Add IStudentSubmissions to an IUserCourse
     * @param group The group
     * @param groupCourse The group course
     */
    private async fillLinksGroup(group: Group, groupCourse: IGroupCourse, assignments?: IAssignment[]): Promise<void> {
        if (!groupCourse.link) {
            return;
        }
        if (!assignments) {
            assignments = await this.getAssignments(groupCourse.course.getId());
        }
        if (assignments.length > 0) {
            const submissions = MapHelper.toArray(
                await this.courseProvider.getAllGroupLabInfos(groupCourse.course.getId(), group.getId()));

            for (const a of assignments) {
                const submission = submissions.find((sub) => sub.assignmentid === a.id);
                groupCourse.assignments.push({ assignment: a, latest: submission });
            }
        }
    }
}
