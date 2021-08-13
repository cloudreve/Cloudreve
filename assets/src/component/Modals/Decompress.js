import React, { useState, useCallback } from "react";
import { makeStyles } from "@material-ui/core";
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    DialogContentText,
    CircularProgress,
} from "@material-ui/core";
import { toggleSnackbar, setModalsLoading } from "../../actions/index";
import PathSelector from "../FileManager/PathSelector";
import { useDispatch } from "react-redux";
import API from "../../middleware/Api";
import { filePath } from "../../utils";

const useStyles = makeStyles((theme) => ({
    contentFix: {
        padding: "10px 24px 0px 24px",
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
        marginTop: -12,
        marginLeft: -12,
    },
}));

export default function DecompressDialog(props) {
    const [selectedPath, setSelectedPath] = useState("");
    const [selectedPathName, setSelectedPathName] = useState("");

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );
    const SetModalsLoading = useCallback(
        (status) => {
            dispatch(setModalsLoading(status));
        },
        [dispatch]
    );

    const setMoveTarget = (folder) => {
        const path =
            folder.path === "/"
                ? folder.path + folder.name
                : folder.path + "/" + folder.name;
        setSelectedPath(path);
        setSelectedPathName(folder.name);
    };

    const submitMove = (e) => {
        if (e != null) {
            e.preventDefault();
        }
        SetModalsLoading(true);
        API.post("/file/decompress", {
            src: filePath(props.selected[0]),
            dst: selectedPath === "//" ? "/" : selectedPath,
        })
            .then(() => {
                props.onClose();
                ToggleSnackbar("top", "right", "解压缩任务已创建", "success");
                SetModalsLoading(false);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
                SetModalsLoading(false);
            });
    };

    const classes = useStyles();

    return (
        <Dialog
            open={props.open}
            onClose={props.onClose}
            aria-labelledby="form-dialog-title"
        >
            <DialogTitle id="form-dialog-title">解压送至</DialogTitle>
            <PathSelector
                presentPath={props.presentPath}
                selected={props.selected}
                onSelect={setMoveTarget}
            />

            {selectedPath !== "" && (
                <DialogContent className={classes.contentFix}>
                    <DialogContentText>
                        解压送至 <strong>{selectedPathName}</strong>
                    </DialogContentText>
                </DialogContent>
            )}
            <DialogActions>
                <Button onClick={props.onClose}>取消</Button>
                <div className={classes.wrapper}>
                    <Button
                        onClick={submitMove}
                        color="primary"
                        disabled={selectedPath === "" || props.modalsLoading}
                    >
                        确定
                        {props.modalsLoading && (
                            <CircularProgress
                                size={24}
                                className={classes.buttonProgress}
                            />
                        )}
                    </Button>
                </div>
            </DialogActions>
        </Dialog>
    );
}
