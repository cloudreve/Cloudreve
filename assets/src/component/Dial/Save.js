import React from "react";
import { makeStyles } from "@material-ui/core";
import SaveIcon from "@material-ui/icons/Save";
import CheckIcon from "@material-ui/icons/Check";
import AutoHidden from "./AutoHidden";
import statusHelper from "../../utils/page";
import Fab from "@material-ui/core/Fab";
import Tooltip from "@material-ui/core/Tooltip";
import CircularProgress from "@material-ui/core/CircularProgress";
import { green } from "@material-ui/core/colors";
import clsx from "clsx";

const useStyles = makeStyles((theme) => ({
    fab: {
        margin: 0,
        top: "auto",
        right: 20,
        bottom: 20,
        left: "auto",
        zIndex: 5,
        position: "fixed",
    },
    badge: {
        position: "absolute",
        bottom: 26,
        top: "auto",
        zIndex: 9999,
        right: 7,
    },
    fabProgress: {
        color: green[500],
        position: "absolute",
        top: -6,
        left: -6,
        zIndex: 1,
    },
    wrapper: {
        margin: theme.spacing(1),
        position: "relative",
    },
    buttonSuccess: {
        backgroundColor: green[500],
        "&:hover": {
            backgroundColor: green[700],
        },
    },
}));

export default function SaveButton(props) {
    const classes = useStyles();
    const buttonClassname = clsx({
        [classes.buttonSuccess]: props.status === "success",
    });

    return (
        <AutoHidden enable={statusHelper.isMobile()}>
            <div className={classes.fab}>
                <div className={classes.wrapper}>
                    <Tooltip title={"保存"} placement={"left"}>
                        <Fab
                            onClick={props.onClick}
                            color="primary"
                            className={buttonClassname}
                            disabled={props.status === "loading"}
                            aria-label="add"
                        >
                            {props.status === "success" ? (
                                <CheckIcon />
                            ) : (
                                <SaveIcon />
                            )}
                        </Fab>
                    </Tooltip>
                    {props.status === "loading" && (
                        <CircularProgress
                            size={68}
                            className={classes.fabProgress}
                        />
                    )}
                </div>
            </div>
        </AutoHidden>
    );
}
