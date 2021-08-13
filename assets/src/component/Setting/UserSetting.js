import React, { Component } from "react";
import { connect } from "react-redux";
import PhotoIcon from "@material-ui/icons/InsertPhoto";
import GroupIcon from "@material-ui/icons/Group";
import DateIcon from "@material-ui/icons/DateRange";
import EmailIcon from "@material-ui/icons/Email";
import HomeIcon from "@material-ui/icons/Home";
import LinkIcon from "@material-ui/icons/Phonelink";
import InputIcon from "@material-ui/icons/Input";
import SecurityIcon from "@material-ui/icons/Security";
import NickIcon from "@material-ui/icons/PermContactCalendar";
import LockIcon from "@material-ui/icons/Lock";
import VerifyIcon from "@material-ui/icons/VpnKey";
import ColorIcon from "@material-ui/icons/Palette";
import {
    applyThemes,
    changeViewMethod,
    toggleDaylightMode,
    toggleSnackbar,
} from "../../actions";
import axios from "axios";
import FingerprintIcon from "@material-ui/icons/Fingerprint";
import ToggleButton from "@material-ui/lab/ToggleButton";
import ToggleButtonGroup from "@material-ui/lab/ToggleButtonGroup";
import RightIcon from "@material-ui/icons/KeyboardArrowRight";
import {
    ListItemIcon,
    withStyles,
    Button,
    Divider,
    TextField,
    Avatar,
    Paper,
    Typography,
    List,
    ListItem,
    ListItemSecondaryAction,
    ListItemText,
    ListItemAvatar,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Switch,
} from "@material-ui/core";
import { blue, green, yellow } from "@material-ui/core/colors";
import API from "../../middleware/Api";
import Auth from "../../middleware/Auth";
import { withRouter } from "react-router";
import QRCode from "qrcode-react";
import { Brightness3, ListAlt, PermContactCalendar,Schedule } from "@material-ui/icons";
import { transformTime } from "../../utils";
import Authn from "./Authn";
import { formatLocalTime, timeZone } from "../../utils/datetime";
import TimeZoneDialog from "../Modals/TimeZone";

const styles = (theme) => ({
    layout: {
        width: "auto",
        marginLeft: theme.spacing(3),
        marginRight: theme.spacing(3),
        [theme.breakpoints.up(1100 + theme.spacing(3) * 2)]: {
            width: 700,
            marginLeft: "auto",
            marginRight: "auto",
        },
    },
    sectionTitle: {
        paddingBottom: "10px",
        paddingTop: "30px",
    },
    rightIcon: {
        marginTop: "4px",
        marginRight: "10px",
        color: theme.palette.text.secondary,
    },
    uploadFromFile: {
        backgroundColor: blue[100],
        color: blue[600],
    },
    userGravatar: {
        backgroundColor: yellow[100],
        color: yellow[800],
    },
    policySelected: {
        backgroundColor: green[100],
        color: green[800],
    },
    infoText: {
        marginRight: "17px",
    },
    infoTextWithIcon: {
        marginRight: "17px",
        marginTop: "1px",
    },
    rightIconWithText: {
        marginTop: "0px",
        marginRight: "10px",
        color: theme.palette.text.secondary,
    },
    iconFix: {
        marginRight: "11px",
        marginLeft: "7px",
        minWidth: 40,
    },
    flexContainer: {
        display: "flex",
    },
    desenList: {
        paddingTop: 0,
        paddingBottom: 0,
    },
    flexContainerResponse: {
        display: "flex",
        [theme.breakpoints.down("sm")]: {
            display: "initial",
        },
    },
    desText: {
        marginTop: "10px",
    },
    secondColor: {
        height: "20px",
        width: "20px",
        backgroundColor: theme.palette.secondary.main,
        borderRadius: "50%",
        marginRight: "17px",
    },
    firstColor: {
        height: "20px",
        width: "20px",
        backgroundColor: theme.palette.primary.main,
        borderRadius: "50%",
        marginRight: "6px",
    },
    themeBlock: {
        height: "20px",
        width: "20px",
    },
    paddingBottom: {
        marginBottom: "30px",
    },
    paddingText: {
        paddingRight: theme.spacing(2),
    },
    qrcode: {
        width: 128,
        marginTop: 16,
        marginRight: 16,
    },
});

