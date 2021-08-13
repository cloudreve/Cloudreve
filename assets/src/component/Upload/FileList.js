import React, { Component } from "react";
import CloseIcon from "@material-ui/icons/Close";
import AddIcon from "@material-ui/icons/AddCircleOutline";
import DeleteIcon from "@material-ui/icons/Delete";
import RefreshIcon from "@material-ui/icons/Refresh";
import { isWidthDown } from "@material-ui/core/withWidth";
import { darken, lighten } from "@material-ui/core/styles/colorManipulator";
import {
    withStyles,
    Dialog,
    ListItemText,
    ListItem,
    List,
    Divider,
    AppBar,
    Toolbar,
    IconButton,
    Typography,
    Slide,
    ListItemSecondaryAction,
    withWidth,
    DialogContent,
    Tooltip,
} from "@material-ui/core";
import TypeIcon from "../FileManager/TypeIcon";
import { withTheme } from "@material-ui/core/styles";

const styles = (theme) => ({
    appBar: {
        position: "relative",
    },
    flex: {
        flex: 1,
    },
    progressBar: {
        marginTop: 5,
    },
    minHight: {
        [theme.breakpoints.up("sm")]: {
            minWidth: 500,
        },
        padding: 0,
    },
    dialogContent: {
        padding: 0,
    },
    successStatus: {
        color: "#4caf50",
    },
    errorStatus: {
        color: "#ff5722",
        wordBreak: "break-all",
    },
    listAction: {
        marginLeft: 20,
        marginRight: 20,
    },
    delete: {
        zIndex: 9,
    },
    progressContainer: {
        position: "relative",
    },
    progressContent: {
        position: "relative",
        zIndex: 9,
    },
    progress: {
        transition: "width .4s linear",
        zIndex: 1,
        height: "100%",
        position: "absolute",
        left: 0,
        top: 0,
    },
    fileName: {
        wordBreak: "break-all",
    },
});
class FileList extends Component {
    state = {
        open: false,
        files: [],
    };

    //入队
    enQueue(files) {
        this.setState({
            files: [...this.state.files, ...files],
        });
    }

    deQueue(file) {
        const filesNow = [...this.state.files];
        const fileID = filesNow.findIndex((f) => {
            return f.id === file.id;
        });
        if (fileID !== -1) {
            filesNow.splice(fileID, 1);
            this.setState({
                files: filesNow,
                open: filesNow.length !== 0,
            });
        }
    }

    updateStatus(file) {
        const filesNow = [...this.state.files];
        const fileID = filesNow.findIndex((f) => {
            return f.id === file.id;
        });
        if (!file.errMsg || file.ignoreMsg) {
            if (filesNow[fileID] && !filesNow[fileID].errMsg) {
                filesNow[fileID] = file;
                this.setState({
                    files: filesNow,
                });
            }
        } else {
            file.ignoreMsg = true;
        }
    }

    setComplete(file) {
        const filesNow = [...this.state.files];
        const fileID = filesNow.findIndex((f) => {
            return f.id === file.id;
        });
        if (fileID !== -1) {
            if (filesNow[fileID].status !== 4) {
                filesNow[fileID].status = 5;
                this.setState({
                    files: filesNow,
                });
            }
        }
    }

    setError(file, errMsg) {
        const filesNow = [...this.state.files];
        const fileID = filesNow.findIndex((f) => {
            return f.id === file.id;
        });
        if (fileID !== -1) {
            filesNow[fileID].status = 4;
            filesNow[fileID].errMsg = errMsg;
        } else {
            file.status = 4;
            file.errMsg = errMsg;
            filesNow.push(file);
        }
        this.setState({
            files: filesNow,
        });
    }

    Transition(props) {
        return <Slide direction="up" {...props} />;
    }
    openFileList = () => {
        if (!this.state.open) {
            this.setState({ open: true });
        }
    };

    cancelUpload = (file) => {
        this.props.cancelUpload(file);
        // this.deQueue(file);
    };

    handleClose = () => {
        this.setState({ open: false });
    };

    addNewFile = () => {
        document.getElementsByClassName("uploadFileForm")[0].click();
    };

    getProgressBackground = () => {
        return this.props.theme.palette.type === "light"
            ? lighten(this.props.theme.palette.primary.main, 0.8)
            : darken(this.props.theme.palette.background.paper, 0.2);
    };

