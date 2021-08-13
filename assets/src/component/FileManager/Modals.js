import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";
import {
    closeAllModals,
    toggleSnackbar,
    setModalsLoading,
    refreshFileList,
    refreshStorage,
    openLoadingDialog,
} from "../../actions/index";
import PathSelector from "./PathSelector";
import API, { baseURL } from "../../middleware/Api";
import {
    withStyles,
    Button,
    TextField,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    DialogContentText,
    CircularProgress,
} from "@material-ui/core";
import Loading from "../Modals/Loading";
import CopyDialog from "../Modals/Copy";
import CreatShare from "../Modals/CreateShare";
import { withRouter } from "react-router-dom";
import pathHelper from "../../utils/page";
import DecompressDialog from "../Modals/Decompress";
import CompressDialog from "../Modals/Compress";

const styles = (theme) => ({
    wrapper: {
        margin: theme.spacing(1),
        position: "relative",
    },
    buttonProgress: {
        color: theme.palette.secondary.light,
        position: "absolute",
        top: "50%",
        left: "50%",
        marginTop: -12,
        marginLeft: -12,
    },
    contentFix: {
        padding: "10px 24px 0px 24px",
    },
});

const mapStateToProps = (state) => {
    return {
        path: state.navigator.path,
        selected: state.explorer.selected,
        modalsStatus: state.viewUpdate.modals,
        modalsLoading: state.viewUpdate.modalsLoading,
        dirList: state.explorer.dirList,
        fileList: state.explorer.fileList,
        dndSignale: state.explorer.dndSignal,
        dndTarget: state.explorer.dndTarget,
        dndSource: state.explorer.dndSource,
        loading: state.viewUpdate.modals.loading,
        loadingText: state.viewUpdate.modals.loadingText,
    };
};

const mapDispatchToProps = (dispatch) => {
    return {
        closeAllModals: () => {
            dispatch(closeAllModals());
        },
        toggleSnackbar: (vertical, horizontal, msg, color) => {
            dispatch(toggleSnackbar(vertical, horizontal, msg, color));
        },
        setModalsLoading: (status) => {
            dispatch(setModalsLoading(status));
        },
        refreshFileList: () => {
            dispatch(refreshFileList());
        },
        refreshStorage: () => {
            dispatch(refreshStorage());
        },
        openLoadingDialog: (text) => {
            dispatch(openLoadingDialog(text));
        },
    };
};

class ModalsCompoment extends Component {
    state = {
        newFolderName: "",
        newFileName: "",
        newName: "",
        selectedPath: "",
        selectedPathName: "",
        secretShare: false,
        sharePwd: "",
        shareUrl: "",
        downloadURL: "",
        remoteDownloadPathSelect: false,
        source: "",
        purchaseCallback: null,
    };

    handleInputChange = (e) => {
        this.setState({
            [e.target.id]: e.target.value,
        });
    };

    newNameSuffix = "";
    downloaded = false;

