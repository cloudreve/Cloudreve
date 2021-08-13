import React from "react";
import SentimentVeryDissatisfiedIcon from "@material-ui/icons/SentimentVeryDissatisfied";
import { lighten, makeStyles } from "@material-ui/core/styles";

const useStyles = makeStyles((theme) => ({
    icon: {
        fontSize: "160px",
    },
    emptyContainer: {
        bottom: "0",
        height: "300px",
        margin: "50px auto",
        width: "300px",
        color: lighten(theme.palette.text.disabled, 0.4),
        textAlign: "center",
        paddingTop: "20px",
    },
    emptyInfoBig: {
        fontSize: "25px",
        color: lighten(theme.palette.text.disabled, 0.4),
    },
}));

export default function Notice(props) {
    const classes = useStyles();
    return (
        <div className={classes.emptyContainer}>
            <SentimentVeryDissatisfiedIcon className={classes.icon} />
            <div className={classes.emptyInfoBig}>{props.msg}</div>
        </div>
    );
}
