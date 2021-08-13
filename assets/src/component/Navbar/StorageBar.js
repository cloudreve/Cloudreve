import React, { Component } from "react";
import PropTypes from "prop-types";
import StorageIcon from "@material-ui/icons/Storage";
import { connect } from "react-redux";
import API from "../../middleware/Api";
import { sizeToString } from "../../utils";
import { toggleSnackbar } from "../../actions";

import {
    withStyles,
    LinearProgress,
    Typography,
    Divider,
    Tooltip,
} from "@material-ui/core";
import ButtonBase from "@material-ui/core/ButtonBase";
import { withRouter } from "react-router";

const mapStateToProps = (state) => {
    return {
        refresh: state.viewUpdate.storageRefresh,
        isLogin: state.viewUpdate.isLogin,
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
    iconFix: {
        marginLeft: "32px",
        marginRight: "17px",
        color: theme.palette.text.secondary,
        marginTop: "2px",
    },
    textFix: {
        padding: " 0 0 0 16px",
    },
    storageContainer: {
        display: "flex",
        marginTop: "15px",
        textAlign: "left",
        marginBottom: "11px",
    },
    detail: {
        width: "100%",
        marginRight: "35px",
    },
    info: {
        width: "131px",
        overflow: "hidden",
        textOverflow: "ellipsis",
        [theme.breakpoints.down("xs")]: {
            width: "162px",
        },
        marginTop: "5px",
    },
    bar: {
        marginTop: "5px",
    },
    stickFooter: {
        backgroundColor: theme.palette.background.paper,
    },
});

class StorageBarCompoment extends Component {
    state = {
        percent: 0,
        used: null,
        total: null,
        showExpand: false,
    };

    firstLoad = true;

    componentDidMount = () => {
        if (this.firstLoad && this.props.isLogin) {
            this.firstLoad = !this.firstLoad;
            this.updateStatus();
        }
    };

    componentWillUnmount() {
        this.firstLoad = false;
    }

    UNSAFE_componentWillReceiveProps = (nextProps) => {
        if (
            (this.props.isLogin && this.props.refresh !== nextProps.refresh) ||
            (this.props.isLogin !== nextProps.isLogin && nextProps.isLogin)
        ) {
            this.updateStatus();
        }
    };

    updateStatus = () => {
        let percent = 0;
        API.get("/user/storage")
            .then((response) => {
                if (response.data.used / response.data.total >= 1) {
                    percent = 100;
                    this.props.toggleSnackbar(
                        "top",
                        "right",
                        "您的已用容量已超过容量配额，请尽快删除多余文件或购买容量",
                        "warning"
                    );
                } else {
                    percent = (response.data.used / response.data.total) * 100;
                }
                this.setState({
                    percent: percent,
                    used: sizeToString(response.data.used),
                    total: sizeToString(response.data.total),
                });
            })
            // eslint-disable-next-line @typescript-eslint/no-empty-function
            .catch(() => {});
    };

    render() {
        const { classes } = this.props;
        return (
            <div
                onMouseEnter={() => this.setState({ showExpand: true })}
                onMouseLeave={() => this.setState({ showExpand: false })}
                className={classes.stickFooter}
            >
                <Divider />
                <ButtonBase>
                    <div className={classes.storageContainer}>
                        <StorageIcon className={classes.iconFix} />
                        <div className={classes.detail}>
                            存储空间{"   "}
                            <LinearProgress
                                className={classes.bar}
                                color="secondary"
                                variant="determinate"
                                value={this.state.percent}
                            />
                            <div className={classes.info}>
                                <Tooltip
                                    title={
                                        "已使用 " +
                                        (this.state.used === null
                                            ? " -- "
                                            : this.state.used) +
                                        ", 共 " +
                                        (this.state.total === null
                                            ? " -- "
                                            : this.state.total)
                                    }
                                    placement="top"
                                >
                                    <Typography
                                        variant="caption"
                                        noWrap
                                        color="textSecondary"
                                    >
                                        {this.state.used === null
                                            ? " -- "
                                            : this.state.used}
                                        {" / "}
                                        {this.state.total === null
                                            ? " -- "
                                            : this.state.total}
                                    </Typography>
                                </Tooltip>
                            </div>
                        </div>
                    </div>
                </ButtonBase>
            </div>
        );
    }
}

StorageBarCompoment.propTypes = {
    classes: PropTypes.object.isRequired,
};

const StorageBar = connect(
    mapStateToProps,
    mapDispatchToProps
)(withStyles(styles)(withRouter(StorageBarCompoment)));

export default StorageBar;
