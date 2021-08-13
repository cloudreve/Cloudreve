const Auth = {
    isAuthenticated: false,
    authenticate(cb: any) {
        Auth.SetUser(cb);
        Auth.isAuthenticated = true;
    },
    GetUser() {
        return JSON.parse(localStorage.getItem("user") || "null");
    },
    SetUser(newUser: any) {
        localStorage.setItem("user", JSON.stringify(newUser));
    },
    Check(): boolean {
        if (Auth.isAuthenticated) {
            return true;
        }
        if (localStorage.getItem("user") !== null) {
            return !Auth.GetUser().anonymous;
        }
        return false;
    },
    signout() {
        Auth.isAuthenticated = false;
        const oldUser = Auth.GetUser();
        oldUser.id = 0;
        localStorage.setItem("user", JSON.stringify(oldUser));
    },
    SetPreference(key: string, value: any) {
        let preference = JSON.parse(
            localStorage.getItem("user_preference") || "{}"
        );
        preference = preference == null ? {} : preference;
        preference[key] = value;
        localStorage.setItem("user_preference", JSON.stringify(preference));
    },
    GetPreference(key: string): any | null {
        const preference = JSON.parse(
            localStorage.getItem("user_preference") || "{}"
        );
        if (preference && preference[key]) {
            return preference[key];
        }
        return null;
    },
};

export default Auth;
