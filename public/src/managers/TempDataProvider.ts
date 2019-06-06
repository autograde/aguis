import * as Models from "../models";
import {
    IAssignment,
    ICourseUserLink,
    IError,
    INewGroup,
    IStatusCode,
    ISubmission,
} from "../models";
import { ICourseProvider } from "./CourseManager";

import { IMap, MapHelper, mapify } from "../map";
import { IUserProvider } from "./UserManager";

import { ICourseEnrollment, IUserEnrollment } from "../managers";
import { Course, Group, Enrollment, User, Directory } from "../../proto/ag_pb";
import { isNull } from "util";

interface IDummyUser extends User {
    password: string;
}

interface IGrpcDummyUser {
    user: User;
    password: string;
}

/**
 * The TempDataProvider is a fake implemetation of the backend
 * to be able to simulate the backend for easier developtment
 */
export class TempDataProvider implements IUserProvider, ICourseProvider {

    private localUsers: IMap<IGrpcDummyUser>;
    private localAssignments: IMap<IAssignment>;
    private localCourses: IMap<Course>;
    private localCourseStudent: ICourseUserLink[];
    private localLabInfo: IMap<ISubmission>;
    private localCourseGroups: Group[];

    private currentLoggedIn: User | null = null;

    constructor() {
        this.addLocalAssignments();
        this.addLocalCourses();
        this.addLocalCourseStudent();
        this.addLocalUsers();
        this.addLocalLabInfo();
        this.addLocalCourseGroups();
    }

    public async approveSubmission(submissionid: number): Promise<void> {
        throw new Error("Method not implemented.");
    }

    public async getDirectories(provider: string): Promise<Directory[]> {
        throw new Error("Not implemented");
    }

    public async getAllUser(): Promise<User[]> {
        const users: User[] = [];
        const dummyUsers =  MapHelper.toArray(this.localUsers);
        dummyUsers.forEach((ele) => {
            users.push(ele.user);
        });
        return users;
    }

    public async getCourses(): Promise<Course[]> {
        // return this.localCourses;
        return MapHelper.toArray(this.localCourses);
    }

    public async getCoursesStudent(): Promise<ICourseUserLink[]> {
        return this.localCourseStudent;
    }

    public async getAssignments(courseId: number): Promise<IMap<IAssignment>> {
        const temp: IMap<IAssignment> = [];
        MapHelper.forEach(this.localAssignments, (a, i) => {
            if (a.courseid === courseId) {
                temp[i] = a;
            }
        });
        return temp;
    }

    public async tryLogin(username: string, password: string): Promise<User | null> {
        const user = MapHelper.find(this.localUsers, (u) =>
            u.user.getEmail().toLocaleLowerCase() === username.toLocaleLowerCase());
        if (user && user.password === password) {
            this.currentLoggedIn = user.user;
            return user.user;
        }
        return null;
    }

    public async tryRemoteLogin(provider: string): Promise< User | null> {
        let lookup = "test@testersen.no";
        if (provider === "gitlab") {
            lookup = "bob@bobsen.no";
        }
        const user = MapHelper.find(this.localUsers, (u) =>
            u.user.getEmail().toLocaleLowerCase() === lookup);

        return new Promise< User | null>((resolve, reject) => {
            // Simulate async callback
            setTimeout(() => {
                if (isNull(user)) {
                    this.currentLoggedIn = user;
                    resolve(user);
                } else {
                    this.currentLoggedIn = user.user;
                    resolve(user.user);
                }
                
            }, 500);
        });
    }

    public async logout(user: User): Promise<boolean> {
        return true;
    }

    public async addUserToCourse(user: User, course: Course): Promise<boolean> {
        this.localCourseStudent.push({
            courseId: course.getId(),
            userid: user.getId(),
            state: Enrollment.UserStatus.PENDING,
        });
        return true;
    }

