export * from "./explorer";

export const setNavigator = (path, navigatorLoading) => {
    return {
        type: "SET_NAVIGATOR",
        path,
        navigatorLoading,
    };
};

export const navigateTo = (path) => {
    return (dispatch, getState) => {
        const state = getState();
        const navigatorLoading = path !== state.navigator.path;
        dispatch(setNavigator(path, navigatorLoading));
    };
};

export const navigateUp = () => {
    return (dispatch, getState) => {
        const state = getState();
        const pathSplit = state.navigator.path.split("/");
        pathSplit.pop();
        const newPath = pathSplit.length === 1 ? "/" : pathSplit.join("/");
        const navigatorLoading = newPath !== state.navigator.path;
        dispatch(setNavigator(newPath, navigatorLoading));
    };
};

export const drawerToggleAction = (open) => {
    return {
        type: "DRAWER_TOGGLE",
        open: open,
    };
};

export const dragAndDrop = (source, target) => {
    return {
        type: "DRAG_AND_DROP",
        source: source,
        target: target,
    };
};

export const changeViewMethod = (method) => {
    return {
        type: "CHANGE_VIEW_METHOD",
        method: method,
    };
};

export const toggleDaylightMode = () => {
    return {
        type: "TOGGLE_DAYLIGHT_MODE",
    };
};

export const changeContextMenu = (type, open) => {
    return {
        type: "CHANGE_CONTEXT_MENU",
        menuType: type,
        open: open,
    };
};

export const setNavigatorLoadingStatus = (status) => {
    return {
        type: "SET_NAVIGATOR_LOADING_STATUE",
        status: status,
    };
};

export const setNavigatorError = (status, msg) => {
    return {
        type: "SET_NAVIGATOR_ERROR",
        status: status,
        msg: msg,
    };
};

export const openCreateFolderDialog = () => {
    return {
        type: "OPEN_CREATE_FOLDER_DIALOG",
    };
};

export const openCreateFileDialog = () => {
    return {
        type: "OPEN_CREATE_FILE_DIALOG",
    };
};

export const setUserPopover = (anchor) => {
    return {
        type: "SET_USER_POPOVER",
        anchor: anchor,
    };
};

export const setShareUserPopover = (anchor) => {
    return {
        type: "SET_SHARE_USER_POPOVER",
        anchor: anchor,
    };
};

export const openRenameDialog = () => {
    return {
        type: "OPEN_RENAME_DIALOG",
    };
};

export const openResaveDialog = (key) => {
    return {
        type: "OPEN_RESAVE_DIALOG",
        key: key,
    };
};

export const openMoveDialog = () => {
    return {
        type: "OPEN_MOVE_DIALOG",
    };
};

export const openRemoveDialog = () => {
    return {
        type: "OPEN_REMOVE_DIALOG",
    };
};

export const openShareDialog = () => {
    return {
        type: "OPEN_SHARE_DIALOG",
    };
};

export const applyThemes = (theme) => {
    return {
        type: "APPLY_THEME",
        theme: theme,
    };
};

export const setSessionStatus = (status) => {
    return {
        type: "SET_SESSION_STATUS",
        status: status,
    };
};

export const openMusicDialog = () => {
    return {
        type: "OPEN_MUSIC_DIALOG",
    };
};

export const openRemoteDownloadDialog = () => {
    return {
        type: "OPEN_REMOTE_DOWNLOAD_DIALOG",
    };
};

export const openTorrentDownloadDialog = () => {
    return {
        type: "OPEN_TORRENT_DOWNLOAD_DIALOG",
    };
};

export const openDecompressDialog = () => {
    return {
        type: "OPEN_DECOMPRESS_DIALOG",
    };
};

export const openCompressDialog = () => {
    return {
        type: "OPEN_COMPRESS_DIALOG",
    };
};

export const openGetSourceDialog = () => {
    return {
        type: "OPEN_GET_SOURCE_DIALOG",
    };
};

export const openCopyDialog = () => {
    return {
        type: "OPEN_COPY_DIALOG",
    };
};

export const openLoadingDialog = (text) => {
    return {
        type: "OPEN_LOADING_DIALOG",
        text: text,
    };
};

export const closeAllModals = () => {
    return {
        type: "CLOSE_ALL_MODALS",
    };
};

export const toggleSnackbar = (vertical, horizontal, msg, color) => {
    return {
        type: "TOGGLE_SNACKBAR",
        vertical: vertical,
        horizontal: horizontal,
        msg: msg,
        color: color,
    };
};

export const enableLoadUploader = () => {
    return {
        type: "ENABLE_LOAD_UPLOADER",
    };
};

export const setModalsLoading = (status) => {
    return {
        type: "SET_MODALS_LOADING",
        status: status,
    };
};

export const refreshFileList = () => {
    return {
        type: "REFRESH_FILE_LIST",
    };
};

export const searchMyFile = (keywords) => {
    return {
        type: "SEARCH_MY_FILE",
        keywords: keywords,
    };
};

export const showImgPreivew = (first) => {
    return {
        type: "SHOW_IMG_PREIVEW",
        first: first,
    };
};

export const refreshStorage = () => {
    return {
        type: "REFRESH_STORAGE",
    };
};

export const saveFile = () => {
    return {
        type: "SAVE_FILE",
    };
};

export const setSiteConfig = (config) => {
    return {
        type: "SET_SITE_CONFIG",
        config: config,
    };
};
