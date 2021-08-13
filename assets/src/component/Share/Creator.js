import React from "react";
import { makeStyles } from "@material-ui/core/styles";
import { Avatar, Typography } from "@material-ui/core";
import { useHistory } from "react-router";
import Link from "@material-ui/core/Link";
import { formatLocalTime } from "../../utils/datetime";

const useStyles = makeStyles((theme) => ({
    boxHeader: {
        textAlign: "center",
        padding: 24,
    },
    avatar: {
        backgroundColor: theme.palette.secondary.main,
        margin: "0 auto",
        width: 50,
        height: 50,
        cursor: "pointer",
    },
    shareDes: {
        marginTop: 12,
    },
    shareInfo: {
        color: theme.palette.text.disabled,
        fontSize: 14,
    },
}));

export default function Creator(props) {
    const classes = useStyles();
    const history = useHistory();

    const getSecondDes = () => {
        if (props.share.expire > 0) {
            if (props.share.expire >= 24 * 3600) {
                return (
                    Math.round(props.share.expire / (24 * 3600)) + " 天后到期"
                );
            }
            return Math.round(props.share.expire / 3600) + " 小时后到期";
        }
        return formatLocalTime(props.share.create_date, "YYYY-MM-DD H:mm:ss");
    };

    const userProfile = () => {
        history.push("/profile/" + props.share.creator.key);
        props.onClose && props.onClose();
    };

    return (
        <div className={classes.boxHeader}>
            <Avatar
                className={classes.avatar}
                alt={props.share.creator.nick}
                src={"/api/v3/user/avatar/" + props.share.creator.key + "/l"}
                onClick={() => userProfile()}
            />
            <Typography variant="h6" className={classes.shareDes}>
                {props.isFolder && (
                    <>
                        此分享由{" "}
                        <Link
                            onClick={() => userProfile()}
                            href={"javascript:void(0)"}
                            color="inherit"
                        >
                            {props.share.creator.nick}
                        </Link>{" "}
                        创建
                    </>
                )}
                {!props.isFolder && (
                    <>
                        {" "}
                        <Link
                            onClick={() => userProfile()}
                            href={"javascript:void(0)"}
                            color="inherit"
                        >
                            {props.share.creator.nick}
                        </Link>{" "}
                        向您分享了 1 个文件
                    </>
                )}
            </Typography>
            <Typography className={classes.shareInfo}>
                {props.share.views} 次浏览 • {props.share.downloads} 次下载 •{" "}
                {getSecondDes()}
            </Typography>
        </div>
    );
}
