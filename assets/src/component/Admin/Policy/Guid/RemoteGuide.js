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
import Alert from "@material-ui/lab/Alert";
import React, { useCallback, useState } from "react";
import { useDispatch } from "react-redux";
import { useHistory } from "react-router";
import { toggleSnackbar } from "../../../../actions";
import API from "../../../../middleware/Api";
import { randomStr } from "../../../../utils";
import DomainInput from "../../Common/DomainInput";
import SizeInput from "../../Common/SizeInput";
import MagicVar from "../../Dialogs/MagicVar";

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
    "@global": {
        code: {
            color: "rgba(0, 0, 0, 0.87)",
            display: "inline-block",
            padding: "2px 6px",
            fontSize: "14px",
            fontFamily:
                ' Consolas, "Liberation Mono", Menlo, Courier, monospace',
            borderRadius: "2px",
            backgroundColor: "rgba(255,229,100,0.1)",
        },
        pre: {
            margin: "24px 0",
            padding: "12px 18px",
            overflow: "auto",
            direction: "ltr",
            borderRadius: "4px",
            backgroundColor: "#272c34",
            color: "#fff",
        },
    },
}));

const steps = [
    {
        title: "存储端配置",
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
        title: "完成",
        optional: false,
    },
];

export default function RemoteGuide(props) {
    const classes = useStyles();
    const history = useHistory();

    const [activeStep, setActiveStep] = useState(0);
    const [loading, setLoading] = useState(false);
    const [skipped] = React.useState(new Set());
    const [magicVar, setMagicVar] = useState("");
    const [useCDN, setUseCDN] = useState("false");
    const [policy, setPolicy] = useState(
        props.policy
            ? props.policy
            : {
                  Type: "remote",
                  Name: "",
                  Server: "https://example.com:5212",
                  SecretKey: randomStr(64),
                  DirNameRule: "uploads/{year}/{month}/{day}",
                  AutoRename: "true",
                  FileNameRule: "{randomkey8}_{originname}",
                  IsOriginLinkEnable: "false",
                  BaseURL: "",
                  IsPrivate: "true",
                  MaxSize: "0",
                  OptionsSerialized: {
                      file_type: "",
                  },
              }
    );

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

    const testSlave = () => {
        setLoading(true);

        // 测试路径是否可用
        API.post("/admin/policy/test/slave", {
            server: policy.Server,
            secret: policy.SecretKey,
        })
            .then(() => {
                ToggleSnackbar("top", "right", "通信正常", "success");
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
    };

    const submitPolicy = (e) => {
        e.preventDefault();
        setLoading(true);

        const policyCopy = { ...policy };
        policyCopy.OptionsSerialized = { ...policyCopy.OptionsSerialized };

        // 处理存储策略
        if (useCDN === "false" || policy.IsOriginLinkEnable === "false") {
            policyCopy.BaseURL = "";
        }

        // 类型转换
        policyCopy.AutoRename = policyCopy.AutoRename === "true";
        policyCopy.IsOriginLinkEnable =
            policyCopy.IsOriginLinkEnable === "true";
        policyCopy.MaxSize = parseInt(policyCopy.MaxSize);
        policyCopy.IsPrivate = policyCopy.IsPrivate === "true";
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
                setActiveStep(5);
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
            <Typography variant={"h6"}>
                {props.policy ? "修改" : "添加"}从机存储策略
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
                    <Alert severity="info" style={{ marginBottom: 10 }}>
                        从机存储策略允许你使用同样运行了 Cloudreve
                        的服务器作为存储端， 用户上传下载流量通过 HTTP 直传。
                    </Alert>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>1</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                将和主站相同版本的 Cloudreve
                                程序拷贝至要作为从机的服务器上。
                            </Typography>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>2</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                下方为系统为您随机生成的从机端密钥，一般无需改动，如果有自定义需求，
                                可将您的密钥填入下方：
                            </Typography>
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <InputLabel htmlFor="component-helper">
                                        从机密钥
                                    </InputLabel>
                                    <Input
                                        required
                                        inputProps={{
                                            minlength: 64,
                                        }}
                                        value={policy.SecretKey}
                                        onChange={handleChange("SecretKey")}
                                    />
                                </FormControl>
                            </div>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>3</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                修改从机配置文件。
                                <br />
                                在从机端 Cloudreve 的同级目录下新建
                                <code>conf.ini</code>
                                文件，填入从机配置，启动/重启从机端 Cloudreve。
                                以下为一个可供参考的配置例子，其中密钥部分已帮您填写为上一步所生成的。
                            </Typography>
                            <pre>
                                [System]
                                <br />
                                Mode = slave
                                <br />
                                Listen = :5212
                                <br />
                                <br />
                                [Slave]
                                <br />
                                Secret = {policy.SecretKey}
                                <br />
                                <br />
                                [CORS]
                                <br />
                                AllowOrigins = *<br />
                                AllowMethods = OPTIONS,GET,POST
                                <br />
                                AllowHeaders = *<br />
                            </pre>
                            <Typography variant={"body2"}>
                                从机端配置文件格式大致与主站端相同，区别在于：
                                <ul>
                                    <li>
                                        <code>System</code>
                                        分区下的
                                        <code>mode</code>
                                        字段必须更改为<code>slave</code>
                                    </li>
                                    <li>
                                        必须指定<code>Slave</code>分区下的
                                        <code>Secret</code>
                                        字段，其值为第二步里填写或生成的密钥。
                                    </li>
                                    <li>
                                        必须启动跨域配置，即<code>CORS</code>
                                        字段的内容，
                                        具体可参考上文范例或官方文档。如果配置不正确，用户将无法通过
                                        Web 端向从机上传文件。
                                    </li>
                                </ul>
                            </Typography>
                        </div>
                    </div>

                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}>
                            <div className={classes.stepNumber}>4</div>
                        </div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                填写从机地址。
                                <br />
                                如果主站启用了
                                HTTPS，从机也需要启用，并在下方填入 HTTPS
                                协议的地址。
                            </Typography>
                            <div className={classes.form}>
                                <FormControl fullWidth>
                                    <InputLabel htmlFor="component-helper">
                                        从机地址
                                    </InputLabel>
                                    <Input
                                        fullWidth
                                        required
                                        type={"url"}
                                        value={policy.Server}
                                        onChange={handleChange("Server")}
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
                                完成以上步骤后，你可以点击下方的测试按钮测试通信是否正常。
                            </Typography>
                            <div className={classes.form}>
                                <Button
                                    disabled={loading}
                                    onClick={() => testSlave()}
                                    variant={"outlined"}
                                    color={"primary"}
                                >
                                    测试从机通信
                                </Button>
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

                    <Collapse in={policy.IsOriginLinkEnable === "true"}>
                        <div className={classes.subStepContainer}>
                            <div className={classes.stepNumberContainer}>
                                <div className={classes.stepNumber}>2</div>
                            </div>
                            <div className={classes.subStepContent}>
                                <Typography variant={"body2"}>
                                    是否要对下载/直链使用 CDN？
                                    <br />
                                    开启后，用户访问文件时的 URL
                                    中的域名部分会被替换为 CDN 域名。
                                </Typography>

                                <div className={classes.form}>
                                    <FormControl required component="fieldset">
                                        <RadioGroup
                                            required
                                            value={useCDN}
                                            onChange={(e) => {
                                                if (
                                                    e.target.value === "false"
                                                ) {
                                                    setPolicy({
                                                        ...policy,
                                                        BaseURL: "",
                                                    });
                                                }
                                                setUseCDN(e.target.value);
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
                            </div>
                        </div>

                        <Collapse in={useCDN === "true"}>
                            <div className={classes.subStepContainer}>
                                <div className={classes.stepNumberContainer}>
                                    <div className={classes.stepNumber}>3</div>
                                </div>
                                <div className={classes.subStepContent}>
                                    <Typography variant={"body2"}>
                                        选择协议并填写 CDN 域名
                                    </Typography>

                                    <div className={classes.form}>
                                        <DomainInput
                                            value={policy.BaseURL}
                                            onChange={handleChange("BaseURL")}
                                            required={
                                                policy.IsOriginLinkEnable ===
                                                    "true" && useCDN === "true"
                                            }
                                            label={"CDN 前缀"}
                                        />
                                    </div>
                                </div>
                            </div>
                        </Collapse>
                    </Collapse>

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
                <form
                    className={classes.stepContent}
                    onSubmit={(e) => {
                        e.preventDefault();
                        setActiveStep(4);
                    }}
                >
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
                <form className={classes.stepContent} onSubmit={submitPolicy}>
                    <div className={classes.subStepContainer}>
                        <div className={classes.stepNumberContainer}></div>
                        <div className={classes.subStepContent}>
                            <Typography variant={"body2"}>
                                最后一步，为此存储策略命名：
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
                            color={"default"}
                            className={classes.button}
                            onClick={() => setActiveStep(3)}
                        >
                            上一步
                        </Button>{" "}
                        <Button
                            disabled={loading}
                            type={"submit"}
                            variant={"contained"}
                            color={"primary"}
                        >
                            完成
                        </Button>
                    </div>
                </form>
            )}

            {activeStep === 5 && (
                <>
                    <form className={classes.stepContent}>
                        <Typography>
                            存储策略已{props.policy ? "保存" : "添加"}！
                        </Typography>
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
                isSlave
                onClose={() => setMagicVar("")}
            />
            <MagicVar
                open={magicVar === "path"}
                isSlave
                onClose={() => setMagicVar("")}
            />
        </div>
    );
}
