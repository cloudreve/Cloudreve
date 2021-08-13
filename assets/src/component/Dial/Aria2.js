import React, { useCallback } from "react";
import { openRemoteDownloadDialog } from "../../actions";
import { useDispatch } from "react-redux";
import AutoHidden from "./AutoHidden";
import { makeStyles } from "@material-ui/core";
import Fab from "@material-ui/core/Fab";
import { Add } from "@material-ui/icons";
import Modals from "../FileManager/Modals";

const useStyles = makeStyles(() => ({
    fab: {
        margin: 0,
        top: "auto",
        right: 20,
        bottom: 20,
        left: "auto",
        zIndex: 5,
        position: "fixed",
    },
}));

export default function RemoteDownloadButton() {
    const classes = useStyles();
    const dispatch = useDispatch();

    const OpenRemoteDownloadDialog = useCallback(
        () => dispatch(openRemoteDownloadDialog()),
        [dispatch]
    );

    return (
        <>
            <Modals />
            <AutoHidden enable>
                <Fab
                    className={classes.fab}
                    color="secondary"
                    onClick={() => OpenRemoteDownloadDialog()}
                >
                    <Add />
                </Fab>
            </AutoHidden>
        </>
    );
}
