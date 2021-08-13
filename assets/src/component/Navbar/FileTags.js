import React, { useCallback, useState, Suspense } from "react";
import {
    Divider,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    makeStyles,
    withStyles,
} from "@material-ui/core";
import { Clear, KeyboardArrowRight } from "@material-ui/icons";
import classNames from "classnames";
import FolderShared from "@material-ui/icons/FolderShared";
import UploadIcon from "@material-ui/icons/CloudUpload";
import VideoIcon from "@material-ui/icons/VideoLibraryOutlined";
import ImageIcon from "@material-ui/icons/CollectionsOutlined";
import MusicIcon from "@material-ui/icons/LibraryMusicOutlined";
import DocIcon from "@material-ui/icons/FileCopyOutlined";
import { useHistory, useLocation } from "react-router";
import pathHelper from "../../utils/page";
import MuiExpansionPanel from "@material-ui/core/ExpansionPanel";
import MuiExpansionPanelSummary from "@material-ui/core/ExpansionPanelSummary";
import MuiExpansionPanelDetails from "@material-ui/core/ExpansionPanelDetails";
import { navigateTo, searchMyFile, toggleSnackbar } from "../../actions";
import { useDispatch } from "react-redux";
import Auth from "../../middleware/Auth";
import {
    Circle,
    CircleOutline,
    Heart,
    HeartOutline,
    Hexagon,
    HexagonOutline,
    Hexagram,
    HexagramOutline,
    Rhombus,
    RhombusOutline,
    Square,
    SquareOutline,
    Triangle,
    TriangleOutline,
    FolderHeartOutline,
    TagPlus,
} from "mdi-material-ui";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import IconButton from "@material-ui/core/IconButton";
import API from "../../middleware/Api";

const ExpansionPanel = withStyles({
    root: {
        maxWidth: "100%",
        boxShadow: "none",
        "&:not(:last-child)": {
            borderBottom: 0,
        },
        "&:before": {
            display: "none",
        },
        "&$expanded": { margin: 0 },
    },
    expanded: {},
})(MuiExpansionPanel);

const ExpansionPanelSummary = withStyles({
    root: {
        minHeight: 0,
        padding: 0,

        "&$expanded": {
            minHeight: 0,
        },
    },
    content: {
        maxWidth: "100%",
        margin: 0,
        display: "block",
        "&$expanded": {
            margin: "0",
        },
    },
    expanded: {},
})(MuiExpansionPanelSummary);

const ExpansionPanelDetails = withStyles((theme) => ({
    root: {
        display: "block",
        padding: theme.spacing(0),
    },
}))(MuiExpansionPanelDetails);

const useStyles = makeStyles((theme) => ({
    expand: {
        display: "none",
        transition: ".15s all ease-in-out",
    },
    expanded: {
        display: "block",
        transform: "rotate(90deg)",
    },
    iconFix: {
        marginLeft: "16px",
    },
    hiddenButton: {
        display: "none",
    },
    subMenu: {
        marginLeft: theme.spacing(2),
    },
    overFlow: {
        whiteSpace: "nowrap",
        overflow: "hidden",
        textOverflow: "ellipsis",
    },
}));

const icons = {
    Circle: Circle,
    CircleOutline: CircleOutline,
    Heart: Heart,
    HeartOutline: HeartOutline,
    Hexagon: Hexagon,
    HexagonOutline: HexagonOutline,
    Hexagram: Hexagram,
    HexagramOutline: HexagramOutline,
    Rhombus: Rhombus,
    RhombusOutline: RhombusOutline,
    Square: Square,
    SquareOutline: SquareOutline,
    Triangle: Triangle,
    TriangleOutline: TriangleOutline,
    FolderHeartOutline: FolderHeartOutline,
};

const AddTag = React.lazy(() => import("../Modals/AddTag"));