    /**
     * Get all userlinks for a single course
     * @param course The course userlinks should be retrived from
     * @param state Optinal. The state of the relation, all if not present
     */
    public async getUserLinksForCourse(course: Course, state?: Enrollment.UserStatus[]): Promise<ICourseUserLink[]> {
        const users: ICourseUserLink[] = [];
        for (const c of await this.getCoursesStudent()) {
            if (course.getId() === c.courseId && (state === undefined || c.state === Enrollment.UserStatus.STUDENT)) {
                users.push(c);
            }
        }
        return users;
    }

    public async getUsersAsMap(ids: number[]): Promise<IMap<User>> {
        const returnUsers: IMap<User> = {};
        const allUsers = await this.getAllUser();
        ids.forEach((ele) => {
            const temp = allUsers[ele];
            if (temp) {
                returnUsers[ele] = temp;
            }
        });
        return returnUsers;
    }

    public async getUsersForCourse(course: Course, state?: Enrollment.UserStatus[])
        : Promise<IUserEnrollment[]> {
        const courseStds: ICourseUserLink[] =
            await this.getUserLinksForCourse(course, state);
        const users = await this.getUsersAsMap(courseStds.map((e) => e.userid));
        return courseStds.map<IUserEnrollment>((link) => {
            const user = users[link.userid];
            if (!user) {
                // TODO: See if we should have an error here or not
                throw new Error("Link exist witout a user object");
            }
            return { courseid: link.courseId, userid: link.userid, user, status: link.state };
        });
    }

    public async createNewCourse(course: any): Promise<Course | IError> {
        throw new Error("Method not implemented");
    }

    public async getCourse(id: number): Promise<Course | null> {
        const course: Course | undefined = this.localCourses[id];
        if (course) {
            return course;
        }
        return null;
    }

    public async updateCourse(courseId: number, courseData: Course): Promise<IStatusCode | IError> {
        throw new Error("Method not implemented");
    }

    public async changeUserState(link: ICourseUserLink, state: Enrollment.UserStatus): Promise<boolean> {
        link.state = state;
        return true;
    }

    public async changeAdminRole(user: User): Promise<boolean> {
        user.setIsadmin(!user.getIsadmin());
        return true;
    }

    public async getAllLabInfos(courseId: number): Promise<IMap<ISubmission>> {
        const temp: IMap<ISubmission> = {};
        const assignments = await this.getAssignments(courseId);
        MapHelper.forEach(this.localLabInfo, (ele) => {
            if (assignments[ele.assignmentid]) {
                temp[ele.id] = ele;
            }
        });
        return temp;
    }

    public async getProviders(): Promise<string[]> {
        return ["github"];
    }

    public async grpcGetLoggedInUser(): Promise <User| null > {
        return this.currentLoggedIn;
    }

    public async getLoggedInUser(): Promise< User | null> {
        return this.currentLoggedIn;
    }

    public async getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<ICourseEnrollment[]> {
        const cLinks: ICourseUserLink[] = [];
        const temp = await this.getCoursesStudent();
        for (const c of temp) {
            if (user.getId() === c.userid && (state === undefined || c.state === Enrollment.UserStatus.STUDENT)) {
                cLinks.push(c);
            }
        }
        const courses: ICourseEnrollment[] = [];
        const tempCourses = await this.getCourses();
        for (const link of cLinks) {
            const c = tempCourses[link.courseId];
            if (c) {
                courses.push({ course: c, courseid: link.courseId, userid: link.userid, status: link.state });
            }
        }
        return courses;
    }
    public async createGroup(groupData: INewGroup, courseId: number): Promise<Group | IError> {
        throw new Error("Method not implemented");
    }
    public async getCourseGroups(courseId: number): Promise<Group[]> {
        return this.localCourseGroups;
    }

    public async deleteGroup(groupId: number): Promise<boolean> {
        throw new Error("Method not implemented");
    }

    public async getGroupByUserAndCourse(userid: number, courseid: number): Promise<Group | null> {
        throw new Error("Method not implemented");
    }

