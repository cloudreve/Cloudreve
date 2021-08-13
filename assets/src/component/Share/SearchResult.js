import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../actions";
import OpenIcon from "@material-ui/icons/OpenInNew";
import Pagination from "@material-ui/lab/Pagination";
import FolderIcon from "@material-ui/icons/Folder";

import {
    Tooltip,
    Card,
    Avatar,
    CardHeader,
    Typography,
    Grid,
    IconButton,
} from "@material-ui/core";
import API from "../../middleware/Api";
import TypeIcon from "../FileManager/TypeIcon";
import Select from "@material-ui/core/Select";
import MenuItem from "@material-ui/core/MenuItem";
import FormControl from "@material-ui/core/FormControl";
import { useHistory } from "react-router-dom";
import { makeStyles } from "@material-ui/core/styles";
import { useLocation } from "react-router";
import TimeAgo from "timeago-react";

const useStyles = makeStyles((theme) => ({
    cardContainer: {
        padding: theme.spacing(1),
    },
    card: {
        maxWidth: 400,
        margin: "0 auto",
    },
    actions: {
        display: "flex",
    },
    layout: {
        width: "auto",
        marginTop: "50px",
        marginLeft: theme.spacing(3),
        marginRight: theme.spacing(3),
        [theme.breakpoints.up(1100 + theme.spacing(3) * 2)]: {
            width: 1100,
            marginLeft: "auto",
            marginRight: "auto",
        },
    },
    shareTitle: {
        maxWidth: "200px",
    },
    avatarFile: {
        backgroundColor: theme.palette.primary.light,
    },
    avatarFolder: {
        backgroundColor: theme.palette.secondary.light,
    },
    gird: {
        marginTop: "30px",
    },
    loadMore: {
        textAlign: "right",
        marginTop: "20px",
        marginBottom: "40px",
    },
    badge: {
        marginLeft: theme.spacing(1),
        height: 17,
    },
    orderSelect: {
        textAlign: "right",
        marginTop: 5,
    },
}));

function useQuery() {
    return new URLSearchParams(useLocation().search);
}

export default function SearchResult() {
    const classes = useStyles();
    const dispatch = useDispatch();

    const query = useQuery();
    const location = useLocation();
    const history = useHistory();

    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);
    const [shareList, setShareList] = useState([]);
    const [orderBy, setOrderBy] = useState("created_at DESC");

    const search = useCallback((keywords, page, orderBy) => {
        const order = orderBy.split(" ");
        API.get(
            "/share/search?page=" +
                page +
                "&order_by=" +
                order[0] +
                "&order=" +
                order[1] +
                "&keywords=" +
                encodeURIComponent(keywords)
        )
            .then((response) => {
                if (response.data.items.length === 0) {
                    ToggleSnackbar(
                        "top",
                        "right",
                        "找不到符合条件的分享",
                        "info"
                    );
                }
                setTotal(response.data.total);
                setShareList(response.data.items);
            })
            .catch(() => {
                ToggleSnackbar("top", "right", "加载失败", "error");
            });
    }, []);

    useEffect(() => {
        const keywords = query.get("keywords");
        if (keywords) {
            search(keywords, page, orderBy);
        } else {
            ToggleSnackbar("top", "right", "请输入搜索关键词", "warning");
        }
    }, [location]);

    const handlePageChange = (event, value) => {
        setPage(value);
        const keywords = query.get("keywords");
        search(keywords, value, orderBy);
    };

    const handleOrderChange = (event) => {
        setOrderBy(event.target.value);
        const keywords = query.get("keywords");
        search(keywords, page, event.target.value);
    };

    return (
        <div className={classes.layout}>
            <Grid container>
                <Grid sm={6} xs={6}>
                    <Typography color="textSecondary" variant="h4">
                        搜索结果
                    </Typography>
                </Grid>
                <Grid sm={6} xs={6} className={classes.orderSelect}>
                    <FormControl>
                        <Select
                            color={"secondary"}
                            onChange={handleOrderChange}
                            value={orderBy}
                        >
                            <MenuItem value={"created_at DESC"}>
                                创建日期由晚到早
                            </MenuItem>
                            <MenuItem value={"created_at ASC"}>
                                创建日期由早到晚
                            </MenuItem>
                            <MenuItem value={"downloads DESC"}>
                                下载次数由大到小
                            </MenuItem>
                            <MenuItem value={"downloads ASC"}>
                                下载次数由小到大
                            </MenuItem>
                            <MenuItem value={"views DESC"}>
                                浏览次数由大到小
                            </MenuItem>
                            <MenuItem value={"views ASC"}>
                                浏览次数由小到大
                            </MenuItem>
                        </Select>
                    </FormControl>
                </Grid>
            </Grid>
            <Grid container spacing={24} className={classes.gird}>
                {shareList.map((value) => (
                    <Grid
                        item
                        xs={12}
                        sm={4}
                        key={value.id}
                        className={classes.cardContainer}
                    >
                        <Card className={classes.card}>
                            <CardHeader
                                avatar={
                                    <div>
                                        {!value.is_dir && (
                                            <TypeIcon
                                                fileName={
                                                    value.source
                                                        ? value.source.name
                                                        : ""
                                                }
                                                isUpload
                                            />
                                        )}{" "}
                                        {value.is_dir && (
                                            <Avatar
                                                className={classes.avatarFolder}
                                            >
                                                <FolderIcon />
                                            </Avatar>
                                        )}
                                    </div>
                                }
                                action={
                                    <Tooltip placement="top" title="打开">
                                        <IconButton
                                            onClick={() =>
                                                history.push("/s/" + value.key)
                                            }
                                        >
                                            <OpenIcon />
                                        </IconButton>
                                    </Tooltip>
                                }
                                title={
                                    <Tooltip
                                        placement="top"
                                        title={
                                            value.source
                                                ? value.source.name
                                                : "[原始对象不存在]"
                                        }
                                    >
                                        <Typography
                                            noWrap
                                            className={classes.shareTitle}
                                        >
                                            {value.source
                                                ? value.source.name
                                                : "[原始对象不存在]"}{" "}
                                        </Typography>
                                    </Tooltip>
                                }
                                subheader={
                                    <span>
                                        分享于{" "}
                                        <TimeAgo
                                            datetime={value.create_date}
                                            locale="zh_CN"
                                        />
                                    </span>
                                }
                            />
                        </Card>
                    </Grid>
                ))}
            </Grid>
            <div className={classes.loadMore}>
                <Pagination
                    count={Math.ceil(total / 18)}
                    onChange={handlePageChange}
                    color="secondary"
                />
            </div>{" "}
        </div>
    );
}
