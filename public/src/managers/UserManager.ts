import { IEventData, newEvent } from "../event";
import { User } from "../../proto/ag_pb";

import { ILogger } from "./LogManager";

export interface IUserProvider {
    tryLogin(username: string, password: string): Promise<User | null>;
    logout(user: User): Promise<boolean>;
    getAllUser(): Promise<User[]>;
    tryRemoteLogin(provider: string): Promise<User | null>;
    changeAdminRole(user: User): Promise<boolean>;
    getLoggedInUser(): Promise<User | null>;
    updateUser(user: User): Promise<boolean>;
}

interface IUserLoginEvent extends IEventData {
    user: User;
}

export class UserManager {
    /**
     * This event fires when an user is loged in to the service
     */
    public onLogin = newEvent<IUserLoginEvent>("UserManager.onLogin");
    /**
     * This event fires when an user is loged out of the service
     */
    public onLogout = newEvent<IEventData>("UserManager.onLogout");

    private userProvider: IUserProvider;
    private currentUser: User | null;
   // private usersToken: string | null;

    /**
     * Creates a new instance of the UserManager
     * @param userProvider A user provider to get user information from
     */
    constructor(userProvider: IUserProvider, logger: ILogger) {
        this.userProvider = userProvider;
    }

    /**
     * Returns the current logged in user.
     * If there is no logged in user it returns null
     */
    public getCurrentUser(): User | null {
        return this.currentUser;
    }

    public isValidUser(user: User): boolean {
        return user.getEmail().length > 0
            && user.getName().length > 0
            && user.getStudentid().length > 0;
    }

    /**
     * Trys to login to the service with username and password
     * This is only used for testing
     * @param username The username to try login with
     * @param password The password to try login with
     */
    public async tryLogin(username: string, password: string): Promise<User | null> {
        const result = await this.userProvider.tryLogin(username, password);
        if (result) {
            this.currentUser = result;
            this.onLogin({ target: this, user: this.currentUser });
        }
        return result;
    }

    /**
     * Try to login with a remote service, like github and gitlab.
     * Normaly this function redirects before it returns.
     * @param provider Provider service to login with. Currently supports github and gitlab
     * @returns Returns the user if succsess or null if failed.
     */
    public async tryRemoteLogin(provider: string): Promise<User | null> {
        const result = await this.userProvider.tryRemoteLogin(provider);
        if (result) {
            this.currentUser = result;
            this.onLogin({ target: this, user: this.currentUser });
        }
        return result;
    }

    /**
     * logout from the current logged in session
     */
    public async logout() {
        if (this.currentUser) {
            await this.userProvider.logout(this.currentUser);
            this.currentUser = null;
            this.onLogout({ target: this });
        }
    }

    /**
     * Function to see of a user is an admin or not
     * @param user User to check if is an admin
     * @returns Returns true if admin. False otherwise
     */
    public isAdmin(user: User): boolean {
        if (user.getIsadmin()) {
            return user.getIsadmin();
        }
        return false;
    }

    /**
     * Function to see if a user is a teacher in any courses at all
     * @param user User to check if is an teacher in a courses
     * @returns Returns true if user is teacher in one or more courses
     */
    public async isTeacher(user: User): Promise<boolean> {
        return user.getIsadmin();
    }

    /**
     * Returns all users available at the backend
     * This function is mostly for testing and will change in the future
     * @returns All users at the backend
     */
    public async getAllUser(): Promise<User[]> {
        return await this.userProvider.getAllUser();
    }

    /**
     * This is not implemented
     * @param id the id of a single userobject to be returned
     */
    public async getUser(id: number): Promise<User> {
        throw new Error("Not implemented error");
    }

    /**
     * A way to promote a user to an administrator
     * @param user The user to premote to admin
     */
    public async changeAdminRole(user: User): Promise<boolean> {
        return await this.userProvider.changeAdminRole(user);
    }

    /**
     * Updates a user
     * @param user The user to update with the new information
     */
    public updateUser(user: User): Promise<boolean> {
        return this.userProvider.updateUser(user);
    }

    /**
     * Communicates with the backend to see if there is a logged inn user
     */
    public async checkUserLoggedIn(): Promise<boolean> {
        const usr = await this.userProvider.getLoggedInUser();

        this.currentUser = usr;
        if (usr) {
            return true;
        }
        return false;
    }
}
