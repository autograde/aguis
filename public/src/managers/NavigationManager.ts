import { IEventData, newEvent } from "../event";
import { isViewPage, ViewPage } from "../pages/ViewPage";

interface IPageContainer {
    [name: string]: IPageContainer | ViewPage;
}

interface INavEvent extends IEventData {
    uri: string;
    page: ViewPage;
    subPage: string;
}

interface ILink {
    name: string;
    description?: string;
    uri?: string;
    active?: boolean;
}

class NavigationManager {
    public onNavigate = newEvent<INavEvent>("NavigationManager.onNavigate");

    private pages: IPageContainer = {};
    private errorPages: ViewPage[] = [];
    private defaultPath: string = "";
    private currentPath: string = "";
    private browserHistory: History;

    constructor(history: History) {
        this.browserHistory = history;
        window.addEventListener("popstate", (e: PopStateEvent) => {
            this.navigateTo(location.pathname, true);
        });

    }

    // TODO: Move out to utility
    public getParts(path: string): string[] {
        return this.removeEmptyEntries(path.split("/"));
    }

    // TODO: Move out to utility
    public removeEmptyEntries(array: string[]): string[] {
        const newArray: string[] = [];
        array.map((v) => {
            if (v.length > 0) {
                newArray.push(v);
            }
        });
        return newArray;
    }

    public setDefaultPath(path: string) {
        this.defaultPath = path;
    }

    public navigateTo(path: string, preventPush?: boolean) {
        if (path === "/") {
            this.navigateToDefault();
            return;
        }
        const parts = this.getParts(path);
        let curPage: IPageContainer | ViewPage = this.pages;
        this.currentPath = parts.join("/");
        if (!preventPush) {
            this.browserHistory.pushState({}, "Autograder", "/" + this.currentPath);
        }
        for (let i = 0; i < parts.length; i++) {
            const a = parts[i];
            if (isViewPage(curPage)) {
                this.onNavigate({
                    page: curPage,
                    subPage: parts.slice(i, parts.length).join("/"),
                    target: this,
                    uri: path,
                });
                return;
            } else {
                const cur: IPageContainer | ViewPage = curPage[a];
                if (!cur) {
                    this.onNavigate({ target: this, page: this.getErrorPage(404), subPage: "", uri: path });
                    return;
                    // throw Error("404 Page not found");
                }
                curPage = cur;
            }
        }
        if (isViewPage(curPage)) {
            this.onNavigate({ target: this, page: curPage, uri: path, subPage: "" });
            return;
        } else {
            this.onNavigate({ target: this, page: this.getErrorPage(404), subPage: "", uri: path });
            // throw Error("404 Page not found");
        }
    }

    public navigateToDefault(): void {
        this.navigateTo(this.defaultPath);
    }

    public navigateToError(statusCode: number): void {
        this.onNavigate({ target: this, page: this.getErrorPage(statusCode), subPage: "", uri: statusCode.toString() });
    }

    public registerPage(path: string, page: ViewPage) {
        const parts = this.getParts(path);
        if (parts.length === 0) {
            throw Error("Can't add page to index element");
        }
        page.setPath(parts.join("/"));
        let curObj = this.pages;

        for (let i = 0; i < parts.length - 1; i++) {
            const a = parts[i];
            if (a.length === 0) {
                continue;
            }
            let temp: IPageContainer | ViewPage = curObj[a];
            if (!temp) {
                temp = {};
                curObj[a] = temp;
            } else if (!isViewPage(temp)) {
                temp = curObj[a];
            }

            if (isViewPage(temp)) {
                throw Error("Can't assign a IPageContainer to a ViewPage");
            }
            curObj = temp;
        }
        curObj[parts[parts.length - 1]] = page;

    }

    public registerErrorPage(statusCode: number, page: ViewPage) {
        this.errorPages[statusCode] = page;
    }

    /**
     * Checks to see if the link is part of the current path,
     * or the default page to the given ViewPage. Also mark them as active if they are.
     * @param links The links to check
     * @param viewPage ViewPage to get defaultPage information from
     */
    public checkLinks(links: ILink[], viewPage?: ViewPage): void {
        let checkUrl = this.currentPath;
        if (viewPage && viewPage.pagePath === checkUrl) {
            checkUrl += "/" + viewPage.defaultPage;
        }
        for (const l of links) {
            if (!l.uri) {
                continue;
            }
            const a = this.getParts(l.uri).join("/");
            l.active = a === checkUrl.substr(0, a.length);
        }
    }

    public refresh() {
        this.navigateTo(this.currentPath);
    }

    private getErrorPage(statusCode: number): ViewPage {
        if (this.errorPages[statusCode]) {
            return this.errorPages[statusCode];
        }
        throw Error("Status page: " + statusCode + " is not defined");
    }
}

export { IPageContainer, INavEvent, ILink, NavigationManager };
