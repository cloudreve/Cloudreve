import React, { Suspense, useCallback, useEffect, useState } from "react";
import PageLoading from "../Placeholder/PageLoading";
import { useParams } from "react-router";
import API from "../../middleware/Api";
import { toggleSnackbar } from "../../actions";
import { changeSubTitle } from "../../redux/viewUpdate/action";
import { useDispatch } from "react-redux";
import Notice from "./NotFound";
import LockedFile from "./LockedFile";
import SharedFile from "./SharedFile";
import SharedFolder from "./SharedFolder";

export default function SharePreload() {
    const dispatch = useDispatch();
    const { id } = useParams();

    const [share, setShare] = useState(undefined);
    const [loading, setLoading] = useState(false);
    const [password, setPassword] = useState("");

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
        if (share) {
            if (share.locked) {
                SetSubTitle(share.creator.nick + "的加密分享");
                if (password !== "") {
                    ToggleSnackbar("top", "right", "密码不正确", "warning");
                }
            } else {
                SetSubTitle(share.source.name);
            }
        } else {
            SetSubTitle();
        }
    }, [share, SetSubTitle, ToggleSnackbar]);

    useEffect(() => {
        return () => {
            SetSubTitle();
        };
        // eslint-disable-next-line
    }, []);

    useEffect(() => {
        setLoading(true);
        let withPassword = "";
        if (password !== "") {
            withPassword = "?password=" + password;
        }
        API.get("/share/info/" + id + withPassword)
            .then((response) => {
                setShare(response.data);
                setLoading(false);
            })
            .catch((error) => {
                setLoading(false);
                if (error.code === 404) {
                    setShare(null);
                } else {
                    ToggleSnackbar("top", "right", error.message, "error");
                }
            });
    }, [id, password, ToggleSnackbar]);

    return (
        <Suspense fallback={<PageLoading />}>
            {share === undefined && <PageLoading />}
            {share === null && <Notice msg={"分享不存在或已过期"} />}
            {share && share.locked && (
                <LockedFile
                    loading={loading}
                    setPassowrd={setPassword}
                    share={share}
                />
            )}
            {share && !share.locked && !share.is_dir && (
                <SharedFile share={share} />
            )}
            {share && !share.locked && share.is_dir && (
                <SharedFolder share={share} />
            )}
        </Suspense>
    );
}
