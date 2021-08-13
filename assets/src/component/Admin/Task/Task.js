import { lighten } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Checkbox from "@material-ui/core/Checkbox";
import IconButton from "@material-ui/core/IconButton";
import Link from "@material-ui/core/Link";
import Paper from "@material-ui/core/Paper";
import { makeStyles } from "@material-ui/core/styles";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TablePagination from "@material-ui/core/TablePagination";
import TableRow from "@material-ui/core/TableRow";
import TableSortLabel from "@material-ui/core/TableSortLabel";
import Toolbar from "@material-ui/core/Toolbar";
import Tooltip from "@material-ui/core/Tooltip";
import Typography from "@material-ui/core/Typography";
import { Delete } from "@material-ui/icons";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../../actions";
import { getTaskProgress, getTaskStatus, getTaskType } from "../../../config";
import API from "../../../middleware/Api";
import ShareFilter from "../Dialogs/ShareFilter";
import { formatLocalTime } from "../../../utils/datetime";

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
    container: {
        overflowX: "auto",
    },
    tableContainer: {
        marginTop: 16,
    },
    header: {
        display: "flex",
        justifyContent: "space-between",
    },
    headerRight: {},
    highlight:
        theme.palette.type === "light"
            ? {
                  color: theme.palette.secondary.main,
                  backgroundColor: lighten(theme.palette.secondary.light, 0.85),
              }
            : {
                  color: theme.palette.text.primary,
                  backgroundColor: theme.palette.secondary.dark,
              },
    visuallyHidden: {
        border: 0,
        clip: "rect(0 0 0 0)",
        height: 1,
        margin: -1,
        overflow: "hidden",
        padding: 0,
        position: "absolute",
        top: 20,
        width: 1,
    },
}));

