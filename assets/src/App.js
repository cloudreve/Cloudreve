import React, { Suspense } from "react";
import AuthRoute from "./middleware/AuthRoute";
import Navbar from "./component/Navbar/Navbar.js";
import useMediaQuery from "@material-ui/core/useMediaQuery";
import AlertBar from "./component/Common/Snackbar";
import { createMuiTheme, lighten } from "@material-ui/core/styles";
import { useSelector } from "react-redux";
import { Redirect, Route, Switch, useRouteMatch } from "react-router-dom";
import Auth from "./middleware/Auth";
import { CssBaseline, makeStyles, ThemeProvider } from "@material-ui/core";
import { changeThemeColor } from "./utils";
import NotFound from "./component/Share/NotFound";
// Lazy loads
import LoginForm from "./component/Login/LoginForm";
import FileManager from "./component/FileManager/FileManager.js";
import VideoPreview from "./component/Viewer/Video.js";
import SearchResult from "./component/Share/SearchResult";
import MyShare from "./component/Share/MyShare";
import Download from "./component/Download/Download";
import SharePreload from "./component/Share/SharePreload";
import DocViewer from "./component/Viewer/Doc";
import TextViewer from "./component/Viewer/Text";
import WebDAV from "./component/Setting/WebDAV";
import Tasks from "./component/Setting/Tasks";
import Profile from "./component/Setting/Profile";
import UserSetting from "./component/Setting/UserSetting";
import Register from "./component/Login/Register";
import Activation from "./component/Login/Activication";
import ResetForm from "./component/Login/ResetForm";
import Reset from "./component/Login/Reset";
import PageLoading from "./component/Placeholder/PageLoading";
import CodeViewer from "./component/Viewer/Code";
const PDFViewer = React.lazy(() =>
    import(/* webpackChunkName: "pdf" */ "./component/Viewer/PDF")
);

export default function App() {
    const themeConfig = useSelector((state) => state.siteConfig.theme);
    const isLogin = useSelector((state) => state.viewUpdate.isLogin);
    const prefersDarkMode = useMediaQuery("(prefers-color-scheme: dark)");

    const theme = React.useMemo(() => {
        themeConfig.palette.type = prefersDarkMode ? "dark" : "light";
        const prefer = Auth.GetPreference("theme_mode");
        if (prefer) {
            themeConfig.palette.type = prefer;
        }
        const theme = createMuiTheme({
            ...themeConfig,
            palette: {
                ...themeConfig.palette,
                primary: {
                    ...themeConfig.palette.primary,
                    main:
                        themeConfig.palette.type === "dark"
                            ? lighten(themeConfig.palette.primary.main, 0.3)
                            : themeConfig.palette.primary.main,
                },
            },
        });
        changeThemeColor(
            themeConfig.palette.type === "dark"
                ? theme.palette.background.default
                : theme.palette.primary.main
        );
        return theme;
    }, [prefersDarkMode, themeConfig]);

    const useStyles = makeStyles((theme) => ({
        root: {
            display: "flex",
        },
        content: {
            flexGrow: 1,
            padding: theme.spacing(0),
            minWidth: 0,
        },
        toolbar: theme.mixins.toolbar,
    }));

    const classes = useStyles();

    const { path } = useRouteMatch();
    return (
        <React.Fragment>
            <ThemeProvider theme={theme}>
                <div className={classes.root} id="container">
                    <CssBaseline />
                    <AlertBar />
                    <Navbar />
                    <main className={classes.content}>
                        <div className={classes.toolbar} />
                        <Switch>
                            <AuthRoute exact path={path} isLogin={isLogin}>
                                <Redirect
                                    to={{
                                        pathname: "/home",
                                    }}
                                />
                            </AuthRoute>

                            <AuthRoute path={`${path}home`} isLogin={isLogin}>
                                <FileManager />
                            </AuthRoute>

                            <AuthRoute path={`${path}video`} isLogin={isLogin}>
                                <VideoPreview />
                            </AuthRoute>

                            <AuthRoute path={`${path}text`} isLogin={isLogin}>
                                <TextViewer />
                            </AuthRoute>

                            <AuthRoute path={`${path}doc`} isLogin={isLogin}>
                                <DocViewer />
                            </AuthRoute>

                            <AuthRoute path={`${path}pdf`} isLogin={isLogin}>
                                <Suspense fallback={<PageLoading />}>
                                    <PDFViewer />
                                </Suspense>
                            </AuthRoute>

                            <AuthRoute path={`${path}code`} isLogin={isLogin}>
                                <CodeViewer />
                            </AuthRoute>

                            <AuthRoute path={`${path}aria2`} isLogin={isLogin}>
                                <Download />
                            </AuthRoute>

                            <AuthRoute path={`${path}shares`} isLogin={isLogin}>
                                <MyShare />
                            </AuthRoute>

                            <Route path={`${path}search`} isLogin={isLogin}>
                                <SearchResult />
                            </Route>

                            <Route path={`${path}setting`} isLogin={isLogin}>
                                <UserSetting />
                            </Route>

                            <AuthRoute
                                path={`${path}profile/:id`}
                                isLogin={isLogin}
                            >
                                <Profile />
                            </AuthRoute>

                            <AuthRoute path={`${path}webdav`} isLogin={isLogin}>
                                <WebDAV />
                            </AuthRoute>

                            <AuthRoute path={`${path}tasks`} isLogin={isLogin}>
                                <Tasks />
                            </AuthRoute>

                            <Route path={`${path}login`} exact>
                                <LoginForm />
                            </Route>

                            <Route path={`${path}signup`} exact>
                                <Register />
                            </Route>

                            <Route path={`${path}activate`} exact>
                                <Activation />
                            </Route>

                            <Route path={`${path}reset`} exact>
                                <ResetForm />
                            </Route>

                            <Route path={`${path}forget`} exact>
                                <Reset />
                            </Route>

                            <Route exact path={`${path}s/:id`}>
                                <SharePreload />
                            </Route>

                            <Route path={`${path}s/:id/video(/)*`}>
                                <VideoPreview />
                            </Route>

                            <Route path={`${path}s/:id/doc(/)*`}>
                                <DocViewer />
                            </Route>

                            <Route path={`${path}s/:id/text(/)*`}>
                                <TextViewer />
                            </Route>

                            <Route path={`${path}s/:id/pdf(/)*`}>
                                <Suspense fallback={<PageLoading />}>
                                    <PDFViewer />
                                </Suspense>
                            </Route>

                            <Route path={`${path}s/:id/code(/)*`}>
                                <CodeViewer />
                            </Route>

                            <Route path="*">
                                <NotFound msg={"页面不存在"} />
                            </Route>
                        </Switch>
                    </main>
                </div>
            </ThemeProvider>
        </React.Fragment>
    );
}
