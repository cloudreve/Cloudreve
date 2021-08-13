import React, { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router";
import API from "../../../middleware/Api";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../../actions";
import UserForm from "./UserForm";

export default function EditUserPreload() {
    const [user, setUser] = useState({});

    const { id } = useParams();

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    useEffect(() => {
        setUser({});
        API.get("/admin/user/" + id)
            .then((response) => {
                // 整型转换
                ["Status", "GroupID"].forEach((v) => {
                    response.data[v] = response.data[v].toString();
                });
                setUser(response.data);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    }, [id]);

    return <div>{user.ID !== undefined && <UserForm user={user} />}</div>;
}
