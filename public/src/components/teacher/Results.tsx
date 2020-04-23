import * as React from "react";
import { Assignment, Course } from "../../../proto/ag_pb";
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IStudentLabsForCourse, IStudentLab, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, sortByScore } from "./labHelper";
import { generateGitLink, searchForLabs } from '../../componentHelper';

interface IResultsProp {
    course: Course;
    courseURL: string;
    students: IStudentLabsForCourse[];
    labs: Assignment[];
    onApproveClick: (submissionID: number, approve: boolean) => Promise<boolean>;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<ISubmission | null>;
}

interface IResultsState {
    assignment?: IStudentLab;
    students: IStudentLabsForCourse[];
}

export class Results extends React.Component<IResultsProp, IResultsState> {

    constructor(props: IResultsProp) {
        super(props);

        const currentStudent = this.props.students.length > 0 ? this.props.students[0] : null;
        const courseAssignments = currentStudent ? currentStudent.course.getAssignmentsList() : null;
        if (currentStudent && courseAssignments && courseAssignments.length > 0) {
            this.state = {
                // Only using the first student to fetch assignments.
                assignment: currentStudent.labs[0],
                students: sortByScore(this.props.students, this.props.labs, false),
            };
        } else {
            this.state = {
                assignment: undefined,
                students: sortByScore(this.props.students, this.props.labs, false),
            };
        }
    }

    public render() {
        let studentLab: JSX.Element | null = null;
        const currentStudents = this.props.students.length > 0 ? this.props.students : null;
        if (currentStudents
            && this.state.assignment
            && !this.state.assignment.assignment.getIsgrouplab()
        ) {
            studentLab = <StudentLab
                assignment={this.state.assignment}
                showApprove={true}
                onRebuildClick={
                    async () => {
                        if (this.state.assignment && this.state.assignment.submission) {
                            const ans = await this.props.onRebuildClick(this.state.assignment.assignment.getId(), this.state.assignment.submission.id);
                            if (ans) {
                                this.state.assignment.submission = ans;
                                return true;
                            }
                        }
                        return false;
                    }
                }
                onApproveClick={ async (approve: boolean) => {
                    if (this.state.assignment && this.state.assignment.submission) {
                        const ans = await this.props.onApproveClick(this.state.assignment.submission.id, approve);
                        if (ans) {
                            this.state.assignment.submission.approved = approve;
                        }
                    }
                }}
            />;
        }

        return (
            <div>
                <h1>Result: {this.props.course.getName()}</h1>
                <Row>
                    <div key="resultshead" className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
                            placeholder="Search for students"
                            onChange={(query) => this.handleSearch(query)}
                        />
                        <DynamicTable header={this.getResultHeader()}
                            data={this.state.students}
                            selector={(item: IStudentLabsForCourse) => this.getResultSelector(item)}
                        />
                    </div>
                    <div key="resultsbody" className="col-lg-6 col-md-6 col-sm-12">
                        {studentLab}
                    </div>
                </Row>
            </div>
        );
    }

    private getResultHeader(): string[] {
        let headers: string[] = ["Name"];
        headers = headers.concat(this.props.labs.filter((e) => !e.getIsgrouplab()).map((e) => e.getName()));
        return headers;
    }

    private getResultSelector(student: IStudentLabsForCourse): (string | JSX.Element | ICellElement)[] {
        const user = student.enrollment.getUser();
        const displayName = user ? this.generateUserRepoLink(user.getName(), user.getLogin()) : "";
        let selector: (string | JSX.Element | ICellElement)[] = [displayName];
        selector = selector.concat(student.labs.filter((e, i) => !e.assignment.getIsgrouplab()).map(
            (e, i) => {
                let cellCss: string = "";
                if (e.submission) {
                    cellCss = generateCellClass(e);
                }
                const iCell: ICellElement = {
                    value: <a className={cellCss + " lab-cell-link"}
                        onClick={() => this.handleOnclick(e)}
                        href="#">
                        {e.submission ? (e.submission.score + "%") : "N/A"}</a>,
                    className: cellCss,
                };
                return iCell;
            }));
        return selector;
    }

    private generateUserRepoLink(name: string, userName: string): JSX.Element {
        return <a href={generateGitLink(userName, this.props.courseURL)} target="_blank">{ name }</a>;
    }

    private handleOnclick(item: IStudentLab): void {
        this.setState({
            assignment: item,
        });
    }

    private handleSearch(query: string): void {
        this.setState({
            students: searchForLabs(this.props.students, query),
        });
    }
}