const mapStateToProps = (state) => {
    return {
        title: state.siteConfig.title,
        authn: state.siteConfig.authn,
        viewMethod: state.viewUpdate.explorerViewMethod,
    };
};

const mapDispatchToProps = (dispatch) => {
    return {
        toggleSnackbar: (vertical, horizontal, msg, color) => {
            dispatch(toggleSnackbar(vertical, horizontal, msg, color));
        },
        applyThemes: (color) => {
            dispatch(applyThemes(color));
        },
        toggleDaylightMode: () => {
            dispatch(toggleDaylightMode());
        },
        changeView: (method) => {
            dispatch(changeViewMethod(method));
        },
    };
};

class UserSettingCompoment extends Component {
    constructor(props) {
        super(props);
        this.fileInput = React.createRef();
    }

    state = {
        avatarModal: false,
        nickModal: false,
        changePassword: false,
        loading: "",
        oldPwd: "",
        newPwd: "",
        webdavPwd: "",
        newPwdRepeat: "",
        twoFactor: false,
        authCode: "",
        changeTheme: false,
        chosenTheme: null,
        showWebDavUrl: false,
        showWebDavUserName: false,
        changeWebDavPwd: false,
        groupBackModal: false,
        changePolicy: false,
        changeTimeZone: false,
        settings: {
            uid: 0,
            group_expires: 0,
            policy: {
                current: {
                    name: "-",
                    id: "",
                },
                options: [],
            },
            qq: "",
            homepage: true,
            two_factor: "",
            two_fa_secret: "",
            prefer_theme: "",
            themes: {},
            authn: [],
        },
    };

    handleClose = () => {
        this.setState({
            avatarModal: false,
            nickModal: false,
            changePassword: false,
            loading: "",
            twoFactor: false,
            changeTheme: false,
            showWebDavUrl: false,
            showWebDavUserName: false,
            changeWebDavPwd: false,
            groupBackModal: false,
            changePolicy: false,
        });
    };

    componentDidMount() {
        this.loadSetting();
    }

    toggleViewMethod = () => {
        const newMethod =
            this.props.viewMethod === "icon"
                ? "list"
                : this.props.viewMethod === "list"
                ? "smallIcon"
                : "icon";
        Auth.SetPreference("view_method", newMethod);
        this.props.changeView(newMethod);
    };

    loadSetting = () => {
        API.get("/user/setting")
            .then((response) => {
                const theme = JSON.parse(response.data.themes);
                response.data.themes = theme;
                this.setState({
                    settings: response.data,
                });
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
            });
    };

