import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";
import classNames from "classnames";
import {
    withStyles,
    ButtonBase,
    Typography,
    Tooltip,
    fade,
} from "@material-ui/core";
import TypeIcon from "./TypeIcon";
import { lighten } from "@material-ui/core/styles";
import CheckCircleRoundedIcon from "@material-ui/icons/CheckCircleRounded";
import statusHelper from "../../utils/page";
import Grow from "@material-ui/core/Grow";
import { Folder } from "@material-ui/icons";

const styles = (theme) => ({
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
            "background-color 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms,box-shadow 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms,border 250ms cubic-bezier(0.4, 0, 0.2, 1) 0ms",
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
    checkIcon: {
        color: theme.palette.primary.main,
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

class SmallIconCompoment extends Component {
    state = {};

    render() {
        const { classes } = this.props;
        const isSelected =
            this.props.selected.findIndex((value) => {
                return value === this.props.file;
            }) !== -1;
        const isMobile = statusHelper.isMobile();

        return (
            <ButtonBase
                focusRipple
                className={classNames(
                    {
                        [classes.selected]: isSelected,
                        [classes.notSelected]: !isSelected,
                    },
                    classes.button
                )}
            >
                <div
                    onClick={this.props.onIconClick}
                    className={classNames(classes.icon, {
                        [classes.iconSelected]: isSelected,
                        [classes.iconNotSelected]: !isSelected,
                    })}
                >
                    {(!isSelected || !isMobile) && (
                        <>
                            {this.props.isFolder && <Folder />}
                            {!this.props.isFolder && (
                                <TypeIcon fileName={this.props.file.name} />
                            )}
                        </>
                    )}
                    {isSelected && isMobile && (
                        <Grow in={isSelected && isMobile}>
                            <CheckCircleRoundedIcon
                                className={classes.checkIcon}
                            />
                        </Grow>
                    )}
                </div>
                <Tooltip
                    title={this.props.file.name}
                    aria-label={this.props.file.name}
                >
                    <Typography
                        className={classNames(classes.folderName, {
                            [classes.folderNameSelected]: isSelected,
                            [classes.folderNameNotSelected]: !isSelected,
                        })}
                        variant="body2"
                    >
                        {this.props.file.name}
                    </Typography>
                </Tooltip>
            </ButtonBase>
        );
    }
}

SmallIconCompoment.propTypes = {
    classes: PropTypes.object.isRequired,
    file: PropTypes.object.isRequired,
};

const SmallIcon = connect(
    mapStateToProps,
    mapDispatchToProps
)(withStyles(styles)(SmallIconCompoment));

export default SmallIcon;
