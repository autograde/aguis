import * as React from "react";
import { Enrollment, User } from "../../../proto/ag_pb";
import { BootstrapButton, BootstrapClass, DynamicTable, Search } from "../../components";
import { ILink, NavigationManager, UserManager } from "../../managers";

import { LiDropDownMenu } from "../../components/navigation/LiDropDownMenu";
import { generateGitLink, searchForStudents } from '../../componentHelper';

interface IUserViewerProps {
    users: Enrollment[];
    isCourseList: boolean;
    userMan?: UserManager;
    navMan?: NavigationManager;
    courseURL: string;
    searchable?: boolean;
    actions?: ILink[];
    optionalActions?: (enrol: Enrollment) => ILink[];
    linkType?: ActionType;
    actionClick?: (enrollment: Enrollment, link: ILink) => void;
}

export enum ActionType {
    None,
    Menu,
    InRow,
}

interface IUserViewerState {
    enrollments: Enrollment[];
}

export class UserView extends React.Component<IUserViewerProps, IUserViewerState> {

    public constructor(props: IUserViewerProps) {
        super(props);
        this.state = {
            enrollments: props.users,
        };
    }

    public componentWillReceiveProps(nextProps: Readonly<IUserViewerProps>, nextContext: any): void {
        this.setState({
            enrollments: nextProps.users,
        });
    }

    public render() {
        return <div>
            {this.renderSearch()}
            <DynamicTable
                header={this.getTableHeading()}
                data={this.state.enrollments}
                classType={"table-grp"}
                selector={(item: Enrollment) => this.renderRow(item)}
            />
        </div>;
    }

    private renderSearch() {
        if (this.props.searchable) {
            return <Search className="input-group"
                placeholder="Search for students"
                onChange={(query) => this.handleSearch(query)}
            />;
        }
        return null;
    }

    private getTableHeading(): string[] {
        const heading: string[] = ["Name", "Email", "Student ID"];
        if (this.props.userMan || this.props.actions) {
            heading.push("Role");
        }
        return heading;
    }

    private renderRow(enr: Enrollment): (string | JSX.Element)[] {
        const selector: (string | JSX.Element)[] = [];
        const user = enr.getUser();
        if (!user) {
            return selector;
        }
        if (enr.getStatus() === Enrollment.UserStatus.TEACHER) {
            selector.push(
                <span className="text-muted">
                    <a href={generateGitLink(user.getLogin())} target="_blank">{user.getName()}</a>
                </span>);
        } else {
            selector.push(
                <a href={generateGitLink(user.getLogin())} target="_blank">{user.getName()}</a>);
        }
        selector.push(
            <a href={"mailto:" + enr.getUser()?.getEmail()}>{user?.getEmail()}</a>,
            enr.getUser()?.getStudentid() ?? "",
        );
        const temp = this.renderActions(enr);
        if (Array.isArray(temp) && temp.length > 0) {
            selector.push(<div className="btn-group action-btn">{temp}</div>);
        } else if (!Array.isArray(temp)) {
            selector.push(temp);
        }
        return selector;
    }

    private renderActions(enrol: Enrollment): JSX.Element[] | JSX.Element {
        const actionButtons: JSX.Element[] = [];
        const tempActions = this.getAllLinks(enrol);
        if (tempActions.length > 0) {
            switch (this.props.linkType) {
                case ActionType.Menu:
                    return this.renderDropdownMenu(enrol, tempActions);
                case ActionType.InRow:
                    actionButtons.push(...this.renderActionRow(enrol, tempActions));
                    break;
            }
        }
        return actionButtons;
    }

    private getAllLinks(enrol: Enrollment) {
        const tempActions: ILink[] = [];
        if (this.props.actions) {
            tempActions.push(...this.props.actions);
        }
        if (this.props.optionalActions) {
            tempActions.push(...this.props.optionalActions(enrol));
        }
        return tempActions;
    }

    private renderDropdownMenu(enrol: Enrollment, tempActions: ILink[]) {
        return <ul className="nav nav-pills">
            <LiDropDownMenu
                links={tempActions}
                onClick={(link) => { if (this.props.actionClick) { this.props.actionClick(enrol, link); } }}>
                <span className="glyphicon glyphicon-option-vertical" />
            </LiDropDownMenu>
        </ul>;
    }

    private renderActionRow(enrol: Enrollment, tempActions: ILink[]) {
        return tempActions.map((v, i) => {
            let hoverText = "";
            if (v.uri === "teacher") {
                hoverText = "Promote to teacher";
            } else if (v.uri === "demote") {
                hoverText = "Demote teacher";
            }

            return <BootstrapButton
                key={i}
                classType={v.extra ? v.extra as BootstrapClass : "default"}
                tooltip={hoverText}
                type={v.description}
                onClick={(link) => { if (this.props.actionClick) { this.props.actionClick(enrol, v); } }}
            >{v.name}
            </BootstrapButton>;
        });
    }

    private handleSearch(query: string): void {
        const filteredData = searchForStudents(this.props.users, query);
        this.setState({
            enrollments: filteredData,
        });
    }
}
