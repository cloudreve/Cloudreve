import React, { useCallback } from "react";
import { IconButton, makeStyles } from "@material-ui/core";
import DayIcon from "@material-ui/icons/Brightness7";
import NightIcon from "@material-ui/icons/Brightness4";
import { useDispatch, useSelector } from "react-redux";
import { toggleDaylightMode } from "../../actions";
import Tooltip from "@material-ui/core/Tooltip";
import Auth from "../../middleware/Auth";
import classNames from "classnames";

const useStyles = makeStyles(() => ({
    icon: {
        color: "rgb(255, 255, 255)",
        opacity: "0.54",
    },
}));

const DarkModeSwitcher = ({ position }) => {
    const ThemeType = useSelector(
        (state) => state.siteConfig.theme.palette.type
    );
    const dispatch = useDispatch();
    const ToggleThemeMode = useCallback(() => dispatch(toggleDaylightMode()), [
        dispatch,
    ]);
    const isDayLight = (ThemeType && ThemeType === "light") || !ThemeType;
    const isDark = ThemeType && ThemeType === "dark";
    const toggleMode = () => {
        Auth.SetPreference("theme_mode", isDayLight ? "dark" : "light");
        ToggleThemeMode();
    };
    const classes = useStyles();
    return (
        <Tooltip
            title={isDayLight ? "切换到深色模式" : "切换到浅色模式"}
            placement="bottom"
        >
            <IconButton
                className={classNames({
                    [classes.icon]: "left" === position,
                })}
                onClick={toggleMode}
                color="inherit"
            >
                {isDayLight && <NightIcon />}
                {isDark && <DayIcon />}
            </IconButton>
        </Tooltip>
    );
};

export default DarkModeSwitcher;
