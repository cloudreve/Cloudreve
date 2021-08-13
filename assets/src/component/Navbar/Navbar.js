import React, { Component } from "react";
import PropTypes from "prop-types";
import classNames from "classnames";
import { connect } from "react-redux";
import ShareIcon from "@material-ui/icons/Share";
import BackIcon from "@material-ui/icons/ArrowBack";
import OpenIcon from "@material-ui/icons/OpenInNew";
import DownloadIcon from "@material-ui/icons/CloudDownload";
import OpenFolderIcon from "@material-ui/icons/FolderOpen";
import RenameIcon from "@material-ui/icons/BorderColor";
import MoveIcon from "@material-ui/icons/Input";
import DeleteIcon from "@material-ui/icons/Delete";
import SaveIcon from "@material-ui/icons/Save";
import MenuIcon from "@material-ui/icons/Menu";
import { isPreviewable } from "../../config";
import {
    drawerToggleAction,
    setSelectedTarget,
    navigateTo,
    openCreateFolderDialog,
    changeContextMenu,
    searchMyFile,
    saveFile,
    openMusicDialog,
    showImgPreivew,
    toggleSnackbar,
    openMoveDialog,
    openRemoveDialog,
    openShareDialog,
    openRenameDialog,
    openLoadingDialog,
    setSessionStatus,
    openPreview,
} from "../../actions";
import {
    allowSharePreview,
    checkGetParameters,
    changeThemeColor,
} from "../../utils";
import Uploader from "../Upload/Uploader.js";
import { sizeToString, vhCheck } from "../../utils";
import pathHelper from "../../utils/page";
import SezrchBar from "./SearchBar";
import StorageBar from "./StorageBar";
import UserAvatar from "./UserAvatar";
import UserInfo from "./UserInfo";
import { AccountArrowRight, AccountPlus, LogoutVariant } from "mdi-material-ui";
import { withRouter } from "react-router-dom";
import {
    AppBar,
    Toolbar,
    Typography,
    withStyles,
    withTheme,
    Drawer,
    SwipeableDrawer,
    IconButton,
    Hidden,
    ListItem,
    ListItemIcon,
    ListItemText,
    List,
    Grow,
    Tooltip,
} from "@material-ui/core";
import Auth from "../../middleware/Auth";
import API from "../../middleware/Api";
import FileTag from "./FileTags";
import { Assignment, Devices, MoreHoriz, Settings } from "@material-ui/icons";
import Divider from "@material-ui/core/Divider";
import SubActions from "../FileManager/Navigator/SubActions";

vhCheck();
const drawerWidth = 240;
const drawerWidthMobile = 270;

const mapStateToProps = (state) => {
    return {
        desktopOpen: state.viewUpdate.open,
        selected: state.explorer.selected,
        isMultiple: state.explorer.selectProps.isMultiple,
        withFolder: state.explorer.selectProps.withFolder,
        withFile: state.explorer.selectProps.withFile,
        path: state.navigator.path,
        keywords: state.explorer.keywords,
        title: state.siteConfig.title,
        subTitle: state.viewUpdate.subTitle,
        loadUploader: state.viewUpdate.loadUploader,
        isLogin: state.viewUpdate.isLogin,
    };
};

const mapDispatchToProps = (dispatch) => {
    return {
        handleDesktopToggle: (open) => {
            dispatch(drawerToggleAction(open));
        },
        setSelectedTarget: (targets) => {
            dispatch(setSelectedTarget(targets));
        },
        navigateTo: (path) => {
            dispatch(navigateTo(path));
        },
        openCreateFolderDialog: () => {
            dispatch(openCreateFolderDialog());
        },
        changeContextMenu: (type, open) => {
            dispatch(changeContextMenu(type, open));
        },
        searchMyFile: (keywords) => {
            dispatch(searchMyFile(keywords));
        },
        saveFile: () => {
            dispatch(saveFile());
        },
        openMusicDialog: () => {
            dispatch(openMusicDialog());
        },
        showImgPreivew: (first) => {
            dispatch(showImgPreivew(first));
        },
        toggleSnackbar: (vertical, horizontal, msg, color) => {
            dispatch(toggleSnackbar(vertical, horizontal, msg, color));
        },
        openRenameDialog: () => {
            dispatch(openRenameDialog());
        },
        openMoveDialog: () => {
            dispatch(openMoveDialog());
        },
        openRemoveDialog: () => {
            dispatch(openRemoveDialog());
        },
        openShareDialog: () => {
            dispatch(openShareDialog());
        },
        openLoadingDialog: (text) => {
            dispatch(openLoadingDialog(text));
        },
        setSessionStatus: () => {
            dispatch(setSessionStatus());
        },
        openPreview: () => {
            dispatch(openPreview());
        },
    };
};

