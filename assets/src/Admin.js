import React, { useEffect, useState } from "react";
import { CssBaseline, makeStyles } from "@material-ui/core";
import AlertBar from "./component/Common/Snackbar";
import Dashboard from "./component/Admin/Dashboard";
import { useHistory } from "react-router";
import Auth from "./middleware/Auth";
import { Route, Switch } from "react-router-dom";
import { ThemeProvider } from "@material-ui/styles";
import createMuiTheme from "@material-ui/core/styles/createMuiTheme";
import { zhCN } from "@material-ui/core/locale";

import Index from "./component/Admin/Index";
import SiteInformation from "./component/Admin/Setting/SiteInformation";
import Access from "./component/Admin/Setting/Access";
import Mail from "./component/Admin/Setting/Mail";
import UploadDownload from "./component/Admin/Setting/UploadDownload";
import Theme from "./component/Admin/Setting/Theme";
import Aria2 from "./component/Admin/Setting/Aria2";
import ImageSetting from "./component/Admin/Setting/Image";
import Policy from "./component/Admin/Policy/Policy";
import AddPolicy from "./component/Admin/Policy/AddPolicy";
import EditPolicyPreload from "./component/Admin/Policy/EditPolicy";
import Group from "./component/Admin/Group/Group";
import GroupForm from "./component/Admin/Group/GroupForm";
import EditGroupPreload from "./component/Admin/Group/EditGroup";
import User from "./component/Admin/User/User";
import UserForm from "./component/Admin/User/UserForm";
import EditUserPreload from "./component/Admin/User/EditUser";
import File from "./component/Admin/File/File";
import Share from "./component/Admin/Share/Share";
import Download from "./component/Admin/Task/Download";
import Task from "./component/Admin/Task/Task";
import Import from "./component/Admin/File/Import";
import Captcha from "./component/Admin/Setting/Captcha";

const useStyles = makeStyles((theme) => ({
    root: {
        display: "flex",
    },
    content: {
        flexGrow: 1,
        padding: 0,
        minWidth: 0,
    },
    toolbar: theme.mixins.toolbar,
}));

const theme = createMuiTheme(
    {
        palette: {
            background: {},
        },
    },
    zhCN
);

export default function Admin() {
    const classes = useStyles();
    const history = useHistory();
    const [show, setShow] = useState(false);

    useEffect(() => {
        const user = Auth.GetUser();
        if (user && user.group) {
            if (user.group.id !== 1) {
                history.push("/home");
                return;
            }
            setShow(true);
            return;
        }
        history.push("/login");
        // eslint-disable-next-line
    }, []);

    return (
        <React.Fragment>
            <ThemeProvider theme={theme}>
                <div className={classes.root}>
                    <CssBaseline />
                    <AlertBar />
                    {show && (
                        <Dashboard
                            content={(path) => (
                                <Switch>
                                    <Route path={`${path}/home`} exact>
                                        <Index />
                                    </Route>

                                    <Route path={`${path}/basic`}>
                                        <SiteInformation />
                                    </Route>

                                    <Route path={`${path}/access`}>
                                        <Access />
                                    </Route>

                                    <Route path={`${path}/mail`}>
                                        <Mail />
                                    </Route>

                                    <Route path={`${path}/upload`}>
                                        <UploadDownload />
                                    </Route>

                                    <Route path={`${path}/theme`}>
                                        <Theme />
                                    </Route>

                                    <Route path={`${path}/aria2`}>
                                        <Aria2 />
                                    </Route>

                                    <Route path={`${path}/image`}>
                                        <ImageSetting />
                                    </Route>

                                    <Route path={`${path}/captcha`}>
                                        <Captcha />
                                    </Route>

                                    <Route path={`${path}/policy`} exact>
                                        <Policy />
                                    </Route>

                                    <Route
                                        path={`${path}/policy/add/:type`}
                                        exact
                                    >
                                        <AddPolicy />
                                    </Route>

                                    <Route
                                        path={`${path}/policy/edit/:mode/:id`}
                                        exact
                                    >
                                        <EditPolicyPreload />
                                    </Route>

                                    <Route path={`${path}/group`} exact>
                                        <Group />
                                    </Route>

                                    <Route path={`${path}/group/add`} exact>
                                        <GroupForm />
                                    </Route>

                                    <Route
                                        path={`${path}/group/edit/:id`}
                                        exact
                                    >
                                        <EditGroupPreload />
                                    </Route>

                                    <Route path={`${path}/user`} exact>
                                        <User />
                                    </Route>

                                    <Route path={`${path}/user/add`} exact>
                                        <UserForm />
                                    </Route>

                                    <Route path={`${path}/user/edit/:id`} exact>
                                        <EditUserPreload />
                                    </Route>

                                    <Route path={`${path}/file`} exact>
                                        <File />
                                    </Route>

                                    <Route path={`${path}/file/import`} exact>
                                        <Import />
                                    </Route>

                                    <Route path={`${path}/share`} exact>
                                        <Share />
                                    </Route>

                                    <Route path={`${path}/download`} exact>
                                        <Download />
                                    </Route>

                                    <Route path={`${path}/task`} exact>
                                        <Task />
                                    </Route>
                                </Switch>
                            )}
                        />
                    )}
                </div>
            </ThemeProvider>
        </React.Fragment>
    );
}