export default function FileTag() {
    const classes = useStyles();

    const location = useLocation();
    const history = useHistory();

    const isHomePage = pathHelper.isHomePage(location.pathname);

    const [tagOpen, setTagOpen] = useState(true);
    const [addTagModal, setAddTagModal] = useState(false);
    const [tagHover, setTagHover] = useState(null);
    const [tags, setTags] = useState(
        Auth.GetUser().tags ? Auth.GetUser().tags : []
    );

    const dispatch = useDispatch();
    const SearchMyFile = useCallback((k) => dispatch(searchMyFile(k)), [
        dispatch,
    ]);
    const NavigateTo = useCallback((k) => dispatch(navigateTo(k)), [dispatch]);

    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const getIcon = (icon, color) => {
        if (icons[icon]) {
            const IconComponent = icons[icon];
            return (
                <IconComponent
                    className={[classes.iconFix]}
                    style={
                        color
                            ? {
                                  color: color,
                              }
                            : {}
                    }
                />
            );
        }
        return <Circle className={[classes.iconFix]} />;
    };

    const submitSuccess = (tag) => {
        const newTags = [...tags, tag];
        setTags(newTags);
        const user = Auth.GetUser();
        user.tags = newTags;
        Auth.SetUser(user);
    };

    const submitDelete = (id) => {
        API.delete("/tag/" + id)
            .then(() => {
                const newTags = tags.filter((v) => {
                    return v.id !== id;
                });
                setTags(newTags);
                const user = Auth.GetUser();
                user.tags = newTags;
                Auth.SetUser(user);
            })
            .catch((error) => {
                ToggleSnackbar("top", "right", error.message, "error");
            });
    };

    return (
        <>
            <Suspense fallback={""}>
                <AddTag
                    onSuccess={submitSuccess}
                    open={addTagModal}
                    onClose={() => setAddTagModal(false)}
                />
            </Suspense>
            <ExpansionPanel
                square
                expanded={tagOpen && isHomePage}
                onChange={() => isHomePage && setTagOpen(!tagOpen)}
            >
                <ExpansionPanelSummary
                    aria-controls="panel1d-content"
                    id="panel1d-header"
                >
                    <ListItem
                        button
                        key="我的文件"
                        onClick={() =>
                            !isHomePage && history.push("/home?path=%2F")
                        }
                    >
                        <ListItemIcon>
                            <KeyboardArrowRight
                                className={classNames(
                                    {
                                        [classes.expanded]:
                                            tagOpen && isHomePage,
                                        [classes.iconFix]: true,
                                    },
                                    classes.expand
                                )}
                            />
                            {!(tagOpen && isHomePage) && (
                                <FolderShared className={classes.iconFix} />
                            )}
                        </ListItemIcon>
                        <ListItemText primary="我的文件" />
                    </ListItem>
                    <Divider />
                </ExpansionPanelSummary>

                <ExpansionPanelDetails>
                    <List onMouseLeave={() => setTagHover(null)}>
                        <ListItem
                            button
                            id="pickfiles"
                            className={classes.hiddenButton}
                        >
                            <ListItemIcon>
                                <UploadIcon />
                            </ListItemIcon>
                            <ListItemText />
                        </ListItem>
                        <ListItem
                            button
                            id="pickfolder"
                            className={classes.hiddenButton}
                        >
                            <ListItemIcon>
                                <UploadIcon />
                            </ListItemIcon>
                            <ListItemText />
                        </ListItem>
                        {[
                            {
                                key: "视频",
                                id: "video",
                                icon: (
                                    <VideoIcon
                                        className={[
                                            classes.iconFix,
                                            classes.iconVideo,
                                        ]}
                                    />
                                ),
                            },
                            {
                                key: "图片",
                                id: "image",
                                icon: (
                                    <ImageIcon
                                        className={[
                                            classes.iconFix,
                                            classes.iconImg,
                                        ]}
                                    />
                                ),
                            },
                            {
                                key: "音频",
                                id: "audio",
                                icon: (
                                    <MusicIcon
                                        className={[
                                            classes.iconFix,
                                            classes.iconAudio,
                                        ]}
                                    />
                                ),
                            },
                            {
                                key: "文档",
                                id: "doc",
                                icon: (
                                    <DocIcon
                                        className={[
                                            classes.iconFix,
                                            classes.iconDoc,
                                        ]}
                                    />
                                ),
                            },
                        ].map((v) => (
                            <ListItem
                                button
                                key={v.key}
                                onClick={() => SearchMyFile(v.id + "/internal")}
                            >
                                <ListItemIcon className={classes.subMenu}>
                                    {v.icon}
                                </ListItemIcon>
                                <ListItemText primary={v.key} />
                            </ListItem>
                        ))}
                        {tags.map((v) => (
                            <ListItem
                                button
                                key={v.id}
                                onMouseEnter={() => setTagHover(v.id)}
                                onClick={() => {
                                    if (v.type === 0) {
                                        SearchMyFile("tag/" + v.id);
                                    } else {
                                        NavigateTo(v.expression);
                                    }
                                }}
                            >
                                <ListItemIcon className={classes.subMenu}>
                                    {getIcon(
                                        v.type === 0
                                            ? v.icon
                                            : "FolderHeartOutline",
                                        v.type === 0 ? v.color : null
                                    )}
                                </ListItemIcon>
                                <ListItemText
                                    className={classes.overFlow}
                                    primary={v.name}
                                />

                                {tagHover === v.id && (
                                    <ListItemSecondaryAction
                                        onClick={() => submitDelete(v.id)}
                                    >
                                        <IconButton
                                            size={"small"}
                                            edge="end"
                                            aria-label="delete"
                                        >
                                            <Clear />
                                        </IconButton>
                                    </ListItemSecondaryAction>
                                )}
                            </ListItem>
                        ))}

                        <ListItem button onClick={() => setAddTagModal(true)}>
                            <ListItemIcon className={classes.subMenu}>
                                <TagPlus className={classes.iconFix} />
                            </ListItemIcon>
                            <ListItemText primary={"添加标签..."} />
                        </ListItem>
                    </List>{" "}
                    <Divider />
                </ExpansionPanelDetails>
            </ExpansionPanel>
        </>
    );
}
