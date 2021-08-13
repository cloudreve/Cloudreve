import React from "react";
import { useDragLayer } from "react-dnd";
import Preview from "./Preview";
import { useSelector } from "react-redux";
const layerStyles = {
    position: "fixed",
    pointerEvents: "none",
    zIndex: 100,
    left: 0,
    top: 0,
    width: "100%",
    height: "100%",
};

function getItemStyles(
    initialOffset,
    currentOffset,
    pointerOffset,
    viewMethod
) {
    if (!initialOffset || !currentOffset) {
        return {
            display: "none",
        };
    }
    let { x, y } = currentOffset;
    if (viewMethod === "list") {
        x += pointerOffset.x - initialOffset.x;
        y += pointerOffset.y - initialOffset.y;
    }

    const transform = `translate(${x}px, ${y}px)`;
    return {
        opacity: 0.5,
        transform,
        WebkitTransform: transform,
    };
}
const CustomDragLayer = (props) => {
    const {
        itemType,
        isDragging,
        item,
        initialOffset,
        currentOffset,
        pointerOffset,
    } = useDragLayer((monitor) => ({
        item: monitor.getItem(),
        itemType: monitor.getItemType(),
        initialOffset: monitor.getInitialSourceClientOffset(),
        currentOffset: monitor.getSourceClientOffset(),
        pointerOffset: monitor.getInitialClientOffset(),
        isDragging: monitor.isDragging(),
    }));
    const viewMethod = useSelector(
        (state) => state.viewUpdate.explorerViewMethod
    );
    function renderItem() {
        switch (itemType) {
            case "object":
                return <Preview object={item.object} />;
            default:
                return null;
        }
    }
    if (!isDragging) {
        return null;
    }
    return (
        <div style={layerStyles}>
            <div
                style={getItemStyles(
                    initialOffset,
                    currentOffset,
                    pointerOffset,
                    viewMethod
                )}
            >
                {renderItem()}
            </div>
        </div>
    );
};
export default CustomDragLayer;
