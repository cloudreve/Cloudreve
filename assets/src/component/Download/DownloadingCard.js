import {
    Card,
    CardContent,
    darken,
    IconButton,
    lighten,
    LinearProgress,
    makeStyles,
    Typography,
    useTheme,
} from "@material-ui/core";
import Badge from "@material-ui/core/Badge";
import Button from "@material-ui/core/Button";
import Divider from "@material-ui/core/Divider";
import MuiExpansionPanel from "@material-ui/core/ExpansionPanel";
import MuiExpansionPanelDetails from "@material-ui/core/ExpansionPanelDetails";
import MuiExpansionPanelSummary from "@material-ui/core/ExpansionPanelSummary";
import Grid from "@material-ui/core/Grid";
import withStyles from "@material-ui/core/styles/withStyles";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Tooltip from "@material-ui/core/Tooltip";
import { ExpandMore, HighlightOff } from "@material-ui/icons";
import PermMediaIcon from "@material-ui/icons/PermMedia";
import classNames from "classnames";
import React, { useCallback, useEffect } from "react";
import { useDispatch } from "react-redux";
import TimeAgo from "timeago-react";
import { toggleSnackbar } from "../../actions";
import API from "../../middleware/Api";
import { hex2bin, sizeToString } from "../../utils";
import TypeIcon from "../FileManager/TypeIcon";
import SelectFileDialog from "../Modals/SelectFile";
import { useHistory } from "react-router";
const ExpansionPanel = withStyles({
    root: {
        maxWidth: "100%",
        boxShadow: "none",
        "&:not(:last-child)": {
            borderBottom: 0,
        },
        "&:before": {
            display: "none",
        },
        "&$expanded": {},
    },
    expanded: {},
})(MuiExpansionPanel);

const ExpansionPanelSummary = withStyles({
    root: {
        minHeight: 0,
        padding: 0,

        "&$expanded": {
            minHeight: 56,
        },
    },
    content: {
        maxWidth: "100%",
        margin: 0,
        display: "flex",
        "&$expanded": {
            margin: "0",
        },
    },
    expanded: {},
})(MuiExpansionPanelSummary);

const ExpansionPanelDetails = withStyles((theme) => ({
    root: {
        display: "block",
        padding: theme.spacing(0),
    },
}))(MuiExpansionPanelDetails);

const useStyles = makeStyles((theme) => ({
    card: {
        marginTop: "20px",
        justifyContent: "space-between",
    },
    iconContainer: {
        width: "90px",
        height: "96px",
        padding: " 35px 29px 29px 29px",
        paddingLeft: "35px",
        [theme.breakpoints.down("sm")]: {
            display: "none",
        },
    },
    content: {
        width: "100%",
        minWidth: 0,
        [theme.breakpoints.up("sm")]: {
            borderInlineStart: "1px " + theme.palette.divider + " solid",
        },
    },
    contentSide: {
        minWidth: 0,
        paddingTop: "24px",
        paddingRight: "28px",
        [theme.breakpoints.down("sm")]: {
            display: "none",
        },
    },
    iconBig: {
        fontSize: "30px",
    },
    iconMultiple: {
        fontSize: "30px",
        color: "#607D8B",
    },
    progress: {
        marginTop: 8,
        marginBottom: 4,
    },
    expand: {
        transition: ".15s transform ease-in-out",
    },
    expanded: {
        transform: "rotate(180deg)",
    },
    subFileName: {
        display: "flex",
    },
    subFileIcon: {
        marginRight: "20px",
    },
    scroll: {
        overflowY: "auto",
    },
    action: {
        padding: theme.spacing(2),
        textAlign: "right",
    },
    actionButton: {
        marginLeft: theme.spacing(1),
    },
    info: {
        padding: theme.spacing(2),
    },
    infoTitle: {
        fontWeight: 700,
    },
    infoValue: {
        color: theme.palette.text.secondary,
    },
    bitmap: {
        width: "100%",
        height: "50px",
        backgroundColor: theme.palette.background.default,
    },
}));

