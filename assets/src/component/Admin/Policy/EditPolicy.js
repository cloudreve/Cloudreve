import Paper from "@material-ui/core/Paper";
import { makeStyles } from "@material-ui/core/styles";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { useParams } from "react-router";
import { toggleSnackbar } from "../../../actions";
import API from "../../../middleware/Api";
import COSGuide from "./Guid/COSGuide";
import EditPro from "./Guid/EditPro";
import LocalGuide from "./Guid/LocalGuide";
import OneDriveGuide from "./Guid/OneDriveGuide";
import OSSGuide from "./Guid/OSSGuide";
import QiniuGuide from "./Guid/QiniuGuide";
import RemoteGuide from "./Guid/RemoteGuide";
import UpyunGuide from "./Guid/UpyunGuide";
import S3Guide from "./Guid/S3Guide";

const useStyles = makeStyles((theme) => ({
    root: {
        [theme.breakpoints.up("md")]: {
            marginLeft: 100,
        },
        marginBottom: 40,
    },
    content: {
        padding: theme.spacing(2),
    },
}));

export default function EditPolicyPreload() {
    const classes = useStyles();
    const [type, setType] = useState("");
    const [policy, setPolicy] = useState({});

    const { mode, id } = useParams();

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    useEffect(() => {
        setType("");
        API.get("/admin/policy/" + id)
            .then((response) => {
                response.data.IsOriginLinkEnable = response.data
                    .IsOriginLinkEnable
                    ? "true"
                    : "false";
                response.data.AutoRename = response.data.AutoRename
                    ? "true"
                    : "false";
                response.data.MaxSize = response.data.MaxSize.toString();
                response.data.IsPrivate = response.data.IsPrivate
                    ? "true"
                    : "false";
                response.data.OptionsSerialized.file_type = response.data
                    .OptionsSerialized.file_type
                    ? response.data.OptionsSerialized.file_type.join(",")
                    : "";
                setPolicy(response.data);
                setType(response.data.Type);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    }, [id]);

    return (
        <div>
            <Paper square className={classes.content}>
                {mode === "guide" && (
                    <>
                        {type === "local" && <LocalGuide policy={policy} />}
                        {type === "remote" && <RemoteGuide policy={policy} />}
                        {type === "qiniu" && <QiniuGuide policy={policy} />}
                        {type === "oss" && <OSSGuide policy={policy} />}
                        {type === "upyun" && <UpyunGuide policy={policy} />}
                        {type === "cos" && <COSGuide policy={policy} />}
                        {type === "onedrive" && (
                            <OneDriveGuide policy={policy} />
                        )}
                        {type === "s3" && <S3Guide policy={policy} />}
                    </>
                )}

                {mode === "pro" && type !== "" && <EditPro policy={policy} />}
            </Paper>
        </div>
    );
}
