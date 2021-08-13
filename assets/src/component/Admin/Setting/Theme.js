import Button from "@material-ui/core/Button";
import FormControl from "@material-ui/core/FormControl";
import FormHelperText from "@material-ui/core/FormHelperText";
import IconButton from "@material-ui/core/IconButton";
import InputLabel from "@material-ui/core/InputLabel";
import Link from "@material-ui/core/Link";
import MenuItem from "@material-ui/core/MenuItem";
import Select from "@material-ui/core/Select";
import { makeStyles } from "@material-ui/core/styles";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { Delete } from "@material-ui/icons";
import Alert from "@material-ui/lab/Alert";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../../actions";
import API from "../../../middleware/Api";
import CreateTheme from "../Dialogs/CreateTheme";

const useStyles = makeStyles((theme) => ({
    root: {
        [theme.breakpoints.up("md")]: {
            marginLeft: 100,
        },
        marginBottom: 40,
    },
    form: {
        maxWidth: 500,
        marginTop: 20,
        marginBottom: 20,
    },
    formContainer: {
        [theme.breakpoints.up("md")]: {
            padding: "0px 24px 0 24px",
        },
    },
    colorContainer: {
        display: "flex",
    },
    colorDot: {
        width: 20,
        height: 20,
        borderRadius: "50%",
        marginLeft: 6,
    },
}));

