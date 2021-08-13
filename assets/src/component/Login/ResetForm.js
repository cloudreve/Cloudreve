import React, { useCallback, useState } from "react";
import { useDispatch } from "react-redux";
import { makeStyles } from "@material-ui/core";
import { toggleSnackbar } from "../../actions/index";
import { useHistory } from "react-router-dom";
import API from "../../middleware/Api";
import {
    Button,
    FormControl,
    Divider,
    Link,
    Input,
    InputLabel,
    Paper,
    Avatar,
    Typography,
} from "@material-ui/core";
import { useLocation } from "react-router";
import KeyIcon from "@material-ui/icons/VpnKeyOutlined";
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
    link: {
        marginTop: "20px",
        display: "flex",
        width: "100%",
        justifyContent: "space-between",
    },
}));

function useQuery() {
    return new URLSearchParams(useLocation().search);
}

function ResetForm() {
    const query = useQuery();
    const [input, setInput] = useState({
        password: "",
        password_repeat: "",
    });
    const [loading, setLoading] = useState(false);
    const handleInputChange = (name) => (e) => {
        setInput({
            ...input,
            [name]: e.target.value,
        });
    };
    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );
    const history = useHistory();

    const submit = (e) => {
        e.preventDefault();
        if (input.password !== input.password_repeat) {
            ToggleSnackbar("top", "right", "两次密码输入不一致", "warning");
            return;
        }
        setLoading(true);
        API.patch("/user/reset", {
            secret: query.get("sign"),
            id: query.get("id"),
            Password: input.password,
        })
            .then(() => {
                setLoading(false);
                history.push("/login");
                ToggleSnackbar("top", "right", "密码已重设", "success");
            })
            .catch((error) => {
                setLoading(false);
                ToggleSnackbar("top", "right", error.message, "warning");
            });
    };

    const classes = useStyles();

    return (
        <div className={classes.layout}>
            <Paper className={classes.paper}>
                <Avatar className={classes.avatar}>
                    <KeyIcon />
                </Avatar>
                <Typography component="h1" variant="h5">
                    找回密码
                </Typography>
                <form className={classes.form} onSubmit={submit}>
                    <FormControl margin="normal" required fullWidth>
                        <InputLabel htmlFor="email">新密码</InputLabel>
                        <Input
                            id="pwd"
                            type="password"
                            name="pwd"
                            onChange={handleInputChange("password")}
                            autoComplete
                            value={input.password}
                            autoFocus
                        />
                    </FormControl>
                    <FormControl margin="normal" required fullWidth>
                        <InputLabel htmlFor="email">重复新密码</InputLabel>
                        <Input
                            id="pwdRepeat"
                            type="password"
                            name="pwdRepeat"
                            onChange={handleInputChange("password_repeat")}
                            autoComplete
                            value={input.password_repeat}
                            autoFocus
                        />
                    </FormControl>
                    <Button
                        type="submit"
                        fullWidth
                        variant="contained"
                        color="primary"
                        disabled={loading}
                        className={classes.submit}
                    >
                        重设密码
                    </Button>{" "}
                </form>{" "}
                <Divider />
                <div className={classes.link}>
                    <div>
                        <Link href={"/#/login"}>返回登录</Link>
                    </div>
                    <div>
                        <Link href={"/#/signup"}>注册账号</Link>
                    </div>
                </div>
            </Paper>
        </div>
    );
}

export default ResetForm;
