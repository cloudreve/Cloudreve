import Button from "@material-ui/core/Button";
import Collapse from "@material-ui/core/Collapse";
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import Input from "@material-ui/core/Input";
import InputLabel from "@material-ui/core/InputLabel";
import Link from "@material-ui/core/Link";
import Radio from "@material-ui/core/Radio";
import RadioGroup from "@material-ui/core/RadioGroup";
import Step from "@material-ui/core/Step";
import StepLabel from "@material-ui/core/StepLabel";
import Stepper from "@material-ui/core/Stepper";
import { lighten, makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { useHistory } from "react-router";
import { toggleSnackbar } from "../../../../actions";
import API from "../../../../middleware/Api";
import SizeInput from "../../Common/SizeInput";
import AlertDialog from "../../Dialogs/Alert";
import MagicVar from "../../Dialogs/MagicVar";
import DomainInput from "../../Common/DomainInput";

const useStyles = makeStyles((theme) => ({
    stepContent: {
        padding: "16px 32px 16px 32px",
    },
    form: {
        maxWidth: 400,
        marginTop: 20,
    },
    formContainer: {
        [theme.breakpoints.up("md")]: {
            padding: "0px 24px 0 24px",
        },
    },
    subStepContainer: {
        display: "flex",
        marginBottom: 20,
        padding: 10,
        transition: theme.transitions.create("background-color", {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
        }),
        "&:focus-within": {
            backgroundColor: theme.palette.background.default,
        },
    },
    stepNumber: {
        width: 20,
        height: 20,
        backgroundColor: lighten(theme.palette.secondary.light, 0.2),
        color: theme.palette.secondary.contrastText,
        textAlign: "center",
        borderRadius: " 50%",
    },
    stepNumberContainer: {
        marginRight: 10,
    },
    stepFooter: {
        marginTop: 32,
    },
    button: {
        marginRight: theme.spacing(1),
    },
    viewButtonLabel: { textTransform: "none" },
    "@global": {
        code: {
            color: "rgba(0, 0, 0, 0.87)",
            display: "inline-block",
            padding: "2px 6px",
            fontFamily:
                ' Consolas, "Liberation Mono", Menlo, Courier, monospace',
            borderRadius: "2px",
            backgroundColor: "rgba(255,229,100,0.1)",
        },
    },
}));

const steps = [
    {
        title: "应用授权",
        optional: false,
    },
    {
        title: "上传路径",
        optional: false,
    },
    {
        title: "直链设置",
        optional: false,
    },
    {
        title: "上传限制",
        optional: false,
    },
    {
        title: "账号授权",
        optional: false,
    },
    {
        title: "完成",
        optional: false,
    },
];

export default function OneDriveGuide(props) {
    const classes = useStyles();
    const history = useHistory();

    const [activeStep, setActiveStep] = useState(0);
    const [loading, setLoading] = useState(false);
    const [skipped] = React.useState(new Set());
    const [magicVar, setMagicVar] = useState("");
    const [useCDN, setUseCDN] = useState(
        props.policy && props.policy.OptionsSerialized.od_proxy
            ? props.policy.OptionsSerialized.od_proxy !== ""
            : false
    );
    const [useSharePoint, setUseSharePoint] = useState(
        props.policy && props.policy.OptionsSerialized.od_driver
            ? props.policy.OptionsSerialized.od_driver !== ""
            : false
    );
    const [policy, setPolicy] = useState(
        props.policy
            ? props.policy
            : {
                  Type: "onedrive",
                  Name: "",
                  BucketName: "",
                  SecretKey: "",
                  AccessKey: "",
                  BaseURL: "",
                  Server: "https://graph.microsoft.com/v1.0",
                  IsPrivate: "true",
                  DirNameRule: "uploads/{year}/{month}/{day}",
                  AutoRename: "true",
                  FileNameRule: "{randomkey8}_{originname}",
                  IsOriginLinkEnable: "false",
                  MaxSize: "0",
                  OptionsSerialized: {
                      file_type: "",
                      od_redirect: "",
                      od_proxy: "",
                      od_driver: "",
                  },
              }
    );
    const [policyID, setPolicyID] = useState(
        props.policy ? props.policy.ID : 0
    );
    const [httpsAlert, setHttpsAlert] = useState(false);

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

    const isStepSkipped = (step) => {
        return skipped.has(step);
    };

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    useEffect(() => {
        API.post("/admin/setting", {
            keys: ["siteURL"],
        })
            .then((response) => {
                if (!response.data.siteURL.startsWith("https://")) {
                    setHttpsAlert(true);
                }
                if (policy.OptionsSerialized.od_redirect === "") {
                    setPolicy({
                        ...policy,
                        OptionsSerialized: {
                            ...policy.OptionsSerialized,
                            od_redirect: new URL(
                                "/api/v3/callback/onedrive/auth",
                                response.data.siteURL
                            ).toString(),
                        },
                    });
                }
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    }, []);

    const statOAuth = () => {
        setLoading(true);
        API.get("/admin/policy/" + policyID + "/oauth")
            .then((response) => {
                window.location.href = response.data;
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
                setLoading(false);
            });
    };

    const submitPolicy = (e) => {
        e.preventDefault();
        setLoading(true);

        const policyCopy = { ...policy };
        policyCopy.OptionsSerialized = { ...policyCopy.OptionsSerialized };

        // baseURL处理
        if (policyCopy.Server === "https://graph.microsoft.com/v1.0") {
            policyCopy.BaseURL =
                "https://login.microsoftonline.com/common/oauth2/v2.0";
        } else {
            policyCopy.BaseURL = "https://login.chinacloudapi.cn/common/oauth2";
        }

        if (!useCDN) {
            policyCopy.OptionsSerialized.od_proxy = "";
        }

        if (!useSharePoint) {
            policyCopy.OptionsSerialized.od_driver = "";
        }

        // 类型转换
        policyCopy.AutoRename = policyCopy.AutoRename === "true";
        policyCopy.IsOriginLinkEnable =
            policyCopy.IsOriginLinkEnable === "true";
        policyCopy.IsPrivate = policyCopy.IsPrivate === "true";
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
            .then((response) => {
                ToggleSnackbar(
                    "top",
                    "right",
                    "存储策略已" + (props.policy ? "保存" : "添加"),
                    "success"
                );
                setActiveStep(4);
                setPolicyID(response.data);
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
            <AlertDialog
                open={httpsAlert}
                onClose={() => setHttpsAlert(false)}
                title={"警告"}
                msg={
                    "您必须启用 HTTPS 才能使用 OneDrive/SharePoint 存储策略；启用后同步更改 参数设置 - 站点信息 - 站点URL。"
                }
            />
            <Typography variant={"h6"}>
                {props.policy ? "修改" : "添加"} OneDrive/SharePoint 存储策略
            </Typography>
            <Stepper activeStep={activeStep}>
                {steps.map((label, index) => {
                    const stepProps = {};
                    const labelProps = {};
                    if (label.optional) {
                        labelProps.optional = (
                            <Typography variant="caption">可选</Typography>
                        );
                    }
                    if (isStepSkipped(index)) {
                        stepProps.completed = false;
                    }
                    return (
                        <Step key={label.title} {...stepProps}>
                            <StepLabel {...labelProps}>{label.title}</StepLabel>
                        </Step>
                    );
                })}
            </Stepper>

            {activeStep === 0 && (
                <form
                    className={classes.stepContent}
                    onSubmit={(e) => {
                        e.preventDefault();
                        setActiveStep(1);
                    }}
                >
                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>1</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                前往
                                <Link
                                    href={
                                        "https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/Overview"
                                    }
                                    target={"_blank"}
                                >
                                    Azure Active Directory 控制台 (国际版账号)
                                </Link>{" "}
                                或者{" "}
                                <Link
                                    href={
                                        "https://portal.azure.cn/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/Overview"
                                    }
                                    target={"_blank"}
                                >
                                    Azure Active Directory 控制台 (世纪互联账号)
                                </Link>
                                并登录，登录后进入
                                <code>Azure Active Directory</code>管理面板。
                            </Typography>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>2</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                进入左侧 <code>应用注册</code> 菜单，并点击{" "}
                                <code>新注册</code> 按钮。
                            </Typography>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>3</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                填写应用注册表单。其中，名称可任取；
                                <code>受支持的帐户类型</code> 选择为
                                <code>
                                    任何组织目录(任何 Azure AD 目录 -
                                    多租户)中的帐户
                                </code>
                                ；<code>重定向 URI (可选)</code>
                                请选择<code>Web</code>，并填写
                                <code>
                                    {policy.OptionsSerialized.od_redirect}
                                </code>
                                ； 其他保持默认即可
                            </Typography>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>4</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                创建完成后进入应用管理的<code>概览</code>
                                页面，复制<code>应用程序(客户端) ID</code>
                                并填写在下方：
                            </Typography>
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <InputLabel htmlFor="component-helper">
                                        应用程序(客户端) ID
                                    </InputLabel>
                                    <Input
                                        required
                                        value={policy.BucketName}
                                        onChange={handleChange("BucketName")}
                                    />
                                </FormControl>
                            </div>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>5</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                进入应用管理页面左侧的<code>证书和密码</code>
                                菜单，点击
                                <code>新建客户端密码</code>
                                按钮，<code>截止期限</code>选择为
                                <code>从不</code>
                                。创建完成后将客户端密码的值填写在下方：
                            </Typography>
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <InputLabel htmlFor="component-helper">
                                        客户端密码
                                    </InputLabel>
                                    <Input
                                        required
                                        value={policy.SecretKey}
                                        onChange={handleChange("SecretKey")}
                                    />
                                </FormControl>
                            </div>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>6</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                选择您的 Microsoft 365 账号类型：
                            </Typography>
                            <div className={classes.form}>
                                <FormControl required component="fieldset">
                                    <RadioGroup
                                        required
                                        value={policy.Server}
                                        onChange={handleChange("Server")}
                                        row
                                    >
                                        <FormControlLabel
                                            value={
                                                "https://graph.microsoft.com/v1.0"
                                            }
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="国际版"
                                        />
                                        <FormControlLabel
                                            value={
                                                "https://microsoftgraph.chinacloudapi.cn/v1.0"
                                            }
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="世纪互联版"
                                        />
                                    </RadioGroup>
                                </FormControl>
                            </div>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>7</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                是否将文件存放在 SharePoint 中？
                            </Typography>
                            <div className={classes.form}>
                                <FormControl required component="fieldset">
                                    <RadioGroup
                                        required
                                        value={useSharePoint.toString()}
                                        onChange={(e) => {
                                            setUseSharePoint(
                                                e.target.value === "true"
                                            );
                                        }}
                                        row
                                    >
                                        <FormControlLabel
                                            value={"true"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="存到指定 SharePoint 中"
                                        />
                                        <FormControlLabel
                                            value={"false"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="存到账号默认 OneDrive 驱动器中"
                                        />
                                    </RadioGroup>
                                </FormControl>
                            </div>
                            <Collapse in={useSharePoint}>
                                <div className={classes.form}>
                                    <FormControl fullWidth>
                                        <InputLabel htmlFor="component-helper">
                                            SharePoint 站点地址
                                        </InputLabel>
                                        <Input
                                            placeholder={
                                                "https://example.sharepoint.com/sites/demo"
                                            }
                                            value={
                                                policy.OptionsSerialized
                                                    .od_driver
                                            }
                                            onChange={handleOptionChange(
                                                "od_driver"
                                            )}
                                            required={useSharePoint}
                                            label={"SharePoint 站点地址"}
                                        />
                                    </FormControl>
                                </div>
                            </Collapse>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>8</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                是否要在文件下载时替换为使用自建的反代服务器？
                            </Typography>
                            <div className={classes.form}>
                                <FormControl required component="fieldset">
                                    <RadioGroup
                                        required
                                        value={useCDN.toString()}
                                        onChange={(e) => {
                                            setUseCDN(
                                                e.target.value === "true"
                                            );
                                        }}
                                        row
                                    >
                                        <FormControlLabel
                                            value={"true"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="使用"
                                        />
                                        <FormControlLabel
                                            value={"false"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="不使用"
                                        />
                                    </RadioGroup>
                                </FormControl>
                            </div>
                            <Collapse in={useCDN}>
                                <div className={classes.form}>
                                    <FormControl fullWidth>
                                        <DomainInput
                                            value={
                                                policy.OptionsSerialized
                                                    .od_proxy
                                            }
                                            onChange={handleOptionChange(
                                                "od_proxy"
                                            )}
                                            required={useCDN}
                                            label={"反代服务器地址"}
                                        />
                                    </FormControl>
                                </div>
                            </Collapse>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>9</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                为此存储策略命名：
                            </Typography>
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <InputLabel htmlFor="component-helper">
                                        存储策略名
                                    </InputLabel>
                                    <Input
                                        required
                                        value={policy.Name}
                                        onChange={handleChange("Name")}
                                    />
                                </FormControl>
                            </div>
                        </div>
                    </div>

                    <div className={classes.stepFooter}>
                        <Button
                            disabled={loading}
                            type={"submit"}
                            variant={"contained"}
                            color={"primary"}
                        >
                            下一步
                        </Button>
                    </div>
                </form>
            )}

            {activeStep === 1 && (
                <form
                    className={classes.stepContent}
                    onSubmit={(e) => {
                        e.preventDefault();
                        setActiveStep(2);
                    }}
                >
                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>1</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                请在下方输入文件的存储目录路径，可以为绝对路径或相对路径（相对于
                                从机的
                                Cloudreve）。路径中可以使用魔法变量，文件在上传时会自动替换这些变量为相应值；
                                可用魔法变量可参考{" "}
                                <Link
                                    color={"secondary"}
                                    onClick={(e) => {
                                        e.preventDefault();
                                        setMagicVar("path");
                                    }}
                                >
                                    路径魔法变量列表
                                </Link>{" "}
                                。
                            </Typography>
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <InputLabel htmlFor="component-helper">
                                        存储目录
                                    </InputLabel>
                                    <Input
                                        required
                                        value={policy.DirNameRule}
                                        onChange={handleChange("DirNameRule")}
                                    />
                                </FormControl>
                            </div>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>2</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                是否需要对存储的物理文件进行重命名？此处的重命名不会影响最终呈现给用户的
                                文件名。文件名也可使用魔法变量，
                                可用魔法变量可参考{" "}
                                <Link
                                    color={"secondary"}
                                    onClick={(e) => {
                                        e.preventDefault();
                                        setMagicVar("file");
                                    }}
                                >
                                    文件名魔法变量列表
                                </Link>{" "}
                                。
                            </Typography>
                            <div className={classes.form}>
                                <FormControl required component="fieldset">
                                    <RadioGroup
                                        aria-label="gender"
                                        name="gender1"
                                        value={policy.AutoRename}
                                        onChange={handleChange("AutoRename")}
                                        row
                                    >
                                        <FormControlLabel
                                            value={"true"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="开启重命名"
                                        />
                                        <FormControlLabel
                                            value={"false"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="不开启"
                                        />
                                    </RadioGroup>
                                </FormControl>
                            </div>

                            <Collapse in={policy.AutoRename === "true"}>
                                <div className={classes.form}>
                                    <FormControl fullWidth>
                                        <InputLabel htmlFor="component-helper">
                                            命名规则
                                        </InputLabel>
                                        <Input
                                            required={
                                                policy.AutoRename === "true"
                                            }
                                            value={policy.FileNameRule}
                                            onChange={handleChange(
                                                "FileNameRule"
                                            )}
                                        />
                                    </FormControl>
                                </div>
                            </Collapse>
                        </div>
                    </div>

                    <div className={classes.stepFooter}>
                        <Button
                            color={"default"}
                            className={classes.button}
                            onClick={() => setActiveStep(0)}
                        >
                            上一步
                        </Button>
                        <Button
                            disabled={loading}
                            type={"submit"}
                            variant={"contained"}
                            color={"primary"}
                        >
                            下一步
                        </Button>
                    </div>
                </form>
            )}

            {activeStep === 2 && (
                <form
                    className={classes.stepContent}
                    onSubmit={(e) => {
                        e.preventDefault();
                        setActiveStep(3);
                    }}
                >
                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>1</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                是否允许获取文件永久直链？
                                <br />
                                开启后，用户可以请求获得能直接访问到文件内容的直链，适用于图床应用或自用。
                            </Typography>

                            <div className={classes.form}>
                                <FormControl required component="fieldset">
                                    <RadioGroup
                                        required
                                        value={policy.IsOriginLinkEnable}
                                        onChange={(e) => {
                                            handleChange("IsOriginLinkEnable")(
                                                e
                                            );
                                        }}
                                        row
                                    >
                                        <FormControlLabel
                                            value={"true"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="允许"
                                        />
                                        <FormControlLabel
                                            value={"false"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="禁止"
                                        />
                                    </RadioGroup>
                                </FormControl>
                            </div>
                        </div>
                    </div>

                    <div className={classes.stepFooter}>
                        <Button
                            color={"default"}
                            className={classes.button}
                            onClick={() => setActiveStep(1)}
                        >
                            上一步
                        </Button>{" "}
                        <Button
                            disabled={loading}
                            type={"submit"}
                            variant={"contained"}
                            color={"primary"}
                        >
                            下一步
                        </Button>
                    </div>
                </form>
            )}

            {activeStep === 3 && (
                <form className={classes.stepContent} onSubmit={submitPolicy}>
                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>1</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                是否限制上传的单文件大小？
                            </Typography>

                            <div className={classes.form}>
                                <FormControl required component="fieldset">
                                    <RadioGroup
                                        required
                                        value={
                                            policy.MaxSize === "0"
                                                ? "false"
                                                : "true"
                                        }
                                        onChange={(e) => {
                                            if (e.target.value === "true") {
                                                setPolicy({
                                                    ...policy,
                                                    MaxSize: "10485760",
                                                });
                                            } else {
                                                setPolicy({
                                                    ...policy,
                                                    MaxSize: "0",
                                                });
                                            }
                                        }}
                                        row
                                    >
                                        <FormControlLabel
                                            value={"true"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="限制"
                                        />
                                        <FormControlLabel
                                            value={"false"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="不限制"
                                        />
                                    </RadioGroup>
                                </FormControl>
                            </div>
                        </div>
                    </div>

                    <Collapse in={policy.MaxSize !== "0"}>
                        <div className={classes.subStepContainer}>
                            <div className={classes.stepNumberContainer}>
                                <div className={classes.stepNumber}>2</div>
                            </div>
                            <div className={classes.subStepContent}>
                                <Typography variant={"body2"}>
                                    输入限制：
                                </Typography>
                                <div className={classes.form}>
                                    <SizeInput
                                        value={policy.MaxSize}
                                        onChange={handleChange("MaxSize")}
                                        min={0}
                                        max={9223372036854775807}
                                        label={"单文件大小限制"}
                                    />
                                </div>
                            </div>
                        </div>
                    </Collapse>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>
                                {policy.MaxSize !== "0" ? "3" : "2"}
                            </div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                是否限制上传文件扩展名？
                            </Typography>

                            <div className={classes.form}>
                                <FormControl required component="fieldset">
                                    <RadioGroup
                                        required
                                        value={
                                            policy.OptionsSerialized
                                                .file_type === ""
                                                ? "false"
                                                : "true"
                                        }
                                        onChange={(e) => {
                                            if (e.target.value === "true") {
                                                setPolicy({
                                                    ...policy,
                                                    OptionsSerialized: {
                                                        ...policy.OptionsSerialized,
                                                        file_type:
                                                            "jpg,png,mp4,zip,rar",
                                                    },
                                                });
                                            } else {
                                                setPolicy({
                                                    ...policy,
                                                    OptionsSerialized: {
                                                        ...policy.OptionsSerialized,
                                                        file_type: "",
                                                    },
                                                });
                                            }
                                        }}
                                        row
                                    >
                                        <FormControlLabel
                                            value={"true"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="限制"
                                        />
                                        <FormControlLabel
                                            value={"false"}
                                            control={
                                                <Radio color={"primary"} />
                                            }
                                            label="不限制"
                                        />
                                    </RadioGroup>
                                </FormControl>
                            </div>
                        </div>
                    </div>

                    <Collapse in={policy.OptionsSerialized.file_type !== ""}>
                        <div className={classes.subStepContainer}>
                            <div className={classes.stepNumberContainer}>
                                <div className={classes.stepNumber}>
                                    {policy.MaxSize !== "0" ? "4" : "3"}
                                </div>
                            </div>
                            <div className={classes.subStepContent}>
                                <Typography variant={"body2"}>
                                    输入允许上传的文件扩展名，多个请以半角逗号 ,
                                    隔开
                                </Typography>
                                <div className={classes.form}>
                                    <FormControl fullWidth>
                                        <InputLabel htmlFor="component-helper">
                                            扩展名列表
                                        </InputLabel>
                                        <Input
                                            value={
                                                policy.OptionsSerialized
                                                    .file_type
                                            }
                                            onChange={handleOptionChange(
                                                "file_type"
                                            )}
                                        />
                                    </FormControl>
                                </div>
                            </div>
                        </div>
                    </Collapse>

                    <div className={classes.stepFooter}>
                        <Button
                            color={"default"}
                            className={classes.button}
                            onClick={() => setActiveStep(2)}
                        >
                            上一步
                        </Button>{" "}
                        <Button
                            disabled={loading}
                            type={"submit"}
                            variant={"contained"}
                            color={"primary"}
                        >
                            下一步
                        </Button>
                    </div>
                </form>
            )}

            {activeStep === 4 && (
                <form className={classes.stepContent}>
                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer} />
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                存储策略已{props.policy ? "保存" : "添加"}
                                ，但是你需要点击下方按钮，并使用 OneDrive
                                登录授权以完成初始化后才能使用。
                                日后你可以在存储策略列表页面重新进行授权。
                            </Typography>
                            <div className={classes.form}>
                                <Button
                                    disabled={loading}
                                    color={"secondary"}
                                    variant={"contained"}
                                    className={classes.button}
                                    onClick={statOAuth}
                                    classes={{ label: classes.viewButtonLabel }}
                                >
                                    转到授权页面
                                </Button>
                            </div>
                        </div>
                    </div>
                    <div className={classes.stepFooter}></div>
                </form>
            )}

            {activeStep === 5 && (
                <>
                    <form className={classes.stepContent}>
                        <Typography>存储策略已添加！</Typography>
                        <Typography variant={"body2"} color={"textSecondary"}>
                            要使用此存储策略，请到用户组管理页面，为相应用户组绑定此存储策略。
                        </Typography>
                    </form>
                    <div className={classes.stepFooter}>
                        <Button
                            color={"primary"}
                            className={classes.button}
                            onClick={() => history.push("/admin/policy")}
                        >
                            返回存储策略列表
                        </Button>
                    </div>
                </>
            )}

            <MagicVar
                open={magicVar === "file"}
                isFile
                onClose={() => setMagicVar("")}
            />
            <MagicVar
                open={magicVar === "path"}
                onClose={() => setMagicVar("")}
            />
        </div>
    );
}
