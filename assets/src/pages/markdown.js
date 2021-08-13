import React from "react";
import ReactDOM from "react-dom";

import { Provider } from "react-redux";
import { createStore } from "redux";
import cloureveApp from "../reducers";
import MarkdownApp from "./markdown.app";
const defaultStatus = {
    navigator: {
        path: window.path,
        refresh: true,
    },
    viewUpdate: {
        open: window.isHomePage,
        explorerViewMethod: "icon",
        sortMethod: "timePos",
        contextType: "none",
        menuOpen: false,
        navigatorLoading: true,
        navigatorError: false,
        navigatorErrorMsg: null,
        modalsLoading: false,
        storageRefresh: false,
        modals: {
            createNewFolder: false,
            rename: false,
            move: false,
            remove: false,
            share: false,
            music: false,
            remoteDownload: false,
            torrentDownload: false,
        },
        snackbar: {
            toggle: false,
            vertical: "top",
            horizontal: "center",
            msg: "",
            color: "",
        },
    },
    explorer: {
        fileList: [],
        fileSave: false,
        dirList: [],
        selected: [{ path: "/", name: window.fileInfo.name, type: "file" }],
        selectProps: {
            isMultiple: false,
            withFolder: false,
            withFile: true,
        },
        imgPreview: {
            first: null,
            other: [],
        },
        keywords: null,
    },
};

const store = createStore(cloureveApp, defaultStatus);
ReactDOM.render(
    <Provider store={store}>
        <MarkdownApp />
    </Provider>,
    document.getElementById("root")
);
