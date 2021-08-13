import React from "react";
import { Facebook } from "react-content-loader";
import { makeStyles, useTheme } from "@material-ui/core/styles";

const useStyles = makeStyles((theme) => ({
    loader: {
        width: "80%",
        [theme.breakpoints.up("md")]: {
            width: " 50%",
        },

        marginTop: 30,
    },
}));

const MyLoader = (props) => {
    return (
        <Facebook
            backgroundColor={props.dark ? "#333" : "#f5f6f7"}
            foregroundColor={props.dark ? "#636363" : "#eee"}
            className={props.className}
        />
    );
};

function PageLoading() {
    const theme = useTheme();
    const classes = useStyles();

    return (
        <div
            style={{
                textAlign: "center",
            }}
        >
            <MyLoader
                dark={theme.palette.type === "dark"}
                className={classes.loader}
            />
        </div>
    );
}

export default PageLoading;