    render() {
        const { classes } = this.props;
        const { width } = this.props;

        this.props.inRef({
            openFileList: this.openFileList.bind(this),
            enQueue: this.enQueue.bind(this),
            updateStatus: this.updateStatus.bind(this),
            setComplete: this.setComplete.bind(this),
            setError: this.setError.bind(this),
        });

        return (
            <Dialog
                fullScreen={isWidthDown("sm", width)}
                open={this.state.open}
                onClose={this.handleClose}
                TransitionComponent={this.Transition}
            >
                <AppBar className={classes.appBar}>
                    <Toolbar>
                        <IconButton
                            color="inherit"
                            onClick={this.handleClose}
                            aria-label="Close"
                        >
                            <CloseIcon />
                        </IconButton>
                        <Typography
                            variant="h6"
                            color="inherit"
                            className={classes.flex}
                        >
                            上传队列
                        </Typography>
                        <IconButton color="inherit" onClick={this.addNewFile}>
                            <AddIcon />
                        </IconButton>
                    </Toolbar>
                </AppBar>
                <DialogContent className={classes.dialogContent}>
                    <List className={classes.minHight}>
                        {this.state.files.map((item, i) => (
                            <div key={i} className={classes.progressContainer}>
                                {item.status === 2 && (
                                    <div
                                        style={{
                                            backgroundColor: this.getProgressBackground(),
                                            width: item.percent + "%",
                                        }}
                                        className={classes.progress}
                                    />
                                )}
                                <ListItem
                                    className={classes.progressContent}
                                    button
                                >
                                    <TypeIcon fileName={item.name} isUpload />
                                    {item.status === 1 && (
                                        <ListItemText
                                            className={classes.listAction}
                                            primary={
                                                <span
                                                    className={classes.fileName}
                                                >
                                                    {item.name}
                                                </span>
                                            }
                                            secondary={<div>排队中...</div>}
                                        />
                                    )}
                                    {item.status === 2 && (
                                        <ListItemText
                                            className={classes.listAction}
                                            primary={
                                                <span
                                                    className={classes.fileName}
                                                >
                                                    {item.name}
                                                </span>
                                            }
                                            secondary={
                                                <div>
                                                    {item.percent <= 99 && (
                                                        <>
                                                            {window.plupload
                                                                .formatSize(
                                                                    item.speed
                                                                )
                                                                .toUpperCase()}
                                                            /s 已上传{" "}
                                                            {window.plupload
                                                                .formatSize(
                                                                    item.loaded
                                                                )
                                                                .toUpperCase()}{" "}
                                                            , 共{" "}
                                                            {window.plupload
                                                                .formatSize(
                                                                    item.size
                                                                )
                                                                .toUpperCase()}{" "}
                                                            - {item.percent}%{" "}
                                                        </>
                                                    )}
                                                    {item.percent > 99 && (
                                                        <div>处理中...</div>
                                                    )}
                                                </div>
                                            }
                                        />
                                    )}
                                    {item.status === 3 && (
                                        <ListItemText
                                            className={classes.listAction}
                                            primary={
                                                <span
                                                    className={classes.fileName}
                                                >
                                                    {item.name}
                                                </span>
                                            }
                                            secondary={item.status}
                                        />
                                    )}
                                    {item.status === 4 && (
                                        <ListItemText
                                            className={classes.listAction}
                                            primary={
                                                <span
                                                    className={classes.fileName}
                                                >
                                                    {item.name}
                                                </span>
                                            }
                                            secondary={
                                                <div
                                                    className={
                                                        classes.errorStatus
                                                    }
                                                >
                                                    {item.errMsg}
                                                    <br />
                                                </div>
                                            }
                                        />
                                    )}
                                    {item.status === 5 && (
                                        <ListItemText
                                            className={classes.listAction}
                                            primary={
                                                <span
                                                    className={classes.fileName}
                                                >
                                                    {item.name}
                                                </span>
                                            }
                                            secondary={
                                                <div
                                                    className={
                                                        classes.successStatus
                                                    }
                                                >
                                                    已完成
                                                    <br />
                                                </div>
                                            }
                                        />
                                    )}
                                    <ListItemSecondaryAction
                                        className={classes.delete}
                                    >
                                        {item.status !== 4 && (
                                            <IconButton
                                                aria-label="Delete"
                                                onClick={() =>
                                                    this.cancelUpload(item)
                                                }
                                            >
                                                <DeleteIcon />
                                            </IconButton>
                                        )}
                                        {item.status === 4 && (
                                            <Tooltip title={"重试"}>
                                                <IconButton
                                                    aria-label="Delete"
                                                    onClick={() =>
                                                        this.reQueue(item)
                                                    }
                                                >
                                                    <RefreshIcon />
                                                </IconButton>
                                            </Tooltip>
                                        )}
                                    </ListItemSecondaryAction>
                                </ListItem>
                                <Divider />
                            </div>
                        ))}
                    </List>
                </DialogContent>
            </Dialog>
        );
    }
}
FileList.propTypes = {};

export default withStyles(styles)(withWidth()(withTheme(FileList)));
