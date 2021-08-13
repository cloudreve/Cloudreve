import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";

import FolderIcon from "@material-ui/icons/Folder";
import classNames from "classnames";
import { sizeToString } from "../../utils/index";
import {
    withStyles,
    TableCell,
    TableRow,
    Typography,
    fade,
} from "@material-ui/core";
import TypeIcon from "./TypeIcon";
import { lighten } from "@material-ui/core/styles";
import pathHelper from "../../utils/page";
import { withRouter } from "react-router";
import KeyboardReturnIcon from "@material-ui/icons/KeyboardReturn";
import CheckCircleRoundedIcon from "@material-ui/icons/CheckCircleRounded";
import statusHelper from "../../utils/page";
import Grow from "@material-ui/core/Grow";
import dayjs from "dayjs";
import { formatLocalTime } from "../../utils/datetime";

const styles = (theme) => ({
    selected: {
        "&:hover": {},
        backgroundColor: fade(theme.palette.primary.main, 0.18),
    },

    selectedShared: {
        "&:hover": {},
        backgroundColor: fade(theme.palette.primary.main, 0.18),
    },

    notSelected: {
        "&:hover": {
            backgroundColor: theme.palette.background.default,
        },
    },
    icon: {
        verticalAlign: "middle",
        marginRight: "20px",
        color: theme.palette.text.secondary,
    },
    tableIcon: {
        marginRight: "20px",
        verticalAlign: "middle",
    },
    folderNameSelected: {
        color:
            theme.palette.type === "dark" ? "#fff" : theme.palette.primary.dark,
        fontWeight: "500",
        userSelect: "none",
    },
    folderNameNotSelected: {
        color: theme.palette.text.secondary,
        userSelect: "none",
    },
    folderName: {
        marginRight: "20px",
        display: "flex",
        alignItems: "center",
    },
    hideAuto: {
        [theme.breakpoints.down("sm")]: {
            display: "none",
        },
    },
    tableRow: {
        padding: "10px 16px",
    },
    checkIcon: {
        color: theme.palette.primary.main,
    },
    active: {
        backgroundColor: fade(theme.palette.primary.main, 0.1),
    },
});

const mapStateToProps = (state) => {
    return {
        selected: state.explorer.selected,
    };
};

const mapDispatchToProps = () => {
    return {};
};

class TableRowCompoment extends Component {
    state = {};

    render() {
        const { classes } = this.props;
        const isShare = pathHelper.isSharePage(this.props.location.pathname);

        let icon;
        if (this.props.file.type === "dir") {
            icon = <FolderIcon className={classes.icon} />;
        } else if (this.props.file.type === "up") {
            icon = <KeyboardReturnIcon className={classes.icon} />;
        } else {
            icon = (
                <TypeIcon
                    className={classes.tableIcon}
                    fileName={this.props.file.name}
                />
            );
        }
        const isSelected =
            this.props.selected.findIndex((value) => {
                return value === this.props.file;
            }) !== -1;
        const isMobile = statusHelper.isMobile();

        return (
            <TableRow
                ref={this.props.pref}
                onContextMenu={this.props.contextMenu}
                onClick={this.props.handleClick}
                onDoubleClick={this.props.handleDoubleClick.bind(this)}
                className={classNames({
                    [classes.selected]: isSelected && !isShare,
                    [classes.selectedShared]: isSelected && isShare,
                    [classes.notSelected]: !isSelected,
                    [classes.active]: this.props.isActive,
                })}
            >
                <TableCell
                    ref={this.props.dref}
                    component="th"
                    scope="row"
                    className={classes.tableRow}
                >
                    <Typography
                        variant="body2"
                        className={classNames(classes.folderName, {
                            [classes.folderNameSelected]: isSelected,
                            [classes.folderNameNotSelected]: !isSelected,
                        })}
                    >
                        <div
                            onClick={
                                this.props.file.type !== "up"
                                    ? this.props.onIconClick
                                    : null
                            }
                        >
                            {(!isSelected || !isMobile) && icon}
                            {isSelected && isMobile && (
                                <Grow in={isSelected && isMobile}>
                                    <CheckCircleRoundedIcon
                                        className={classNames(
                                            classes.checkIcon,
                                            classes.icon
                                        )}
                                    />
                                </Grow>
                            )}
                        </div>
                        {this.props.file.name}
                    </Typography>
                </TableCell>
                <TableCell
                    className={classNames(classes.hideAuto, classes.tableRow)}
                >
                    <Typography
                        variant="body2"
                        className={classNames(classes.folderName, {
                            [classes.folderNameSelected]: isSelected,
                            [classes.folderNameNotSelected]: !isSelected,
                        })}
                    >
                        {" "}
                        {this.props.file.type !== "dir" &&
                            this.props.file.type !== "up" &&
                            sizeToString(this.props.file.size)}
                    </Typography>
                </TableCell>
                <TableCell
                    className={classNames(classes.hideAuto, classes.tableRow)}
                >
                    <Typography
                        variant="body2"
                        className={classNames(classes.folderName, {
                            [classes.folderNameSelected]: isSelected,
                            [classes.folderNameNotSelected]: !isSelected,
                        })}
                    >
                        {" "}
                        {formatLocalTime(
                            this.props.file.date,
                            "YYYY-MM-DD H:mm:ss"
                        )}
                    </Typography>
                </TableCell>
            </TableRow>
        );
    }
}

TableRowCompoment.propTypes = {
    classes: PropTypes.object.isRequired,
    file: PropTypes.object.isRequired,
};

const TableItem = connect(
    mapStateToProps,
    mapDispatchToProps
)(withStyles(styles)(withRouter(TableRowCompoment)));

export default TableItem;
