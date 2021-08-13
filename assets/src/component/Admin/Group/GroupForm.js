import Button from "@material-ui/core/Button";
import Collapse from "@material-ui/core/Collapse";
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import FormHelperText from "@material-ui/core/FormHelperText";
import Input from "@material-ui/core/Input";
import InputLabel from "@material-ui/core/InputLabel";
import MenuItem from "@material-ui/core/MenuItem";
import Select from "@material-ui/core/Select";
import { makeStyles } from "@material-ui/core/styles";
import Switch from "@material-ui/core/Switch";
import Typography from "@material-ui/core/Typography";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { useHistory } from "react-router";
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

// function getStyles(name, personName, theme) {
//     return {
//         fontWeight:
//             personName.indexOf(name) === -1
//                 ? theme.typography.fontWeightRegular
//                 : theme.typography.fontWeightMedium
//     };
// }

export default function GroupForm(props) {
    const classes = useStyles();
    const [loading, setLoading] = useState(false);
    const [group, setGroup] = useState(
        props.group
            ? props.group
            : {
                  ID: 0,
                  Name: "",
                  MaxStorage: "1073741824", // 转换类型
                  ShareEnabled: "true", // 转换类型
                  WebDAVEnabled: "true", // 转换类型
                  SpeedLimit: "0", // 转换类型
                  PolicyList: 1, // 转换类型,至少选择一个
                  OptionsSerialized: {
                      // 批量转换类型
                      share_download: "true",
                      aria2_options: "{}", // json decode
                      compress_size: "0",
                      decompress_size: "0",
                  },
              }
    );
    const [policies, setPolicies] = useState({});

    const history = useHistory();

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    useEffect(() => {
        API.post("/admin/policy/list", {
            page: 1,
            page_size: 10000,
            order_by: "id asc",
            conditions: {},
        })
            .then((response) => {
                const res = {};
                response.data.items.forEach((v) => {
                    res[v.ID] = v.Name;
                });
                setPolicies(res);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    }, []);

    const handleChange = (name) => (event) => {
        setGroup({
            ...group,
            [name]: event.target.value,
        });
    };

    const handleCheckChange = (name) => (event) => {
        const value = event.target.checked ? "true" : "false";
        setGroup({
            ...group,
            [name]: value,
        });
    };

    const handleOptionCheckChange = (name) => (event) => {
        const value = event.target.checked ? "true" : "false";
        setGroup({
            ...group,
            OptionsSerialized: {
                ...group.OptionsSerialized,
                [name]: value,
            },
        });
    };

    const handleOptionChange = (name) => (event) => {
        setGroup({
            ...group,
            OptionsSerialized: {
                ...group.OptionsSerialized,
                [name]: event.target.value,
            },
        });
    };

    const submit = (e) => {
        e.preventDefault();
        const groupCopy = {
            ...group,
            OptionsSerialized: { ...group.OptionsSerialized },
        };

        // 布尔值转换
        ["ShareEnabled", "WebDAVEnabled"].forEach((v) => {
            groupCopy[v] = groupCopy[v] === "true";
        });
        [
            "archive_download",
            "archive_task",
            "one_time_download",
            "share_download",
            "aria2",
        ].forEach((v) => {
            if (groupCopy.OptionsSerialized[v] !== undefined) {
                groupCopy.OptionsSerialized[v] =
                    groupCopy.OptionsSerialized[v] === "true";
            }
        });

        // 整型转换
        ["MaxStorage", "SpeedLimit"].forEach((v) => {
            groupCopy[v] = parseInt(groupCopy[v]);
        });
        ["compress_size", "decompress_size"].forEach((v) => {
            if (groupCopy.OptionsSerialized[v] !== undefined) {
                groupCopy.OptionsSerialized[v] = parseInt(
                    groupCopy.OptionsSerialized[v]
                );
            }
        });
        groupCopy.PolicyList = [parseInt(groupCopy.PolicyList)];
        // JSON转换
        try {
            groupCopy.OptionsSerialized.aria2_options = JSON.parse(
                groupCopy.OptionsSerialized.aria2_options
            );
        } catch (e) {
            ToggleSnackbar("top", "right", "Aria2 设置项格式错误", "warning");
            return;
        }

        setLoading(true);
        API.post("/admin/group", {
            group: groupCopy,
        })
            .then(() => {
                history.push("/admin/group");
                ToggleSnackbar(
                    "top",
                    "right",
                    "用户组已" + (props.group ? "保存" : "添加"),
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

    return (
        <div>
            <form onSubmit={submit}>
                <div className={classes.root}>
                    <Typography variant="h6" gutterBottom>
                        {group.ID === 0 && "新建用户组"}
                        {group.ID !== 0 && "编辑 " + group.Name}
                    </Typography>

                    <div className={classes.formContainer}>
                        {group.ID !== 3 && (
                            <>
                                <div className={classes.form}>
                                    <FormControl fullWidth>
                                        <InputLabel htmlFor="component-helper">
                                            用户组名
                                        </InputLabel>
                                        <Input
                                            value={group.Name}
                                            onChange={handleChange("Name")}
                                            required
                                        />
                                        <FormHelperText id="component-helper-text">
                                            用户组的名称
                                        </FormHelperText>
                                    </FormControl>
                                </div>

                                <div className={classes.form}>
                                    <FormControl fullWidth>
                                        <InputLabel htmlFor="component-helper">
                                            存储策略
                                        </InputLabel>
                                        <Select
                                            labelId="demo-mutiple-chip-label"
                                            id="demo-mutiple-chip"
                                            value={group.PolicyList}
                                            onChange={handleChange(
                                                "PolicyList"
                                            )}
                                            input={
                                                <Input id="select-multiple-chip" />
                                            }
                                        >
                                            {Object.keys(policies).map(
                                                (pid) => (
                                                    <MenuItem
                                                        key={pid}
                                                        value={pid}
                                                    >
                                                        {policies[pid]}
                                                    </MenuItem>
                                                )
                                            )}
                                        </Select>
                                        <FormHelperText id="component-helper-text">
                                            指定用户组的存储策略。
                                        </FormHelperText>
                                    </FormControl>
                                </div>

                                <div className={classes.form}>
                                    <FormControl fullWidth>
                                        <SizeInput
                                            value={group.MaxStorage}
                                            onChange={handleChange(
                                                "MaxStorage"
                                            )}
                                            min={0}
                                            max={9223372036854775807}
                                            label={"初始容量"}
                                            required
                                        />
                                    </FormControl>
                                    <FormHelperText id="component-helper-text">
                                        用户组下的用户初始可用最大容量
                                    </FormHelperText>
                                </div>
                            </>
                        )}

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <SizeInput
                                    value={group.SpeedLimit}
                                    onChange={handleChange("SpeedLimit")}
                                    min={0}
                                    max={9223372036854775807}
                                    label={"下载限速"}
                                    suffix={"/s"}
                                    required
                                />
                            </FormControl>
                            <FormHelperText id="component-helper-text">
                                填写为 0 表示不限制。开启限制后，
                                此用户组下的用户下载所有支持限速的存储策略下的文件时，下载最大速度会被限制。
                            </FormHelperText>
                        </div>

                        {group.ID !== 3 && (
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <FormControlLabel
                                        control={
                                            <Switch
                                                checked={
                                                    group.ShareEnabled ===
                                                    "true"
                                                }
                                                onChange={handleCheckChange(
                                                    "ShareEnabled"
                                                )}
                                            />
                                        }
                                        label="允许创建分享"
                                    />
                                    <FormHelperText id="component-helper-text">
                                        关闭后，用户无法创建分享链接
                                    </FormHelperText>
                                </FormControl>
                            </div>
                        )}

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={
                                                group.OptionsSerialized
                                                    .share_download === "true"
                                            }
                                            onChange={handleOptionCheckChange(
                                                "share_download"
                                            )}
                                        />
                                    }
                                    label="允许下载分享"
                                />
                                <FormHelperText id="component-helper-text">
                                    关闭后，用户无法下载别人创建的文件分享
                                </FormHelperText>
                            </FormControl>
                        </div>

                        {group.ID !== 3 && (
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <FormControlLabel
                                        control={
                                            <Switch
                                                checked={
                                                    group.WebDAVEnabled ===
                                                    "true"
                                                }
                                                onChange={handleCheckChange(
                                                    "WebDAVEnabled"
                                                )}
                                            />
                                        }
                                        label="WebDAV"
                                    />
                                    <FormHelperText id="component-helper-text">
                                        关闭后，用户无法通过 WebDAV
                                        协议连接至网盘
                                    </FormHelperText>
                                </FormControl>
                            </div>
                        )}

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={
                                                group.OptionsSerialized
                                                    .one_time_download ===
                                                "true"
                                            }
                                            onChange={handleOptionCheckChange(
                                                "one_time_download"
                                            )}
                                        />
                                    }
                                    label="禁止多次下载请求"
                                />
                                <FormHelperText id="component-helper-text">
                                    只针对本机存储策略有效。开启后，用户无法使用多线程下载工具。
                                </FormHelperText>
                            </FormControl>
                        </div>

                        {group.ID !== 3 && (
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <FormControlLabel
                                        control={
                                            <Switch
                                                checked={
                                                    group.OptionsSerialized
                                                        .aria2 === "true"
                                                }
                                                onChange={handleOptionCheckChange(
                                                    "aria2"
                                                )}
                                            />
                                        }
                                        label="离线下载"
                                    />
                                    <FormHelperText id="component-helper-text">
                                        是否允许用户创建离线下载任务
                                    </FormHelperText>
                                </FormControl>
                            </div>
                        )}

                        <Collapse in={group.OptionsSerialized.aria2 === "true"}>
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <InputLabel htmlFor="component-helper">
                                        Aria2 任务参数
                                    </InputLabel>
                                    <Input
                                        multiline
                                        value={
                                            group.OptionsSerialized
                                                .aria2_options
                                        }
                                        onChange={handleOptionChange(
                                            "aria2_options"
                                        )}
                                    />
                                    <FormHelperText id="component-helper-text">
                                        此用户组创建离线下载任务时额外携带的参数，以
                                        JSON
                                        编码后的格式书写，您可也可以将这些设置写在
                                        Aria2 配置文件里，可用参数请查阅官方文档
                                    </FormHelperText>
                                </FormControl>
                            </div>
                        </Collapse>

                        <div className={classes.form}>
                            <FormControl fullWidth>
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={
                                                group.OptionsSerialized
                                                    .archive_download === "true"
                                            }
                                            onChange={handleOptionCheckChange(
                                                "archive_download"
                                            )}
                                        />
                                    }
                                    label="打包下载"
                                />
                                <FormHelperText id="component-helper-text">
                                    是否允许用户多选文件打包下载
                                </FormHelperText>
                            </FormControl>
                        </div>

                        {group.ID !== 3 && (
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <FormControlLabel
                                        control={
                                            <Switch
                                                checked={
                                                    group.OptionsSerialized
                                                        .archive_task === "true"
                                                }
                                                onChange={handleOptionCheckChange(
                                                    "archive_task"
                                                )}
                                            />
                                        }
                                        label="压缩/解压缩 任务"
                                    />
                                    <FormHelperText id="component-helper-text">
                                        是否用户创建 压缩/解压缩 任务
                                    </FormHelperText>
                                </FormControl>
                            </div>
                        )}

                        <Collapse
                            in={group.OptionsSerialized.archive_task === "true"}
                        >
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <SizeInput
                                        value={
                                            group.OptionsSerialized
                                                .compress_size
                                        }
                                        onChange={handleOptionChange(
                                            "compress_size"
                                        )}
                                        min={0}
                                        max={9223372036854775807}
                                        label={"待压缩文件最大大小"}
                                    />
                                </FormControl>
                                <FormHelperText id="component-helper-text">
                                    用户可创建的压缩任务的文件最大总大小，填写为
                                    0 表示不限制
                                </FormHelperText>
                            </div>

                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <SizeInput
                                        value={
                                            group.OptionsSerialized
                                                .decompress_size
                                        }
                                        onChange={handleOptionChange(
                                            "decompress_size"
                                        )}
                                        min={0}
                                        max={9223372036854775807}
                                        label={"待解压文件最大大小"}
                                    />
                                </FormControl>
                                <FormHelperText id="component-helper-text">
                                    用户可创建的解压缩任务的文件最大总大小，填写为
                                    0 表示不限制
                                </FormHelperText>
                            </div>
                        </Collapse>
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
