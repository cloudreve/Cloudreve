import React, { useCallback, useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import LockOutlinedIcon from "@material-ui/icons/LockOutlined";
import {
    Avatar,
    Button,
    Divider,
    FormControl,
    Input,
    InputLabel,
    Link,
    makeStyles,
    Paper,
    Typography,
} from "@material-ui/core";
import {
    applyThemes,
    setSessionStatus,
    toggleSnackbar,
} from "../../actions/index";
import { useHistory } from "react-router-dom";
import API from "../../middleware/Api";
import Auth from "../../middleware/Auth";
import { bufferDecode, bufferEncode } from "../../utils/index";
import { enableUploaderLoad } from "../../middleware/Init";
import { Fingerprint, VpnKey } from "@material-ui/icons";
import VpnIcon from "@material-ui/icons/VpnKeyOutlined";
import { useLocation } from "react-router";
import { ICPFooter } from "../Common/ICPFooter";
import { useCaptcha } from "../../hooks/useCaptcha";

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
    form: {
        width: "100%", // Fix IE 11 issue.
        marginTop: theme.spacing(1),
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
    captchaContainer: {
        display: "flex",
        marginTop: "10px",
        [theme.breakpoints.down("sm")]: {
            display: "block",
        },
    },
    captchaPlaceholder: {
        width: 200,
    },
    buttonContainer: {
        display: "flex",
    },
    authnLink: {
        textAlign: "center",
        marginTop: 16,
    },
}));

function useQuery() {
    return new URLSearchParams(useLocation().search);
}