export default function DownloadingCard(props) {
    const canvasRef = React.createRef();
    const classes = useStyles();
    const theme = useTheme();
    const history = useHistory();

    const [expanded, setExpanded] = React.useState("");
    const [task, setTask] = React.useState(props.task);
    const [loading, setLoading] = React.useState(false);
    const [selectDialogOpen, setSelectDialogOpen] = React.useState(false);
    const [selectFileOption, setSelectFileOption] = React.useState([]);

    const handleChange = (panel) => (event, newExpanded) => {
        setExpanded(newExpanded ? panel : false);
    };

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    useEffect(() => {
        setTask(props.task);
    }, [props.task]);

    useEffect(() => {
        if (task.info.bitfield === "") {
            return;
        }
        let result = "";
        task.info.bitfield.match(/.{1,2}/g).forEach((str) => {
            result += hex2bin(str);
        });
        const canvas = canvasRef.current;
        const context = canvas.getContext("2d");
        context.clearRect(0, 0, canvas.width, canvas.height);
        context.strokeStyle = theme.palette.primary.main;
        for (let i = 0; i < canvas.width; i++) {
            let bit =
                result[
                    Math.round(((i + 1) / canvas.width) * task.info.numPieces)
                ];
            bit = bit ? bit : result.slice(-1);
            if (bit === "1") {
                context.beginPath();
                context.moveTo(i, 0);
                context.lineTo(i, canvas.height);
                context.stroke();
            }
        }
        // eslint-disable-next-line
    }, [task.info.bitfield, task.info.numPieces, theme]);

    const getPercent = (completed, total) => {
        if (total === 0) {
            return 0;
        }
        return (completed / total) * 100;
    };

    const activeFiles = useCallback(() => {
        return task.info.files.filter((v) => v.selected === "true");
    }, [task.info.files]);

    const deleteFile = (index) => {
        setLoading(true);
        const current = activeFiles();
        const newIndex = [];
        const newFiles = [];
        // eslint-disable-next-line
        current.map((v) => {
            if (v.index !== index && v.selected) {
                newIndex.push(parseInt(v.index));
                newFiles.push({
                    ...v,
                    selected: "true",
                });
            } else {
                newFiles.push({
                    ...v,
                    selected: "false",
                });
            }
        });
        API.put("/aria2/select/" + task.info.gid, {
            indexes: newIndex,
        })
            .then(() => {
                setTask({
                    ...task,
                    info: {
                        ...task.info,
                        files: newFiles,
                    },
                });
                ToggleSnackbar("top", "right", "文件已删除", "success");
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
    };

    const getDownloadName = useCallback(() => {
        if (task.info.bittorrent.info.name !== "") {
            return task.info.bittorrent.info.name;
        }
        return task.name === "." ? "[未知]" : task.name;
    }, [task]);

    const getIcon = useCallback(() => {
        if (task.info.bittorrent.mode === "multi") {
            return (
                <Badge badgeContent={activeFiles().length} color="secondary">
                    <PermMediaIcon className={classes.iconMultiple} />
                </Badge>
            );
        } else {
            return (
                <TypeIcon
                    className={classes.iconBig}
                    fileName={getDownloadName(task)}
                />
            );
        }
        // eslint-disable-next-line
    }, [task, classes]);

    const cancel = () => {
        setLoading(true);
        API.delete("/aria2/task/" + task.info.gid)
            .then(() => {
                ToggleSnackbar(
                    "top",
                    "right",
                    "任务已取消，状态会在稍后更新",
                    "success"
                );
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
    };

    const changeSelectedFile = (fileIndex) => {
        setLoading(true);
        API.put("/aria2/select/" + task.info.gid, {
            indexes: fileIndex,
        })
            .then(() => {
                ToggleSnackbar(
                    "top",
                    "right",
                    "操作成功，状态会在稍后更新",
                    "success"
                );
                setSelectDialogOpen(false);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
    };

    return (
        <Card className={classes.card}>
            <SelectFileDialog
                open={selectDialogOpen}
                onClose={() => setSelectDialogOpen(false)}
                modalsLoading={loading}
                files={selectFileOption}
                onSubmit={changeSelectedFile}
            />
            <ExpansionPanel
                square
                expanded={expanded === task.info.gid}
                onChange={handleChange(task.info.gid)}
            >
                <ExpansionPanelSummary
                    aria-controls="panel1d-content"
                    id="panel1d-header"
                >
                    <div className={classes.iconContainer}>{getIcon()}</div>
                    <CardContent className={classes.content}>
                        <Typography color="primary" noWrap>
                            <Tooltip title={getDownloadName()}>
                                <span>{getDownloadName()}</span>
                            </Tooltip>
                        </Typography>
                        <LinearProgress
                            color="secondary"
                            variant="determinate"
                            className={classes.progress}
                            value={getPercent(task.downloaded, task.total)}
                        />
                        <Typography
                            variant="body2"
                            color="textSecondary"
                            noWrap
                        >
                            {task.total > 0 && (
                                <span>
                                    {getPercent(
                                        task.downloaded,
                                        task.total
                                    ).toFixed(2)}
                                    % -{" "}
                                    {task.downloaded === 0
                                        ? "0Bytes"
                                        : sizeToString(task.downloaded)}
                                    /
                                    {task.total === 0
                                        ? "0Bytes"
                                        : sizeToString(task.total)}{" "}
                                    -{" "}
                                    {task.speed === "0"
                                        ? "0B/s"
                                        : sizeToString(task.speed) + "/s"}
                                </span>
                            )}
                            {task.total === 0 && <span> - </span>}
                        </Typography>
                    </CardContent>
                    <CardContent className={classes.contentSide}>
                        <IconButton>
                            <ExpandMore
                                className={classNames(
                                    {
                                        [classes.expanded]:
                                            expanded === task.info.gid,
                                    },
                                    classes.expand
                                )}
                            />
                        </IconButton>
                    </CardContent>
                </ExpansionPanelSummary>
                <ExpansionPanelDetails>
                    <Divider />
                    {task.info.bittorrent.mode === "multi" && (
                        <div className={classes.scroll}>
                            <Table size="small">
                                <TableBody>
                                    {activeFiles().map((value) => {
                                        return (
                                            <TableRow
                                                key={value.index}
                                                style={{
                                                    background:
                                                        "linear-gradient(to right, " +
                                                        (theme.palette.type ===
                                                        "dark"
                                                            ? darken(
                                                                  theme.palette
                                                                      .primary
                                                                      .main,
                                                                  0.4
                                                              )
                                                            : lighten(
                                                                  theme.palette
                                                                      .primary
                                                                      .main,
                                                                  0.85
                                                              )) +
                                                        " 0%," +
                                                        (theme.palette.type ===
                                                        "dark"
                                                            ? darken(
                                                                  theme.palette
                                                                      .primary
                                                                      .main,
                                                                  0.4
                                                              )
                                                            : lighten(
                                                                  theme.palette
                                                                      .primary
                                                                      .main,
                                                                  0.85
                                                              )) +
                                                        " " +
                                                        getPercent(
                                                            value.completedLength,
                                                            value.length
                                                        ).toFixed(0) +
                                                        "%," +
                                                        theme.palette.background
                                                            .paper +
                                                        " " +
                                                        getPercent(
                                                            value.completedLength,
                                                            value.length
                                                        ).toFixed(0) +
                                                        "%," +
                                                        theme.palette.background
                                                            .paper +
                                                        " 100%)",
                                                }}
                                            >
                                                <TableCell
                                                    component="th"
                                                    scope="row"
                                                >
                                                    <Typography
                                                        className={
                                                            classes.subFileName
                                                        }
                                                    >
                                                        <TypeIcon
                                                            className={
                                                                classes.subFileIcon
                                                            }
                                                            fileName={
                                                                value.path
                                                            }
                                                        />
                                                        {value.path}
                                                    </Typography>
                                                </TableCell>
                                                <TableCell
                                                    component="th"
                                                    scope="row"
                                                >
                                                    <Typography noWrap>
                                                        {" "}
                                                        {sizeToString(
                                                            value.length
                                                        )}
                                                    </Typography>
                                                </TableCell>
                                                <TableCell
                                                    component="th"
                                                    scope="row"
                                                >
                                                    <Typography noWrap>
                                                        {getPercent(
                                                            value.completedLength,
                                                            value.length
                                                        ).toFixed(2)}
                                                        %
                                                    </Typography>
                                                </TableCell>
                                                <TableCell>
                                                    <Tooltip title="删除此文件">
                                                        <IconButton
                                                            onClick={() =>
                                                                deleteFile(
                                                                    value.index
                                                                )
                                                            }
                                                            disabled={loading}
                                                            size={"small"}
                                                        >
                                                            <HighlightOff />
                                                        </IconButton>
                                                    </Tooltip>
                                                </TableCell>
                                            </TableRow>
                                        );
                                    })}
                                </TableBody>
                            </Table>
                        </div>
                    )}

                    <div className={classes.action}>
                        <Button
                            className={classes.actionButton}
                            variant="outlined"
                            color="secondary"
                            onClick={() =>
                                history.push(
                                    "/#/home?path=" +
                                        encodeURIComponent(task.dst)
                                )
                            }
                        >
                            打开存放目录
                        </Button>
                        {task.info.bittorrent.mode === "multi" && (
                            <Button
                                className={classes.actionButton}
                                variant="outlined"
                                color="secondary"
                                disabled={loading}
                                onClick={() => {
                                    setSelectDialogOpen(true);
                                    setSelectFileOption([
                                        ...props.task.info.files,
                                    ]);
                                }}
                            >
                                选择要下载的文件
                            </Button>
                        )}
                        <Button
                            className={classes.actionButton}
                            onClick={cancel}
                            variant="contained"
                            color="secondary"
                            disabled={loading}
                        >
                            取消任务
                        </Button>
                    </div>
                    <Divider />
                    <div className={classes.info}>
                        {task.info.bitfield !== "" && (
                            <canvas
                                width={"700"}
                                height={"100"}
                                ref={canvasRef}
                                className={classes.bitmap}
                            />
                        )}

                        <Grid container>
                            <Grid container xs={12} sm={4}>
                                <Grid item xs={4} className={classes.infoTitle}>
                                    更新于：
                                </Grid>
                                <Grid item xs={8} className={classes.infoValue}>
                                    <TimeAgo
                                        datetime={task.update}
                                        locale="zh_CN"
                                    />
                                </Grid>
                            </Grid>
                            <Grid container xs={12} sm={4}>
                                <Grid item xs={4} className={classes.infoTitle}>
                                    上传大小：
                                </Grid>
                                <Grid item xs={8} className={classes.infoValue}>
                                    {sizeToString(task.info.uploadLength)}
                                </Grid>
                            </Grid>
                            <Grid container xs={12} sm={4}>
                                <Grid item xs={4} className={classes.infoTitle}>
                                    上传速度：
                                </Grid>
                                <Grid item xs={8} className={classes.infoValue}>
                                    {sizeToString(task.info.uploadSpeed)} / s
                                </Grid>
                            </Grid>
                            {task.info.bittorrent.mode !== "" && (
                                <>
                                    <Grid container xs={12} sm={8}>
                                        <Grid
                                            item
                                            sm={2}
                                            xs={4}
                                            className={classes.infoTitle}
                                        >
                                            InfoHash：
                                        </Grid>
                                        <Grid
                                            item
                                            sm={10}
                                            xs={8}
                                            style={{
                                                wordBreak: "break-all",
                                            }}
                                            className={classes.infoValue}
                                        >
                                            {task.info.infoHash}
                                        </Grid>
                                    </Grid>
                                    <Grid container xs={12} sm={4}>
                                        <Grid
                                            item
                                            xs={4}
                                            className={classes.infoTitle}
                                        >
                                            做种者：
                                        </Grid>
                                        <Grid
                                            item
                                            xs={8}
                                            className={classes.infoValue}
                                        >
                                            {task.info.numSeeders}
                                        </Grid>
                                    </Grid>
                                    <Grid container xs={12} sm={4}>
                                        <Grid
                                            item
                                            xs={4}
                                            className={classes.infoTitle}
                                        >
                                            做种中：
                                        </Grid>
                                        <Grid
                                            item
                                            xs={8}
                                            className={classes.infoValue}
                                        >
                                            {task.info.seeder === "true"
                                                ? "是"
                                                : "否"}
                                        </Grid>
                                    </Grid>
                                </>
                            )}
                            <Grid container xs={12} sm={4}>
                                <Grid item xs={4} className={classes.infoTitle}>
                                    分片大小：
                                </Grid>
                                <Grid item xs={8} className={classes.infoValue}>
                                    {sizeToString(task.info.pieceLength)}
                                </Grid>
                            </Grid>
                            <Grid container xs={12} sm={4}>
                                <Grid item xs={4} className={classes.infoTitle}>
                                    分片数量：
                                </Grid>
                                <Grid item xs={8} className={classes.infoValue}>
                                    {task.info.numPieces}
                                </Grid>
                            </Grid>
                        </Grid>
                    </div>
                </ExpansionPanelDetails>
            </ExpansionPanel>
        </Card>
    );
}