const styles = (theme) => ({
    appBar: {
        marginLeft: drawerWidth,
        [theme.breakpoints.down("xs")]: {
            marginLeft: drawerWidthMobile,
        },
        zIndex: theme.zIndex.drawer + 1,
        transition: " background-color 250ms",
    },

    drawer: {
        width: 0,
        flexShrink: 0,
    },
    drawerDesktop: {
        width: drawerWidth,
        flexShrink: 0,
    },
    icon: {
        marginRight: theme.spacing(2),
    },
    menuButton: {
        marginRight: 20,
        [theme.breakpoints.up("sm")]: {
            display: "none",
        },
    },
    menuButtonDesktop: {
        marginRight: 20,
        [theme.breakpoints.down("sm")]: {
            display: "none",
        },
    },
    menuIcon: {
        marginRight: 20,
    },
    toolbar: theme.mixins.toolbar,
    drawerPaper: {
        width: drawerWidthMobile,
    },
    drawerPaperDesktop: {
        width: drawerWidth,
    },
    upDrawer: {
        overflowX: "hidden",
    },
    drawerOpen: {
        width: drawerWidth,
        transition: theme.transitions.create("width", {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.enteringScreen,
        }),
    },
    drawerClose: {
        transition: theme.transitions.create("width", {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
        }),
        overflowX: "hidden",
        width: 0,
    },
    content: {
        flexGrow: 1,
        padding: theme.spacing(3),
    },
    grow: {
        flexGrow: 1,
    },
    badge: {
        top: 1,
        right: -15,
    },
    nested: {
        paddingLeft: theme.spacing(4),
    },
    sectionForFile: {
        display: "flex",
    },
    extendedIcon: {
        marginRight: theme.spacing(1),
    },
    addButton: {
        marginLeft: "40px",
        marginTop: "25px",
        marginBottom: "15px",
    },
    fabButton: {
        borderRadius: "100px",
    },
    badgeFix: {
        right: "10px",
    },
    iconFix: {
        marginLeft: "16px",
    },
    dividerFix: {
        marginTop: "8px",
    },
    folderShareIcon: {
        verticalAlign: "sub",
        marginRight: "5px",
    },
    shareInfoContainer: {
        display: "flex",
        marginTop: "15px",
        marginBottom: "20px",
        marginLeft: "28px",
        textDecoration: "none",
    },
    shareAvatar: {
        width: "40px",
        height: "40px",
    },
    stickFooter: {
        bottom: "0px",
        position: "absolute",
        backgroundColor: theme.palette.background.paper,
        width: "100%",
    },
    ownerInfo: {
        marginLeft: "10px",
        width: "150px",
    },
    minStickDrawer: {
        overflowY: "auto",
        [theme.breakpoints.up("sm")]: {
            height: "calc(var(--vh, 100vh) - 145px)",
        },

        [theme.breakpoints.down("sm")]: {
            minHeight: "calc(var(--vh, 100vh) - 360px)",
        },
    },
});
class NavbarCompoment extends Component {
    constructor(props) {
        super(props);
        this.state = {
            mobileOpen: false,
        };
        this.UploaderRef = React.createRef();
    }

    UNSAFE_componentWillMount() {
        this.unlisten = this.props.history.listen(() => {
            this.setState(() => ({ mobileOpen: false }));
        });
    }
    componentWillUnmount() {
        this.unlisten();
    }

    componentDidMount() {
        changeThemeColor(
            this.props.selected.length <= 1 &&
                !(!this.props.isMultiple && this.props.withFile)
                ? this.props.theme.palette.primary.main
                : this.props.theme.palette.background.default
        );
    }

