import React, { useCallback, useEffect, useState } from "react";
import { makeStyles, useTheme } from "@material-ui/core/styles";
import { MenuBook } from "@material-ui/icons";
import { Typography } from "@material-ui/core";
import Divider from "@material-ui/core/Divider";
import Paper from "@material-ui/core/Paper";
import TextLoading from "../Placeholder/TextLoading";
import API from "../../middleware/Api";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../actions";
import Editor from "for-editor";

const useStyles = makeStyles((theme) => ({
    readMeContainer: {
        marginTop: 30,
        [theme.breakpoints.down("sm")]: {
            marginTop: theme.spacing(2),
        },
    },
    readMeHeader: {
        padding: "10px 16px",
        display: "flex",
        color: theme.palette.text.secondary,
    },
    readMeIcon: {
        marginRight: 8,
    },
    content: {},
    "@global": {
        //如果嵌套主题，则应该定位[class * =“MuiButton-root”]。
        ".for-container": {
            border: "none!important",
        },
        ".for-container .for-editor .for-editor-edit": {
            height: "0!important",
        },
        ".for-container > div:first-child": {
            borderTopLeftRadius: "0!important",
            borderTopRightRadius: "0!important",
        },
        ".for-container .for-editor .for-panel .for-preview": {
            backgroundColor: theme.palette.background.paper + "!important",
            color: theme.palette.text.primary + "!important",
        },
        ".for-container .for-markdown-preview pre": {
            backgroundColor: theme.palette.background.default + "!important",
            color:
                theme.palette.type === "dark"
                    ? "#fff !important"
                    : "rgba(0, 0, 0, 0.87);!important",
        },

        ".for-container .for-markdown-preview code": {
            backgroundColor: theme.palette.background.default + "!important",
        },
        ".for-container .for-markdown-preview a": {
            color:
                theme.palette.type === "dark"
                    ? "#67aeff !important"
                    : "#0366d6 !important",
        },
        ".for-container .for-markdown-preview table th": {
            backgroundColor: theme.palette.background.default + "!important",
        },
    },
}));

export default function ReadMe(props) {
    const classes = useStyles();
    const theme = useTheme();

    const [loading, setLoading] = useState(true);
    const [content, setContent] = useState("");

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const $vm = React.createRef();

    useEffect(() => {
        setLoading(true);
        const previewPath =
            props.file.path === "/"
                ? props.file.path + props.file.name
                : props.file.path + "/" + props.file.name;
        API.get(
            "/share/readme/" +
                props.share.key +
                "?path=" +
                encodeURIComponent(previewPath)
        )
            .then((response) => {
                setContent(response.rawData.toString());
            })
            .catch((error) => {
                ToggleSnackbar(
                    "top",
                    "right",
                    "无法读取 README 内容，" + error.message,
                    "error"
                );
            })
            .then(() => {
                setLoading(false);
            });
        // eslint-disable-next-line
    }, [props.share, props.file]);

    return (
        <Paper className={classes.readMeContainer}>
            <div className={classes.readMeHeader}>
                <MenuBook className={classes.readMeIcon} />
                <Typography>{props.file.name}</Typography>
            </div>

            <Divider />
            <div className={classes.content}>
                {loading && <TextLoading />}
                {!loading && (
                    <Editor
                        ref={$vm}
                        style={{
                            boxShadow: "none",
                            borderRadius: 4,
                            backgroundColor: theme.palette.background.paper,
                        }}
                        height={"100%"}
                        value={content}
                        onChange={(value) => setContent(value)}
                        preview
                        toolbar={{}}
                    />
                )}
            </div>
        </Paper>
    );
}
