import { Link, makeStyles } from "@material-ui/core";
import React, { useEffect, useState } from "react";
import { useSelector } from "react-redux";
import { useLocation } from "react-router";
import pageHelper from "../../utils/page";

const useStyles = makeStyles(() => ({
    icp: {
        padding: "8px 24px",
        position: "absolute",
        bottom: 0,
    },
}));

export const ICPFooter = () => {
    const siteICPId = useSelector((state) => state.siteConfig.siteICPId);
    const classes = useStyles();
    const location = useLocation();
    const [show, setShow] = useState(true);

    useEffect(() => {
        // 只在分享和登录界面显示
        const isSharePage = pageHelper.isSharePage(location.pathname);
        const isLoginPage = pageHelper.isLoginPage(location.pathname);
        setShow(siteICPId && (isSharePage || isLoginPage));
    }, [siteICPId, location]);

    if (!show) {
        return <></>;
    }
    return (
        <div className={classes.icp}>
            {`备案号: `}
            <Link
                href="https://beian.miit.gov.cn/"
                rel="noopener noreferrer"
                target="_blank"
            >
                {siteICPId}
            </Link>
        </div>
    );
};
