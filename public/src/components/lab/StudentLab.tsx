import * as React from "react";
import { Course } from "../../../proto/ag_pb";
import { IStudentSubmission } from "../../models";
import { LabResultView } from "../../pages/views/LabResultView";

interface IStudentLabProps {
    course: Course;
    assignment: IStudentSubmission;
    showApprove: boolean;
    onApproveClick: () => void;
    onRebuildClick: (submissionID: number) => Promise<boolean>;
}

export class StudentLab extends React.Component<IStudentLabProps> {
    public render() {
        return <LabResultView
            course={this.props.course}
            labInfo={this.props.assignment}
            onApproveClick={this.props.onApproveClick}
            onRebuildClick={this.props.onRebuildClick}
            showApprove={this.props.showApprove}>
        </LabResultView>;
    }
}
