import Button from "@material-ui/core/Button";
import IconButton from "@material-ui/core/IconButton";
import Paper from "@material-ui/core/Paper";
import { makeStyles } from "@material-ui/core/styles";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TablePagination from "@material-ui/core/TablePagination";
import TableRow from "@material-ui/core/TableRow";
import Tooltip from "@material-ui/core/Tooltip";
import { Delete, Edit } from "@material-ui/icons";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { useHistory, useLocation } from "react-router";
import { toggleSnackbar } from "../../../actions";
import API from "../../../middleware/Api";
import { sizeToString } from "../../../utils";

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
}));

const columns = [
    { id: "#", label: "#", minWidth: 50 },
    { id: "name", label: "名称", minWidth: 170 },
    { id: "type", label: "存储策略", minWidth: 170 },
    {
        id: "count",
        label: "下属用户数",
        minWidth: 50,
        align: "right",
    },
    {
        id: "size",
        label: "最大容量",
        minWidth: 100,
        align: "right",
    },
    {
        id: "action",
        label: "操作",
        minWidth: 170,
        align: "right",
    },
];

function useQuery() {
    return new URLSearchParams(useLocation().search);
}

export default function Group() {
    const classes = useStyles();
    const [groups, setGroups] = useState([]);
    const [statics, setStatics] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [total, setTotal] = useState(0);
    const [policies, setPolicies] = React.useState({});

    const location = useLocation();
    const history = useHistory();
    const query = useQuery();

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const loadList = () => {
        API.post("/admin/group/list", {
            page: page,
            page_size: pageSize,
            order_by: "id desc",
        })
            .then((response) => {
                setGroups(response.data.items);
                setStatics(response.data.statics);
                setTotal(response.data.total);
                setPolicies(response.data.policies);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    };

    useEffect(() => {
        if (query.get("code") === "0") {
            ToggleSnackbar("top", "right", "授权成功", "success");
        } else if (query.get("msg") && query.get("msg") !== "") {
            ToggleSnackbar(
                "top",
                "right",
                query.get("msg") + ", " + query.get("err"),
                "warning"
            );
        }
    }, [location]);

    useEffect(() => {
        loadList();
    }, [page, pageSize]);

    const deletePolicy = (id) => {
        API.delete("/admin/group/" + id)
            .then(() => {
                loadList();
                ToggleSnackbar("top", "right", "用户组已删除", "success");
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    };

    return (
        <div>
            <div className={classes.header}>
                <Button
                    color={"primary"}
                    onClick={() => history.push("/admin/group/add")}
                    variant={"contained"}
                >
                    新建用户组
                </Button>
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
                <TableContainer className={classes.container}>
                    <Table aria-label="sticky table" size={"small"}>
                        <TableHead>
                            <TableRow style={{ height: 52 }}>
                                {columns.map((column) => (
                                    <TableCell
                                        key={column.id}
                                        align={column.align}
                                        style={{ minWidth: column.minWidth }}
                                    >
                                        {column.label}
                                    </TableCell>
                                ))}
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {groups.map((row) => (
                                <TableRow hover key={row.ID}>
                                    <TableCell>{row.ID}</TableCell>
                                    <TableCell>{row.Name}</TableCell>
                                    <TableCell>
                                        {row.PolicyList !== null &&
                                            row.PolicyList.map((pid, key) => {
                                                let res = "";
                                                if (policies[pid]) {
                                                    res += policies[pid].Name;
                                                }
                                                if (
                                                    key !==
                                                    row.PolicyList.length - 1
                                                ) {
                                                    res += " / ";
                                                }
                                                return res;
                                            })}
                                    </TableCell>
                                    <TableCell align={"right"}>
                                        {statics[row.ID] !== undefined &&
                                            statics[row.ID].toLocaleString()}
                                    </TableCell>
                                    <TableCell align={"right"}>
                                        {statics[row.ID] !== undefined &&
                                            sizeToString(row.MaxStorage)}
                                    </TableCell>
                                    <TableCell align={"right"}>
                                        <Tooltip title={"编辑"}>
                                            <IconButton
                                                onClick={() =>
                                                    history.push(
                                                        "/admin/group/edit/" +
                                                            row.ID
                                                    )
                                                }
                                                size={"small"}
                                            >
                                                <Edit />
                                            </IconButton>
                                        </Tooltip>
                                        <Tooltip title={"删除"}>
                                            <IconButton
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