    UNSAFE_componentWillReceiveProps = (nextProps) => {
        if (this.props.dndSignale !== nextProps.dndSignale) {
            this.dragMove(nextProps.dndSource, nextProps.dndTarget);
            return;
        }
        if (this.props.loading !== nextProps.loading) {
            // 打包下载
            if (nextProps.loading === true) {
                if (nextProps.loadingText === "打包中...") {
                    if (
                        pathHelper.isSharePage(this.props.location.pathname) &&
                        this.props.share &&
                        this.props.share.score > 0
                    ) {
                        this.scoreHandler(this.archiveDownload);
                        return;
                    }
                    this.archiveDownload();
                } else if (nextProps.loadingText === "获取下载地址...") {
                    if (
                        pathHelper.isSharePage(this.props.location.pathname) &&
                        this.props.share &&
                        this.props.share.score > 0
                    ) {
                        this.scoreHandler(this.Download);
                        return;
                    }
                    this.Download();
                }
            }
            return;
        }
        if (this.props.modalsStatus.rename !== nextProps.modalsStatus.rename) {
            const name = nextProps.selected[0].name;
            this.setState({
                newName: name,
            });
            return;
        }
        if (
            this.props.modalsStatus.getSource !==
                nextProps.modalsStatus.getSource &&
            nextProps.modalsStatus.getSource === true
        ) {
            API.get("/file/source/" + this.props.selected[0].id)
                .then((response) => {
                    this.setState({
                        source: response.data.url,
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
        }
    };

    scoreHandler = (callback) => {
        callback();
    };

    Download = () => {
        let reqURL = "";
        if (this.props.selected[0].key) {
            const downloadPath =
                this.props.selected[0].path === "/"
                    ? this.props.selected[0].path + this.props.selected[0].name
                    : this.props.selected[0].path +
                      "/" +
                      this.props.selected[0].name;
            reqURL =
                "/share/download/" +
                this.props.selected[0].key +
                "?path=" +
                encodeURIComponent(downloadPath);
        } else {
            reqURL = "/file/download/" + this.props.selected[0].id;
        }

        API.put(reqURL)
            .then((response) => {
                window.location.assign(response.data);
                this.onClose();
                this.downloaded = true;
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.onClose();
            });
    };

    archiveDownload = () => {
        const dirs = [],
            items = [];
        this.props.selected.map((value) => {
            if (value.type === "dir") {
                dirs.push(value.id);
            } else {
                items.push(value.id);
            }
            return null;
        });

        let reqURL = "/file/archive";
        const postBody = {
            items: items,
            dirs: dirs,
        };
        if (pathHelper.isSharePage(this.props.location.pathname)) {
            reqURL = "/share/archive/" + window.shareInfo.key;
            postBody["path"] = this.props.selected[0].path;
        }

        API.post(reqURL, postBody)
            .then((response) => {
                if (response.rawData.code === 0) {
                    this.onClose();
                    window.location.assign(response.data);
                } else {
                    this.props.toggleSnackbar(
                        "top",
                        "right",
                        response.rawData.msg,
                        "warning"
                    );
                }
                this.onClose();
                this.props.refreshStorage();
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.onClose();
            });
    };

    submitRemove = (e) => {
        e.preventDefault();
        this.props.setModalsLoading(true);
        const dirs = [],
            items = [];
        // eslint-disable-next-line
        this.props.selected.map((value) => {
            if (value.type === "dir") {
                dirs.push(value.id);
            } else {
                items.push(value.id);
            }
        });
        API.delete("/object", {
            data: {
                items: items,
                dirs: dirs,
            },
        })
            .then((response) => {
                if (response.rawData.code === 0) {
                    this.onClose();
                    setTimeout(this.props.refreshFileList, 500);
                } else {
                    this.props.toggleSnackbar(
                        "top",
                        "right",
                        response.rawData.msg,
                        "warning"
                    );
                }
                this.props.setModalsLoading(false);
                this.props.refreshStorage();
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.props.setModalsLoading(false);
            });
    };

    submitMove = (e) => {
        if (e != null) {
            e.preventDefault();
        }
        this.props.setModalsLoading(true);
        const dirs = [],
            items = [];
        // eslint-disable-next-line
        this.props.selected.map((value) => {
            if (value.type === "dir") {
                dirs.push(value.id);
            } else {
                items.push(value.id);
            }
        });
        API.patch("/object", {
            action: "move",
            src_dir: this.props.selected[0].path,
            src: {
                dirs: dirs,
                items: items,
            },
            dst: this.DragSelectedPath
                ? this.DragSelectedPath
                : this.state.selectedPath === "//"
                ? "/"
                : this.state.selectedPath,
        })
            .then(() => {
                this.onClose();
                this.props.refreshFileList();
                this.props.setModalsLoading(false);
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.props.setModalsLoading(false);
            })
            .then(() => {
                this.props.closeAllModals();
            });
    };

    dragMove = (source, target) => {
        if (this.props.selected.length === 0) {
            this.props.selected[0] = source;
        }
        let doMove = true;

        // eslint-disable-next-line
        this.props.selected.map((value) => {
            // 根据ID过滤
            if (value.id === target.id && value.type === target.type) {
                doMove = false;
                // eslint-disable-next-line
                return;
            }
            // 根据路径过滤
            if (
                value.path ===
                target.path + (target.path === "/" ? "" : "/") + target.name
            ) {
                doMove = false;
                // eslint-disable-next-line
                return;
            }
        });
        if (doMove) {
            this.DragSelectedPath =
                target.path === "/"
                    ? target.path + target.name
                    : target.path + "/" + target.name;
            this.props.openLoadingDialog("处理中...");
            this.submitMove();
        }
    };

    submitRename = (e) => {
        e.preventDefault();
        this.props.setModalsLoading(true);
        const newName = this.state.newName;

        const src = {
            dirs: [],
            items: [],
        };

        if (this.props.selected[0].type === "dir") {
            src.dirs[0] = this.props.selected[0].id;
        } else {
            src.items[0] = this.props.selected[0].id;
        }

        // 检查重名
        if (
            this.props.dirList.findIndex((value) => {
                return value.name === newName;
            }) !== -1 ||
            this.props.fileList.findIndex((value) => {
                return value.name === newName;
            }) !== -1
        ) {
            this.props.toggleSnackbar(
                "top",
                "right",
                "新名称与已有文件重复",
                "warning"
            );
            this.props.setModalsLoading(false);
        } else {
            API.post("/object/rename", {
                action: "rename",
                src: src,
                new_name: newName,
            })
                .then(() => {
                    this.onClose();
                    this.props.refreshFileList();
                    this.props.setModalsLoading(false);
                })
                .catch((error) => {
                    this.props.toggleSnackbar(
                        "top",
                        "right",
                        error.message,
                        "error"
                    );
                    this.props.setModalsLoading(false);
                });
        }
    };

    submitCreateNewFolder = (e) => {
        e.preventDefault();
        this.props.setModalsLoading(true);
        if (
            this.props.dirList.findIndex((value) => {
                return value.name === this.state.newFolderName;
            }) !== -1
        ) {
            this.props.toggleSnackbar(
                "top",
                "right",
                "文件夹名称重复",
                "warning"
            );
            this.props.setModalsLoading(false);
        } else {
            API.put("/directory", {
                path:
                    (this.props.path === "/" ? "" : this.props.path) +
                    "/" +
                    this.state.newFolderName,
            })
                .then(() => {
                    this.onClose();
                    this.props.refreshFileList();
                    this.props.setModalsLoading(false);
                })
                .catch((error) => {
                    this.props.setModalsLoading(false);

                    this.props.toggleSnackbar(
                        "top",
                        "right",
                        error.message,
                        "error"
                    );
                });
        }
        //this.props.toggleSnackbar();
    };

    submitCreateNewFile = (e) => {
        e.preventDefault();
        this.props.setModalsLoading(true);
        if (
            this.props.dirList.findIndex((value) => {
                return value.name === this.state.newFileName;
            }) !== -1
        ) {
            this.props.toggleSnackbar(
                "top",
                "right",
                "文件名称重复",
                "warning"
            );
            this.props.setModalsLoading(false);
        } else {
            API.post("/file/create", {
                path:
                    (this.props.path === "/" ? "" : this.props.path) +
                    "/" +
                    this.state.newFileName,
            })
                .then(() => {
                    this.onClose();
                    this.props.refreshFileList();
                    this.props.setModalsLoading(false);
                })
                .catch((error) => {
                    this.props.setModalsLoading(false);

                    this.props.toggleSnackbar(
                        "top",
                        "right",
                        error.message,
                        "error"
                    );
                });
        }
        //this.props.toggleSnackbar();
    };

    submitTorrentDownload = (e) => {
        e.preventDefault();
        this.props.setModalsLoading(true);
        API.post("/aria2/torrent/" + this.props.selected[0].id, {
            dst:
                this.state.selectedPath === "//"
                    ? "/"
                    : this.state.selectedPath,
        })
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "任务已创建",
                    "success"
                );
                this.onClose();
                this.props.setModalsLoading(false);
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.props.setModalsLoading(false);
            });
    };

    submitDownload = (e) => {
        e.preventDefault();
        this.props.setModalsLoading(true);
        API.post("/aria2/url", {
            url: this.state.downloadURL,
            dst:
                this.state.selectedPath === "//"
                    ? "/"
                    : this.state.selectedPath,
        })
            .then(() => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    "任务已创建",
                    "success"
                );
                this.onClose();
                this.props.setModalsLoading(false);
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "error"
                );
                this.props.setModalsLoading(false);
            });
    };

    setMoveTarget = (folder) => {
        const path =
            folder.path === "/"
                ? folder.path + folder.name
                : folder.path + "/" + folder.name;
        this.setState({
            selectedPath: path,
            selectedPathName: folder.name,
        });
    };

    remoteDownloadNext = () => {
        this.props.closeAllModals();
        this.setState({
            remoteDownloadPathSelect: true,
        });
    };

    onClose = () => {
        this.setState({
            newFolderName: "",
            newFileName: "",
            newName: "",
            selectedPath: "",
            selectedPathName: "",
            secretShare: false,
            sharePwd: "",
            downloadURL: "",
            shareUrl: "",
            remoteDownloadPathSelect: false,
            source: "",
        });
        this.newNameSuffix = "";
        this.props.closeAllModals();
    };

    handleChange = (name) => (event) => {
        this.setState({ [name]: event.target.checked });
    };

    render() {
        const { classes } = this.props;

        return (
            <div>
                <Loading />
                <Dialog
                    open={this.props.modalsStatus.getSource}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                >
                    <DialogTitle id="form-dialog-title">
                        获取文件外链
                    </DialogTitle>

                    <DialogContent>
                        <form onSubmit={this.submitCreateNewFolder}>
                            <TextField
                                autoFocus
                                margin="dense"
                                id="newFolderName"
                                label="外链地址"
                                type="text"
                                value={this.state.source}
                                fullWidth
                            />
                        </form>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.onClose}>关闭</Button>
                    </DialogActions>
                </Dialog>
                <Dialog
                    open={this.props.modalsStatus.createNewFolder}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                >
                    <DialogTitle id="form-dialog-title">新建文件夹</DialogTitle>

                    <DialogContent>
                        <form onSubmit={this.submitCreateNewFolder}>
                            <TextField
                                autoFocus
                                margin="dense"
                                id="newFolderName"
                                label="文件夹名称"
                                type="text"
                                value={this.state.newFolderName}
                                onChange={(e) => this.handleInputChange(e)}
                                fullWidth
                            />
                        </form>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.onClose}>取消</Button>
                        <div className={classes.wrapper}>
                            <Button
                                onClick={this.submitCreateNewFolder}
                                color="primary"
                                disabled={
                                    this.state.newFolderName === "" ||
                                    this.props.modalsLoading
                                }
                            >
                                创建
                                {this.props.modalsLoading && (
                                    <CircularProgress
                                        size={24}
                                        className={classes.buttonProgress}
                                    />
                                )}
                            </Button>
                        </div>
                    </DialogActions>
                </Dialog>

                <Dialog
                    open={this.props.modalsStatus.createNewFile}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                >
                    <DialogTitle id="form-dialog-title">新建文件</DialogTitle>

                    <DialogContent>
                        <form onSubmit={this.submitCreateNewFile}>
                            <TextField
                                autoFocus
                                margin="dense"
                                id="newFileName"
                                label="文件名称"
                                type="text"
                                value={this.state.newFileName}
                                onChange={(e) => this.handleInputChange(e)}
                                fullWidth
                            />
                        </form>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.onClose}>取消</Button>
                        <div className={classes.wrapper}>
                            <Button
                                onClick={this.submitCreateNewFile}
                                color="primary"
                                disabled={
                                    this.state.newFileName === "" ||
                                    this.props.modalsLoading
                                }
                            >
                                创建
                                {this.props.modalsLoading && (
                                    <CircularProgress
                                        size={24}
                                        className={classes.buttonProgress}
                                    />
                                )}
                            </Button>
                        </div>
                    </DialogActions>
                </Dialog>

                <Dialog
                    open={this.props.modalsStatus.rename}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                    maxWidth="sm"
                    fullWidth={true}
                >
                    <DialogTitle id="form-dialog-title">重命名</DialogTitle>
                    <DialogContent>
                        <DialogContentText>
                            输入{" "}
                            <strong>
                                {this.props.selected.length === 1
                                    ? this.props.selected[0].name
                                    : ""}
                            </strong>{" "}
                            的新名称：
                        </DialogContentText>
                        <form onSubmit={this.submitRename}>
                            <TextField
                                autoFocus
                                margin="dense"
                                id="newName"
                                label="新名称"
                                type="text"
                                value={this.state.newName}
                                onChange={(e) => this.handleInputChange(e)}
                                fullWidth
                            />
                        </form>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.onClose}>取消</Button>
                        <div className={classes.wrapper}>
                            <Button
                                onClick={this.submitRename}
                                color="primary"
                                disabled={
                                    this.state.newName === "" ||
                                    this.props.modalsLoading
                                }
                            >
                                确定
                                {this.props.modalsLoading && (
                                    <CircularProgress
                                        size={24}
                                        className={classes.buttonProgress}
                                    />
                                )}
                            </Button>
                        </div>
                    </DialogActions>
                </Dialog>
                <CopyDialog
                    open={this.props.modalsStatus.copy}
                    onClose={this.onClose}
                    presentPath={this.props.path}
                    selected={this.props.selected}
                    modalsLoading={this.props.modalsLoading}
                />

                <Dialog
                    open={this.props.modalsStatus.move}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                >
                    <DialogTitle id="form-dialog-title">移动至</DialogTitle>
                    <PathSelector
                        presentPath={this.props.path}
                        selected={this.props.selected}
                        onSelect={this.setMoveTarget}
                    />

                    {this.state.selectedPath !== "" && (
                        <DialogContent className={classes.contentFix}>
                            <DialogContentText>
                                移动至{" "}
                                <strong>{this.state.selectedPathName}</strong>
                            </DialogContentText>
                        </DialogContent>
                    )}
                    <DialogActions>
                        <Button onClick={this.onClose}>取消</Button>
                        <div className={classes.wrapper}>
                            <Button
                                onClick={this.submitMove}
                                color="primary"
                                disabled={
                                    this.state.selectedPath === "" ||
                                    this.props.modalsLoading
                                }
                            >
                                确定
                                {this.props.modalsLoading && (
                                    <CircularProgress
                                        size={24}
                                        className={classes.buttonProgress}
                                    />
                                )}
                            </Button>
                        </div>
                    </DialogActions>
                </Dialog>
                <Dialog
                    open={this.props.modalsStatus.remove}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                >
                    <DialogTitle id="form-dialog-title">删除对象</DialogTitle>

                    <DialogContent>
                        <DialogContentText>
                            确定要删除
                            {this.props.selected.length === 1 && (
                                <strong> {this.props.selected[0].name} </strong>
                            )}
                            {this.props.selected.length > 1 && (
                                <span>
                                    这{this.props.selected.length}个对象
                                </span>
                            )}
                            吗？
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.onClose}>取消</Button>
                        <div className={classes.wrapper}>
                            <Button
                                onClick={this.submitRemove}
                                color="primary"
                                disabled={this.props.modalsLoading}
                            >
                                确定
                                {this.props.modalsLoading && (
                                    <CircularProgress
                                        size={24}
                                        className={classes.buttonProgress}
                                    />
                                )}
                            </Button>
                        </div>
                    </DialogActions>
                </Dialog>

                <CreatShare
                    open={this.props.modalsStatus.share}
                    onClose={this.onClose}
                    modalsLoading={this.props.modalsLoading}
                    setModalsLoading={this.props.setModalsLoading}
                    selected={this.props.selected}
                />

                <Dialog
                    open={this.props.modalsStatus.music}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                >
                    <DialogTitle id="form-dialog-title">音频播放</DialogTitle>

                    <DialogContent>
                        <DialogContentText>
                            {this.props.selected.length !== 0 && (
                                <audio
                                    controls
                                    src={
                                        pathHelper.isSharePage(
                                            this.props.location.pathname
                                        )
                                            ? baseURL +
                                              "/share/preview/" +
                                              this.props.selected[0].key +
                                              (this.props.selected[0].key
                                                  ? "?path=" +
                                                    encodeURIComponent(
                                                        this.props.selected[0]
                                                            .path === "/"
                                                            ? this.props
                                                                  .selected[0]
                                                                  .path +
                                                                  this.props
                                                                      .selected[0]
                                                                      .name
                                                            : this.props
                                                                  .selected[0]
                                                                  .path +
                                                                  "/" +
                                                                  this.props
                                                                      .selected[0]
                                                                      .name
                                                    )
                                                  : "")
                                            : baseURL +
                                              "/file/preview/" +
                                              this.props.selected[0].id
                                    }
                                />
                            )}
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.onClose}>关闭</Button>
                    </DialogActions>
                </Dialog>
                <Dialog
                    open={this.props.modalsStatus.remoteDownload}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                    fullWidth
                >
                    <DialogTitle id="form-dialog-title">
                        新建离线下载任务
                    </DialogTitle>

                    <DialogContent>
                        <DialogContentText>
                            <TextField
                                label="文件地址"
                                autoFocus
                                fullWidth
                                id="downloadURL"
                                onChange={this.handleInputChange}
                                placeholder="输入文件下载地址，支持 HTTP(s)/FTP/磁力链"
                            />
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.onClose}>关闭</Button>
                        <Button
                            onClick={this.remoteDownloadNext}
                            color="primary"
                            disabled={
                                this.props.modalsLoading ||
                                this.state.downloadURL === ""
                            }
                        >
                            下一步
                        </Button>
                    </DialogActions>
                </Dialog>
                <Dialog
                    open={this.state.remoteDownloadPathSelect}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                >
                    <DialogTitle id="form-dialog-title">
                        选择存储位置
                    </DialogTitle>
                    <PathSelector
                        presentPath={this.props.path}
                        selected={this.props.selected}
                        onSelect={this.setMoveTarget}
                    />

                    {this.state.selectedPath !== "" && (
                        <DialogContent className={classes.contentFix}>
                            <DialogContentText>
                                下载至{" "}
                                <strong>{this.state.selectedPathName}</strong>
                            </DialogContentText>
                        </DialogContent>
                    )}
                    <DialogActions>
                        <Button onClick={this.onClose}>取消</Button>
                        <div className={classes.wrapper}>
                            <Button
                                onClick={this.submitDownload}
                                color="primary"
                                disabled={
                                    this.state.selectedPath === "" ||
                                    this.props.modalsLoading
                                }
                            >
                                创建任务
                                {this.props.modalsLoading && (
                                    <CircularProgress
                                        size={24}
                                        className={classes.buttonProgress}
                                    />
                                )}
                            </Button>
                        </div>
                    </DialogActions>
                </Dialog>
                <Dialog
                    open={this.props.modalsStatus.torrentDownload}
                    onClose={this.onClose}
                    aria-labelledby="form-dialog-title"
                >
                    <DialogTitle id="form-dialog-title">
                        选择存储位置
                    </DialogTitle>
                    <PathSelector
                        presentPath={this.props.path}
                        selected={this.props.selected}
                        onSelect={this.setMoveTarget}
                    />

                    {this.state.selectedPath !== "" && (
                        <DialogContent className={classes.contentFix}>
                            <DialogContentText>
                                下载至{" "}
                                <strong>{this.state.selectedPathName}</strong>
                            </DialogContentText>
                        </DialogContent>
                    )}
                    <DialogActions>
                        <Button onClick={this.onClose}>取消</Button>
                        <div className={classes.wrapper}>
                            <Button
                                onClick={this.submitTorrentDownload}
                                color="primary"
                                disabled={
                                    this.state.selectedPath === "" ||
                                    this.props.modalsLoading
                                }
                            >
                                创建任务
                                {this.props.modalsLoading && (
                                    <CircularProgress
                                        size={24}
                                        className={classes.buttonProgress}
                                    />
                                )}
                            </Button>
                        </div>
                    </DialogActions>
                </Dialog>

                <DecompressDialog
                    open={this.props.modalsStatus.decompress}
                    onClose={this.onClose}
                    presentPath={this.props.path}
                    selected={this.props.selected}
                    modalsLoading={this.props.modalsLoading}
                />
                <CompressDialog
                    open={this.props.modalsStatus.compress}
                    onClose={this.onClose}
                    presentPath={this.props.path}
                    selected={this.props.selected}
                    modalsLoading={this.props.modalsLoading}
                />
            </div>
        );
    }
}

ModalsCompoment.propTypes = {
    classes: PropTypes.object.isRequired,
};

const Modals = connect(
    mapStateToProps,
    mapDispatchToProps
)(withStyles(styles)(withRouter(ModalsCompoment)));

export default Modals;
