import React, { useCallback } from "react";
import {
    Checkbox,
    FormControl,
    makeStyles,
    TextField,
} from "@material-ui/core";
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    CircularProgress,
} from "@material-ui/core";
import { toggleSnackbar } from "../../actions/index";
import { useDispatch } from "react-redux";
import API from "../../middleware/Api";
import List from "@material-ui/core/List";
import ListItemText from "@material-ui/core/ListItemText";
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import LockIcon from "@material-ui/icons/Lock";
import TimerIcon from "@material-ui/icons/Timer";
import CasinoIcon from "@material-ui/icons/Casino";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import Divider from "@material-ui/core/Divider";
import MuiExpansionPanel from "@material-ui/core/ExpansionPanel";
import MuiExpansionPanelSummary from "@material-ui/core/ExpansionPanelSummary";
import MuiExpansionPanelDetails from "@material-ui/core/ExpansionPanelDetails";
import Typography from "@material-ui/core/Typography";
import withStyles from "@material-ui/core/styles/withStyles";
import InputLabel from "@material-ui/core/InputLabel";
import { Visibility, VisibilityOff } from "@material-ui/icons";
import IconButton from "@material-ui/core/IconButton";
import InputAdornment from "@material-ui/core/InputAdornment";
import OutlinedInput from "@material-ui/core/OutlinedInput";
import Tooltip from "@material-ui/core/Tooltip";
import MenuItem from "@material-ui/core/MenuItem";
import Select from "@material-ui/core/Select";
import EyeIcon from "@material-ui/icons/RemoveRedEye";
import ToggleIcon from "material-ui-toggle-icon";

const useStyles = makeStyles((theme) => ({
    widthAnimation: {},
    shareUrl: {
        minWidth: "400px",
    },
    wrapper: {
        margin: theme.spacing(1),
        position: "relative",
    },
    buttonProgress: {
        color: theme.palette.secondary.light,
        position: "absolute",
        top: "50%",
        left: "50%",
    },
    flexCenter: {
        alignItems: "center",
    },
    noFlex: {
        display: "block",
    },
    scoreCalc: {
        marginTop: 10,
    },
}));

const ExpansionPanel = withStyles({
    root: {
        border: "0px solid rgba(0, 0, 0, .125)",
        boxShadow: "none",
        "&:not(:last-child)": {
            borderBottom: 0,
        },
        "&:before": {
            display: "none",
        },
        "&$expanded": {
            margin: "auto",
        },
    },
    expanded: {},
})(MuiExpansionPanel);

const ExpansionPanelSummary = withStyles({
    root: {
        padding: 0,
        "&$expanded": {},
    },
    content: {
        margin: 0,
        display: "initial",
        "&$expanded": {
            margin: "0 0",
        },
    },
    expanded: {},
})(MuiExpansionPanelSummary);

const ExpansionPanelDetails = withStyles((theme) => ({
    root: {
        padding: 24,
        backgroundColor: theme.palette.background.default,
    },
}))(MuiExpansionPanelDetails);

