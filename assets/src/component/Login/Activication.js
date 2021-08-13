import React, { useCallback, useState, useEffect } from "react";
import { useDispatch } from "react-redux";
import { makeStyles } from "@material-ui/core";
import { toggleSnackbar } from "../../actions/index";
import { useHistory } from "react-router-dom";
import API from "../../middleware/Api";
import { Button, Paper, Avatar, Typography } from "@material-ui/core";
import EmailIcon from "@material-ui/icons/EmailOutlined";
import { useLocation } from "react-router";
const useStyles = makeStyles((theme) => ({
    layout: {
        width: "auto",
        marginTop: "110px",
        marginLeft: theme.spacing(3),
        marginRight: theme.spacing(3),
        [theme.breakpoints.up("sm")]: {
            width: 400,
            marginLeft: "auto",
            marginRight: "auto",
        },
        marginBottom: 110,
    },
    paper: {
        marginTop: theme.spacing(8),
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        padding: `${theme.spacing(2)}px ${theme.spacing(3)}px ${theme.spacing(
            3
        )}px`,
    },
    avatar: {
        margin: theme.spacing(1),
        backgroundColor: theme.palette.secondary.main,
    },
    submit: {
        marginTop: theme.spacing(3),
    },
}));

function useQuery() {
    return new URLSearchParams(useLocation().search);
}

function Activation() {
    const query = useQuery();
    const location = useLocation();

    const [success, setSuccess] = useState(false);
    const [email, setEmail] = useState("");

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );
    const history = useHistory();

    const classes = useStyles();

    useEffect(() => {
        API.get(
            "/user/activate/" + query.get("id") + "?sign=" + query.get("sign")
        )
            .then((response) => {
                setEmail(response.data);
                setSuccess(true);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "warning");
                history.push("/login");
            });
        // eslint-disable-next-line
    }, [location]);

    return (
        <div className={classes.layout}>
            {success && (
                <Paper className={classes.paper}>
                    <Avatar className={classes.avatar}>
                        <EmailIcon />
                    </Avatar>
                    <Typography component="h1" variant="h5">
                        激活成功
                    </Typography>
                    <Typography style={{ marginTop: "20px" }}>
                        您的账号已被成功激活。
                    </Typography>
                    <Button
                        type="submit"
                        fullWidth
                        variant="contained"
                        color="primary"
                        className={classes.submit}
                        onClick={() => history.push("/login?username=" + email)}
                    >
                        返回登录
                    </Button>
                </Paper>
            )}
        </div>
    );
}

export default Activation;