export default function Theme() {
    const classes = useStyles();
    const [loading, setLoading] = useState(false);
    const [theme, setTheme] = useState({});
    const [options, setOptions] = useState({
        themes: "{}",
        defaultTheme: "",
        home_view_method: "icon",
        share_view_method: "list",
    });
    const [themeConfig, setThemeConfig] = useState({});
    const [themeConfigError, setThemeConfigError] = useState({});
    const [create, setCreate] = useState(false);

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const deleteTheme = (color) => {
        if (color === options.defaultTheme) {
            ToggleSnackbar("top", "right", "不能删除默认配色", "warning");
            return;
        }
        if (Object.keys(theme).length <= 1) {
            ToggleSnackbar("top", "right", "请至少保留一个配色方案", "warning");
            return;
        }
        const themeCopy = { ...theme };
        delete themeCopy[color];
        const resStr = JSON.stringify(themeCopy);
        setOptions({
            ...options,
            themes: resStr,
        });
    };

    const addTheme = (newTheme) => {
        setCreate(false);
        if (theme[newTheme.palette.primary.main] !== undefined) {
            ToggleSnackbar(
                "top",
                "right",
                "主色调不能与已有配色重复",
                "warning"
            );
            return;
        }
        const res = {
            ...theme,
            [newTheme.palette.primary.main]: newTheme,
        };
        const resStr = JSON.stringify(res);
        setOptions({
            ...options,
            themes: resStr,
        });
    };

    useEffect(() => {
        const res = JSON.parse(options.themes);
        const themeString = {};

        Object.keys(res).forEach((k) => {
            themeString[k] = JSON.stringify(res[k]);
        });

        setTheme(res);
        setThemeConfig(themeString);
    }, [options.themes]);

    const handleChange = (name) => (event) => {
        setOptions({
            ...options,
            [name]: event.target.value,
        });
    };

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
                        主题配色
                    </Typography>
                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <Table aria-label="simple table">
                                <TableHead>
                                    <TableRow>
                                        <TableCell>关键色</TableCell>
                                        <TableCell>色彩配置</TableCell>
                                        <TableCell>操作</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {Object.keys(theme).map((k) => (
                                        <TableRow key={k}>
                                            <TableCell
                                                component="th"
                                                scope="row"
                                            >
                                                <div
                                                    className={
                                                        classes.colorContainer
                                                    }
                                                >
                                                    <div
                                                        style={{
                                                            backgroundColor:
                                                                theme[k].palette
                                                                    .primary
                                                                    .main,
                                                        }}
                                                        className={
                                                            classes.colorDot
                                                        }
                                                    />
                                                    <div
                                                        style={{
                                                            backgroundColor:
                                                                theme[k].palette
                                                                    .secondary
                                                                    .main,
                                                        }}
                                                        className={
                                                            classes.colorDot
                                                        }
                                                    />
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <TextField
                                                    error={themeConfigError[k]}
                                                    helperText={
                                                        themeConfigError[k] &&
                                                        "格式不正确"
                                                    }
                                                    fullWidth
                                                    multiline
                                                    onChange={(e) => {
                                                        setThemeConfig({
                                                            ...themeConfig,
                                                            [k]: e.target.value,
                                                        });
                                                    }}
                                                    onBlur={(e) => {
                                                        try {
                                                            const res = JSON.parse(
                                                                e.target.value
                                                            );
                                                            if (
                                                                !(
                                                                    "palette" in
                                                                    res
                                                                ) ||
                                                                !(
                                                                    "primary" in
                                                                    res.palette
                                                                ) ||
                                                                !(
                                                                    "main" in
                                                                    res.palette
                                                                        .primary
                                                                ) ||
                                                                !(
                                                                    "secondary" in
                                                                    res.palette
                                                                ) ||
                                                                !(
                                                                    "main" in
                                                                    res.palette
                                                                        .secondary
                                                                )
                                                            ) {
                                                                throw e;
                                                            }
                                                            setTheme({
                                                                ...theme,
                                                                [k]: res,
                                                            });
                                                        } catch (e) {
                                                            setThemeConfigError(
                                                                {
                                                                    ...themeConfigError,
                                                                    [k]: true,
                                                                }
                                                            );
                                                            return;
                                                        }
                                                        setThemeConfigError({
                                                            ...themeConfigError,
                                                            [k]: false,
                                                        });
                                                    }}
                                                    value={themeConfig[k]}
                                                />
                                            </TableCell>
                                            <TableCell>
                                                <IconButton
                                                    onClick={() =>
                                                        deleteTheme(k)
                                                    }
                                                >
                                                    <Delete />
                                                </IconButton>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                            <div>
                                <Button
                                    variant="outlined"
                                    color="primary"
                                    style={{ marginTop: 8 }}
                                    onClick={() => setCreate(true)}
                                >
                                    新建配色方案
                                </Button>
                            </div>
                            <Alert severity="info" style={{ marginTop: 8 }}>
                                <Typography variant="body2">
                                    完整的配置项可在{" "}
                                    <Link
                                        href={
                                            "https://material-ui.com/zh/customization/default-theme/"
                                        }
                                        target={"_blank"}
                                    >
                                        默认主题 - Material-UI
                                    </Link>{" "}
                                    查阅。
                                </Typography>
                            </Alert>
                        </div>

                        <div className={classes.form}>
                            <FormControl>
                                <InputLabel htmlFor="component-helper">
                                    默认配色
                                </InputLabel>
                                <Select
                                    value={options.defaultTheme}
                                    onChange={handleChange("defaultTheme")}
                                >
                                    {Object.keys(theme).map((k) => (
                                        <MenuItem key={k} value={k}>
                                            <div
                                                className={
                                                    classes.colorContainer
                                                }
                                            >
                                                <div
                                                    style={{
                                                        backgroundColor:
                                                            theme[k].palette
                                                                .primary.main,
                                                    }}
                                                    className={classes.colorDot}
                                                />
                                                <div
                                                    style={{
                                                        backgroundColor:
                                                            theme[k].palette
                                                                .secondary.main,
                                                    }}
                                                    className={classes.colorDot}
                                                />
                                            </div>
                                        </MenuItem>
                                    ))}
                                </Select>
                                <FormHelperText id="component-helper-text">
                                    用户未指定偏好配色时，站点默认使用的配色方案
                                </FormHelperText>
                            </FormControl>
                        </div>
                    </div>
                </div>

                <div className={classes.root}>
                    <Typography variant="h6" gutterBottom>
                        界面
                    </Typography>

                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl>
                                <InputLabel htmlFor="component-helper">
                                    个人文件列表默认样式
                                </InputLabel>
                                <Select
                                    value={options.home_view_method}
                                    onChange={handleChange("home_view_method")}
                                    required
                                >
                                    <MenuItem value={"icon"}>大图标</MenuItem>
                                    <MenuItem value={"smallIcon"}>
                                        小图标
                                    </MenuItem>
                                    <MenuItem value={"list"}>列表</MenuItem>
                                </Select>
                                <FormHelperText id="component-helper-text">
                                    用户未指定偏好样式时，个人文件页面列表默认样式
                                </FormHelperText>
                            </FormControl>
                        </div>
                    </div>

                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl>
                                <InputLabel htmlFor="component-helper">
                                    目录分享页列表默认样式
                                </InputLabel>
                                <Select
                                    value={options.share_view_method}
                                    onChange={handleChange("share_view_method")}
                                    required
                                >
                                    <MenuItem value={"icon"}>大图标</MenuItem>
                                    <MenuItem value={"smallIcon"}>
                                        小图标
                                    </MenuItem>
                                    <MenuItem value={"list"}>列表</MenuItem>
                                </Select>
                                <FormHelperText id="component-helper-text">
                                    用户未指定偏好样式时，目录分享页面的默认样式
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
                </div>
            </form>

            <CreateTheme
                onSubmit={addTheme}
                open={create}
                onClose={() => setCreate(false)}
            />
        </div>
    );
}
