import React, { useCallback, useEffect, useState } from "react";
import { makeStyles } from "@material-ui/core";
import { useDispatch, useSelector } from "react-redux";
import Drawer from "@material-ui/core/Drawer";
import Toolbar from "@material-ui/core/Toolbar";
import { Clear, Folder } from "@material-ui/icons";
import Divider from "@material-ui/core/Divider";
import { setSideBar } from "../../../redux/explorer/action";
import TypeIcon from "../TypeIcon";
import Typography from "@material-ui/core/Typography";
import IconButton from "@material-ui/core/IconButton";
import Grid from "@material-ui/core/Grid";
import API from "../../../middleware/Api";
import { navigateTo, toggleSnackbar } from "../../../actions";
import { filename, sizeToString } from "../../../utils";
import Link from "@material-ui/core/Link";
import Tooltip from "@material-ui/core/Tooltip";
import TimeAgo from "timeago-react";
import ListLoading from "../../Placeholder/ListLoading";
import Hidden from "@material-ui/core/Hidden";
import Dialog from "@material-ui/core/Dialog";
import Slide from "@material-ui/core/Slide";
import AppBar from "@material-ui/core/AppBar";
import { formatLocalTime } from "../../../utils/datetime";

const drawerWidth = 350;

const useStyles = makeStyles((theme) => ({
    drawer: {
        width: drawerWidth,
        flexShrink: 0,
    },
    drawerPaper: {
        width: drawerWidth,
        boxShadow:
            "0px 8px 10px -5px rgb(0 0 0 / 20%), 0px 16px 24px 2px rgb(0 0 0 / 14%), 0px 6px 30px 5px rgb(0 0 0 / 12%)",
    },
    drawerContainer: {
        overflow: "auto",
    },
    header: {
        display: "flex",
        padding: theme.spacing(3),
        placeContent: "space-between",
    },
    fileIcon: { width: 33, height: 33 },
    fileIconSVG: { fontSize: 20 },
    folderIcon: {
        color: theme.palette.text.secondary,
        width: 33,
        height: 33,
    },
    fileName: {
        marginLeft: theme.spacing(2),
        marginRight: theme.spacing(2),
        wordBreak: "break-all",
        flexGrow: 2,
    },
    closeIcon: {
        placeSelf: "flex-start",
        marginTop: 2,
    },
    propsContainer: {
        padding: theme.spacing(3),
    },
    propsLabel: {
        color: theme.palette.text.secondary,
        padding: theme.spacing(1),
    },
    propsTime: {
        color: theme.palette.text.disabled,
        padding: theme.spacing(1),
    },
    propsValue: {
        padding: theme.spacing(1),
        wordBreak: "break-all",
    },
    appBar: {
        position: "relative",
    },
    title: {
        marginLeft: theme.spacing(2),
        flex: 1,
    },
}));

const Transition = React.forwardRef(function Transition(props, ref) {
    return <Slide direction="up" ref={ref} {...props} />;
});

