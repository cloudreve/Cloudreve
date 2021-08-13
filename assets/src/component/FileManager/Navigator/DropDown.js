import React from "react";
import DropDownItem from "./DropDownItem";

export default function DropDown(props) {
    let timer;
    let first = props.folders.length;
    const status = [];
    for (let index = 0; index < props.folders.length; index++) {
        status[index] = false;
    }

    const setActiveStatus = (id, value) => {
        status[id] = value;
        if (value) {
            clearTimeout(timer);
        } else {
            let shouldClose = true;
            status.forEach((element) => {
                if (element) {
                    shouldClose = false;
                }
            });
            if (shouldClose) {
                if (first <= 0) {
                    timer = setTimeout(() => {
                        props.onClose();
                    }, 100);
                } else {
                    first--;
                }
            }
        }
        console.log(status);
    };

    return (
        <>
            {props.folders.map((folder, id) => (
                <DropDownItem
                    key={id}
                    path={"/" + props.folders.slice(0, id).join("/")}
                    navigateTo={props.navigateTo}
                    id={id}
                    setActiveStatus={setActiveStatus}
                    folder={folder}
                />
            ))}
        </>
    );
}
