import React, { useEffect } from "react";
import ExpandMore from "@material-ui/icons/ExpandMore";
import { Button } from "@material-ui/core";
import { makeStyles } from "@material-ui/core";
import { useDrop } from "react-dnd";
import classNames from "classnames";
import MoreIcon from "@material-ui/icons/MoreHoriz";

const useStyles = makeStyles((theme) => ({
    expandMore: {
        color: "#8d8d8d",
    },
    active: {
        boxShadow: "0 0 0 2px " + theme.palette.primary.light,
    },
    button: {
        textTransform: "none",
    },
}));

export default function PathButton(props) {
    const inputRef = React.useRef(null);

    const [{ canDrop, isOver }, drop] = useDrop({
        accept: "object",
        drop: () => {
            if (props.more) {
                inputRef.current.click();
            } else {
                return {
                    folder: {
                        id: -1,
                        path: props.path,
                        name: props.folder === "/" ? "" : props.folder,
                    },
                };
            }
        },
        collect: (monitor) => ({
            isOver: monitor.isOver(),
            canDrop: monitor.canDrop(),
        }),
    });

    const isActive = canDrop && isOver;

    useEffect(() => {
        if (props.more && isActive) {
            inputRef.current.click();
        }
        // eslint-disable-next-line
    }, [isActive]);

    const classes = useStyles();
    return (
        <span onClick={props.onClick} ref={inputRef}>
            <Button
                ref={drop}
                className={classNames(
                    {
                        [classes.active]: isActive,
                    },
                    classes.button
                )}
                component="span"
                title={props.title}
            >
                {props.more && <MoreIcon />}
                {!props.more && (
                    <>
                        {props.folder}
                        {props.last && (
                            <ExpandMore className={classes.expandMore} />
                        )}
                    </>
                )}
            </Button>
        </span>
    );
}
