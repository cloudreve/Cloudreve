import React, { Component } from "react";
import PropTypes from "prop-types";
import FolderIcon from "@material-ui/icons/Folder";
import RightIcon from "@material-ui/icons/KeyboardArrowRight";
import UpIcon from "@material-ui/icons/ArrowUpward";
import { connect } from "react-redux";
import classNames from "classnames";
import { toggleSnackbar } from "../../actions/index";

import {
    MenuList,
    MenuItem,
    IconButton,
    ListItemIcon,
    ListItemText,
    withStyles,
    ListItemSecondaryAction,
} from "@material-ui/core";
import API from "../../middleware/Api";

const mapStateToProps = (state) => {
    return {
        keywords: state.explorer.keywords,
    };
};

const mapDispatchToProps = (dispatch) => {
    return {
        toggleSnackbar: (vertical, horizontal, msg, color) => {
            dispatch(toggleSnackbar(vertical, horizontal, msg, color));
        },
    };
};

const styles = (theme) => ({
    iconWhite: {
        color: theme.palette.common.white,
    },
    selected: {
        backgroundColor: theme.palette.primary.main + "!important",
        "& $primary, & $icon": {
            color: theme.palette.common.white,
        },
    },
    primary: {},
    icon: {},
    buttonIcon: {},
    selector: {
        minWidth: "300px",
    },
    container: {
        maxHeight: "330px",
        overflowY: " auto",
    },
});

class PathSelectorCompoment extends Component {
    state = {
        presentPath: "/",
        dirList: [],
        selectedTarget: null,
    };

    componentDidMount = () => {
        const toBeLoad = this.props.presentPath;
        this.enterFolder(this.props.keywords === "" ? toBeLoad : "/");
    };

    back = () => {
        const paths = this.state.presentPath.split("/");
        paths.pop();
        const toBeLoad = paths.join("/");
        this.enterFolder(toBeLoad === "" ? "/" : toBeLoad);
    };

    enterFolder = (toBeLoad) => {
        API.get(
            (this.props.api ? this.props.api : "/directory") +
                encodeURIComponent(toBeLoad)
        )
            .then((response) => {
                const dirList = response.data.objects.filter((x) => {
                    return (
                        x.type === "dir" &&
                        this.props.selected.findIndex((value) => {
                            return (
                                value.name === x.name && value.path === x.path
                            );
                        }) === -1
                    );
                });
                if (toBeLoad === "/") {
                    dirList.unshift({ name: "/", path: "" });
                }
                this.setState({
                    presentPath: toBeLoad,
                    dirList: dirList,
                    selectedTarget: null,
                });
            })
            .catch((error) => {
                this.props.toggleSnackbar(
                    "top",
                    "right",
                    error.message,
                    "warning"
                );
            });
    };

    handleSelect = (index) => {
        this.setState({ selectedTarget: index });
        this.props.onSelect(this.state.dirList[index]);
    };

    render() {
        const { classes } = this.props;

        return (
            <div className={classes.container}>
                <MenuList className={classes.selector}>
                    {this.state.presentPath !== "/" && (
                        <MenuItem onClick={this.back}>
                            <ListItemIcon>
                                <UpIcon />
                            </ListItemIcon>
                            <ListItemText primary="返回上一层" />
                        </MenuItem>
                    )}
                    {this.state.dirList.map((value, index) => (
                        <MenuItem
                            classes={{
                                selected: classes.selected,
                            }}
                            key={index}
                            selected={this.state.selectedTarget === index}
                            onClick={() => this.handleSelect(index)}
                        >
                            <ListItemIcon className={classes.icon}>
                                <FolderIcon />
                            </ListItemIcon>
                            <ListItemText
                                classes={{ primary: classes.primary }}
                                primary={value.name}
                                primaryTypographyProps={{
                                    style: { whiteSpace: "normal" },
                                }}
                            />
                            {value.name !== "/" && (
                                <ListItemSecondaryAction
                                    className={classes.buttonIcon}
                                >
                                    <IconButton
                                        className={classNames({
                                            [classes.iconWhite]:
                                                this.state.selectedTarget ===
                                                index,
                                        })}
                                        onClick={() =>
                                            this.enterFolder(
                                                value.path === "/"
                                                    ? value.path + value.name
                                                    : value.path +
                                                          "/" +
                                                          value.name
                                            )
                                        }
                                    >
                                        <RightIcon />
                                    </IconButton>
                                </ListItemSecondaryAction>
                            )}
                        </MenuItem>
                    ))}
                </MenuList>
            </div>
        );
    }
}

PathSelectorCompoment.propTypes = {
    classes: PropTypes.object.isRequired,
    presentPath: PropTypes.string.isRequired,
    selected: PropTypes.array.isRequired,
};

export default connect(
    mapStateToProps,
    mapDispatchToProps
)(withStyles(styles)(PathSelectorCompoment));