export default function CreatShare(props) {
    const dispatch = useDispatch();
    const classes = useStyles();

    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const [expanded, setExpanded] = React.useState(false);
    const [shareURL, setShareURL] = React.useState("");
    const [values, setValues] = React.useState({
        password: "",
        downloads: 1,
        expires: 24 * 3600,
        showPassword: false,
    });
    const [shareOption, setShareOption] = React.useState({
        password: false,
        expire: false,
        preview: true,
    });

    const handleChange = (prop) => (event) => {
        // 输入密码
        if (prop === "password") {
            if (event.target.value === "") {
                setShareOption({ ...shareOption, password: false });
            } else {
                setShareOption({ ...shareOption, password: true });
            }
        }

        setValues({ ...values, [prop]: event.target.value });
    };

    const handleClickShowPassword = () => {
        setValues({ ...values, showPassword: !values.showPassword });
    };

    const handleMouseDownPassword = (event) => {
        event.preventDefault();
    };

    const randomPassword = () => {
        setShareOption({ ...shareOption, password: true });
        setValues({
            ...values,
            password: Math.random().toString(36).substr(2).slice(2, 8),
            showPassword: true,
        });
    };

    const handleExpand = (panel) => (event, isExpanded) => {
        setExpanded(isExpanded ? panel : false);
    };

    const handleCheck = (prop) => () => {
        if (!shareOption[prop]) {
            handleExpand(prop)(null, true);
        }
        if (prop === "password" && shareOption[prop]) {
            setValues({
                ...values,
                password: "",
            });
        }
        setShareOption({ ...shareOption, [prop]: !shareOption[prop] });
    };

    const onClose = () => {
        props.onClose();
        setTimeout(() => {
            setShareURL("");
        }, 500);
    };

    const submitShare = (e) => {
        e.preventDefault();
        props.setModalsLoading(true);
        const submitFormBody = {
            id: props.selected[0].id,
            is_dir: props.selected[0].type === "dir",
            password: values.password,
            downloads: shareOption.expire ? values.downloads : -1,
            expire: values.expires,
            preview: shareOption.preview,
        };

        API.post("/share", submitFormBody)
            .then((response) => {
                setShareURL(response.data);
                setValues({
                    password: "",
                    downloads: 1,
                    expires: 24 * 3600,
                    showPassword: false,
                });
                setShareOption({
                    password: false,
                    expire: false,
                });
                props.setModalsLoading(false);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
                props.setModalsLoading(false);
            });
    };

    const handleFocus = (event) => event.target.select();

    return (
        <Dialog
            open={props.open}
            onClose={onClose}
            aria-labelledby="form-dialog-title"
            className={classes.widthAnimation}
            maxWidth="xs"
            fullWidth
        >
            <DialogTitle id="form-dialog-title">创建分享链接</DialogTitle>

            {shareURL === "" && (
                <>
                    <Divider />
                    <List>
                        <ExpansionPanel
                            expanded={expanded === "password"}
                            onChange={handleExpand("password")}
                        >
                            <ExpansionPanelSummary
                                aria-controls="panel1a-content"
                                id="panel1a-header"
                            >
                                <ListItem button>
                                    <ListItemIcon>
                                        <LockIcon />
                                    </ListItemIcon>
                                    <ListItemText primary="使用密码保护" />
                                    <ListItemSecondaryAction>
                                        <Checkbox
                                            checked={shareOption.password}
                                            onChange={handleCheck("password")}
                                        />
                                    </ListItemSecondaryAction>
                                </ListItem>
                            </ExpansionPanelSummary>
                            <ExpansionPanelDetails>
                                <FormControl
                                    variant="outlined"
                                    color="secondary"
                                    fullWidth
                                >
                                    <InputLabel htmlFor="filled-adornment-password">
                                        分享密码
                                    </InputLabel>
                                    <OutlinedInput
                                        fullWidth
                                        id="outlined-adornment-password"
                                        type={
                                            values.showPassword
                                                ? "text"
                                                : "password"
                                        }
                                        value={values.password}
                                        onChange={handleChange("password")}
                                        endAdornment={
                                            <InputAdornment position="end">
                                                <Tooltip title="随机生成">
                                                    <IconButton
                                                        aria-label="toggle password visibility"
                                                        onClick={randomPassword}
                                                        edge="end"
                                                    >
                                                        <CasinoIcon />
                                                    </IconButton>
                                                </Tooltip>
                                                <IconButton
                                                    aria-label="toggle password visibility"
                                                    onClick={
                                                        handleClickShowPassword
                                                    }
                                                    onMouseDown={
                                                        handleMouseDownPassword
                                                    }
                                                    edge="end"
                                                >
                                                    <ToggleIcon
                                                        on={values.showPassword}
                                                        onIcon={<Visibility />}
                                                        offIcon={
                                                            <VisibilityOff />
                                                        }
                                                    />
                                                </IconButton>
                                            </InputAdornment>
                                        }
                                        labelWidth={70}
                                    />
                                </FormControl>
                            </ExpansionPanelDetails>
                        </ExpansionPanel>
                        <ExpansionPanel
                            expanded={expanded === "expire"}
                            onChange={handleExpand("expire")}
                        >
                            <ExpansionPanelSummary
                                aria-controls="panel1a-content"
                                id="panel1a-header"
                            >
                                <ListItem button>
                                    <ListItemIcon>
                                        <TimerIcon />
                                    </ListItemIcon>
                                    <ListItemText primary="自动过期" />
                                    <ListItemSecondaryAction>
                                        <Checkbox
                                            checked={shareOption.expire}
                                            onChange={handleCheck("expire")}
                                        />
                                    </ListItemSecondaryAction>
                                </ListItem>
                            </ExpansionPanelSummary>
                            <ExpansionPanelDetails
                                className={classes.flexCenter}
                            >
                                <FormControl
                                    style={{
                                        marginRight: 10,
                                    }}
                                >
                                    <Select
                                        labelId="demo-simple-select-label"
                                        id="demo-simple-select"
                                        value={values.downloads}
                                        onChange={handleChange("downloads")}
                                    >
                                        <MenuItem value={1}>1 次下载</MenuItem>
                                        <MenuItem value={2}>2 次下载</MenuItem>
                                        <MenuItem value={3}>3 次下载</MenuItem>
                                        <MenuItem value={4}>4 次下载</MenuItem>
                                        <MenuItem value={5}>5 次下载</MenuItem>
                                        <MenuItem value={20}>
                                            20 次下载
                                        </MenuItem>
                                        <MenuItem value={50}>
                                            50 次下载
                                        </MenuItem>
                                        <MenuItem value={100}>
                                            100 次下载
                                        </MenuItem>
                                    </Select>
                                </FormControl>
                                <Typography>或者</Typography>
                                <FormControl
                                    style={{
                                        marginRight: 10,
                                        marginLeft: 10,
                                    }}
                                >
                                    <Select
                                        labelId="demo-simple-select-label"
                                        id="demo-simple-select"
                                        value={values.expires}
                                        onChange={handleChange("expires")}
                                    >
                                        <MenuItem value={300}>5 分钟</MenuItem>
                                        <MenuItem value={3600}>1 小时</MenuItem>
                                        <MenuItem value={24 * 3600}>
                                            1 天
                                        </MenuItem>
                                        <MenuItem value={7 * 24 * 3600}>
                                            7 天
                                        </MenuItem>
                                        <MenuItem value={30 * 24 * 3600}>
                                            30 天
                                        </MenuItem>
                                    </Select>
                                </FormControl>
                                <Typography>后过期</Typography>
                            </ExpansionPanelDetails>
                        </ExpansionPanel>
                        <ExpansionPanel
                            expanded={expanded === "preview"}
                            onChange={handleExpand("preview")}
                        >
                            <ExpansionPanelSummary
                                aria-controls="panel1a-content"
                                id="panel1a-header"
                            >
                                <ListItem button>
                                    <ListItemIcon>
                                        <LockIcon />
                                    </ListItemIcon>
                                    <ListItemText primary="允许预览" />
                                    <ListItemSecondaryAction>
                                        <Checkbox
                                            checked={shareOption.preview}
                                            onChange={handleCheck("preview")}
                                        />
                                    </ListItemSecondaryAction>
                                </ListItem>
                            </ExpansionPanelSummary>
                            <ExpansionPanelDetails>
                                <Typography>
                                    是否允许在分享页面预览文件内容
                                </Typography>
                            </ExpansionPanelDetails>
                        </ExpansionPanel>
                    </List>
                    <Divider />
                </>
            )}
            {shareURL !== "" && (
                <DialogContent>
                    <TextField
                        onFocus={handleFocus}
                        autoFocus
                        inputProps={{ readonly: true }}
                        label="分享链接"
                        value={shareURL}
                        variant="outlined"
                        fullWidth
                    />
                </DialogContent>
            )}

            <DialogActions>
                <Button onClick={onClose}>关闭</Button>

                {shareURL === "" && (
                    <div className={classes.wrapper}>
                        <Button
                            onClick={submitShare}
                            color="secondary"
                            disabled={props.modalsLoading}
                        >
                            创建分享链接
                            {props.modalsLoading && (
                                <CircularProgress
                                    size={24}
                                    className={classes.buttonProgress}
                                />
                            )}
                        </Button>
                    </div>
                )}
            </DialogActions>
        </Dialog>
    );
}
