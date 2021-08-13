import Button from "@material-ui/core/Button";
import FormControl from "@material-ui/core/FormControl";
import FormHelperText from "@material-ui/core/FormHelperText";
import Input from "@material-ui/core/Input";
import InputLabel from "@material-ui/core/InputLabel";
import Link from "@material-ui/core/Link";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import Alert from "@material-ui/lab/Alert";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../../actions";
import API from "../../../middleware/Api";

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
}));

export default function Aria2() {
    const classes = useStyles();
    const [loading, setLoading] = useState(false);
    const [options, setOptions] = useState({
        aria2_rpcurl: "",
        aria2_token: "",
        aria2_temp_path: "",
        aria2_options: "",
        aria2_interval: "0",
        aria2_call_timeout: "0",
    });

    const handleChange = (name) => (event) => {
        setOptions({
            ...options,
            [name]: event.target.value,
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

    const reload = () => {
        API.get("/admin/reload/aria2")
            // eslint-disable-next-line @typescript-eslint/no-empty-function
            .then(() => {})
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            // eslint-disable-next-line @typescript-eslint/no-empty-function
            .then(() => {});
    };

    const test = () => {
        setLoading(true);
        API.post("/admin/aria2/test", {
            server: options.aria2_rpcurl,
            token: options.aria2_token,
        })
            .then((response) => {
                ToggleSnackbar(
                    "top",
                    "right",
                    "连接成功，Aria2 版本为：" + response.data,
                    "success"
                );
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
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
            <form onSubmit={submit}>
                <div className={classes.root}>
                    <Typography variant="h6" gutterBottom>
                        Aria2
                    </Typography>

                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <Alert severity="info" style={{ marginTop: 8 }}>
                                <Typography variant="body2">
                                    Cloudreve 的离线下载功能由{" "}
                                    <Link
                                        href={"https://aria2.github.io/"}
                                        target={"_blank"}
                                    >
                                        Aria2
                                    </Link>{" "}
                                    驱动。如需使用，请在同一设备上以和运行
                                    Cloudreve 相同的用户身份启动 Aria2， 并在
                                    Aria2 的配置文件中开启 RPC
                                    服务。更多信息及指引请参考文档的{" "}
                                    <Link
                                        href={
                                            "https://docs.cloudreve.org/use/aria2"
                                        }
                                        target={"_blank"}
                                    >
                                        离线下载
                                    </Link>{" "}
                                    章节。
                                </Typography>
                            </Alert>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    RPC 服务器地址
                                </InputLabel>
                                <Input
                                    type={"url"}
                                    value={options.aria2_rpcurl}
                                    onChange={handleChange("aria2_rpcurl")}
                                />
                                <FormHelperText id="component-helper-text">
                                    包含端口的完整 RPC
                                    服务器地址，例如：http://127.0.0.1:6800/，留空表示不启用
                                    Aria2 服务
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    RPC Secret
                                </InputLabel>
                                <Input
                                    value={options.aria2_token}
                                    onChange={handleChange("aria2_token")}
                                />
                                <FormHelperText id="component-helper-text">
                                    RPC 授权令牌，与 Aria2
                                    配置文件中保持一致，未设置请留空。
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    临时下载目录
                                </InputLabel>
                                <Input
                                    value={options.aria2_temp_path}
                                    onChange={handleChange("aria2_temp_path")}
                                />
                                <FormHelperText id="component-helper-text">
                                    离线下载临时下载目录的
                                    <strong>绝对路径</strong>，Cloudreve
                                    进程需要此目录的读、写、执行权限。
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    状态刷新间隔 (秒)
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        step: 1,
                                        min: 1,
                                    }}
                                    required
                                    value={options.aria2_interval}
                                    onChange={handleChange("aria2_interval")}
                                />
                                <FormHelperText id="component-helper-text">
                                    Cloudreve 向 Aria2 请求刷新任务状态的间隔。
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    RPC 调用超时 (秒)
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        step: 1,
                                        min: 1,
                                    }}
                                    required
                                    value={options.aria2_call_timeout}
                                    onChange={handleChange(
                                        "aria2_call_timeout"
                                    )}
                                />
                                <FormHelperText id="component-helper-text">
                                    调用 RPC 服务时最长等待时间
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    全局任务参数
                                </InputLabel>
                                <Input
                                    multiline
                                    required
                                    value={options.aria2_options}
                                    onChange={handleChange("aria2_options")}
                                />
                                <FormHelperText id="component-helper-text">
                                    创建下载任务时携带的额外设置参数，以 JSON
                                    编码后的格式书写，您可也可以将这些设置写在
                                    Aria2 配置文件里，可用参数请查阅官方文档
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
                    <Button
                        style={{ marginLeft: 8 }}
                        disabled={loading}
                        onClick={() => test()}
                        variant={"outlined"}
                        color={"secondary"}
                    >
                        测试连接
                    </Button>
                </div>
            </form>
        </div>
    );
}