    UNSAFE_componentWillReceiveProps = (nextProps) => {
        if (
            (this.props.selected.length <= 1 &&
                !(!this.props.isMultiple && this.props.withFile)) !==
            (nextProps.selected.length <= 1 &&
                !(!nextProps.isMultiple && nextProps.withFile))
        ) {
            changeThemeColor(
                !(
                    this.props.selected.length <= 1 &&
                    !(!this.props.isMultiple && this.props.withFile)
                )
                    ? this.props.theme.palette.type === "dark"
                        ? this.props.theme.palette.background.default
                        : this.props.theme.palette.primary.main
                    : this.props.theme.palette.background.default
            );
        }
    };

    handleDrawerToggle = () => {
        this.setState((state) => ({ mobileOpen: !state.mobileOpen }));
    };

    loadUploader = () => {
        if (pathHelper.isHomePage(this.props.location.pathname)) {
            return (
                <>
                    {this.props.loadUploader && this.props.isLogin && (
                        <Uploader />
                    )}
                </>
            );
        }
    };

    openDownload = () => {
        if (!allowSharePreview()) {
            this.props.toggleSnackbar(
                "top",
                "right",
                "未登录用户无法预览",
                "warning"
            );
            return;
        }
        this.props.openLoadingDialog("获取下载地址...");
    };

    archiveDownload = () => {
        this.props.openLoadingDialog("打包中...");
    };

