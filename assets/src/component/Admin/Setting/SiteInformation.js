import Button from "@material-ui/core/Button";
import FormControl from "@material-ui/core/FormControl";
import FormHelperText from "@material-ui/core/FormHelperText";
import Input from "@material-ui/core/Input";
import InputLabel from "@material-ui/core/InputLabel";
import MenuItem from "@material-ui/core/MenuItem";
import Select from "@material-ui/core/Select";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
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

export default function SiteInformation() {
    const classes = useStyles();
    const [loading, setLoading] = useState(false);
    const [options, setOptions] = useState({
        siteURL: "",
        siteName: "",
        siteTitle: "",
        siteDes: "",
        siteICPId: "",
        siteScript: "",
        pwa_small_icon: "",
        pwa_medium_icon: "",
        pwa_large_icon: "",
        pwa_display: "",
        pwa_theme_color: "",
        pwa_background_color: "",
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
                        基本信息
                    </Typography>
                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    主标题
                                </InputLabel>
                                <Input
                                    value={options.siteName}
                                    onChange={handleChange("siteName")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    站点的主标题
                                </FormHelperText>
                            </FormControl>
                        </div>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    副标题
                                </InputLabel>
                                <Input
                                    value={options.siteTitle}
                                    onChange={handleChange("siteTitle")}
                                />
                                <FormHelperText id="component-helper-text">
                                    站点的副标题
                                </FormHelperText>
                            </FormControl>
                        </div>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    站点描述
                                </InputLabel>
                                <Input
                                    value={options.siteDes}
                                    onChange={handleChange("siteDes")}
                                />
                                <FormHelperText id="component-helper-text">
                                    站点描述信息，可能会在分享页面摘要内展示
                                </FormHelperText>
                            </FormControl>
                        </div>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    站点URL
                                </InputLabel>
                                <Input
                                    type={"url"}
                                    value={options.siteURL}
                                    onChange={handleChange("siteURL")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    非常重要，请确保与实际情况一致。使用云存储策略、支付平台时，请填入可以被外网访问的地址。
                                </FormHelperText>
                            </FormControl>
                        </div>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    网站备案号
                                </InputLabel>
                                <Input
                                    value={options.siteICPId}
                                    onChange={handleChange("siteICPId")}
                                />
                                <FormHelperText id="component-helper-text">
                                    工信部网站ICP备案号
                                </FormHelperText>
                            </FormControl>
                        </div>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    页脚代码
                                </InputLabel>
                                <Input
                                    multiline
                                    value={options.siteScript}
                                    onChange={handleChange("siteScript")}
                                />
                                <FormHelperText id="component-helper-text">
                                    在页面底部插入的自定义HTML代码
                                </FormHelperText>
                            </FormControl>
                        </div>
                    </div>
                </div>
                <div className={classes.root}>
                    <Typography variant="h6" gutterBottom>
                        渐进式应用 (PWA)
                    </Typography>
                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    小图标
                                </InputLabel>
                                <Input
                                    value={options.pwa_small_icon}
                                    onChange={handleChange("pwa_small_icon")}
                                />
                                <FormHelperText id="component-helper-text">
                                    扩展名为 ico 的小图标地址
                                </FormHelperText>
                            </FormControl>
                        </div>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    中图标
                                </InputLabel>
                                <Input
                                    value={options.pwa_medium_icon}
                                    onChange={handleChange("pwa_medium_icon")}
                                />
                                <FormHelperText id="component-helper-text">
                                    192x192 的中等图标地址，png 格式
                                </FormHelperText>
                            </FormControl>
                        </div>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    大图标
                                </InputLabel>
                                <Input
                                    value={options.pwa_large_icon}
                                    onChange={handleChange("pwa_large_icon")}
                                />
                                <FormHelperText id="component-helper-text">
                                    512x512 的大图标地址，png 格式
                                </FormHelperText>
                            </FormControl>
                        </div>
                        <div className={classes.form}>
                            <FormControl>
                                <InputLabel htmlFor="component-helper">
                                    展示模式
                                </InputLabel>
                                <Select
                                    value={options.pwa_display}
                                    onChange={handleChange("pwa_display")}
                                >
                                    <MenuItem value={"fullscreen"}>
                                        fullscreen
                                    </MenuItem>
                                    <MenuItem value={"standalone"}>
                                        standalone
                                    </MenuItem>
                                    <MenuItem value={"minimal-ui"}>
                                        minimal-ui
                                    </MenuItem>
                                    <MenuItem value={"browser"}>
                                        browser
                                    </MenuItem>
                                </Select>
                                <FormHelperText id="component-helper-text">
                                    PWA 应用添加后的展示模式
                                </FormHelperText>
                            </FormControl>
                        </div>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    主题色
                                </InputLabel>
                                <Input
                                    value={options.pwa_theme_color}
                                    onChange={handleChange("pwa_theme_color")}
                                />
                                <FormHelperText id="component-helper-text">
                                    CSS 色值，影响 PWA
                                    启动画面上状态栏、内容页中状态栏、地址栏的颜色
                                </FormHelperText>
                            </FormControl>
                        </div>
                    </div>
                    <div className={classes.formContainer}>
                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    背景色
                                </InputLabel>
                                <Input
                                    value={options.pwa_background_color}
                                    onChange={handleChange(
                                        "pwa_background_color"
                                    )}
                                />
                                <FormHelperText id="component-helper-text">
                                    CSS 色值
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
        </div>
    );
}