    public async updateGroupStatus(groupId: number, status: Group.GroupStatus): Promise<boolean> {
        throw new Error("Method not implemented");
    }
    public async getGroup(gid: number): Promise<Group | null> {
        throw new Error("Method not implemented");
    }
    public async updateGroup(groupData: Group): Promise<IStatusCode | IError> {
        throw new Error("Method not implemented");
    }
    public async getAllGroupLabInfos(courseId: number, groupID: number): Promise<IMap<Models.ISubmission>> {
        throw new Error("Method not implemented.");
    }

    public async refreshCoursesFor(courseid: number): Promise<any> {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                resolve({});
            }, 10);
        });
    }

    public async updateUser(user: User): Promise<boolean> {

        const tempUser = this.localUsers[user.getId()];
        if (tempUser) {
            tempUser.user.setName(user.getName());
            tempUser.user.setEmail(user.getEmail());
            tempUser.user.setStudentid(user.getStudentid());
            tempUser.user.setIsadmin(user.getIsadmin());
        }

        return Promise.resolve(true);
    }

    public async getCourseInformationURL(cid: number): Promise<string> {
        throw new Error("Method not implemented.");
    }

    public async getRepositoryURL(cid: number, type: number): Promise<string> {
        throw new Error("Method not implemented.");
    }

    private addLocalUsers() {
        const dummyUsers: IGrpcDummyUser[] = [];
        
        const tempUser: User = new User();
        
        tempUser.setId(999);
        tempUser.setName("Test Testersen");
        tempUser.setEmail("test@testersen.no");
        tempUser.setStudentid("9999");
        tempUser.setIsadmin(true);
        let tempDummy = {user: tempUser, password: "1234"} as IGrpcDummyUser;
        dummyUsers.push(tempDummy);

        tempUser.setId(1000);
        tempUser.setName("Admin Admin");
        tempUser.setEmail("admin@admin");
        tempUser.setStudentid("1000");
        tempUser.setIsadmin(true);
        let tempDummy1 = {user: tempUser, password: "1234"} as IGrpcDummyUser;
        dummyUsers.push(tempDummy1);

        tempUser.setId(1);
        tempUser.setName("Per Pettersen");
        tempUser.setEmail("per@pettersen.no");
        tempUser.setStudentid("1234");
        tempUser.setIsadmin(true);
        let tempDummy2 = {user: tempUser, password: "1234"} as IGrpcDummyUser;
        dummyUsers.push(tempDummy2);

        tempUser.setId(2);
        tempUser.setName("Bob Bobsen");
        tempUser.setEmail("bob@bobsen.no");
        tempUser.setStudentid("1234");
        tempUser.setIsadmin(true);
        let tempDummy3 = {user: tempUser, password: "1234"} as IGrpcDummyUser;
        dummyUsers.push(tempDummy3);

        tempUser.setId(3);
        tempUser.setName("Petter Pan");
        tempUser.setEmail("petter@pan.no");
        tempUser.setStudentid("1234");
        tempUser.setIsadmin(true);
        let tempDummy4 = {user: tempUser, password: "1234"} as IGrpcDummyUser;
        dummyUsers.push(tempDummy4);

        this.localUsers = mapify(dummyUsers, (ele) => ele.user.getId());
    }

    private addLocalAssignments() {
        this.localAssignments = mapify([
            {
                id: 0,
                courseid: 0,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 1,
                courseid: 0,
                name: "Lab 2",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 2,
                courseid: 0,
                name: "Lab 3",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 3,
                courseid: 0,
                name: "Lab 4",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 4,
                courseid: 1,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 5,
                courseid: 1,
                name: "Lab 2",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 6,
                courseid: 1,
                name: "Lab 3",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 7,
                courseid: 2,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 8,
                courseid: 2,
                name: "Lab 2",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 9,
                courseid: 3,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 10,
                courseid: 4,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
        ] as IAssignment[], (ele) => ele.id);
    }

    private addLocalCourses() {

        const tempCourses: Course[] = [];
        const course0 = new Course();
        const course1 = new Course();
        const course2 = new Course();
        const course3 = new Course();
        const course4 = new Course();

        course0.setId(0);
        course0.setName("Object Oriented Programming");
        course0.setCode("DAT100");
        course0.setTag("Spring");
        course0.setYear(2017);
        course0.setProvider("github");
        course0.setDirectoryid(23650610);

        course1.setId(1);
        course1.setName("Algorithms and Datastructures");
        course1.setCode("DAT200");
        course1.setTag("Spring");
        course1.setYear(2017);
        course1.setProvider("github");
        course1.setDirectoryid(23650611);

        course2.setId(2);
        course2.setName("Databases");
        course2.setCode("DAT220");
        course2.setTag("Spring");
        course2.setYear(2017);
        course2.setProvider("github");
        course2.setDirectoryid(23650612);

        course3.setId(3);
        course3.setName("Communication Technology");
        course3.setCode("DAT230");
        course3.setTag("Spring");
        course3.setYear(2017);
        course3.setProvider("github");
        course3.setDirectoryid(23650613);

        course4.setId(4);
        course4.setName("Operating Systems");
        course4.setCode("DAT320");
        course4.setTag("Spring");
        course4.setYear(2017);
        course4.setProvider("github");
        course4.setDirectoryid(23650614);

        tempCourses.push(course0);
        tempCourses.push(course1);
        tempCourses.push(course2);
        tempCourses.push(course3);
        tempCourses.push(course4);
       
        this.localCourses = mapify(tempCourses, (ele) => ele.getId());
    }

    private addLocalCourseStudent() {
        this.localCourseStudent = [
            { courseId: 0, userid: 999, state: 2 },
            { courseId: 1, userid: 999, state: 2 },
            { courseId: 0, userid: 1, state: 0 },
            { courseId: 0, userid: 2, state: 0 },
        ] as ICourseUserLink[];
    }

    private addLocalLabInfo() {
        this.localLabInfo = mapify<ISubmission>([
            {
                id: 1,
                assignmentid: 0,
                userid: 999,
                buildId: 1,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 1",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
            {
                id: 2,
                assignmentid: 1,
                userid: 999,
                buildId: 2,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 2",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
            {
                id: 3,
                assignmentid: 2,
                userid: 999,
                buildId: 3,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 3",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
            {
                id: 4,
                assignmentid: 3,
                userid: 999,
                buildId: 4,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 4",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
            {
                id: 5,
                assignmentid: 4,
                userid: 999,
                buildId: 5,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 5",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
        ] as ISubmission[], (ele: ISubmission) => {
            return ele.id;
        });
    }

    private getLocalDirectories(): Directory[] {
        const localDirectory = new Directory();
        localDirectory.setId(23650610);
        localDirectory.setPath("dat520-2017");
        localDirectory.setAvatar("https://avatars2.githubusercontent.com/u/23650610?v=3");

        const localDirectories: Directory[] = [];
        localDirectories.push(localDirectory);
        return localDirectories;
        /*return (
            [
                {
                    id: 23650610,
                    path: "dat520-2017",
                    avatar: "https://avatars2.githubusercontent.com/u/23650610?v=3",
                },
            ]
        );*/
    }

    private addLocalCourseGroups(): void {

        /*

        this.localCourseGroups = [
            {
                id: 1,
                name: "Group1",
                status: Group.GroupStatus.APPROVED,
                courseid: 1,
                users: [
                    {
                        id: 1,
                        email: "test@example.com",
                        name: "Student 1",
                        isadmin: false,
                        studentid: "12345",
                        avatarurl: "",
                    },
                    {
                        id: 2,
                        email: "test2@example.com",
                        name: "Student 2",
                        isadmin: false,
                        studentid: "12346",
                        avatarurl: "",
                    },
                ],
            },
            {
                id: 2,
                name: "Group2",
                status: Group.GroupStatus.PENDING_GROUP,
                courseid: 1,
                users: [
                    {
                        id: 3,
                        email: "tes3t@example.com",
                        name: "Student 3",
                        isadmin: false,
                        studentid: "12347",
                        avatarurl: "",
                    },
                    {
                        id: 4,
                        email: "test4@example.com",
                        name: "Student 4",
                        isadmin: false,
                        studentid: "12348",
                        avatarurl: "",
                    },
                ],
            },
        ];*/
    }

}