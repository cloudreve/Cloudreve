import React from "react";
import SmallIcon from "../SmallIcon";
import FileIcon from "../FileIcon";
import { useSelector } from "react-redux";
import { makeStyles } from "@material-ui/core";
import Folder from "../Folder";

const useStyles = makeStyles(() => ({
    dragging: {
        width: "200px",
    },
    cardDragged: {
        position: "absolute",
        "transform-origin": "bottom left",
    },
}));

const diliverIcon = (object, viewMethod, classes) => {
    if (object.type === "dir") {
        return (
            <div className={classes.dragging}>
                <SmallIcon file={object} isFolder />
            </div>
        );
    }
    if (object.type === "file" && viewMethod === "icon") {
        return (
            <div className={classes.dragging}>
                <FileIcon file={object} />
            </div>
        );
    }
    if (
        (object.type === "file" && viewMethod === "smallIcon") ||
        viewMethod === "list"
    ) {
        return (
            <div className={classes.dragging}>
                <SmallIcon file={object} />
            </div>
        );
    }
};

const Preview = (props) => {
    const selected = useSelector((state) => state.explorer.selected);
    const viewMethod = useSelector(
        (state) => state.viewUpdate.explorerViewMethod
    );
    const classes = useStyles();
    return (
        <>
            {selected.length === 0 &&
                diliverIcon(props.object, viewMethod, classes)}
            {selected.length > 0 && (
                <>
                    {selected.slice(0, 3).map((card, i) => (
                        <div
                            key={card.id}
                            className={classes.cardDragged}
                            style={{
                                zIndex: selected.length - i,
                                transform: `rotateZ(${-i * 2.5}deg)`,
                            }}
                        >
                            {diliverIcon(card, viewMethod, classes)}
                        </div>
                    ))}
                </>
            )}
        </>
    );
};
export default Preview;
