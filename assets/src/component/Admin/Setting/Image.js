import Button from "@material-ui/core/Button";
import FormControl from "@material-ui/core/FormControl";
import FormHelperText from "@material-ui/core/FormHelperText";
import Input from "@material-ui/core/Input";
import InputLabel from "@material-ui/core/InputLabel";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../../actions";
import API from "../../../middleware/Api";
import SizeInput from "../Common/SizeInput";

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

export default function ImageSetting() {
    const classes = useStyles();
    const [loading, setLoading] = useState(false);
    const [options, setOptions] = useState({
        gravatar_server: "",
        avatar_path: "",
        avatar_size: "",
        avatar_size_l: "",
        avatar_size_m: "",
        avatar_size_s: "",
        thumb_width: "",
        thumb_height: "",
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
                        头像
                    </Typography>
                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    Gravatar 服务器
                                </InputLabel>
                                <Input
                                    type={"url"}
                                    value={options.gravatar_server}
                                    onChange={handleChange("gravatar_server")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    Gravatar 服务器地址，可选择使用国内镜像
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    头像存储路径
                                </InputLabel>
                                <Input
                                    value={options.avatar_path}
                                    onChange={handleChange("avatar_path")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    用户上传自定义头像的存储路径
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <SizeInput
                                    value={options.avatar_size}
                                    onChange={handleChange("avatar_size")}
                                    required
                                    min={0}
                                    max={2147483647}
                                    label={"头像文件大小限制"}
                                />
                                <FormHelperText id="component-helper-text">
                                    用户可上传头像文件的最大大小
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl>
                                <InputLabel htmlFor="component-helper">
                                    小头像尺寸
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        min: 1,
                                        step: 1,
                                    }}
                                    value={options.avatar_size_s}
                                    onChange={handleChange("avatar_size_s")}
                                    required
                                />
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl>
                                <InputLabel htmlFor="component-helper">
                                    中头像尺寸
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        min: 1,
                                        step: 1,
                                    }}
                                    value={options.avatar_size_m}
                                    onChange={handleChange("avatar_size_m")}
                                    required
                                />
                            </FormControl>
                        </div>

                        <div className={classes.form}>
                            <FormControl>
                                <InputLabel htmlFor="component-helper">
                                    大头像尺寸
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        min: 1,
                                        step: 1,
                                    }}
                                    value={options.avatar_size_l}
                                    onChange={handleChange("avatar_size_l")}
                                    required
                                />
                            </FormControl>
                        </div>
                    </div>
                </div>

                <div className={classes.root}>
                    <Typography variant="h6" gutterBottom>
                        缩略图
                    </Typography>

                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl>
                                <InputLabel htmlFor="component-helper">
                                    宽度
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        min: 1,
                                        step: 1,
                                    }}
                                    value={options.thumb_width}
                                    onChange={handleChange("thumb_width")}
                                    required
                                />
                            </FormControl>
                        </div>
                    </div>

                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl>
                                <InputLabel htmlFor="component-helper">
                                    高度
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        min: 1,
                                        step: 1,
                                    }}
                                    value={options.thumb_height}
                                    onChange={handleChange("thumb_height")}
                                    required
                                />
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
                </div>
            </form>
        </div>
    );
}
