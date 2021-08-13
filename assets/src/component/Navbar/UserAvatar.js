import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";
import SettingIcon from "@material-ui/icons/Settings";
import UserAvatarPopover from "./UserAvatarPopover";
import { AccountCircle } from "mdi-material-ui";
import { setUserPopover } from "../../actions";
import Auth from "../../middleware/Auth";
import {
    withStyles,
    Grow,
    Avatar,
    IconButton,
    Tooltip,
} from "@material-ui/core";
import { withRouter } from "react-router-dom";
import pathHelper from "../../utils/page";
import DarkModeSwitcher from "./DarkModeSwitcher";
import { Home } from "@material-ui/icons";

const mapStateToProps = (state) => {
    return {
        selected: state.explorer.selected,
        isMultiple: state.explorer.selectProps.isMultiple,
        withFolder: state.explorer.selectProps.withFolder,
        withFile: state.explorer.selectProps.withFile,
        isLogin: state.viewUpdate.isLogin,
    };
};

const mapDispatchToProps = (dispatch) => {
    return {
        setUserPopover: (anchor) => {
            dispatch(setUserPopover(anchor));
        },
    };
};

const styles = (theme) => ({
    mobileHidden: {
        [theme.breakpoints.down("xs")]: {
            display: "none",
        },
        whiteSpace: "nowrap",
    },
    avatar: {
        width: "30px",
        height: "30px",
    },
    header: {
        display: "flex",
        padding: "20px 20px 20px 20px",
    },
    largeAvatar: {
        height: "90px",
        width: "90px",
    },
    info: {
        marginLeft: "10px",
        width: "139px",
    },
    badge: {
        marginTop: "10px",
    },
    visitorMenu: {
        width: 200,
    },
});

class UserAvatarCompoment extends Component {
    state = {
        anchorEl: null,
    };

    showUserInfo = (e) => {
        this.props.setUserPopover(e.currentTarget);
    };

    handleClose = () => {
        this.setState({
            anchorEl: null,
        });
    };

    openURL = (url) => {
        window.location.href = url;
    };

    returnHome = () => {
        window.location.href = "/home";
    };

    render() {
        const { classes } = this.props;
        const loginCheck = Auth.Check(this.props.isLogin);
        const user = Auth.GetUser(this.props.isLogin);
        const isAdminPage = pathHelper.isAdminPage(
            this.props.location.pathname
        );

        return (
            <div className={classes.mobileHidden}>
                <Grow
                    in={
                        this.props.selected.length <= 1 &&
                        !(!this.props.isMultiple && this.props.withFile)
                    }
                >
                    <div>
                        {!isAdminPage && (
                            <>
                                <DarkModeSwitcher position="top" />
                                {loginCheck && (
                                    <>
                                        <Tooltip
                                            title={"设置"}
                                            placement="bottom"
                                        >
                                            <IconButton
                                                onClick={() =>
                                                    this.props.history.push(
                                                        "/setting?"
                                                    )
                                                }
                                                color="inherit"
                                            >
                                                <SettingIcon />
                                            </IconButton>
                                        </Tooltip>
                                    </>
                                )}
                            </>
                        )}
                        {isAdminPage && (
                            <Tooltip title={"返回主页"} placement="bottom">
                                <IconButton
                                    color="inherit"
                                    onClick={this.returnHome}
                                >
                                    <Home />
                                </IconButton>
                            </Tooltip>
                        )}
                        <IconButton color="inherit" onClick={this.showUserInfo}>
                            {!loginCheck && <AccountCircle />}
                            {loginCheck && (
                                <Avatar
                                    src={
                                        "/api/v3/user/avatar/" + user.id + "/s"
                                    }
                                    className={classes.avatar}
                                />
                            )}
                        </IconButton>{" "}
                    </div>
                </Grow>
                <UserAvatarPopover />
            </div>
        );
    }
}

UserAvatarCompoment.propTypes = {
    classes: PropTypes.object.isRequired,
};

const UserAvatar = connect(
    mapStateToProps,
    mapDispatchToProps
)(withStyles(styles)(withRouter(UserAvatarCompoment)));

export default UserAvatar;
