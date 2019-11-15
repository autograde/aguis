import * as React from "react";
import { Assignment } from "../../../proto/ag_pb";
import { Row } from "../../components"
import { ISubmission } from "../../models";

interface ILastBuildInfo {
    submission: ISubmission;
    assignment: Assignment;
    }

interface ILastBuildInfoState {
    rebuilding: boolean;
}

export class LastBuildInfo extends React.Component<ILastBuildInfo, ILastBuildInfoState> {

    constructor(props: ILastBuildInfo) {
        super(props);
        this.state = {
            rebuilding: false,
         };
    }
    public render() {
        return (
            <div>
                <Row>
                <div className="col-lg-12">
                    <table className="table">
                        <thead><tr><th colSpan={2}>Lab Information </th></tr></thead>
                        <tbody>
        <tr><td>Delivered</td><td>{this.getDeliveredTime(this.props.submission.buildDate)}</td></tr>
        <tr><td>Deadline</td><td>{this.props.assignment.getDeadline()}</td></tr>
                            <tr><td>Slipdays</td><td>5</td></tr>
        <tr><td>Execution time</td><td>{this.props.submission.executetionTime / 1000} s</td></tr>
                        </tbody>
                    </table>
                </div>
            </Row>

            <Row>
                <div className="col-lg-12">
                    <table className="table">
                        <thead><tr><th>Tests: </th><th>Passed</th><th>Failed</th></tr></thead>
                        <tbody>
    <tr><td></td><td>{this.props.submission.passedTests}</td><td>{this.props.submission.failedTests}</td></tr>
                        </tbody>
                    </table>
                </div>
            </Row>

            </div>
        );
    }

    private getDeliveredTime(date: Date): string {
        return date ? date.toDateString() : "-";
    }

}
