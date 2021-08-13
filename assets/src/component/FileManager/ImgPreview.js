import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";
import { baseURL } from "../../middleware/Api";
import { showImgPreivew } from "../../actions/index";
import { imgPreviewSuffix } from "../../config";
import { withStyles } from "@material-ui/core";
import pathHelper from "../../utils/page";
import { withRouter } from "react-router";
import { PhotoSlider } from "react-photo-view";
import "react-photo-view/dist/index.css";
import * as explorer from "../../redux/explorer/reducer";

const styles = () => ({});

const mapStateToProps = (state) => {
    return {
        first: state.explorer.imgPreview.first,
        other: state.explorer.imgPreview.other,
    };
};

const mapDispatchToProps = (dispatch) => {
    return {
        showImgPreivew: (first) => {
            dispatch(showImgPreivew(first));
        },
    };
};

class ImagPreviewComponent extends Component {
    state = {
        items: [],
        photoIndex: 0,
        isOpen: false,
    };

    UNSAFE_componentWillReceiveProps = (nextProps) => {
        const items = [];
        let firstOne = 0;
        if (nextProps.first.id !== "") {
            if (
                pathHelper.isSharePage(this.props.location.pathname) &&
                !nextProps.first.path
            ) {
                const newImg = {
                    intro: nextProps.first.name,
                    src: baseURL + "/share/preview/" + nextProps.first.key,
                };
                firstOne = 0;
                items.push(newImg);
                this.setState({
                    photoIndex: firstOne,
                    items: items,
                    isOpen: true,
                });
                return;
            }
            // eslint-disable-next-line
            nextProps.other.map((value) => {
                const fileType = value.name.split(".").pop().toLowerCase();
                if (imgPreviewSuffix.indexOf(fileType) !== -1) {
                    let src = "";
                    if (pathHelper.isSharePage(this.props.location.pathname)) {
                        src = baseURL + "/share/preview/" + value.key;
                        src =
                            src +
                            "?path=" +
                            encodeURIComponent(
                                value.path === "/"
                                    ? value.path + value.name
                                    : value.path + "/" + value.name
                            );
                    } else {
                        src = baseURL + "/file/preview/" + value.id;
                    }
                    const newImg = {
                        intro: value.name,
                        src: src,
                    };
                    if (
                        value.path === nextProps.first.path &&
                        value.name === nextProps.first.name
                    ) {
                        firstOne = items.length;
                    }
                    items.push(newImg);
                }
            });
            this.setState({
                photoIndex: firstOne,
                items: items,
                isOpen: true,
            });
        }
    };

    handleClose = () => {
        this.props.showImgPreivew(explorer.initState.imgPreview.first);
        this.setState({
            isOpen: false,
        });
    };

    render() {
        const { photoIndex, isOpen, items } = this.state;

        return (
            <div>
                {isOpen && (
                    <PhotoSlider
                        images={items}
                        visible={isOpen}
                        onClose={() => this.handleClose()}
                        index={photoIndex}
                        onIndexChange={(n) =>
                            this.setState({
                                photoIndex: n,
                            })
                        }
                    />
                )}
            </div>
        );
    }
}

ImagPreviewComponent.propTypes = {
    classes: PropTypes.object.isRequired,
};

const ImgPreivew = connect(
    mapStateToProps,
    mapDispatchToProps
)(withStyles(styles)(withRouter(ImagPreviewComponent)));

export default ImgPreivew;