    signOut = () => {
        API.delete("/user/session/")
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "您已退出登录",
                    "success"
                );
                Auth.signout();
                window.location.reload();
                this.props.setSessionStatus(false);
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "warning"
                );
            })
            .finally(() => {
                this.handleClose();
            });
    };

    render() {
        const { classes } = this.props;
        const user = Auth.GetUser(this.props.isLogin);
        const isHomePage = pathHelper.isHomePage(this.props.location.pathname);
        const isSharePage = pathHelper.isSharePage(
            this.props.location.pathname
        );

        const drawer = (
            <div id="container" className={classes.upDrawer}>
                {pathHelper.isMobile() && <UserInfo />}

                {Auth.Check(this.props.isLogin) && (
                    <>
                        <div className={classes.minStickDrawer}>
                            <FileTag />
                            <List>
                                <ListItem
                                    button
                                    key="我的分享"
                                    onClick={() =>
                                        this.props.history.push("/shares?")
                                    }
                                >
                                    <ListItemIcon>
                                        <ShareIcon
                                            className={classes.iconFix}
                                        />
                                    </ListItemIcon>
                                    <ListItemText primary="我的分享" />
                                </ListItem>
                                <ListItem
                                    button
                                    key="离线下载"
                                    onClick={() =>
                                        this.props.history.push("/aria2?")
                                    }
                                >
                                    <ListItemIcon>
                                        <DownloadIcon
                                            className={classes.iconFix}
                                        />
                                    </ListItemIcon>
                                    <ListItemText primary="离线下载" />
                                </ListItem>
                                {user.group.webdav && (
                                    <ListItem
                                        button
                                        key="WebDAV"
                                        onClick={() =>
                                            this.props.history.push("/webdav?")
                                        }
                                    >
                                        <ListItemIcon>
                                            <Devices
                                                className={classes.iconFix}
                                            />
                                        </ListItemIcon>
                                        <ListItemText primary="WebDAV" />
                                    </ListItem>
                                )}

                                <ListItem
                                    button
                                    key="任务队列"
                                    onClick={() =>
                                        this.props.history.push("/tasks?")
                                    }
                                >
                                    <ListItemIcon>
                                        <Assignment
                                            className={classes.iconFix}
                                        />
                                    </ListItemIcon>
                                    <ListItemText primary="任务队列" />
                                </ListItem>
                            </List>
                        </div>

                        {pathHelper.isMobile() && (
                            <>
                                <Divider />
                                <List>
                                    <ListItem
                                        button
                                        key="个人设置"
                                        onClick={() =>
                                            this.props.history.push("/setting?")
                                        }
                                    >
                                        <ListItemIcon>
                                            <Settings
                                                className={classes.iconFix}
                                            />
                                        </ListItemIcon>
                                        <ListItemText primary="个人设置" />
                                    </ListItem>

                                    <ListItem
                                        button
                                        key="退出登录"
                                        onClick={this.signOut}
                                    >
                                        <ListItemIcon>
                                            <LogoutVariant
                                                className={classes.iconFix}
                                            />
                                        </ListItemIcon>
                                        <ListItemText primary="退出登录" />
                                    </ListItem>
                                </List>
                            </>
                        )}
                        <div>
                            <StorageBar></StorageBar>
                        </div>
                    </>
                )}

                {!Auth.Check(this.props.isLogin) && (
                    <div>
                        <ListItem
                            button
                            key="登录"
                            onClick={() => this.props.history.push("/login")}
                        >
                            <ListItemIcon>
                                <AccountArrowRight
                                    className={classes.iconFix}
                                />
                            </ListItemIcon>
                            <ListItemText primary="登录" />
                        </ListItem>
                        <ListItem
                            button
                            key="注册"
                            onClick={() => this.props.history.push("/signup")}
                        >
                            <ListItemIcon>
                                <AccountPlus className={classes.iconFix} />
                            </ListItemIcon>
                            <ListItemText primary="注册" />
                        </ListItem>
                    </div>
                )}
            </div>
        );
        const iOS =
            process.browser && /iPad|iPhone|iPod/.test(navigator.userAgent);
        return (
            <div>
                <AppBar
                    position="fixed"
                    className={classes.appBar}
                    color={
                        this.props.theme.palette.type !== "dark" &&
                        this.props.selected.length <= 1 &&
                        !(!this.props.isMultiple && this.props.withFile)
                            ? "primary"
                            : "default"
                    }
                >
                    <Toolbar>
                        {this.props.selected.length <= 1 &&
                            !(
                                !this.props.isMultiple && this.props.withFile
                            ) && (
                                <IconButton
                                    color="inherit"
                                    aria-label="Open drawer"
                                    onClick={this.handleDrawerToggle}
                                    className={classes.menuButton}
                                >
                                    <MenuIcon />
                                </IconButton>
                            )}
                        {this.props.selected.length <= 1 &&
                            !(
                                !this.props.isMultiple && this.props.withFile
                            ) && (
                                <IconButton
                                    color="inherit"
                                    aria-label="Open drawer"
                                    onClick={() =>
                                        this.props.handleDesktopToggle(
                                            !this.props.desktopOpen
                                        )
                                    }
                                    className={classes.menuButtonDesktop}
                                >
                                    <MenuIcon />
                                </IconButton>
                            )}
                        {(this.props.selected.length > 1 ||
                            (!this.props.isMultiple && this.props.withFile)) &&
                            (isHomePage ||
                                pathHelper.isSharePage(
                                    this.props.location.pathname
                                )) && (
                                <Grow
                                    in={
                                        this.props.selected.length > 1 ||
                                        (!this.props.isMultiple &&
                                            this.props.withFile)
                                    }
                                >
                                    <IconButton
                                        color="inherit"
                                        className={classes.menuIcon}
                                        onClick={() =>
                                            this.props.setSelectedTarget([])
                                        }
                                    >
                                        <BackIcon />
                                    </IconButton>
                                </Grow>
                            )}
                        {this.props.selected.length <= 1 &&
                            !(
                                !this.props.isMultiple && this.props.withFile
                            ) && (
                                <Typography
                                    variant="h6"
                                    color="inherit"
                                    noWrap
                                    onClick={() => {
                                        this.props.history.push("/");
                                    }}
                                >
                                    {this.props.subTitle
                                        ? this.props.subTitle
                                        : this.props.title}
                                </Typography>
                            )}

                        {!this.props.isMultiple &&
                            this.props.withFile &&
                            !pathHelper.isMobile() && (
                                <Typography variant="h6" color="inherit" noWrap>
                                    {this.props.selected[0].name}{" "}
                                    {(isHomePage ||
                                        pathHelper.isSharePage(
                                            this.props.location.pathname
                                        )) &&
                                        "(" +
                                            sizeToString(
                                                this.props.selected[0].size
                                            ) +
                                            ")"}
                                </Typography>
                            )}

                        {this.props.selected.length > 1 &&
                            !pathHelper.isMobile() && (
                                <Typography variant="h6" color="inherit" noWrap>
                                    {this.props.selected.length}个对象
                                </Typography>
                            )}
                        {this.props.selected.length <= 1 &&
                            !(
                                !this.props.isMultiple && this.props.withFile
                            ) && <SezrchBar />}
                        <div className={classes.grow} />
                        {(this.props.selected.length > 1 ||
                            (!this.props.isMultiple && this.props.withFile)) &&
                            !isHomePage &&
                            !pathHelper.isSharePage(
                                this.props.location.pathname
                            ) &&
                            Auth.Check(this.props.isLogin) &&
                            !checkGetParameters("share") && (
                                <div className={classes.sectionForFile}>
                                    <Tooltip title="保存">
                                        <IconButton
                                            color="inherit"
                                            onClick={() =>
                                                this.props.saveFile()
                                            }
                                        >
                                            <SaveIcon />
                                        </IconButton>
                                    </Tooltip>
                                </div>
                            )}
                        {(this.props.selected.length > 1 ||
                            (!this.props.isMultiple && this.props.withFile)) &&
                            (isHomePage || isSharePage) && (
                                <div className={classes.sectionForFile}>
                                    {!this.props.isMultiple &&
                                        this.props.withFile &&
                                        isPreviewable(
                                            this.props.selected[0].name
                                        ) && (
                                            <Grow
                                                in={
                                                    !this.props.isMultiple &&
                                                    this.props.withFile &&
                                                    isPreviewable(
                                                        this.props.selected[0]
                                                            .name
                                                    )
                                                }
                                            >
                                                <Tooltip title="打开">
                                                    <IconButton
                                                        color="inherit"
                                                        onClick={() =>
                                                            this.props.openPreview()
                                                        }
                                                    >
                                                        <OpenIcon />
                                                    </IconButton>
                                                </Tooltip>
                                            </Grow>
                                        )}
                                    {!this.props.isMultiple &&
                                        this.props.withFile && (
                                            <Grow
                                                in={
                                                    !this.props.isMultiple &&
                                                    this.props.withFile
                                                }
                                            >
                                                <Tooltip title="下载">
                                                    <IconButton
                                                        color="inherit"
                                                        onClick={() =>
                                                            this.openDownload()
                                                        }
                                                    >
                                                        <DownloadIcon />
                                                    </IconButton>
                                                </Tooltip>
                                            </Grow>
                                        )}
                                    {(this.props.isMultiple ||
                                        this.props.withFolder) &&
                                        user.group.allowArchiveDownload && (
                                            <Grow
                                                in={
                                                    (this.props.isMultiple ||
                                                        this.props
                                                            .withFolder) &&
                                                    user.group
                                                        .allowArchiveDownload
                                                }
                                            >
                                                <Tooltip title="打包下载">
                                                    <IconButton
                                                        color="inherit"
                                                        onClick={() =>
                                                            this.archiveDownload()
                                                        }
                                                    >
                                                        <DownloadIcon />
                                                    </IconButton>
                                                </Tooltip>
                                            </Grow>
                                        )}

                                    {!this.props.isMultiple &&
                                        this.props.withFolder && (
                                            <Grow
                                                in={
                                                    !this.props.isMultiple &&
                                                    this.props.withFolder
                                                }
                                            >
                                                <Tooltip title="进入目录">
                                                    <IconButton
                                                        color="inherit"
                                                        onClick={() =>
                                                            this.props.navigateTo(
                                                                this.props
                                                                    .path ===
                                                                    "/"
                                                                    ? this.props
                                                                          .path +
                                                                          this
                                                                              .props
                                                                              .selected[0]
                                                                              .name
                                                                    : this.props
                                                                          .path +
                                                                          "/" +
                                                                          this
                                                                              .props
                                                                              .selected[0]
                                                                              .name
                                                            )
                                                        }
                                                    >
                                                        <OpenFolderIcon />
                                                    </IconButton>
                                                </Tooltip>
                                            </Grow>
                                        )}
                                    {!this.props.isMultiple &&
                                        !pathHelper.isMobile() &&
                                        !isSharePage && (
                                            <Grow in={!this.props.isMultiple}>
                                                <Tooltip title="分享">
                                                    <IconButton
                                                        color="inherit"
                                                        onClick={() =>
                                                            this.props.openShareDialog()
                                                        }
                                                    >
                                                        <ShareIcon />
                                                    </IconButton>
                                                </Tooltip>
                                            </Grow>
                                        )}
                                    {!this.props.isMultiple && !isSharePage && (
                                        <Grow in={!this.props.isMultiple}>
                                            <Tooltip title="重命名">
                                                <IconButton
                                                    color="inherit"
                                                    onClick={() =>
                                                        this.props.openRenameDialog()
                                                    }
                                                >
                                                    <RenameIcon />
                                                </IconButton>
                                            </Tooltip>
                                        </Grow>
                                    )}
                                    {!isSharePage && (
                                        <div style={{ display: "flex" }}>
                                            {!pathHelper.isMobile() && (
                                                <Grow
                                                    in={
                                                        this.props.selected
                                                            .length !== 0 &&
                                                        !pathHelper.isMobile()
                                                    }
                                                >
                                                    <Tooltip title="移动">
                                                        <IconButton
                                                            color="inherit"
                                                            onClick={() =>
                                                                this.props.openMoveDialog()
                                                            }
                                                        >
                                                            <MoveIcon />
                                                        </IconButton>
                                                    </Tooltip>
                                                </Grow>
                                            )}

                                            <Grow
                                                in={
                                                    this.props.selected
                                                        .length !== 0
                                                }
                                            >
                                                <Tooltip title="删除">
                                                    <IconButton
                                                        color="inherit"
                                                        onClick={() =>
                                                            this.props.openRemoveDialog()
                                                        }
                                                    >
                                                        <DeleteIcon />
                                                    </IconButton>
                                                </Tooltip>
                                            </Grow>

                                            {pathHelper.isMobile() && (
                                                <Grow
                                                    in={
                                                        this.props.selected
                                                            .length !== 0 &&
                                                        pathHelper.isMobile()
                                                    }
                                                >
                                                    <Tooltip title="更多操作">
                                                        <IconButton
                                                            color="inherit"
                                                            onClick={() =>
                                                                this.props.changeContextMenu(
                                                                    "file",
                                                                    true
                                                                )
                                                            }
                                                        >
                                                            <MoreHoriz />
                                                        </IconButton>
                                                    </Tooltip>
                                                </Grow>
                                            )}
                                        </div>
                                    )}
                                </div>
                            )}
                        {this.props.selected.length <= 1 &&
                            !(
                                !this.props.isMultiple && this.props.withFile
                            ) && <UserAvatar />}
                        {this.props.selected.length <= 1 &&
                            !(!this.props.isMultiple && this.props.withFile) &&
                            isHomePage &&
                            pathHelper.isMobile() && <SubActions inherit />}
                    </Toolbar>
                </AppBar>
                {this.loadUploader()}

                <Hidden smUp implementation="css">
                    <SwipeableDrawer
                        container={this.props.container}
                        variant="temporary"
                        classes={{
                            paper: classes.drawerPaper,
                        }}
                        anchor="left"
                        open={this.state.mobileOpen}
                        onClose={this.handleDrawerToggle}
                        onOpen={() =>
                            this.setState(() => ({ mobileOpen: true }))
                        }
                        disableDiscovery={iOS}
                        ModalProps={{
                            keepMounted: true, // Better open performance on mobile.
                        }}
                    >
                        {drawer}
                    </SwipeableDrawer>
                </Hidden>
                <Hidden xsDown implementation="css">
                    <Drawer
                        classes={{
                            paper: classes.drawerPaperDesktop,
                        }}
                        className={classNames(classes.drawer, {
                            [classes.drawerOpen]: this.props.desktopOpen,
                            [classes.drawerClose]: !this.props.desktopOpen,
                        })}
                        variant="persistent"
                        anchor="left"
                        open={this.props.desktopOpen}
                    >
                        <div className={classes.toolbar} />
                        {drawer}
                    </Drawer>
                </Hidden>
            </div>
        );
    }
}
NavbarCompoment.propTypes = {
    classes: PropTypes.object.isRequired,
    theme: PropTypes.object.isRequired,
};

const Navbar = connect(
    mapStateToProps,
    mapDispatchToProps
)(withTheme(withStyles(styles)(withRouter(NavbarCompoment))));

export default Navbar;
