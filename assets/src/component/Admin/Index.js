import Avatar from "@material-ui/core/Avatar";
import Button from "@material-ui/core/Button";
import Chip from "@material-ui/core/Chip";
import { blue, green, red, yellow } from "@material-ui/core/colors";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogTitle from "@material-ui/core/DialogTitle";
import Divider from "@material-ui/core/Divider";
import Grid from "@material-ui/core/Grid";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import Paper from "@material-ui/core/Paper";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import {
    Description,
    Favorite,
    FileCopy,
    Forum,
    GitHub,
    Home,
    Launch,
    Lock,
    People,
    Public,
    Telegram,
} from "@material-ui/icons";
import axios from "axios";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import {
    CartesianGrid,
    Legend,
    Line,
    LineChart,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";
import { ResponsiveContainer } from "recharts/lib/component/ResponsiveContainer";
import TimeAgo from "timeago-react";
import { toggleSnackbar } from "../../actions";
import API from "../../middleware/Api";
import pathHelper from "../../utils/page";

const useStyles = makeStyles((theme) => ({
    paper: {
        padding: theme.spacing(3),
        height: "100%",
    },
    logo: {
        width: 70,
    },
    logoContainer: {
        padding: theme.spacing(3),
        display: "flex",
    },
    title: {
        marginLeft: 16,
    },
    cloudreve: {
        fontSize: 25,
        color: theme.palette.text.secondary,
    },
    version: {
        color: theme.palette.text.hint,
    },
    links: {
        padding: theme.spacing(3),
    },
    iconRight: {
        minWidth: 0,
    },
    userIcon: {
        backgroundColor: blue[100],
        color: blue[600],
    },
    fileIcon: {
        backgroundColor: yellow[100],
        color: yellow[800],
    },
    publicIcon: {
        backgroundColor: green[100],
        color: green[800],
    },
    secretIcon: {
        backgroundColor: red[100],
        color: red[800],
    },
}));

export default function Index() {
    const classes = useStyles();
    const [lineData, setLineData] = useState([]);
    const [news, setNews] = useState([]);
    const [newsUsers, setNewsUsers] = useState({});
    const [open, setOpen] = React.useState(false);
    const [siteURL, setSiteURL] = React.useState("");
    const [statistics, setStatistics] = useState({
        fileTotal: 0,
        userTotal: 0,
        publicShareTotal: 0,
        secretShareTotal: 0,
    });
    const [version, setVersion] = useState({
        backend: "-",
    });

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const ResetSiteURL = () => {
        setOpen(false);
        API.patch("/admin/setting", {
            options: [
                {
                    key: "siteURL",
                    value: window.location.origin,
                },
            ],
        })
            .then(() => {
                setSiteURL(window.location.origin);
                ToggleSnackbar("top", "right", "设置已更改", "success");
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    };

    useEffect(() => {
        API.get("/admin/summary")
            .then((response) => {
                const data = [];
                response.data.date.forEach((v, k) => {
                    data.push({
                        name: v,
                        file: response.data.files[k],
                        user: response.data.users[k],
                        share: response.data.shares[k],
                    });
                });
                setLineData(data);
                setStatistics({
                    fileTotal: response.data.fileTotal,
                    userTotal: response.data.userTotal,
                    publicShareTotal: response.data.publicShareTotal,
                    secretShareTotal: response.data.secretShareTotal,
                });
                setVersion(response.data.version);
                setSiteURL(response.data.siteURL);
                if (
                    response.data.siteURL === "" ||
                    response.data.siteURL !== window.location.origin
                ) {
                    setOpen(true);
                }
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });

        axios
            .get("/api/v3/admin/news")
            .then((response) => {
                setNews(response.data.data);
                const res = {};
                response.data.included.forEach((v) => {
                    if (v.type === "users") {
                        res[v.id] = v.attributes;
                    }
                });
                setNewsUsers(res);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    }, []);

    return (
        <Grid container spacing={3}>
            <Dialog
                open={open}
                onClose={() => setOpen(false)}
                aria-labelledby="alert-dialog-title"
                aria-describedby="alert-dialog-description"
            >
                <DialogTitle id="alert-dialog-title">
                    {"确定站点URL设置"}
                </DialogTitle>
                <DialogContent>
                    <DialogContentText id="alert-dialog-description">
                        <Typography>
                            {siteURL === "" &&
                                "您尚未设定站点URL，是否要将其设定为当前的 " +
                                    window.location.origin +
                                    " ?"}
                            {siteURL !== "" &&
                                "您设置的站点URL与当前实际不一致，是否要将其设定为当前的 " +
                                    window.location.origin +
                                    " ?"}
                        </Typography>
                        <Typography>
                            此设置非常重要，请确保其与您站点的实际地址一致。你可以在
                            参数设置 - 站点信息 中更改此设置。
                        </Typography>
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setOpen(false)} color="default">
                        忽略
                    </Button>
                    <Button onClick={() => ResetSiteURL()} color="primary">
                        更改
                    </Button>
                </DialogActions>
            </Dialog>
            <Grid alignContent={"stretch"} item xs={12} md={8} lg={9}>
                <Paper className={classes.paper}>
                    <Typography variant="button" display="block" gutterBottom>
                        趋势
                    </Typography>
                    <ResponsiveContainer
                        width="100%"
                        aspect={pathHelper.isMobile() ? 4.0 / 3.0 : 3.0 / 1.0}
                    >
                        <LineChart width={1200} height={300} data={lineData}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="name" />
                            <YAxis />
                            <Tooltip />
                            <Legend />
                            <Line
                                name={"文件"}
                                type="monotone"
                                dataKey="file"
                                stroke="#3f51b5"
                            />
                            <Line
                                name={"用户"}
                                type="monotone"
                                dataKey="user"
                                stroke="#82ca9d"
                            />
                            <Line
                                name={"分享"}
                                type="monotone"
                                dataKey="share"
                                stroke="#e91e63"
                            />
                        </LineChart>
                    </ResponsiveContainer>
                </Paper>
            </Grid>
            <Grid item xs={12} md={4} lg={3}>
                <Paper className={classes.paper}>
                    <Typography variant="button" display="block" gutterBottom>
                        总计
                    </Typography>
                    <Divider />
                    <List className={classes.root}>
                        <ListItem>
                            <ListItemAvatar>
                                <Avatar className={classes.userIcon}>
                                    <People />
                                </Avatar>
                            </ListItemAvatar>
                            <ListItemText
                                primary={statistics.userTotal}
                                secondary="注册用户"
                            />
                        </ListItem>
                        <ListItem>
                            <ListItemAvatar>
                                <Avatar className={classes.fileIcon}>
                                    <FileCopy />
                                </Avatar>
                            </ListItemAvatar>
                            <ListItemText
                                primary={statistics.fileTotal}
                                secondary="文件总数"
                            />
                        </ListItem>
                        <ListItem>
                            <ListItemAvatar>
                                <Avatar className={classes.publicIcon}>
                                    <Public />
                                </Avatar>
                            </ListItemAvatar>
                            <ListItemText
                                primary={statistics.publicShareTotal}
                                secondary="公开分享总数"
                            />
                        </ListItem>
                        <ListItem>
                            <ListItemAvatar>
                                <Avatar className={classes.secretIcon}>
                                    <Lock />
                                </Avatar>
                            </ListItemAvatar>
                            <ListItemText
                                primary={statistics.secretShareTotal}
                                secondary="私密分享总数"
                            />
                        </ListItem>
                    </List>
                </Paper>
            </Grid>
            <Grid item xs={12} md={4} lg={3}>
                <Paper>
                    <div className={classes.logoContainer}>
                        <img
                            alt="网盘"
                            className={classes.logo}
                            src={"/static/img/cloudreve.svg"}
                        />
                        <div className={classes.title}>
                            <Typography className={classes.cloudreve}>
                                Cloudreve
                            </Typography>
                            <Typography className={classes.version}>
                                {version.backend}{" "}
                                {version.is_pro === "true" && (
                                    <Chip size="small" label="Pro" />
                                )}
                            </Typography>
                        </div>
                    </div>
                    <Divider />
                    <div>
                        <List component="nav" aria-label="main mailbox folders">
                            <ListItem
                                button
                                onClick={() =>
                                    window.open("https://cloudreve.org")
                                }
                            >
                                <ListItemIcon>
                                    <Home />
                                </ListItemIcon>
                                <ListItemText primary="主页" />
                                <ListItemIcon className={classes.iconRight}>
                                    <Launch />
                                </ListItemIcon>
                            </ListItem>
                            <ListItem
                                button
                                onClick={() =>
                                    window.open(
                                        "https://github.com/cloudreve/cloudreve"
                                    )
                                }
                            >
                                <ListItemIcon>
                                    <GitHub />
                                </ListItemIcon>
                                <ListItemText primary="GitHub" />
                                <ListItemIcon className={classes.iconRight}>
                                    <Launch />
                                </ListItemIcon>
                            </ListItem>
                            <ListItem
                                button
                                onClick={() =>
                                    window.open("https://docs.cloudreve.org/")
                                }
                            >
                                <ListItemIcon>
                                    <Description />
                                </ListItemIcon>
                                <ListItemText primary="文档" />
                                <ListItemIcon className={classes.iconRight}>
                                    <Launch />
                                </ListItemIcon>
                            </ListItem>
                            <ListItem
                                button
                                onClick={() =>
                                    window.open("https://forum.cloudreve.org")
                                }
                            >
                                <ListItemIcon>
                                    <Forum />
                                </ListItemIcon>
                                <ListItemText primary="讨论社区" />
                                <ListItemIcon className={classes.iconRight}>
                                    <Launch />
                                </ListItemIcon>
                            </ListItem>
                            <ListItem
                                button
                                onClick={() =>
                                    window.open(
                                        "https://t.me/cloudreve_official"
                                    )
                                }
                            >
                                <ListItemIcon>
                                    <Telegram />
                                </ListItemIcon>
                                <ListItemText primary="Telegram 群组" />
                                <ListItemIcon className={classes.iconRight}>
                                    <Launch />
                                </ListItemIcon>
                            </ListItem>
                            <ListItem
                                button
                                onClick={() =>
                                    window.open(
                                        "https://docs.cloudreve.org/use/pro/jie-shao"
                                    )
                                }
                            >
                                <ListItemIcon style={{ color: "#ff789d" }}>
                                    <Favorite />
                                </ListItemIcon>
                                <ListItemText primary="升级到捐助版" />
                                <ListItemIcon className={classes.iconRight}>
                                    <Launch />
                                </ListItemIcon>
                            </ListItem>
                        </List>
                    </div>
                </Paper>
            </Grid>
            <Grid item xs={12} md={8} lg={9}>
                <Paper className={classes.paper}>
                    <List>
                        {news &&
                            news.map((v) => (
                                <>
                                    <ListItem
                                        button
                                        alignItems="flex-start"
                                        onClick={() =>
                                            window.open(
                                                "https://forum.cloudreve.org/d/" +
                                                    v.id
                                            )
                                        }
                                    >
                                        <ListItemAvatar>
                                            <Avatar
                                                alt="Travis Howard"
                                                src={
                                                    newsUsers[
                                                        v.relationships
                                                            .startUser.data.id
                                                    ] &&
                                                    newsUsers[
                                                        v.relationships
                                                            .startUser.data.id
                                                    ].avatarUrl
                                                }
                                            />
                                        </ListItemAvatar>
                                        <ListItemText
                                            primary={v.attributes.title}
                                            secondary={
                                                <React.Fragment>
                                                    <Typography
                                                        component="span"
                                                        variant="body2"
                                                        className={
                                                            classes.inline
                                                        }
                                                        color="textPrimary"
                                                    >
                                                        {newsUsers[
                                                            v.relationships
                                                                .startUser.data
                                                                .id
                                                        ] &&
                                                            newsUsers[
                                                                v.relationships
                                                                    .startUser
                                                                    .data.id
                                                            ].username}{" "}
                                                    </Typography>
                                                    发表于{" "}
                                                    <TimeAgo
                                                        datetime={
                                                            v.attributes
                                                                .startTime
                                                        }
                                                        locale="zh_CN"
                                                    />
                                                </React.Fragment>
                                            }
                                        />
                                    </ListItem>
                                    <Divider />
                                </>
                            ))}
                    </List>
                </Paper>
            </Grid>
        </Grid>
    );
}
