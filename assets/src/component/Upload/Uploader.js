import React, { Component } from "react";
import uploaderLoader from "../../loader";
import { connect } from "react-redux";
import { refreshFileList, refreshStorage, toggleSnackbar } from "../../actions";
import FileList from "./FileList.js";
import Auth from "../../middleware/Auth";
import UploadButton from "../Dial/Create.js";
import { basename, pathJoin } from "../../utils";

let loaded = false;

const mapStateToProps = (state) => {
    return {
        path: state.navigator.path,
        keywords: state.explorer.keywords,
    };
};

const mapDispatchToProps = (dispatch) => {
    return {
        refreshFileList: () => {
            dispatch(refreshFileList());
        },
        refreshStorage: () => {
            dispatch(refreshStorage());
        },
        toggleSnackbar: (vertical, horizontal, msg, color) => {
            dispatch(toggleSnackbar(vertical, horizontal, msg, color));
        },
    };
};

class UploaderComponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            queued: 0,
        };
    }

    setRef(val) {
        window.fileList = val;
    }

    cancelUpload(file) {
        this.uploader.removeFile(file);
    }

    reQueue = (file) => {
        this.uploader.addFile(file.getSource());
        this.props.toggleSnackbar(
            "top",
            "right",
            "文件已经重新加入上传队列",
            "info"
        );
    };

    getChunkSize(policyType) {
        if (policyType === "qiniu") {
            return 4 * 1024 * 1024;
        }
        if (policyType === "onedrive") {
            return 100 * 1024 * 1024;
        }
        return 0;
    }

    fileAdd = (up, files) => {
        const path = window.currntPath ? window.currntPath : this.props.path;
        if (
            this.props.keywords === "" &&
            window.location.pathname.toLowerCase().startsWith("/home")
        ) {
            window.fileList["openFileList"]();
            const enqueFiles = files
                // 不上传Mac下的布局文件 .DS_Store
                .filter((file) => {
                    const isDsStore = file.name.toLowerCase() === ".ds_store";
                    if (isDsStore) {
                        up.removeFile(file);
                    }
                    return !isDsStore;
                })
                .map((file) => {
                    const source = file.getSource();
                    if (source.relativePath && source.relativePath !== "") {
                        file.path = basename(
                            pathJoin([path, source.relativePath])
                        );
                        window.pathCache[file.id] = basename(
                            pathJoin([path, source.relativePath])
                        );
                    } else {
                        window.pathCache[file.id] = path;
                        file.path = path;
                    }
                    return file;
                });
            window.fileList["enQueue"](enqueFiles);
        } else {
            window.plupload.each(files, (files) => {
                up.removeFile(files);
            });
        }
    };

    UNSAFE_componentWillReceiveProps({ isScriptLoaded, isScriptLoadSucceed }) {
        if (isScriptLoaded && !this.props.isScriptLoaded) {
            // load finished
            if (isScriptLoadSucceed) {
                if (loaded) {
                    return;
                }
                loaded = true;
                const user = Auth.GetUser();
                this.uploader = window.Qiniu.uploader({
                    runtimes: "html5",
                    browse_button: ["pickfiles", "pickfolder"],
                    container: "container",
                    drop_element: "container",
                    max_file_size:
                        user.policy.maxSize === "0.00mb"
                            ? 0
                            : user.policy.maxSize,
                    dragdrop: true,
                    chunk_size: this.getChunkSize(user.policy.saveType),
                    filters: {
                        mime_types:
                            user.policy.allowedType === null ||
                            user.policy.allowedType.length === 0
                                ? []
                                : [
                                      {
                                          title: "files",
                                          extensions: user.policy.allowedType.join(
                                              ","
                                          ),
                                      },
                                  ],
                    },
                    // iOS不能多选？
                    multi_selection: true,
                    uptoken_url: "/api/v3/file/upload/credential",
                    uptoken: user.policy.saveType === "local" ? "token" : null,
                    domain: "s",
                    max_retries: 0,
                    get_new_uptoken: true,
                    auto_start: true,
                    log_level: 5,
                    init: {
                        FilesAdded: this.fileAdd,

                        // eslint-disable-next-line @typescript-eslint/no-empty-function
                        BeforeUpload: function () {},
                        QueueChanged: (up) => {
                            this.setState({ queued: up.total.queued });
                        },
                        UploadProgress: (up, file) => {
                            window.fileList["updateStatus"](file);
                        },
                        UploadComplete: (up, file) => {
                            if (file.length === 0) {
                                return;
                            }
                            console.log(
                                "UploadComplete",
                                file[0].status,
                                file[0]
                            );
                            for (let i = 0; i < file.length; i++) {
                                if (file[i].status === 5) {
                                    window.fileList["setComplete"](file[i]);
                                }
                            }
                            // 无异步操作的策略，直接刷新
                            if (
                                user.policy.saveType !== "onedrive" &&
                                user.policy.saveType !== "cos"
                            ) {
                                this.props.refreshFileList();
                                this.props.refreshStorage();
                            }
                        },
                        Fresh: () => {
                            this.props.refreshFileList();
                            this.props.refreshStorage();
                        },
                        // eslint-disable-next-line @typescript-eslint/no-empty-function
                        FileUploaded: function () {},
                        Error: (up, err, errTip) => {
                            window.fileList["openFileList"]();
                            window.fileList["setError"](err.file, errTip);
                        },
                        // eslint-disable-next-line @typescript-eslint/no-empty-function
                        FilesRemoved: () => {},
                    },
                });
                // this.fileList["openFileList"]();
            } else this.onError();
        }
    }

    // eslint-disable-next-line @typescript-eslint/no-empty-function
    onError() {}

    openFileList = () => {
        window.fileList["openFileList"]();
    };

    render() {
        return (
            <div>
                <FileList
                    inRef={this.setRef.bind(this)}
                    cancelUpload={this.cancelUpload.bind(this)}
                    reQueue={this.reQueue.bind(this)}
                />
                {this.props.keywords === "" && (
                    <UploadButton
                        Queued={this.state.queued}
                        openFileList={this.openFileList}
                    />
                )}
            </div>
        );
    }
}

const Uploader = connect(mapStateToProps, mapDispatchToProps, null, {
    forwardRef: true,
})(uploaderLoader()(UploaderComponent));

export default Uploader;
