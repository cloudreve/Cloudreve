import React from "react";
import { useDrop } from "react-dnd";
import Folder from "../Folder";
import classNames from "classnames";
import TableItem from "../TableRow";
export default function FolderDropWarpper({
    isListView,
    folder,
    onIconClick,
    contextMenu,
    handleClick,
    handleDoubleClick,
    className,
    pref,
}) {
    const [{ canDrop, isOver }, drop] = useDrop({
        accept: "object",
        drop: () => ({ folder }),
        collect: (monitor) => ({
            isOver: monitor.isOver(),
            canDrop: monitor.canDrop(),
        }),
    });
    const isActive = canDrop && isOver;
    if (!isListView) {
        return (
            <div ref={drop}>
                <Folder
                    folder={folder}
                    onIconClick={onIconClick}
                    isActive={isActive}
                />
            </div>
        );
    }
    return (
        <TableItem
            pref={pref}
            dref={drop}
            className={className}
            onIconClick={onIconClick}
            contextMenu={contextMenu}
            handleClick={handleClick}
            handleDoubleClick={handleDoubleClick}
            file={folder}
            isActive={isActive}
        />
    );
}
