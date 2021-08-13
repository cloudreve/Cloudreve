import React, { useState, useCallback, useEffect } from "react";
import { FormLabel, makeStyles } from "@material-ui/core";
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
import TextField from "@material-ui/core/TextField";
import Select from "@material-ui/core/Select";
import MenuItem from "@material-ui/core/MenuItem";
import {
    refreshTimeZone,
    timeZone,
    validateTimeZone,
} from "../../utils/datetime";
import FormControl from "@material-ui/core/FormControl";
import Auth from "../../middleware/Auth";

const useStyles = makeStyles((theme) => ({}));

export default function TimeZoneDialog(props) {
    const [timeZoneValue, setTimeZoneValue] = useState(timeZone);
    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const saveZoneInfo = () => {
        if (!validateTimeZone(timeZoneValue)) {
            ToggleSnackbar("top", "right", "无效的时区名称", "warning");
            return;
        }
        Auth.SetPreference("timeZone", timeZoneValue);
        refreshTimeZone();
        props.onClose();
    };

    const classes = useStyles();

    return (
        <Dialog
            open={props.open}
            onClose={props.onClose}
            aria-labelledby="form-dialog-title"
        >
            <DialogTitle id="form-dialog-title">更改时区</DialogTitle>

            <DialogContent>
                <FormControl>
                    <TextField
                        label={"IANA 时区名称标识"}
                        value={timeZoneValue}
                        onChange={(e) => setTimeZoneValue(e.target.value)}
                    />
                </FormControl>
            </DialogContent>

            <DialogActions>
                <Button onClick={props.onClose}>取消</Button>
                <div className={classes.wrapper}>
                    <Button
                        color="primary"
                        disabled={timeZoneValue === ""}
                        onClick={() => saveZoneInfo()}
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
