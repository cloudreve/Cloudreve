import { lighten } from "@material-ui/core";
import Badge from "@material-ui/core/Badge";
import Button from "@material-ui/core/Button";
import Checkbox from "@material-ui/core/Checkbox";
import IconButton from "@material-ui/core/IconButton";
import Link from "@material-ui/core/Link";
import Paper from "@material-ui/core/Paper";
import { makeStyles, useTheme } from "@material-ui/core/styles";
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
import { Block, Delete, Edit, FilterList } from "@material-ui/icons";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { useHistory } from "react-router";
import { toggleSnackbar } from "../../../actions";
import API from "../../../middleware/Api";
import { sizeToString } from "../../../utils";
import UserFilter from "../Dialogs/UserFilter";

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

export default function Group() {
    const classes = useStyles();
    const [users, setUsers] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [total, setTotal] = useState(0);
    const [filter, setFilter] = useState({});
    const [search, setSearch] = useState({});
    const [orderBy, setOrderBy] = useState(["id", "desc"]);
    const [filterDialog, setFilterDialog] = useState(false);
    const [selected, setSelected] = useState([]);
    const [loading, setLoading] = useState(false);

    const history = useHistory();
    const theme = useTheme();

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const loadList = () => {
        API.post("/admin/user/list", {
            page: page,
            page_size: pageSize,
            order_by: orderBy.join(" "),
            conditions: filter,
            searches: search,
        })
            .then((response) => {
                setUsers(response.data.items);
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
        API.post("/admin/user/delete", { id: [id] })
            .then(() => {
                loadList();
                ToggleSnackbar("top", "right", "用户已删除", "success");
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
        API.post("/admin/user/delete", { id: selected })
            .then(() => {
                loadList();
                ToggleSnackbar("top", "right", "用户已删除", "success");
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            })
            .then(() => {
                setLoading(false);
            });
    };

    const block = (id) => {
        setLoading(true);
        API.patch("/admin/user/ban/" + id)
            .then((response) => {
                setUsers(
                    users.map((v) => {
                        if (v.ID === id) {
                            const newUser = { ...v, Status: response.data };
                            return newUser;
                        }
                        return v;
                    })
                );
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
            const newSelecteds = users.map((n) => n.ID);
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

    const isSelected = (id) => selected.indexOf(id) !== -1;

    return (
        <div>
            <UserFilter
                filter={filter}
                open={filterDialog}
                onClose={() => setFilterDialog(false)}
                setSearch={setSearch}
                setFilter={setFilter}
            />
            <div className={classes.header}>
                <Button
                    style={{ alignSelf: "center" }}
                    color={"primary"}
                    onClick={() => history.push("/admin/user/add")}
                    variant={"contained"}
                >
                    新建用户
                </Button>
                <div className={classes.headerRight}>
                    <Tooltip title="过滤">
                        <IconButton
                            style={{ marginRight: 8 }}
                            onClick={() => setFilterDialog(true)}
                        >
                            <Badge
                                color="secondary"
                                variant="dot"
                                invisible={
                                    Object.keys(search).length === 0 &&
                                    Object.keys(filter).length === 0
                                }
                            >
                                <FilterList />
                            </Badge>
                        </IconButton>
                    </Tooltip>
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
                                            selected.length < users.length
                                        }
                                        checked={
                                            users.length > 0 &&
                                            selected.length === users.length
                                        }
                                        onChange={handleSelectAllClick}
                                        inputProps={{
                                            "aria-label": "select all desserts",
                                        }}
                                    />
                                </TableCell>
                                <TableCell style={{ minWidth: 59 }}>
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
                                <TableCell style={{ minWidth: 120 }}>
                                    <TableSortLabel
                                        active={orderBy[0] === "nick"}
                                        direction={orderBy[1]}
                                        onClick={() =>
                                            setOrderBy([
                                                "nick",
                                                orderBy[1] === "asc"
                                                    ? "desc"
                                                    : "asc",
                                            ])
                                        }
                                    >
                                        昵称
                                        {orderBy[0] === "nick" ? (
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
                                <TableCell style={{ minWidth: 170 }}>
                                    <TableSortLabel
                                        active={orderBy[0] === "email"}
                                        direction={orderBy[1]}
                                        onClick={() =>
                                            setOrderBy([
                                                "email",
                                                orderBy[1] === "asc"
                                                    ? "desc"
                                                    : "asc",
                                            ])
                                        }
                                    >
                                        Email
                                        {orderBy[0] === "email" ? (
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
                                <TableCell style={{ minWidth: 70 }}>
                                    用户组
                                </TableCell>
                                <TableCell style={{ minWidth: 50 }}>
                                    状态
                                </TableCell>
                                <TableCell
                                    align={"right"}
                                    style={{ minWidth: 80 }}
                                >
                                    <TableSortLabel
                                        active={orderBy[0] === "storage"}
                                        direction={orderBy[1]}
                                        onClick={() =>
                                            setOrderBy([
                                                "storage",
                                                orderBy[1] === "asc"
                                                    ? "desc"
                                                    : "asc",
                                            ])
                                        }
                                    >
                                        已用空间
                                        {orderBy[0] === "storage" ? (
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
                                <TableCell style={{ minWidth: 100 }}>
                                    操作
                                </TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {users.map((row) => (
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
                                    <TableCell>{row.Nick}</TableCell>
                                    <TableCell>{row.Email}</TableCell>
                                    <TableCell>
                                        <Link
                                            href={
                                                "/admin/group/edit/" +
                                                row.Group.ID
                                            }
                                        >
                                            {row.Group.Name}
                                        </Link>
                                    </TableCell>
                                    <TableCell>
                                        {row.Status === 0 && (
                                            <Typography
                                                style={{
                                                    color:
                                                        theme.palette.success
                                                            .main,
                                                }}
                                                variant={"body2"}
                                            >
                                                正常
                                            </Typography>
                                        )}
                                        {row.Status === 1 && (
                                            <Typography
                                                color={"textSecondary"}
                                                variant={"body2"}
                                            >
                                                未激活
                                            </Typography>
                                        )}
                                        {row.Status === 2 && (
                                            <Typography
                                                color={"error"}
                                                variant={"body2"}
                                            >
                                                被封禁
                                            </Typography>
                                        )}
                                        {row.Status === 3 && (
                                            <Typography
                                                color={"error"}
                                                variant={"body2"}
                                            >
                                                超额封禁
                                            </Typography>
                                        )}
                                    </TableCell>
                                    <TableCell align={"right"}>
                                        {sizeToString(row.Storage)}
                                    </TableCell>
                                    <TableCell>
                                        <Tooltip title={"编辑"}>
                                            <IconButton
                                                onClick={() =>
                                                    history.push(
                                                        "/admin/user/edit/" +
                                                            row.ID
                                                    )
                                                }
                                                size={"small"}
                                            >
                                                <Edit />
                                            </IconButton>
                                        </Tooltip>
                                        <Tooltip title={"封禁/解封"}>
                                            <IconButton
                                                disabled={loading}
                                                onClick={() => block(row.ID)}
                                                size={"small"}
                                            >
                                                <Block />
                                            </IconButton>
                                        </Tooltip>
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
                    rowsPerPageOptions={[10, 25, 100]}
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
