import React, { useState } from "react";
import { makeStyles } from "@material-ui/core";
import { Dialog } from "@material-ui/core";
import DialogTitle from "@material-ui/core/DialogTitle";
import DialogActions from "@material-ui/core/DialogActions";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";
import { FolderOpenOutlined, LabelOutlined } from "@material-ui/icons";
import PathSelector from "../FileManager/PathSelector";
const useStyles = makeStyles((theme) => ({
    formGroup: {
        display: "flex",
        marginTop: theme.spacing(1),
    },
    formIcon: {
        marginTop: 21,
        marginRight: 19,
        color: theme.palette.text.secondary,
    },
    input: {
        width: 250,
    },
    dialogContent: {
        paddingTop: 24,
        paddingRight: 24,
        paddingBottom: 8,
        paddingLeft: 24,
    },
    button: {
        marginTop: 8,
    },
}));

export default function CreateWebDAVAccount(props) {
    const [value, setValue] = useState({
        name: "",
        path: "/",
    });
    const [pathSelectDialog, setPathSelectDialog] = React.useState(false);
    const [selectedPath, setSelectedPath] = useState("");
    // eslint-disable-next-line
    const [selectedPathName, setSelectedPathName] = useState("");
    const classes = useStyles();

    const setMoveTarget = (folder) => {
        const path =
            folder.path === "/"
                ? folder.path + folder.name
                : folder.path + "/" + folder.name;
        setSelectedPath(path);
        setSelectedPathName(folder.name);
    };

    const handleInputChange = (name) => (e) => {
        setValue({
            ...value,
            [name]: e.target.value,
        });
    };

    const selectPath = () => {
        setValue({
            ...value,
            path: selectedPath === "//" ? "/" : selectedPath,
        });
        setPathSelectDialog(false);
    };

    return (
        <Dialog
            open={props.open}
            onClose={props.onClose}
            aria-labelledby="form-dialog-title"
        >
            <Dialog
                open={pathSelectDialog}
                onClose={() => setPathSelectDialog(false)}
                aria-labelledby="form-dialog-title"
            >
                <DialogTitle id="form-dialog-title">选择目录</DialogTitle>
                <PathSelector
                    presentPath="/"
                    selected={[]}
                    onSelect={setMoveTarget}
                />

                <DialogActions>
                    <Button onClick={() => setPathSelectDialog(false)}>
                        取消
                    </Button>
                    <Button
                        onClick={selectPath}
                        color="primary"
                        disabled={selectedPath === ""}
                    >
                        确定
                    </Button>
                </DialogActions>
            </Dialog>
            <div className={classes.dialogContent}>
                <div className={classes.formContainer}>
                    <div className={classes.formGroup}>
                        <div className={classes.formIcon}>
                            <LabelOutlined />
                        </div>

                        <TextField
                            className={classes.input}
                            value={value.name}
                            onChange={handleInputChange("name")}
                            label="备注名"
                        />
                    </div>
                    <div className={classes.formGroup}>
                        <div className={classes.formIcon}>
                            <FolderOpenOutlined />
                        </div>
                        <div>
                            <TextField
                                value={value.path}
                                onChange={handleInputChange("path")}
                                className={classes.input}
                                label="相对根目录"
                            />
                            <br />
                            <Button
                                className={classes.button}
                                color="primary"
                                onClick={() => setPathSelectDialog(true)}
                            >
                                选择目录
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
            <DialogActions>
                <Button onClick={props.onClose}>取消</Button>
                <Button
                    disabled={value.path === "" || value.name === ""}
                    color="primary"
                    onClick={() => props.callback(value)}
                >
                    确定
                </Button>
            </DialogActions>
        </Dialog>
    );
}