    useGravatar = () => {
        this.setState({
            loading: "gravatar",
        });
        API.put("/user/setting/avatar")
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "头像已更新，刷新后生效",
                    "success"
                );
                this.setState({
                    loading: "",
                });
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.setState({
                    loading: "",
                });
            });
    };

    changeNick = () => {
        this.setState({
            loading: "nick",
        });
        API.patch("/user/setting/nick", {
            nick: this.state.nick,
        })
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "昵称已更改，刷新后生效",
                    "success"
                );
                this.setState({
                    loading: "",
                });
                this.handleClose();
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.setState({
                    loading: "",
                });
            });
    };

    uploadAvatar = () => {
        this.setState({
            loading: "avatar",
        });
        const formData = new FormData();
        formData.append("avatar", this.fileInput.current.files[0]);
        API.post("/user/setting/avatar", formData, {
            headers: {
                "Content-Type": "multipart/form-data",
            },
        })
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "头像已更新，刷新后生效",
                    "success"
                );
                this.setState({
                    loading: "",
                });
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.setState({
                    loading: "",
                });
            });
    };

    handleToggle = () => {
        API.patch("/user/setting/homepage", {
            status: !this.state.settings.homepage,
        })
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "设置已保存",
                    "success"
                );
                this.setState({
                    settings: {
                        ...this.state.settings,
                        homepage: !this.state.settings.homepage,
                    },
                });
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
            });
    };

    changhePwd = () => {
        if (this.state.newPwd !== this.state.newPwdRepeat) {
            this.props.toggleSnackbar(
                "top",
                "right",
                "两次密码输入不一致",
                "warning"
            );
            return;
        }
        this.setState({
            loading: "changePassword",
        });
        API.patch("/user/setting/password", {
            old: this.state.oldPwd,
            new: this.state.newPwd,
        })
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "密码已更新",
                    "success"
                );
                this.setState({
                    loading: "",
                });
                this.handleClose();
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.setState({
                    loading: "",
                });
            });
    };

    changeTheme = () => {
        this.setState({
            loading: "changeTheme",
        });
        API.patch("/user/setting/theme", {
            theme: this.state.chosenTheme,
        })
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "主题配色已更换",
                    "success"
                );
                this.props.applyThemes(this.state.chosenTheme);
                this.setState({
                    loading: "",
                });
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.setState({
                    loading: "",
                });
            });
    };

    changheWebdavPwd = () => {
        this.setState({
            loading: "changheWebdavPwd",
        });
        axios
            .post("/Member/setWebdavPwd", {
                pwd: this.state.webdavPwd,
            })
            .then((response) => {
                if (response.data.error === "1") {
                    this.props.toggleSnackbar(
                        "top",
                        "right",
                        response.data.msg,
                        "error"
                    );
                    this.setState({
                        loading: "",
                    });
                } else {
                    this.props.toggleSnackbar(
                        "top",
                        "right",
                        response.data.msg,
                        "success"
                    );
                    this.setState({
                        loading: "",
                        changeWebDavPwd: false,
                    });
                }
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.setState({
                    loading: "",
                });
            });
    };

    init2FA = () => {
        if (this.state.settings.two_factor) {
            this.setState({ twoFactor: true });
            return;
        }
        API.get("/user/setting/2fa")
            .then((response) => {
                this.setState({
                    two_fa_secret: response.data,
                    twoFactor: true,
                });
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
            });
    };

    twoFactor = () => {
        this.setState({
            loading: "twoFactor",
        });
        API.patch("/user/setting/2fa", {
            code: this.state.authCode,
        })
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "设定已保存",
                    "success"
                );
                this.setState({
                    loading: "",
                    settings: {
                        ...this.state.settings,
                        two_factor: !this.state.settings.two_factor,
                    },
                });
                this.handleClose();
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.setState({
                    loading: "",
                });
            });
    };

    handleChange = (name) => (event) => {
        this.setState({ [name]: event.target.value });
    };

    handleAlignment = (event, chosenTheme) => this.setState({ chosenTheme });

    toggleThemeMode = (current) => {
        if (current !== null) {
            this.props.toggleDaylightMode();
            Auth.SetPreference("theme_mode", null);
        }
    };

    render() {
        const { classes } = this.props;
        const user = Auth.GetUser();
        const dark = Auth.GetPreference("theme_mode");

        return (
            <div>
                <div className={classes.layout}>
                    <Typography
                        className={classes.sectionTitle}
                        variant="subtitle2"
                    >
                        个人资料
                    </Typography>
                    <Paper>
                        <List className={classes.desenList}>
                            <ListItem
                                button
                                onClick={() =>
                                    this.setState({ avatarModal: true })
                                }
                            >
                                <ListItemAvatar>
                                    <Avatar
                                        src={
                                            "/api/v3/user/avatar/" +
                                            user.id +
                                            "/l"
                                        }
                                    />
                                </ListItemAvatar>
                                <ListItemText primary="头像" />
                                <ListItemSecondaryAction>
                                    <RightIcon className={classes.rightIcon} />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem button>
                                <ListItemIcon className={classes.iconFix}>
                                    <PermContactCalendar />
                                </ListItemIcon>
                                <ListItemText primary="UID" />

                                <ListItemSecondaryAction>
                                    <Typography
                                        className={classes.infoTextWithIcon}
                                        color="textSecondary"
                                    >
                                        {this.state.settings.uid}
                                    </Typography>
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem
                                button
                                onClick={() =>
                                    this.setState({ nickModal: true })
                                }
                            >
                                <ListItemIcon className={classes.iconFix}>
                                    <NickIcon />
                                </ListItemIcon>
                                <ListItemText primary="昵称" />

                                <ListItemSecondaryAction
                                    onClick={() =>
                                        this.setState({ nickModal: true })
                                    }
                                    className={classes.flexContainer}
                                >
                                    <Typography
                                        className={classes.infoTextWithIcon}
                                        color="textSecondary"
                                    >
                                        {user.nickname}
                                    </Typography>
                                    <RightIcon
                                        className={classes.rightIconWithText}
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem button>
                                <ListItemIcon className={classes.iconFix}>
                                    <EmailIcon />
                                </ListItemIcon>
                                <ListItemText primary="Email" />

                                <ListItemSecondaryAction>
                                    <Typography
                                        className={classes.infoText}
                                        color="textSecondary"
                                    >
                                        {user.user_name}
                                    </Typography>
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem button>
                                <ListItemIcon className={classes.iconFix}>
                                    <GroupIcon />
                                </ListItemIcon>
                                <ListItemText primary="用户组" />

                                <ListItemSecondaryAction>
                                    <Typography
                                        className={classes.infoText}
                                        color="textSecondary"
                                    >
                                        {user.group.name}
                                    </Typography>
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem button>
                                <ListItemIcon className={classes.iconFix}>
                                    <DateIcon />
                                </ListItemIcon>
                                <ListItemText primary="注册时间" />

                                <ListItemSecondaryAction>
                                    <Typography
                                        className={classes.infoText}
                                        color="textSecondary"
                                    >
                                        {formatLocalTime(
                                            user.created_at,
                                            "YYYY-MM-DD H:mm:ss"
                                        )}
                                    </Typography>
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Paper>
                    <Typography
                        className={classes.sectionTitle}
                        variant="subtitle2"
                    >
                        安全隐私
                    </Typography>
                    <Paper>
                        <List className={classes.desenList}>
                            <ListItem button>
                                <ListItemIcon className={classes.iconFix}>
                                    <HomeIcon />
                                </ListItemIcon>
                                <ListItemText primary="个人主页" />

                                <ListItemSecondaryAction>
                                    <Switch
                                        onChange={this.handleToggle}
                                        checked={this.state.settings.homepage}
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem
                                button
                                onClick={() =>
                                    this.setState({ changePassword: true })
                                }
                            >
                                <ListItemIcon className={classes.iconFix}>
                                    <LockIcon />
                                </ListItemIcon>
                                <ListItemText primary="登录密码" />

                                <ListItemSecondaryAction
                                    className={classes.flexContainer}
                                >
                                    <RightIcon className={classes.rightIcon} />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem button onClick={() => this.init2FA()}>
                                <ListItemIcon className={classes.iconFix}>
                                    <VerifyIcon />
                                </ListItemIcon>
                                <ListItemText primary="二步验证" />

                                <ListItemSecondaryAction
                                    className={classes.flexContainer}
                                >
                                    <Typography
                                        className={classes.infoTextWithIcon}
                                        color="textSecondary"
                                    >
                                        {!this.state.settings.two_factor
                                            ? "未开启"
                                            : "已开启"}
                                    </Typography>
                                    <RightIcon
                                        className={classes.rightIconWithText}
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Paper>

                    <Authn
                        list={this.state.settings.authn}
                        add={(credential) => {
                            this.setState({
                                settings: {
                                    ...this.state.settings,
                                    authn: [
                                        ...this.state.settings.authn,
                                        credential,
                                    ],
                                },
                            });
                        }}
                        remove={(id) => {
                            let credentials = [...this.state.settings.authn];
                            credentials = credentials.filter((v) => {
                                return v.id !== id;
                            });
                            this.setState({
                                settings: {
                                    ...this.state.settings,
                                    authn: credentials,
                                },
                            });
                        }}
                    />

                    <Typography
                        className={classes.sectionTitle}
                        variant="subtitle2"
                    >
                        个性化
                    </Typography>
                    <Paper>
                        <List className={classes.desenList}>
                            <ListItem
                                button
                                onClick={() =>
                                    this.setState({ changeTheme: true })
                                }
                            >
                                <ListItemIcon className={classes.iconFix}>
                                    <ColorIcon />
                                </ListItemIcon>
                                <ListItemText primary="主题配色" />

                                <ListItemSecondaryAction
                                    className={classes.flexContainer}
                                >
                                    <div className={classes.firstColor}></div>
                                    <div className={classes.secondColor}></div>
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem
                                button
                                onClick={() => this.toggleThemeMode(dark)}
                            >
                                <ListItemIcon className={classes.iconFix}>
                                    <Brightness3 />
                                </ListItemIcon>
                                <ListItemText primary="黑暗模式" />

                                <ListItemSecondaryAction
                                    className={classes.flexContainer}
                                >
                                    <Typography
                                        className={classes.infoTextWithIcon}
                                        color="textSecondary"
                                    >
                                        {dark &&
                                            (dark === "dark"
                                                ? "偏好开启"
                                                : "偏好关闭")}
                                        {dark === null && "跟随系统"}
                                    </Typography>
                                    <RightIcon
                                        className={classes.rightIconWithText}
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem
                                button
                                onClick={() => this.toggleViewMethod()}
                            >
                                <ListItemIcon className={classes.iconFix}>
                                    <ListAlt />
                                </ListItemIcon>
                                <ListItemText primary="文件列表" />

                                <ListItemSecondaryAction
                                    className={classes.flexContainer}
                                >
                                    <Typography
                                        className={classes.infoTextWithIcon}
                                        color="textSecondary"
                                    >
                                        {this.props.viewMethod === "icon" &&
                                            "大图标"}
                                        {this.props.viewMethod === "list" &&
                                            "列表"}
                                        {this.props.viewMethod ===
                                            "smallIcon" && "小图标"}
                                    </Typography>
                                    <RightIcon
                                        className={classes.rightIconWithText}
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem
                                onClick={() =>
                                    this.setState({ changeTimeZone: true })
                                }
                                button
                            >
                                <ListItemIcon className={classes.iconFix}>
                                    <Schedule />
                                </ListItemIcon>
                                <ListItemText primary="时区" />

                                <ListItemSecondaryAction
                                    className={classes.flexContainer}
                                >
                                    <Typography
                                        className={classes.infoTextWithIcon}
                                        color="textSecondary"
                                    >
                                        {timeZone}
                                    </Typography>
                                    <RightIcon
                                        className={classes.rightIconWithText}
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Paper>
                    {user.group.webdav && (
                        <div>
                            <Typography
                                className={classes.sectionTitle}
                                variant="subtitle2"
                            >
                                WebDAV
                            </Typography>
                            <Paper>
                                <List className={classes.desenList}>
                                    <ListItem
                                        button
                                        onClick={() =>
                                            this.setState({
                                                showWebDavUrl: true,
                                            })
                                        }
                                    >
                                        <ListItemIcon
                                            className={classes.iconFix}
                                        >
                                            <LinkIcon />
                                        </ListItemIcon>
                                        <ListItemText primary="连接地址" />

                                        <ListItemSecondaryAction
                                            className={classes.flexContainer}
                                        >
                                            <RightIcon
                                                className={classes.rightIcon}
                                            />
                                        </ListItemSecondaryAction>
                                    </ListItem>
                                    <Divider />
                                    <ListItem
                                        button
                                        onClick={() =>
                                            this.setState({
                                                showWebDavUserName: true,
                                            })
                                        }
                                    >
                                        <ListItemIcon
                                            className={classes.iconFix}
                                        >
                                            <InputIcon />
                                        </ListItemIcon>
                                        <ListItemText primary="用户名" />

                                        <ListItemSecondaryAction
                                            className={classes.flexContainer}
                                        >
                                            <RightIcon
                                                className={classes.rightIcon}
                                            />
                                        </ListItemSecondaryAction>
                                    </ListItem>
                                    <Divider />
                                    <ListItem
                                        button
                                        onClick={() =>
                                            this.props.history.push("/webdav?")
                                        }
                                    >
                                        <ListItemIcon
                                            className={classes.iconFix}
                                        >
                                            <SecurityIcon />
                                        </ListItemIcon>
                                        <ListItemText primary="账号管理" />

                                        <ListItemSecondaryAction
                                            className={classes.flexContainer}
                                        >
                                            <RightIcon
                                                className={classes.rightIcon}
                                            />
                                        </ListItemSecondaryAction>
                                    </ListItem>
                                </List>
                            </Paper>
                        </div>
                    )}
                    <div className={classes.paddingBottom}></div>
                </div>
                <TimeZoneDialog
                    onClose={() => this.setState({ changeTimeZone: false })}
                    open={this.state.changeTimeZone}
                />
                <Dialog
                    open={this.state.avatarModal}
                    onClose={this.handleClose}
                >
                    <DialogTitle>修改头像</DialogTitle>
                    <List>
                        <ListItem
                            button
                            component="label"
                            disabled={this.state.loading === "avatar"}
                        >
                            <input
                                type="file"
                                accept="image/*"
                                style={{ display: "none" }}
                                ref={this.fileInput}
                                onChange={this.uploadAvatar}
                            />
                            <ListItemAvatar>
                                <Avatar className={classes.uploadFromFile}>
                                    <PhotoIcon />
                                </Avatar>
                            </ListItemAvatar>
                            <ListItemText primary="从文件上传" />
                        </ListItem>
                        <ListItem
                            button
                            onClick={this.useGravatar}
                            disabled={this.state.loading === "gravatar"}
                        >
                            <ListItemAvatar>
                                <Avatar className={classes.userGravatar}>
                                    <FingerprintIcon />
                                </Avatar>
                            </ListItemAvatar>
                            <ListItemText
                                className={classes.paddingText}
                                primary="使用 Gravatar 头像 "
                            />
                        </ListItem>
                    </List>
                    <DialogActions>
                        <Button onClick={this.handleClose} color="primary">
                            取消
                        </Button>
                    </DialogActions>
                </Dialog>
                <Dialog open={this.state.nickModal} onClose={this.handleClose}>
                    <DialogTitle>修改昵称</DialogTitle>
                    <DialogContent>
                        <TextField
                            id="standard-name"
                            label="昵称"
                            className={classes.textField}
                            value={this.state.nick}
                            onChange={this.handleChange("nick")}
                            margin="normal"
                            autoFocus
                        />
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.handleClose} color="default">
                            取消
                        </Button>
                        <Button
                            onClick={this.changeNick}
                            color="primary"
                            disabled={
                                this.state.loading === "nick" ||
                                this.state.nick === ""
                            }
                        >
                            保存
                        </Button>
                    </DialogActions>
                </Dialog>
                <Dialog
                    open={this.state.changePassword}
                    onClose={this.handleClose}
                >
                    <DialogTitle>修改登录密码</DialogTitle>
                    <DialogContent>
                        <div>
                            <TextField
                                id="standard-name"
                                label="原密码"
                                type="password"
                                className={classes.textField}
                                value={this.state.oldPwd}
                                onChange={this.handleChange("oldPwd")}
                                margin="normal"
                                autoFocus
                            />
                        </div>
                        <div>
                            <TextField
                                id="standard-name"
                                label="新密码"
                                type="password"
                                className={classes.textField}
                                value={this.state.newPwd}
                                onChange={this.handleChange("newPwd")}
                                margin="normal"
                            />
                        </div>
                        <div>
                            <TextField
                                id="standard-name"
                                label="确认新密码"
                                type="password"
                                className={classes.textField}
                                value={this.state.newPwdRepeat}
                                onChange={this.handleChange("newPwdRepeat")}
                                margin="normal"
                            />
                        </div>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.handleClose} color="default">
                            取消
                        </Button>
                        <Button
                            onClick={this.changhePwd}
                            color="primary"
                            disabled={
                                this.state.loading === "changePassword" ||
                                this.state.oldPwd === "" ||
                                this.state.newPwdRepeat === "" ||
                                this.state.newPwd === ""
                            }
                        >
                            保存
                        </Button>
                    </DialogActions>
                </Dialog>
                <Dialog open={this.state.twoFactor} onClose={this.handleClose}>
                    <DialogTitle>
                        {this.state.settings.two_factor ? "关闭" : "启用"}
                        二步验证
                    </DialogTitle>
                    <DialogContent>
                        <div className={classes.flexContainerResponse}>
                            {!this.state.settings.two_factor && (
                                <div className={classes.qrcode}>
                                    <QRCode
                                        value={
                                            "otpauth://totp/" +
                                            this.props.title +
                                            "?secret=" +
                                            this.state.two_fa_secret
                                        }
                                    />
                                </div>
                            )}

                            <div className={classes.desText}>
                                {!this.state.settings.two_factor && (
                                    <Typography>
                                        请使用任意二步验证APP或者支持二步验证的密码管理软件扫描左侧二维码添加本站。扫描完成后请填写二步验证APP给出的6位验证码以开启二步验证。
                                    </Typography>
                                )}
                                {this.state.settings.two_factor && (
                                    <Typography>
                                        请验证当前二步验证代码。
                                    </Typography>
                                )}
                                <TextField
                                    id="standard-name"
                                    label="6位验证码"
                                    type="number"
                                    className={classes.textField}
                                    value={this.state.authCode}
                                    onChange={this.handleChange("authCode")}
                                    margin="normal"
                                    autoFocus
                                    fullWidth
                                />
                            </div>
                        </div>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.handleClose} color="default">
                            取消
                        </Button>
                        <Button
                            onClick={this.twoFactor}
                            color="primary"
                            disabled={
                                this.state.loading === "twoFactor" ||
                                this.state.authCode === ""
                            }
                        >
                            {this.state.settings.two_factor ? "关闭" : "启用"}
                            二步验证
                        </Button>
                    </DialogActions>
                </Dialog>
                <Dialog
                    open={this.state.changeTheme}
                    onClose={this.handleClose}
                >
                    <DialogTitle>更改主题配色</DialogTitle>
                    <DialogContent>
                        <ToggleButtonGroup
                            value={this.state.chosenTheme}
                            exclusive
                            onChange={this.handleAlignment}
                        >
                            {Object.keys(this.state.settings.themes).map(
                                (value, key) => (
                                    <ToggleButton value={value} key={key}>
                                        <div
                                            className={classes.themeBlock}
                                            style={{ backgroundColor: value }}
                                        />
                                    </ToggleButton>
                                )
                            )}
                        </ToggleButtonGroup>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.handleClose} color="default">
                            取消
                        </Button>
                        <Button
                            onClick={this.changeTheme}
                            color="primary"
                            disabled={
                                this.state.loading === "changeTheme" ||
                                this.state.chosenTheme === null
                            }
                        >
                            保存
                        </Button>
                    </DialogActions>
                </Dialog>
                <Dialog
                    open={this.state.showWebDavUrl}
                    onClose={this.handleClose}
                >
                    <DialogTitle>WebDAV连接地址</DialogTitle>
                    <DialogContent>
                        <TextField
                            id="standard-name"
                            className={classes.textField}
                            value={window.location.origin + "/dav"}
                            margin="normal"
                            autoFocus
                        />
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.handleClose} color="default">
                            关闭
                        </Button>
                    </DialogActions>
                </Dialog>
                <Dialog
                    open={this.state.showWebDavUserName}
                    onClose={this.handleClose}
                >
                    <DialogTitle>WebDAV用户名</DialogTitle>
                    <DialogContent>
                        <TextField
                            id="standard-name"
                            className={classes.textField}
                            value={user.user_name}
                            margin="normal"
                            autoFocus
                        />
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.handleClose} color="default">
                            关闭
                        </Button>
                    </DialogActions>
                </Dialog>
            </div>
        );
    }
}

const UserSetting = connect(
    mapStateToProps,
    mapDispatchToProps
)(withStyles(styles)(withRouter(UserSettingCompoment)));

export default UserSetting;