export default function SideDrawer() {
    const dispatch = useDispatch();
    const sideBarOpen = useSelector((state) => state.explorer.sideBarOpen);
    const selected = useSelector((state) => state.explorer.selected);
    const SetSideBar = useCallback((open) => dispatch(setSideBar(open)), [
        dispatch,
    ]);
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );
    const NavigateTo = useCallback((k) => dispatch(navigateTo(k)), [dispatch]);
    const keywords = useSelector((state) => state.explorer.keywords);
    const [target, setTarget] = useState(null);
    const [details, setDetails] = useState(null);
    const loadProps = (object) => {
        API.get(
            "/object/property/" +
                object.id +
                "?trace_root=" +
                (keywords !== "" ? "true" : "false") +
                "&is_folder=" +
                (object.type === "dir").toString()
        )
            .then((response) => {
                setDetails(response.data);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    };

    useEffect(() => {
        setDetails(null);
        if (sideBarOpen) {
            if (selected.length !== 1) {
                SetSideBar(false);
            } else {
                setTarget(selected[0]);
                loadProps(selected[0]);
            }
        }
    }, [selected, sideBarOpen]);

    const classes = useStyles();
    const propsItem = [
        {
            label: "大小",
            value: (d, t) =>
                sizeToString(d.size) +
                "  (" +
                d.size.toLocaleString() +
                " 字节)",
            show: (d) => true,
        },
        {
            label: "存储策略",
            value: (d, t) => d.policy,
            show: (d) => d.type === "file",
        },
        {
            label: "包含目录",
            value: (d, t) => d.child_folder_num.toLocaleString() + " " + "个",
            show: (d) => d.type === "dir",
        },
        {
            label: "包含文件",
            value: (d, t) => d.child_file_num.toLocaleString() + " " + "个",
            show: (d) => d.type === "dir",
        },
        {
            label: "所在目录",
            // eslint-disable-next-line react/display-name
            value: (d, t) => {
                const path = d.path === "" ? t.path : d.path;
                const name = filename(path);
                return (
                    <Tooltip title={path}>
                        <Link
                            href={"javascript:void"}
                            onClick={() => NavigateTo(path)}
                        >
                            {name === "" ? "根目录" : name}
                        </Link>
                    </Tooltip>
                );
            },
            show: (d) => true,
        },
        {
            label: "修改于",
            value: (d, t) =>
                formatLocalTime(d.updated_at, "YYYY/MM/DD  H:mm:ss"),
            show: (d) => true,
        },
        {
            label: "创建于",
            value: (d) => formatLocalTime(d.created_at, "YYYY/MM/DD  H:mm:ss"),
            show: (d) => true,
        },
    ];
    const content = (
        <Grid container className={classes.propsContainer}>
            {!details && <ListLoading />}
            {details && (
                <>
                    {propsItem.map((item) => {
                        if (item.show(target)) {
                            return (
                                <>
                                    <Grid
                                        item
                                        xs={5}
                                        className={classes.propsLabel}
                                    >
                                        {item.label}
                                    </Grid>
                                    <Grid
                                        item
                                        xs={7}
                                        className={classes.propsValue}
                                    >
                                        {item.value(details, target)}
                                    </Grid>
                                </>
                            );
                        }
                    })}
                    {target.type === "dir" && (
                        <Grid item xs={12} className={classes.propsTime}>
                            统计于{" "}
                            <TimeAgo
                                datetime={details.query_date}
                                locale="zh_CN"
                            />
                        </Grid>
                    )}
                </>
            )}
        </Grid>
    );
    return (
        <>
            <Hidden smUp>
                <Dialog
                    fullScreen
                    open={sideBarOpen}
                    TransitionComponent={Transition}
                >
                    {target && (
                        <>
                            <AppBar className={classes.appBar}>
                                <Toolbar>
                                    <IconButton
                                        edge="start"
                                        color="inherit"
                                        onClick={() => SetSideBar(false)}
                                        aria-label="close"
                                    >
                                        <Clear />
                                    </IconButton>
                                    <Typography
                                        variant="h6"
                                        className={classes.title}
                                    >
                                        {target.name}
                                    </Typography>
                                </Toolbar>
                            </AppBar>
                            {content}
                        </>
                    )}
                </Dialog>
            </Hidden>
            <Hidden xsDown>
                <Drawer
                    className={classes.drawer}
                    variant="persistent"
                    classes={{
                        paper: classes.drawerPaper,
                    }}
                    open={sideBarOpen}
                    anchor="right"
                >
                    <Toolbar />
                    <div className={classes.drawerContainer}>
                        {target && (
                            <>
                                <div className={classes.header}>
                                    {target.type === "dir" && (
                                        <Folder
                                            className={classes.folderIcon}
                                        />
                                    )}
                                    {target.type !== "dir" && (
                                        <TypeIcon
                                            isUpload
                                            className={classes.fileIcon}
                                            iconClassName={classes.fileIconSVG}
                                            fileName={target.name}
                                        />
                                    )}
                                    <div className={classes.fileName}>
                                        <Typography variant="h6" gutterBottom>
                                            {target.name}
                                        </Typography>
                                    </div>
                                    <IconButton
                                        onClick={() => SetSideBar(false)}
                                        className={classes.closeIcon}
                                        aria-label="close"
                                        size={"small"}
                                    >
                                        <Clear />
                                    </IconButton>
                                </div>
                            </>
                        )}
                        <Divider />
                        {content}
                    </div>
                </Drawer>
            </Hidden>
        </>
    );
}
