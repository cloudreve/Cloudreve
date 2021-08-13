import React from "react";
import Auth from "./Auth";
import { Route, Redirect } from "react-router-dom";

function AuthRoute({ children, ...rest }) {
    return (
        <Route
            {...rest}
            render={({ location }) =>
                Auth.Check(rest.isLogin) ? (
                    children
                ) : (
                    <Redirect
                        to={{
                            pathname: "/login",
                            state: { from: location },
                        }}
                    />
                )
            }
        />
    );
}

export default AuthRoute;
