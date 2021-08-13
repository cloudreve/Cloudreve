import React from "react";
import FolderIcon from "@material-ui/icons/Folder";
import classNames from "classnames";
import {
    ButtonBase,
    Typography,
    Tooltip,
    makeStyles,
    fade,
} from "@material-ui/core";
import { useSelector } from "react-redux";
import { lighten } from "@material-ui/core/styles";
import statusHelper from "../../utils/page";
import TypeIcon from "./TypeIcon";
import CheckCircleRoundedIcon from "@material-ui/icons/CheckCircleRounded";
const useStyles = makeStyles((theme) => ({
    container: {
        padding: "7px",
    },

    selected: {
        "&:hover": {
            border: "1px solid #d0d0d0",
        },
        backgroundColor: fade(
            theme.palette.primary.main,
            theme.palette.type === "dark" ? 0.3 : 0.18
        ),
    },

    notSelected: {
        "&:hover": {
            backgroundColor: theme.palette.background.default,
            border: "1px solid #d0d0d0",
        },
        backgroundColor: theme.palette.background.paper,
    },

    button: {
        height: "50px",
        border: "1px solid " + theme.palette.divider,
        width: "100%",
        borderRadius: "6px",
        boxSizing: "border-box",
        transition:
            "background-color 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms,box-shadow 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms",
        display: "flex",
        justifyContent: "left",
        alignItems: "initial",
    },
    icon: {
        margin: "10px 10px 10px 16px",
        height: "30px",
        minWidth: "30px",
        backgroundColor: theme.palette.background.paper,
        borderRadius: "90%",
        paddingTop: "3px",
        color: theme.palette.text.secondary,
    },
    folderNameSelected: {
        color:
            theme.palette.type === "dark" ? "#fff" : theme.palette.primary.dark,
        fontWeight: "500",
    },
    folderNameNotSelected: {
        color: theme.palette.text.secondary,
    },
    folderName: {
        marginTop: "15px",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
        overflow: "hidden",
        marginRight: "20px",
    },
    active: {
        boxShadow: "0 0 0 2px " + theme.palette.primary.light,
    },
    checkIcon: {
        color: theme.palette.primary.main,
    },
}));

export default function Folder({ folder, isActive, onIconClick }) {
    const selected = useSelector((state) => state.explorer.selected);
    const classes = useStyles();
    const isMobile = statusHelper.isMobile();
    const isSelected =
        selected.findIndex((value) => {
            return value === folder;
        }) !== -1;

    return (
        <ButtonBase
            focusRipple
            className={classNames(
                {
                    [classes.selected]: isSelected,
                    [classes.notSelected]: !isSelected,
                    [classes.active]: isActive,
                },
                classes.button
            )}
        >
            <div
                onClick={onIconClick}
                className={classNames(classes.icon, {
                    [classes.iconSelected]: isSelected,
                    [classes.iconNotSelected]: !isSelected,
                })}
            >
                {(!isSelected || !isMobile) && <FolderIcon />}
                {isSelected && isMobile && (
                    <CheckCircleRoundedIcon className={classes.checkIcon} />
                )}
            </div>
            <Tooltip title={folder.name} aria-label={folder.name}>
                <Typography
                    variant="body2"
                    className={classNames(classes.folderName, {
                        [classes.folderNameSelected]: isSelected,
                        [classes.folderNameNotSelected]: !isSelected,
                    })}
                >
                    {folder.name}
                </Typography>
            </Tooltip>
        </ButtonBase>
    );
}
