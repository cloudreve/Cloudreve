import Button from "@material-ui/core/Button";
import Card from "@material-ui/core/Card";
import CardActionArea from "@material-ui/core/CardActionArea";
import CardContent from "@material-ui/core/CardContent";
import CardMedia from "@material-ui/core/CardMedia";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogTitle from "@material-ui/core/DialogTitle";
import Grid from "@material-ui/core/Grid";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import React from "react";
import { useHistory } from "react-router";

const useStyles = makeStyles((theme) => ({
    cardContainer: {
        display: "flex",
    },
    cover: {
        width: 100,
        height: 60,
    },
    card: {},
    content: {
        flex: "1 0 auto",
    },
    bg: {
        backgroundColor: theme.palette.background.default,
        padding: "24px 24px",
    },
    dialogFooter: {
        justifyContent: "space-between",
    },
}));

const policies = [
    {
        name: "本机存储",
        img: "local.png",
        path: "/admin/policy/add/local",
    },
    {
        name: "从机存储",
        img: "remote.png",
        path: "/admin/policy/add/remote",
    },
    {
        name: "七牛",
        img: "qiniu.png",
        path: "/admin/policy/add/qiniu",
    },
    {
        name: "阿里云 OSS",
        img: "oss.png",
        path: "/admin/policy/add/oss",
    },
    {
        name: "又拍云",
        img: "upyun.png",
        path: "/admin/policy/add/upyun",
    },
    {
        name: "腾讯云 COS",
        img: "cos.png",
        path: "/admin/policy/add/cos",
    },
    {
        name: "OneDrive",
        img: "onedrive.png",
        path: "/admin/policy/add/onedrive",
    },
    {
        name: "Amazon S3",
        img: "s3.png",
        path: "/admin/policy/add/s3",
    },
];

export default function AddPolicy({ open, onClose }) {
    const classes = useStyles();

    const location = useHistory();

    return (
        <Dialog
            open={open}
            onClose={onClose}
            aria-labelledby="alert-dialog-title"
            aria-describedby="alert-dialog-description"
            maxWidth={"sm"}
            fullWidth
        >
            <DialogTitle id="alert-dialog-title">选择存储方式</DialogTitle>
            <DialogContent dividers className={classes.bg}>
                <Grid container spacing={2}>
                    {policies.map((v) => (
                        <Grid item sm={12} md={6} key={v.path}>
                            <Card className={classes.card}>
                                <CardActionArea
                                    onClick={() => {
                                        location.push(v.path);
                                        onClose();
                                    }}
                                    className={classes.cardContainer}
                                >
                                    <CardMedia
                                        className={classes.cover}
                                        image={"/static/img/" + v.img}
                                    />
                                    <CardContent className={classes.content}>
                                        <Typography
                                            variant="subtitle1"
                                            color="textSecondary"
                                        >
                                            {v.name}
                                        </Typography>
                                    </CardContent>
                                </CardActionArea>
                            </Card>
                        </Grid>
                    ))}
                </Grid>
            </DialogContent>
            <DialogActions className={classes.dialogFooter}>
                <Button
                    onClick={() =>
                        window.open(
                            "https://docs.cloudreve.org/use/policy/compare"
                        )
                    }
                    color="primary"
                >
                    存储策略对比
                </Button>
                <Button onClick={onClose} color="primary">
                    取消
                </Button>
            </DialogActions>
        </Dialog>
    );
}
