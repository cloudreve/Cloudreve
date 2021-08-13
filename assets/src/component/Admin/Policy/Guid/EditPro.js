import Button from "@material-ui/core/Button";
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import Input from "@material-ui/core/Input";
import Radio from "@material-ui/core/Radio";
import RadioGroup from "@material-ui/core/RadioGroup";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import Typography from "@material-ui/core/Typography";
import React, { useCallback, useState } from "react";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../../../actions";
import API from "../../../../middleware/Api";

export default function EditPro(props) {
    const [, setLoading] = useState(false);
    const [policy, setPolicy] = useState(props.policy);

    const handleChange = (name) => (event) => {
        setPolicy({
            ...policy,
            [name]: event.target.value,
        });
    };

    const handleOptionChange = (name) => (event) => {
        setPolicy({
            ...policy,
            OptionsSerialized: {
                ...policy.OptionsSerialized,
                [name]: event.target.value,
            },
        });
    };

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const submitPolicy = (e) => {
        e.preventDefault();
        setLoading(true);

        const policyCopy = { ...policy };
        policyCopy.OptionsSerialized = { ...policyCopy.OptionsSerialized };

        // 类型转换
        policyCopy.AutoRename = policyCopy.AutoRename === "true";
        policyCopy.IsPrivate = policyCopy.IsPrivate === "true";
        policyCopy.IsOriginLinkEnable =
            policyCopy.IsOriginLinkEnable === "true";
        policyCopy.MaxSize = parseInt(policyCopy.MaxSize);
        policyCopy.OptionsSerialized.file_type = policyCopy.OptionsSerialized.file_type.split(
            ","
        );
        if (
            policyCopy.OptionsSerialized.file_type.length === 1 &&
            policyCopy.OptionsSerialized.file_type[0] === ""
        ) {
            policyCopy.OptionsSerialized.file_type = [];
        }

        API.post("/admin/policy", {
            policy: policyCopy,
        })
            .then(() => {
                ToggleSnackbar(
                    "top",
                    "right",
                    "存储策略已" + (props.policy ? "保存" : "添加"),
                    "success"
                );
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });

        setLoading(false);
    };

    return (
        <div>
            <Typography variant={"h6"}>编辑存储策略</Typography>
            <TableContainer>
                <form onSubmit={submitPolicy}>
                    <Table aria-label="simple table">
                        <TableHead>
                            <TableRow>
                                <TableCell>设置项</TableCell>
                                <TableCell>值</TableCell>
                                <TableCell>描述</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    ID
                                </TableCell>
                                <TableCell>{policy.ID}</TableCell>
                                <TableCell>存储策略编号</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    类型
                                </TableCell>
                                <TableCell>{policy.Type}</TableCell>
                                <TableCell>存储策略类型</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    名称
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            required
                                            value={policy.Name}
                                            onChange={handleChange("Name")}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>存储策名称</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    Server
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            value={policy.Server}
                                            onChange={handleChange("Server")}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>存储端 Endpoint</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    BucketName
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            value={policy.BucketName}
                                            onChange={handleChange(
                                                "BucketName"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>存储桶标识</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    私有空间
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <RadioGroup
                                            required
                                            value={policy.IsPrivate}
                                            onChange={handleChange("IsPrivate")}
                                            row
                                        >
                                            <FormControlLabel
                                                value={"true"}
                                                control={
                                                    <Radio color={"primary"} />
                                                }
                                                label="是"
                                            />
                                            <FormControlLabel
                                                value={"false"}
                                                control={
                                                    <Radio color={"primary"} />
                                                }
                                                label="否"
                                            />
                                        </RadioGroup>
                                    </FormControl>
                                </TableCell>
                                <TableCell>是否为私有空间</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    文件资源根URL
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            value={policy.BaseURL}
                                            onChange={handleChange("BaseURL")}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>
                                    预览/获取文件外链时生成URL的前缀
                                </TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    AccessKey
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            rowsMax={10}
                                            value={policy.AccessKey}
                                            onChange={handleChange("AccessKey")}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>AccessKey / 刷新Token</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    SecretKey
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            rowsMax={10}
                                            value={policy.SecretKey}
                                            onChange={handleChange("SecretKey")}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>SecretKey</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    最大单文件尺寸 (Bytes)
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            type={"number"}
                                            inputProps={{
                                                min: 0,
                                                step: 1,
                                            }}
                                            value={policy.MaxSize}
                                            onChange={handleChange("MaxSize")}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>
                                    最大可上传的文件尺寸，填写为0表示不限制
                                </TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    自动重命名
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <RadioGroup
                                            required
                                            value={policy.AutoRename}
                                            onChange={handleChange(
                                                "AutoRename"
                                            )}
                                            row
                                        >
                                            <FormControlLabel
                                                value={"true"}
                                                control={
                                                    <Radio color={"primary"} />
                                                }
                                                label="是"
                                            />
                                            <FormControlLabel
                                                value={"false"}
                                                control={
                                                    <Radio color={"primary"} />
                                                }
                                                label="否"
                                            />
                                        </RadioGroup>
                                    </FormControl>
                                </TableCell>
                                <TableCell>
                                    是否根据规则对上传物理文件重命名
                                </TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    存储路径
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={policy.DirNameRule}
                                            onChange={handleChange(
                                                "DirNameRule"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>文件物理存储路径</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    存储文件名
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={policy.FileNameRule}
                                            onChange={handleChange(
                                                "FileNameRule"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>文件物理存储文件名</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    允许获取外链
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <RadioGroup
                                            required
                                            value={policy.IsOriginLinkEnable}
                                            onChange={handleChange(
                                                "IsOriginLinkEnable"
                                            )}
                                            row
                                        >
                                            <FormControlLabel
                                                value={"true"}
                                                control={
                                                    <Radio color={"primary"} />
                                                }
                                                label="是"
                                            />
                                            <FormControlLabel
                                                value={"false"}
                                                control={
                                                    <Radio color={"primary"} />
                                                }
                                                label="否"
                                            />
                                        </RadioGroup>
                                    </FormControl>
                                </TableCell>
                                <TableCell>
                                    是否允许获取外链。注意，某些存储策略类型不支持，即使在此开启，获取的外链也无法使用。
                                </TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    又拍云防盗链 Token
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={
                                                policy.OptionsSerialized.token
                                            }
                                            onChange={handleOptionChange(
                                                "token"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>仅对又拍云存储策略有效</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    允许文件扩展名
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={
                                                policy.OptionsSerialized
                                                    .file_type
                                            }
                                            onChange={handleOptionChange(
                                                "file_type"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>留空表示不限制</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    允许的 MimeType
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={
                                                policy.OptionsSerialized
                                                    .mimetype
                                            }
                                            onChange={handleOptionChange(
                                                "mimetype"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>仅对七牛存储策略有效</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    OneDrive 重定向地址
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={
                                                policy.OptionsSerialized
                                                    .od_redirect
                                            }
                                            onChange={handleOptionChange(
                                                "od_redirect"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>一般添加后无需修改</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    OneDrive 反代服务器地址
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={
                                                policy.OptionsSerialized
                                                    .od_proxy
                                            }
                                            onChange={handleOptionChange(
                                                "od_proxy"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>
                                    仅对 OneDrive 存储策略有效
                                </TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    OneDrive/SharePoint 驱动器资源标识
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={
                                                policy.OptionsSerialized
                                                    .od_driver
                                            }
                                            onChange={handleOptionChange(
                                                "od_driver"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>
                                    仅对 OneDrive
                                    存储策略有效，留空则使用用户的默认 OneDrive
                                    驱动器
                                </TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    Amazon S3 Region
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={
                                                policy.OptionsSerialized.region
                                            }
                                            onChange={handleOptionChange(
                                                "region"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>
                                    仅对 Amazon S3 存储策略有效
                                </TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    内网 EndPoint
                                </TableCell>
                                <TableCell>
                                    <FormControl>
                                        <Input
                                            multiline
                                            value={
                                                policy.OptionsSerialized
                                                    .server_side_endpoint
                                            }
                                            onChange={handleOptionChange(
                                                "server_side_endpoint"
                                            )}
                                        />
                                    </FormControl>
                                </TableCell>
                                <TableCell>仅对 OSS 存储策略有效</TableCell>
                            </TableRow>
                        </TableBody>
                    </Table>
                    <Button
                        type={"submit"}
                        color={"primary"}
                        variant={"contained"}
                        style={{ margin: 8 }}
                    >
                        保存更改
                    </Button>
                </form>
            </TableContainer>
        </div>
    );
}
