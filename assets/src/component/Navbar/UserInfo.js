import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";
import { setUserPopover } from "../../actions";
import { withStyles, Typography } from "@material-ui/core";
import Auth from "../../middleware/Auth";
import DarkModeSwitcher from "./DarkModeSwitcher";
import Avatar from "@material-ui/core/Avatar";

const mapStateToProps = (state) => {
    return {
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
    userNav: {
        height: "170px",
        backgroundColor: theme.palette.primary.main,
        padding: "20px 20px 2em",
        backgroundImage:
            "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1600 900'%3E%3Cpolygon fill='" +
            theme.palette.primary.light.replace("#", "%23") +
            "' points='957 450 539 900 1396 900'/%3E%3Cpolygon fill='" +
            theme.palette.primary.dark.replace("#", "%23") +
            "' points='957 450 872.9 900 1396 900'/%3E%3Cpolygon fill='" +
            theme.palette.secondary.main.replace("#", "%23") +
            "' points='-60 900 398 662 816 900'/%3E%3Cpolygon fill='" +
            theme.palette.secondary.dark.replace("#", "%23") +
            "' points='337 900 398 662 816 900'/%3E%3Cpolygon fill='" +
            theme.palette.secondary.light.replace("#", "%23") +
            "' points='1203 546 1552 900 876 900'/%3E%3Cpolygon fill='" +
            theme.palette.secondary.main.replace("#", "%23") +
            "' points='1203 546 1552 900 1162 900'/%3E%3Cpolygon fill='" +
            theme.palette.primary.dark.replace("#", "%23") +
            "' points='641 695 886 900 367 900'/%3E%3Cpolygon fill='" +
            theme.palette.primary.main.replace("#", "%23") +
            "' points='587 900 641 695 886 900'/%3E%3Cpolygon fill='" +
            theme.palette.secondary.light.replace("#", "%23") +
            "' points='1710 900 1401 632 1096 900'/%3E%3Cpolygon fill='" +
            theme.palette.secondary.dark.replace("#", "%23") +
            "' points='1710 900 1401 632 1365 900'/%3E%3Cpolygon fill='" +
            theme.palette.secondary.main.replace("#", "%23") +
            "' points='1210 900 971 687 725 900'/%3E%3Cpolygon fill='" +
            theme.palette.secondary.dark.replace("#", "%23") +
            "' points='943 900 1210 900 971 687'/%3E%3C/svg%3E\")",
        backgroundSize: "cover",
    },
    avatar: {
        display: "block",
        width: "70px",
        height: "70px",
        border: " 2px solid #fff",
        borderRadius: "50%",
        overflow: "hidden",
        boxShadow:
            "0 2px 5px 0 rgba(0,0,0,0.16), 0 2px 10px 0 rgba(0,0,0,0.12)",
    },
    avatarImg: {
        width: "66px",
        height: "66px",
    },
    nickName: {
        color: "#fff",
        marginLeft: "10px",
        marginTop: "15px",
        fontSize: "17px",
    },
    flexAvatar: {
        display: "flex",
        justifyContent: "space-between",
        alignItems: "end",
    },
    groupName: {
        marginLeft: "10px",
        color: "#ffffff",
        opacity: "0.54",
    },
    storageCircle: {
        width: "200px",
    },
});

class UserInfoCompoment extends Component {
    showUserInfo = (e) => {
        this.props.setUserPopover(e.currentTarget);
    };

    render() {
        const { classes } = this.props;
        const isLogin = Auth.Check(this.props.isLogin);
        const user = Auth.GetUser(this.props.isLogin);

        return (
            <div className={classes.userNav}>
                <div className={classes.flexAvatar}>
                    {/* eslint-disable-next-line */}
                    <a onClick={this.showUserInfo} className={classes.avatar}>
                        {isLogin && (
                            <Avatar
                                src={"/api/v3/user/avatar/" + user.id + "/l"}
                                className={classes.avatarImg}
                            />
                        )}
                        {!isLogin && (
                            <Avatar
                                src={"/api/v3/user/avatar/0/l"}
                                className={classes.avatarImg}
                            />
                        )}
                    </a>
                    <DarkModeSwitcher position="left" />
                </div>
                <div className={classes.storageCircle}>
                    <Typography
                        className={classes.nickName}
                        component="h2"
                        noWrap
                    >
                        {isLogin ? user.nickname : "未登录"}
                    </Typography>
                    <Typography
                        className={classes.groupName}
                        component="h2"
                        color="textSecondary"
                        noWrap
                    >
                        {isLogin ? user.group.name : "游客"}
                    </Typography>
                </div>
            </div>
        );
    }
}

UserInfoCompoment.propTypes = {
    classes: PropTypes.object.isRequired,
};

const UserInfo = connect(
    mapStateToProps,
    mapDispatchToProps
)(withStyles(styles)(UserInfoCompoment));

export default UserInfo;
