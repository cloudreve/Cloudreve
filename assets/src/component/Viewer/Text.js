import React, { useCallback, useEffect, useState } from "react";
import { Paper } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import { useLocation, useParams, useRouteMatch } from "react-router";
import API from "../../middleware/Api";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../actions";
import { changeSubTitle } from "../../redux/viewUpdate/action";
import Editor from "for-editor";
import SaveButton from "../Dial/Save";
import pathHelper from "../../utils/page";
import TextLoading from "../Placeholder/TextLoading";
const useStyles = makeStyles((theme) => ({
    layout: {
        width: "auto",
        marginTop: "30px",
        marginLeft: theme.spacing(3),
        marginRight: theme.spacing(3),
        [theme.breakpoints.up(1100 + theme.spacing(3) * 2)]: {
            width: 1100,
            marginLeft: "auto",
            marginRight: "auto",
        },
        marginBottom: 50,
    },
    player: {
        borderRadius: "4px",
    },
    root: {
        backgroundColor: "white",
        borderRadius: "8px",
    },
    "@global": {
        ".for-toolbar": {
            overflowX: "auto!important",
        },
    },
}));

function useQuery() {
    return new URLSearchParams(useLocation().search);
}

export default function TextViewer() {
    const [content, setContent] = useState("");
    const [status, setStatus] = useState("");
    const [loading, setLoading] = useState(true);
    const math = useRouteMatch();
    const $vm = React.createRef();
    const location = useLocation();
    const query = useQuery();
    const { id } = useParams();

    const dispatch = useDispatch();
    const SetSubTitle = useCallback(
        (title) => dispatch(changeSubTitle(title)),
        [dispatch]
    );
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    useEffect(() => {
        if (!pathHelper.isSharePage(location.pathname)) {
            const path = query.get("p").split("/");
            SetSubTitle(path[path.length - 1]);
        } else {
            SetSubTitle(query.get("name"));
        }
        // eslint-disable-next-line
    }, [math.params[0], location]);

    useEffect(() => {
        let requestURL = "/file/content/" + query.get("id");
        if (pathHelper.isSharePage(location.pathname)) {
            requestURL = "/share/content/" + id;
            if (query.get("share_path") !== "") {
                requestURL +=
                    "?path=" + encodeURIComponent(query.get("share_path"));
            }
        }

        setLoading(true);
        API.get(requestURL, { responseType: "arraybuffer" })
            .then((response) => {
                const buffer = new Buffer(response.rawData, "binary");
                const textdata = buffer.toString(); // for string
                setContent(textdata);
            })
            .catch((error) => {
                ToggleSnackbar(
                    "top",
                    "right",
                    "无法读取文件内容，" + error.message,
                    "error"
                );
            })
            .then(() => {
                setLoading(false);
            });
        // eslint-disable-next-line
    }, [math.params[0]]);

    const toBase64 = (file) =>
        new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.onload = () => resolve(reader.result);
            reader.onerror = (error) => reject(error);
        });

    const save = () => {
        setStatus("loading");
        API.put("/file/update/" + query.get("id"), content)
            .then(() => {
                setStatus("success");
                setTimeout(() => setStatus(""), 2000);
            })
            .catch((error) => {
                setStatus("");
                ToggleSnackbar("top", "right", error.message, "error");
            });
    };

    const addImg = async ($file) => {
        $vm.current.$img2Url($file.name, await toBase64($file));
        console.log($file);
    };

    const classes = useStyles();
    return (
        <div className={classes.layout}>
            <Paper className={classes.root} elevation={1}>
                {loading && <TextLoading />}
                {!loading && (
                    <Editor
                        ref={$vm}
                        value={content}
                        onSave={() => save()}
                        addImg={($file) => addImg($file)}
                        onChange={(value) => setContent(value)}
                        toolbar={{
                            h1: true, // h1
                            h2: true, // h2
                            h3: true, // h3
                            h4: true, // h4
                            img: true, // 图片
                            link: true, // 链接
                            code: true, // 代码块
                            preview: true, // 预览
                            expand: true, // 全屏
                            /* v0.0.9 */
                            undo: true, // 撤销
                            redo: true, // 重做
                            save: false, // 保存
                            /* v0.2.3 */
                            subfield: true, // 单双栏模式
                        }}
                    />
                )}
            </Paper>
            {!pathHelper.isSharePage(location.pathname) && (
                <SaveButton onClick={save} status={status} />
            )}
        </div>
    );
}