export default function Task() {
    const classes = useStyles();
    const [tasks, setTasks] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [total, setTotal] = useState(0);
    const [filter, setFilter] = useState({});
    const [users, setUsers] = useState({});
    const [search, setSearch] = useState({});
    const [orderBy, setOrderBy] = useState(["id", "desc"]);
    const [filterDialog, setFilterDialog] = useState(false);
    const [selected, setSelected] = useState([]);
    const [loading, setLoading] = useState(false);

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const loadList = () => {
        API.post("/admin/task/list", {
            page: page,
            page_size: pageSize,
            order_by: orderBy.join(" "),
            conditions: filter,
            searches: search,
        })
            .then((response) => {
                setUsers(response.data.users);
                setTasks(response.data.items);
                setTotal(response.data.total);
                setSelected([]);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    };

    useEffect(() => {
        loadList();
    }, [page, pageSize, orderBy, filter, search]);

    const deletePolicy = (id) => {
        setLoading(true);
        API.post("/admin/task/delete", { id: [id] })
            .then(() => {
                loadList();
                ToggleSnackbar("top", "right", "任务已删除", "success");
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
    };

    const deleteBatch = () => {
        setLoading(true);
        API.post("/admin/task/delete", { id: selected })
            .then(() => {
                loadList();
                ToggleSnackbar("top", "right", "任务已删除", "success");
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
    };

    const handleSelectAllClick = (event) => {
        if (event.target.checked) {
            const newSelecteds = tasks.map((n) => n.ID);
            setSelected(newSelecteds);
            return;
        }
        setSelected([]);
    };

    const handleClick = (event, name) => {
        const selectedIndex = selected.indexOf(name);
        let newSelected = [];

        if (selectedIndex === -1) {
            newSelected = newSelected.concat(selected, name);
        } else if (selectedIndex === 0) {
            newSelected = newSelected.concat(selected.slice(1));
        } else if (selectedIndex === selected.length - 1) {
            newSelected = newSelected.concat(selected.slice(0, -1));
        } else if (selectedIndex > 0) {
            newSelected = newSelected.concat(
                selected.slice(0, selectedIndex),
                selected.slice(selectedIndex + 1)
            );
        }

        setSelected(newSelected);
    };

    const getError = (error) => {
        if (error === "") {
            return "-";
        }
        try {
            const res = JSON.parse(error);
            return res.msg;
        } catch (e) {
            return "未知";
        }
    };

    const isSelected = (id) => selected.indexOf(id) !== -1;

    return (
        <div>
            <ShareFilter
                filter={filter}
                open={filterDialog}
                onClose={() => setFilterDialog(false)}
                setSearch={setSearch}
                setFilter={setFilter}
            />
            <div className={classes.header}>
                <div className={classes.headerRight}>
                    <Button
                        color={"primary"}
                        onClick={() => loadList()}
                        variant={"outlined"}
                    >
                        刷新
                    </Button>
                </div>
            </div>

            <Paper square className={classes.tableContainer}>
                {selected.length > 0 && (
                    <Toolbar className={classes.highlight}>
                        <Typography
                            style={{ flex: "1 1 100%" }}
                            color="inherit"
                            variant="subtitle1"
                        >
                            已选择 {selected.length} 个对象
                        </Typography>
                        <Tooltip title="删除">
                            <IconButton
                                onClick={deleteBatch}
                                disabled={loading}
                                aria-label="delete"
                            >
                                <Delete />
                            </IconButton>
                        </Tooltip>
                    </Toolbar>
                )}
                <TableContainer className={classes.container}>
                    <Table aria-label="sticky table" size={"small"}>
                        <TableHead>
                            <TableRow style={{ height: 52 }}>
                                <TableCell padding="checkbox">
                                    <Checkbox
                                        indeterminate={
                                            selected.length > 0 &&
                                            selected.length < tasks.length
                                        }
                                        checked={
                                            tasks.length > 0 &&
                                            selected.length === tasks.length
                                        }
                                        onChange={handleSelectAllClick}
                                        inputProps={{
                                            "aria-label": "select all desserts",
                                        }}
                                    />
                                </TableCell>
                                <TableCell style={{ minWidth: 10 }}>
                                    <TableSortLabel
                                        active={orderBy[0] === "id"}
                                        direction={orderBy[1]}
                                        onClick={() =>
                                            setOrderBy([
                                                "id",
                                                orderBy[1] === "asc"
                                                    ? "desc"
                                                    : "asc",
                                            ])
                                        }
                                    >
                                        #
                                        {orderBy[0] === "id" ? (
                                            <span
                                                className={
                                                    classes.visuallyHidden
                                                }
                                            >
                                                {orderBy[1] === "desc"
                                                    ? "sorted descending"
                                                    : "sorted ascending"}
                                            </span>
                                        ) : null}
                                    </TableSortLabel>
                                </TableCell>
                                <TableCell style={{ minWidth: 130 }}>
                                    类型
                                </TableCell>
                                <TableCell style={{ minWidth: 90 }}>
                                    状态
                                </TableCell>
                                <TableCell style={{ minWidth: 90 }}>
                                    最后进度
                                </TableCell>
                                <TableCell style={{ minWidth: 150 }}>
                                    错误信息
                                </TableCell>
                                <TableCell style={{ minWidth: 100 }}>
                                    创建者
                                </TableCell>
                                <TableCell style={{ minWidth: 150 }}>
                                    创建于
                                </TableCell>
                                <TableCell style={{ minWidth: 80 }}>
                                    操作
                                </TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {tasks.map((row) => (
                                <TableRow
                                    hover
                                    key={row.ID}
                                    role="checkbox"
                                    selected={isSelected(row.ID)}
                                >
                                    <TableCell padding="checkbox">
                                        <Checkbox
                                            onClick={(event) =>
                                                handleClick(event, row.ID)
                                            }
                                            checked={isSelected(row.ID)}
                                        />
                                    </TableCell>
                                    <TableCell>{row.ID}</TableCell>
                                    <TableCell
                                        style={{ wordBreak: "break-all" }}
                                    >
                                        {getTaskType(row.Type)}
                                    </TableCell>
                                    <TableCell>
                                        {getTaskStatus(row.Status)}
                                    </TableCell>
                                    <TableCell>
                                        {getTaskProgress(
                                            row.Type,
                                            row.Progress
                                        )}
                                    </TableCell>
                                    <TableCell className={classes.noWrap}>
                                        {getError(row.Error)}
                                    </TableCell>
                                    <TableCell>
                                        <Link
                                            href={
                                                "/admin/user/edit/" + row.UserID
                                            }
                                        >
                                            {users[row.UserID]
                                                ? users[row.UserID].Nick
                                                : "未知"}
                                        </Link>
                                    </TableCell>
                                    <TableCell>
                                        {formatLocalTime(
                                            row.CreatedAt,
                                            "YYYY-MM-DD H:mm:ss"
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        <Tooltip title={"删除"}>
                                            <IconButton
                                                disabled={loading}
                                                onClick={() =>
                                                    deletePolicy(row.ID)
                                                }
                                                size={"small"}
                                            >
                                                <Delete />
                                            </IconButton>
                                        </Tooltip>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </TableContainer>
                <TablePagination
                    rowsPerPageOptions={[10, 25, 50, 100]}
                    component="div"
                    count={total}
                    rowsPerPage={pageSize}
                    page={page - 1}
                    onChangePage={(e, p) => setPage(p + 1)}
                    onChangeRowsPerPage={(e) => {
                        setPageSize(e.target.value);
                        setPage(1);
                    }}
                />
            </Paper>
        </div>
    );
}