function LoginForm() {
    const [email, setEmail] = useState("");
    const [pwd, setPwd] = useState("");
    const [loading, setLoading] = useState(false);
    const [useAuthn, setUseAuthn] = useState(false);
    const [twoFA, setTwoFA] = useState(false);
    const [faCode, setFACode] = useState("");

    const loginCaptcha = useSelector((state) => state.siteConfig.loginCaptcha);
    const title = useSelector((state) => state.siteConfig.title);
    const authn = useSelector((state) => state.siteConfig.authn);

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );
    const ApplyThemes = useCallback((theme) => dispatch(applyThemes(theme)), [
        dispatch,
    ]);
    const SetSessionStatus = useCallback(
        (status) => dispatch(setSessionStatus(status)),
        [dispatch]
    );

    const history = useHistory();
    const location = useLocation();
    const {
        captchaLoading,
        isValidate,
        validate,
        CaptchaRender,
        captchaRefreshRef,
        captchaParamsRef,
    } = useCaptcha();
    const query = useQuery();

    const classes = useStyles();

    useEffect(() => {
        setEmail(query.get("username"));
    }, [location]);

    const afterLogin = (data) => {
        Auth.authenticate(data);

        // 设置用户主题色
        if (data["preferred_theme"] !== "") {
            ApplyThemes(data["preferred_theme"]);
        }
        enableUploaderLoad();

        // 设置登录状态
        SetSessionStatus(true);

        history.push("/home");
        ToggleSnackbar("top", "right", "登录成功", "success");

        localStorage.removeItem("siteConfigCache");
    };

    const authnLogin = (e) => {
        e.preventDefault();
        if (!navigator.credentials) {
            ToggleSnackbar("top", "right", "当前浏览器或环境不支持", "warning");

            return;
        }

        setLoading(true);

        API.get("/user/authn/" + email)
            .then((response) => {
                const credentialRequestOptions = response.data;
                console.log(credentialRequestOptions);
                credentialRequestOptions.publicKey.challenge = bufferDecode(
                    credentialRequestOptions.publicKey.challenge
                );
                credentialRequestOptions.publicKey.allowCredentials.forEach(
                    function (listItem) {
                        listItem.id = bufferDecode(listItem.id);
                    }
                );

                return navigator.credentials.get({
                    publicKey: credentialRequestOptions.publicKey,
                });
            })
            .then((assertion) => {
                const authData = assertion.response.authenticatorData;
                const clientDataJSON = assertion.response.clientDataJSON;
                const rawId = assertion.rawId;
                const sig = assertion.response.signature;
                const userHandle = assertion.response.userHandle;

                return API.post(
                    "/user/authn/finish/" + email,
                    JSON.stringify({
                        id: assertion.id,
                        rawId: bufferEncode(rawId),
                        type: assertion.type,
                        response: {
                            authenticatorData: bufferEncode(authData),
                            clientDataJSON: bufferEncode(clientDataJSON),
                            signature: bufferEncode(sig),
                            userHandle: bufferEncode(userHandle),
                        },
                    })
                );
            })
            .then((response) => {
                afterLogin(response.data);
            })
            .catch((error) => {
                console.log(error);
                ToggleSnackbar("top", "right", error.message, "warning");
            })
            .then(() => {
                setLoading(false);
            });
    };

    const login = (e) => {
        e.preventDefault();
        setLoading(true);
        if (!isValidate.current.isValidate && loginCaptcha) {
            validate(() => login(e), setLoading);
            return;
        }
        API.post("/user/session", {
            userName: email,
            Password: pwd,
            ...captchaParamsRef.current,
        })
            .then((response) => {
                setLoading(false);
                if (response.rawData.code === 203) {
                    setTwoFA(true);
                } else {
                    afterLogin(response.data);
                }
            })
            .catch((error) => {
                setLoading(false);
                ToggleSnackbar("top", "right", error.message, "warning");
                captchaRefreshRef.current();
            });
    };

    const twoFALogin = (e) => {
        e.preventDefault();
        setLoading(true);
        API.post("/user/2fa", {
            code: faCode,
        })
            .then((response) => {
                setLoading(false);
                afterLogin(response.data);
            })
            .catch((error) => {
                setLoading(false);
                ToggleSnackbar("top", "right", error.message, "warning");
            });
    };

    return (
        <div className={classes.layout}>
            {!twoFA && (
                <>
                    <Paper className={classes.paper}>
                        <Avatar className={classes.avatar}>
                            <LockOutlinedIcon />
                        </Avatar>
                        <Typography component="h1" variant="h5">
                            登录 {title}
                        </Typography>
                        {!useAuthn && (
                            <form className={classes.form} onSubmit={login}>
                                <FormControl margin="normal" required fullWidth>
                                    <InputLabel htmlFor="email">
                                        电子邮箱
                                    </InputLabel>
                                    <Input
                                        id="email"
                                        type="email"
                                        name="email"
                                        onChange={(e) =>
                                            setEmail(e.target.value)
                                        }
                                        autoComplete
                                        value={email}
                                        autoFocus
                                    />
                                </FormControl>
                                <FormControl margin="normal" required fullWidth>
                                    <InputLabel htmlFor="password">
                                        密码
                                    </InputLabel>
                                    <Input
                                        name="password"
                                        onChange={(e) => setPwd(e.target.value)}
                                        type="password"
                                        id="password"
                                        value={pwd}
                                        autoComplete
                                    />
                                </FormControl>
                                {loginCaptcha && <CaptchaRender />}

                                <Button
                                    type="submit"
                                    fullWidth
                                    variant="contained"
                                    color="primary"
                                    disabled={
                                        loading ||
                                        (loginCaptcha ? captchaLoading : false)
                                    }
                                    className={classes.submit}
                                >
                                    登录
                                </Button>
                            </form>
                        )}
                        {useAuthn && (
                            <form className={classes.form}>
                                <FormControl margin="normal" required fullWidth>
                                    <InputLabel htmlFor="email">
                                        电子邮箱
                                    </InputLabel>
                                    <Input
                                        id="email"
                                        type="email"
                                        name="email"
                                        onChange={(e) =>
                                            setEmail(e.target.value)
                                        }
                                        autoComplete
                                        value={email}
                                        autoFocus
                                    />
                                </FormControl>
                                <Button
                                    type="submit"
                                    fullWidth
                                    variant="contained"
                                    color="primary"
                                    disabled={loading}
                                    onClick={authnLogin}
                                    className={classes.submit}
                                >
                                    下一步
                                </Button>
                            </form>
                        )}
                        <Divider />
                        <div className={classes.link}>
                            <div>
                                <Link href={"/forget"}>忘记密码</Link>
                            </div>
                            <div>
                                <Link href={"/signup"}>注册账号</Link>
                            </div>
                        </div>

                        <ICPFooter />
                    </Paper>

                    {authn && (
                        <div className={classes.authnLink}>
                            <Button
                                color="primary"
                                onClick={() => setUseAuthn(!useAuthn)}
                            >
                                {!useAuthn && (
                                    <>
                                        <Fingerprint
                                            style={{
                                                marginRight: 8,
                                            }}
                                        />
                                        使用外部验证器登录
                                    </>
                                )}
                                {useAuthn && (
                                    <>
                                        <VpnKey
                                            style={{
                                                marginRight: 8,
                                            }}
                                        />
                                        使用密码登录
                                    </>
                                )}
                            </Button>
                        </div>
                    )}
                </>
            )}
            {twoFA && (
                <Paper className={classes.paper}>
                    <Avatar className={classes.avatar}>
                        <VpnIcon />
                    </Avatar>
                    <Typography component="h1" variant="h5">
                        二步验证
                    </Typography>
                    <form className={classes.form} onSubmit={twoFALogin}>
                        <FormControl margin="normal" required fullWidth>
                            <InputLabel htmlFor="code">
                                请输入六位二步验证代码
                            </InputLabel>
                            <Input
                                id="code"
                                type="number"
                                name="code"
                                onChange={(event) =>
                                    setFACode(event.target.value)
                                }
                                autoComplete
                                value={faCode}
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
                            继续登录
                        </Button>{" "}
                    </form>{" "}
                    <Divider />
                </Paper>
            )}
        </div>
    );
}

export default LoginForm;
