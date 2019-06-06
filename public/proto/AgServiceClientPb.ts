/**
 * @fileoverview gRPC-Web generated client stub for 
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


import * as grpcWeb from 'grpc-web';

import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb';

import {
  ActionRequest,
  Assignments,
  Course,
  Courses,
  Directories,
  DirectoryRequest,
  Enrollments,
  Group,
  Groups,
  Providers,
  RecordRequest,
  Repository,
  RepositoryRequest,
  Submission,
  Submissions,
  URLResponse,
  User,
  Users,
  Void} from './ag_pb';

export class AutograderServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: string; };

  constructor (hostname: string,
               credentials: null | { [index: string]: string; },
               options: null | { [index: string]: string; }) {
    if (!options) options = {};
    options['format'] = 'binary';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodInfoGetUser = new grpcWeb.AbstractClientBase.MethodInfo(
    User,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    User.deserializeBinary
  );

  getUser(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: User) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetUser',
      request,
      metadata || {},
      this.methodInfoGetUser,
      callback);
  }

  methodInfoGetUsers = new grpcWeb.AbstractClientBase.MethodInfo(
    Users,
    (request: Void) => {
      return request.serializeBinary();
    },
    Users.deserializeBinary
  );

  getUsers(
    request: Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Users) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetUsers',
      request,
      metadata || {},
      this.methodInfoGetUsers,
      callback);
  }

  methodInfoUpdateUser = new grpcWeb.AbstractClientBase.MethodInfo(
    User,
    (request: User) => {
      return request.serializeBinary();
    },
    User.deserializeBinary
  );

  updateUser(
    request: User,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: User) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateUser',
      request,
      metadata || {},
      this.methodInfoUpdateUser,
      callback);
  }

  methodInfoGetGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    Group,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    Group.deserializeBinary
  );

  getGroup(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Group) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetGroup',
      request,
      metadata || {},
      this.methodInfoGetGroup,
      callback);
  }

  methodInfoGetGroupByUserAndCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Group,
    (request: ActionRequest) => {
      return request.serializeBinary();
    },
    Group.deserializeBinary
  );

  getGroupByUserAndCourse(
    request: ActionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Group) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetGroupByUserAndCourse',
      request,
      metadata || {},
      this.methodInfoGetGroupByUserAndCourse,
      callback);
  }

  methodInfoGetGroups = new grpcWeb.AbstractClientBase.MethodInfo(
    Groups,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    Groups.deserializeBinary
  );

  getGroups(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Groups) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetGroups',
      request,
      metadata || {},
      this.methodInfoGetGroups,
      callback);
  }

  methodInfoCreateGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    Group,
    (request: Group) => {
      return request.serializeBinary();
    },
    Group.deserializeBinary
  );

  createGroup(
    request: Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Group) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/CreateGroup',
      request,
      metadata || {},
      this.methodInfoCreateGroup,
      callback);
  }

  methodInfoUpdateGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: Group) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateGroup(
    request: Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateGroup',
      request,
      metadata || {},
      this.methodInfoUpdateGroup,
      callback);
  }

  methodInfoDeleteGroup = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: Group) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  deleteGroup(
    request: Group,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/DeleteGroup',
      request,
      metadata || {},
      this.methodInfoDeleteGroup,
      callback);
  }

  methodInfoGetCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Course,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    Course.deserializeBinary
  );

  getCourse(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Course) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetCourse',
      request,
      metadata || {},
      this.methodInfoGetCourse,
      callback);
  }

  methodInfoGetCourses = new grpcWeb.AbstractClientBase.MethodInfo(
    Courses,
    (request: Void) => {
      return request.serializeBinary();
    },
    Courses.deserializeBinary
  );

  getCourses(
    request: Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Courses) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetCourses',
      request,
      metadata || {},
      this.methodInfoGetCourses,
      callback);
  }

  methodInfoGetCoursesWithEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(
    Courses,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    Courses.deserializeBinary
  );

  getCoursesWithEnrollment(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Courses) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetCoursesWithEnrollment',
      request,
      metadata || {},
      this.methodInfoGetCoursesWithEnrollment,
      callback);
  }

  methodInfoGetCourseInformationURL = new grpcWeb.AbstractClientBase.MethodInfo(
    URLResponse,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    URLResponse.deserializeBinary
  );

  getCourseInformationURL(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: URLResponse) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetCourseInformationURL',
      request,
      metadata || {},
      this.methodInfoGetCourseInformationURL,
      callback);
  }

  methodInfoCreateCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Course,
    (request: Course) => {
      return request.serializeBinary();
    },
    Course.deserializeBinary
  );

  createCourse(
    request: Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Course) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/CreateCourse',
      request,
      metadata || {},
      this.methodInfoCreateCourse,
      callback);
  }

  methodInfoUpdateCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: Course) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateCourse(
    request: Course,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateCourse',
      request,
      metadata || {},
      this.methodInfoUpdateCourse,
      callback);
  }

  methodInfoRefreshCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Assignments,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    Assignments.deserializeBinary
  );

  refreshCourse(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Assignments) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/RefreshCourse',
      request,
      metadata || {},
      this.methodInfoRefreshCourse,
      callback);
  }

  methodInfoGetEnrollmentsByCourse = new grpcWeb.AbstractClientBase.MethodInfo(
    Enrollments,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    Enrollments.deserializeBinary
  );

  getEnrollmentsByCourse(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Enrollments) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetEnrollmentsByCourse',
      request,
      metadata || {},
      this.methodInfoGetEnrollmentsByCourse,
      callback);
  }

  methodInfoCreateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: ActionRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  createEnrollment(
    request: ActionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/CreateEnrollment',
      request,
      metadata || {},
      this.methodInfoCreateEnrollment,
      callback);
  }

  methodInfoUpdateEnrollment = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: ActionRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateEnrollment(
    request: ActionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateEnrollment',
      request,
      metadata || {},
      this.methodInfoUpdateEnrollment,
      callback);
  }

  methodInfoGetSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(
    Submissions,
    (request: ActionRequest) => {
      return request.serializeBinary();
    },
    Submissions.deserializeBinary
  );

  getSubmissions(
    request: ActionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Submissions) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetSubmissions',
      request,
      metadata || {},
      this.methodInfoGetSubmissions,
      callback);
  }

  methodInfoGetSubmission = new grpcWeb.AbstractClientBase.MethodInfo(
    Submission,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    Submission.deserializeBinary
  );

  getSubmission(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Submission) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetSubmission',
      request,
      metadata || {},
      this.methodInfoGetSubmission,
      callback);
  }

  methodInfoGetGroupSubmissions = new grpcWeb.AbstractClientBase.MethodInfo(
    Submissions,
    (request: ActionRequest) => {
      return request.serializeBinary();
    },
    Submissions.deserializeBinary
  );

  getGroupSubmissions(
    request: ActionRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Submissions) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetGroupSubmissions',
      request,
      metadata || {},
      this.methodInfoGetGroupSubmissions,
      callback);
  }

  methodInfoUpdateSubmission = new grpcWeb.AbstractClientBase.MethodInfo(
    Void,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    Void.deserializeBinary
  );

  updateSubmission(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Void) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/UpdateSubmission',
      request,
      metadata || {},
      this.methodInfoUpdateSubmission,
      callback);
  }

  methodInfoGetAssignments = new grpcWeb.AbstractClientBase.MethodInfo(
    Assignments,
    (request: RecordRequest) => {
      return request.serializeBinary();
    },
    Assignments.deserializeBinary
  );

  getAssignments(
    request: RecordRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Assignments) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetAssignments',
      request,
      metadata || {},
      this.methodInfoGetAssignments,
      callback);
  }

  methodInfoGetRepositoryURL = new grpcWeb.AbstractClientBase.MethodInfo(
    URLResponse,
    (request: RepositoryRequest) => {
      return request.serializeBinary();
    },
    URLResponse.deserializeBinary
  );

  getRepositoryURL(
    request: RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: URLResponse) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetRepositoryURL',
      request,
      metadata || {},
      this.methodInfoGetRepositoryURL,
      callback);
  }

  methodInfoGetProviders = new grpcWeb.AbstractClientBase.MethodInfo(
    Providers,
    (request: Void) => {
      return request.serializeBinary();
    },
    Providers.deserializeBinary
  );

  getProviders(
    request: Void,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Providers) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetProviders',
      request,
      metadata || {},
      this.methodInfoGetProviders,
      callback);
  }

  methodInfoGetDirectories = new grpcWeb.AbstractClientBase.MethodInfo(
    Directories,
    (request: DirectoryRequest) => {
      return request.serializeBinary();
    },
    Directories.deserializeBinary
  );

  getDirectories(
    request: DirectoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Directories) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetDirectories',
      request,
      metadata || {},
      this.methodInfoGetDirectories,
      callback);
  }

  methodInfoGetRepository = new grpcWeb.AbstractClientBase.MethodInfo(
    Repository,
    (request: RepositoryRequest) => {
      return request.serializeBinary();
    },
    Repository.deserializeBinary
  );

  getRepository(
    request: RepositoryRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.Error,
               response: Repository) => void) {
    return this.client_.rpcCall(
      this.hostname_ +
        '/AutograderService/GetRepository',
      request,
      metadata || {},
      this.methodInfoGetRepository,
      callback);
  }

}

