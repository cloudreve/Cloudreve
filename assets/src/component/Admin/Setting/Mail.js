import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogTitle from "@material-ui/core/DialogTitle";
import FormControl from "@material-ui/core/FormControl";
import FormHelperText from "@material-ui/core/FormHelperText";
import Input from "@material-ui/core/Input";
import InputLabel from "@material-ui/core/InputLabel";
import { makeStyles } from "@material-ui/core/styles";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../../actions";
import API from "../../../middleware/Api";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import Switch from "@material-ui/core/Switch";

const useStyles = makeStyles((theme) => ({
    root: {
        [theme.breakpoints.up("md")]: {
            marginLeft: 100,
        },
        marginBottom: 40,
    },
    form: {
        maxWidth: 400,
        marginTop: 20,
        marginBottom: 20,
    },
    formContainer: {
        [theme.breakpoints.up("md")]: {
            padding: "0px 24px 0 24px",
        },
    },
    buttonMargin: {
        marginLeft: 8,
    },
}));

export default function Mail() {
    const classes = useStyles();
    const [loading, setLoading] = useState(false);
    const [test, setTest] = useState(false);
    const [tesInput, setTestInput] = useState("");
    const [options, setOptions] = useState({
        fromName: "",
        fromAdress: "",
        smtpHost: "",
        smtpPort: "",
        replyTo: "",
        smtpUser: "",
        smtpPass: "",
        smtpEncryption: "",
        mail_keepalive: "30",
        mail_activation_template: "",
        mail_reset_pwd_template: "",
    });

    const handleChange = (name) => (event) => {
        setOptions({
            ...options,
            [name]: event.target.value,
        });
    };

    const handleCheckChange = (name) => (event) => {
        let value = event.target.value;
        if (event.target.checked !== undefined) {
            value = event.target.checked ? "1" : "0";
        }
        setOptions({
            ...options,
            [name]: value,
        });
    };

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    useEffect(() => {
        API.post("/admin/setting", {
            keys: Object.keys(options),
        })
            .then((response) => {
                setOptions(response.data);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
        // eslint-disable-next-line
    }, []);

    const sendTestMail = () => {
        setLoading(true);
        API.post("/admin/mailTest", {
            to: tesInput,
        })
            .then(() => {
                ToggleSnackbar("top", "right", "测试邮件已发送", "success");
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
    };

    const reload = () => {
        API.get("/admin/reload/email")
            // eslint-disable-next-line @typescript-eslint/no-empty-function
            .then(() => {})
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            // eslint-disable-next-line @typescript-eslint/no-empty-function
            .then(() => {});
    };

    const submit = (e) => {
        e.preventDefault();
        setLoading(true);
        const option = [];
        Object.keys(options).forEach((k) => {
            option.push({
                key: k,
                value: options[k],
            });
        });
        API.patch("/admin/setting", {
            options: option,
        })
            .then(() => {
                ToggleSnackbar("top", "right", "设置已更改", "success");
                reload();
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
    };

    return (
        <div>
            <Dialog
                open={test}
                onClose={() => setTest(false)}
                aria-labelledby="form-dialog-title"
            >
                <DialogTitle id="form-dialog-title">发件测试</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        <Typography>
                            发送测试邮件前，请先保存已更改的邮件设置；
                        </Typography>
                        <Typography>
                            邮件发送结果不会立即反馈，如果您长时间未收到测试邮件，请检查
                            Cloudreve 在终端输出的错误日志。
                        </Typography>
                    </DialogContentText>
                    <TextField
                        autoFocus
                        margin="dense"
                        id="name"
                        label="收件人地址"
                        value={tesInput}
                        onChange={(e) => setTestInput(e.target.value)}
                        type="email"
                        fullWidth
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setTest(false)} color="default">
                        取消
                    </Button>
                    <Button
                        onClick={() => sendTestMail()}
                        disabled={loading}
                        color="primary"
                    >
                        发送
                    </Button>
                </DialogActions>
            </Dialog>

            <form onSubmit={submit}>
                <div className={classes.root}>
                    <Typography variant="h6" gutterBottom>
                        发信
                    </Typography>

                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    发件人名
                                </InputLabel>
                                <Input
                                    value={options.fromName}
                                    onChange={handleChange("fromName")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    邮件中展示的发件人姓名
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    发件人邮箱
                                </InputLabel>
                                <Input
                                    type={"email"}
                                    required
                                    value={options.fromAdress}
                                    onChange={handleChange("fromAdress")}
                                />
                                <FormHelperText id="component-helper-text">
                                    发件邮箱的地址
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    SMTP 服务器
                                </InputLabel>
                                <Input
                                    value={options.smtpHost}
                                    onChange={handleChange("smtpHost")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    发件服务器地址，不含端口号
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    SMTP 端口
                                </InputLabel>
                                <Input
                                    inputProps={{ min: 1, step: 1 }}
                                    type={"number"}
                                    value={options.smtpPort}
                                    onChange={handleChange("smtpPort")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    发件服务器地址端口号
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    SMTP 用户名
                                </InputLabel>
                                <Input
                                    value={options.smtpUser}
                                    onChange={handleChange("smtpUser")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    发信邮箱用户名，一般与邮箱地址相同
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    SMTP 密码
                                </InputLabel>
                                <Input
                                    type={"password"}
                                    value={options.smtpPass}
                                    onChange={handleChange("smtpPass")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    发信邮箱密码
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    回信邮箱
                                </InputLabel>
                                <Input
                                    value={options.replyTo}
                                    onChange={handleChange("replyTo")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    用户回复系统发送的邮件时，用于接收回信的邮箱
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={
                                                options.smtpEncryption === "1"
                                            }
                                            onChange={handleCheckChange(
                                                "smtpEncryption"
                                            )}
                                        />
                                    }
                                    label="强制使用 SSL 连接"
                                />
                                <FormHelperText id="component-helper-text">
                                    是否强制使用 SSL
                                    加密连接。如果无法发送邮件，可关闭此项，
                                    Cloudreve 会尝试使用 STARTTLS
                                    并决定是否使用加密连接
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    SMTP 连接有效期 (秒)
                                </InputLabel>
                                <Input
                                    inputProps={{ min: 1, step: 1 }}
                                    type={"number"}
                                    value={options.mail_keepalive}
                                    onChange={handleChange("mail_keepalive")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    有效期内建立的 SMTP
                                    连接会被新邮件发送请求复用
                                </FormHelperText>
                            </FormControl>
                        </div>
                    </div>
                </div>

                <div className={classes.root}>
                    <Typography variant="h6" gutterBottom>
                        邮件模板
                    </Typography>

                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    新用户激活
                                </InputLabel>
                                <Input
                                    value={options.mail_activation_template}
                                    onChange={handleChange(
                                        "mail_activation_template"
                                    )}
                                    multiline
                                    rowsMax="10"
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    新用户注册后激活邮件的模板
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    重置密码
                                </InputLabel>
                                <Input
                                    value={options.mail_reset_pwd_template}
                                    onChange={handleChange(
                                        "mail_reset_pwd_template"
                                    )}
                                    multiline
                                    rowsMax="10"
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    密码重置邮件模板
                                </FormHelperText>
                            </FormControl>
                        </div>
                    </div>
                </div>

                <div className={classes.root}>
                    <Button
                        disabled={loading}
                        type={"submit"}
                        variant={"contained"}
                        color={"primary"}
                    >
                        保存
                    </Button>
                    {"   "}
                    <Button
                        className={classes.buttonMargin}
                        variant={"outlined"}
                        color={"primary"}
                        onClick={() => setTest(true)}
                    >
                        发送测试邮件
                    </Button>
                </div>
            </form>
        </div>
    );
}
