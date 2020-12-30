import * as React from "react";
import { Comment } from "../../../proto/ag_pb";


interface ICommentProps {
    comment: Comment,
    onSelect: () => void;
}

interface ICommentState {
    collapsed: boolean,
    editing: boolean,
}

export class IComment extends React.Component<ICommentProps, ICommentState>  {
    constructor(props: ICommentProps) {
        super(props);
        this.state = {
            collapsed: true,
            editing: false,
        }
    }

    public render() {
        if (!(this.props.comment)) {
            return <div></div>;
        }
        return <div className="row col-md-12 user-comment"
            onClick={() => this.props.onSelect()}
        >
            <div className="comment-pic col-md-2"><img src={this.props.comment.getAvatarurl()} className="comment-img" /></div>
            <div className="comment-body col-md-10">

            <div className="row comment-header">
                <span className="coment-author">{this.props.comment.getName()}</span>
                <span className="comment-posted">posted {this.props.comment.getPosted()}</span>
            </div>
            <div className="row comment-message">
                {this.props.comment.getMessage()}
            </div>
            </div>

        </div>
    }
}
