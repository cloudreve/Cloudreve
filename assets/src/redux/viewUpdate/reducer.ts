import { AnyAction } from "redux";
import Auth from "../../middleware/Auth";

export interface ViewUpdateState {
    isLogin: boolean;
    loadUploader: boolean;
    open: boolean;
    explorerViewMethod: string;
    sortMethod:
        | "sizePos"
        | "sizeRes"
        | "namePos"
        | "nameRev"
        | "timePos"
        | "timeRev";
    subTitle: string | null;
    contextType: string;
    contextOpen: boolean;
    menuOpen: boolean;
    navigatorLoading: boolean;
    navigatorError: boolean;
    navigatorErrorMsg: string | null;
    modalsLoading: boolean;
    storageRefresh: boolean;
    userPopoverAnchorEl: any;
    shareUserPopoverAnchorEl: any;
    modals: {
        createNewFolder: boolean;
        createNewFile: boolean;
        rename: boolean;
        move: boolean;
        remove: boolean;
        share: boolean;
        music: boolean;
        remoteDownload: boolean;
        torrentDownload: boolean;
        getSource: boolean;
        copy: boolean;
        resave: boolean;
        compress: boolean;
        decompress: boolean;
        loading: boolean;
        loadingText: string;
    };
    snackbar: {
        toggle: boolean;
        vertical: string;
        horizontal: string;
        msg: string;
        color: string;
    };
}
export const initState: ViewUpdateState = {
    // 是否登录
    isLogin: Auth.Check(),
    loadUploader: false,
    open: false,
    explorerViewMethod: "icon",
    sortMethod: "timePos",
    subTitle: null,
    contextType: "none",
    contextOpen: false,
    menuOpen: false,
    navigatorLoading: true,
    navigatorError: false,
    navigatorErrorMsg: null,
    modalsLoading: false,
    storageRefresh: false,
    userPopoverAnchorEl: null,
    shareUserPopoverAnchorEl: null,
    modals: {
        createNewFolder: false,
        createNewFile: false,
        rename: false,
        move: false,
        remove: false,
        share: false,
        music: false,
        remoteDownload: false,
        torrentDownload: false,
        getSource: false,
        copy: false,
        resave: false,
        compress: false,
        decompress: false,
        loading: false,
        loadingText: "",
    },
    snackbar: {
        toggle: false,
        vertical: "top",
        horizontal: "center",
        msg: "",
        color: "",
    },
};
const viewUpdate = (state: ViewUpdateState = initState, action: AnyAction) => {
    switch (action.type) {
        case "DRAWER_TOGGLE":
            return Object.assign({}, state, {
                open: action.open,
            });
        case "CHANGE_VIEW_METHOD":
            return Object.assign({}, state, {
                explorerViewMethod: action.method,
            });
        case "SET_NAVIGATOR_LOADING_STATUE":
            return Object.assign({}, state, {
                navigatorLoading: action.status,
            });
        case "SET_NAVIGATOR_ERROR":
            return Object.assign({}, state, {
                navigatorError: action.status,
                navigatorErrorMsg: action.msg,
            });
        case "OPEN_CREATE_FOLDER_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    createNewFolder: true,
                }),
                contextOpen: false,
            });
        case "OPEN_CREATE_FILE_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    createNewFile: true,
                }),
                contextOpen: false,
            });
        case "OPEN_RENAME_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    rename: true,
                }),
                contextOpen: false,
            });
        case "OPEN_REMOVE_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    remove: true,
                }),
                contextOpen: false,
            });
        case "OPEN_MOVE_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    move: true,
                }),
                contextOpen: false,
            });
        case "OPEN_RESAVE_DIALOG":
            // window.shareKey = action.key;
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    resave: true,
                }),
                contextOpen: false,
            });
        case "SET_USER_POPOVER":
            return Object.assign({}, state, {
                userPopoverAnchorEl: action.anchor,
            });
        case "SET_SHARE_USER_POPOVER":
            return Object.assign({}, state, {
                shareUserPopoverAnchorEl: action.anchor,
            });
        case "OPEN_SHARE_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    share: true,
                }),
                contextOpen: false,
            });
        case "OPEN_MUSIC_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    music: true,
                }),
                contextOpen: false,
            });
        case "OPEN_REMOTE_DOWNLOAD_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    remoteDownload: true,
                }),
                contextOpen: false,
            });
        case "OPEN_TORRENT_DOWNLOAD_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    torrentDownload: true,
                }),
                contextOpen: false,
            });
        case "OPEN_DECOMPRESS_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    decompress: true,
                }),
                contextOpen: false,
            });
        case "OPEN_COMPRESS_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    compress: true,
                }),
                contextOpen: false,
            });
        case "OPEN_GET_SOURCE_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    getSource: true,
                }),
                contextOpen: false,
            });
        case "OPEN_COPY_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    copy: true,
                }),
                contextOpen: false,
            });
        case "OPEN_LOADING_DIALOG":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    loading: true,
                    loadingText: action.text,
                }),
                contextOpen: false,
            });
        case "CLOSE_CONTEXT_MENU":
            return Object.assign({}, state, {
                contextOpen: false,
            });
        case "CLOSE_ALL_MODALS":
            return Object.assign({}, state, {
                modals: Object.assign({}, state.modals, {
                    createNewFolder: false,
                    createNewFile: false,
                    rename: false,
                    move: false,
                    remove: false,
                    share: false,
                    music: false,
                    remoteDownload: false,
                    torrentDownload: false,
                    getSource: false,
                    resave: false,
                    copy: false,
                    loading: false,
                    compress: false,
                    decompress: false,
                }),
            });
        case "TOGGLE_SNACKBAR":
            return Object.assign({}, state, {
                snackbar: {
                    toggle: !state.snackbar.toggle,
                    vertical: action.vertical,
                    horizontal: action.horizontal,
                    msg: action.msg,
                    color: action.color,
                },
            });
        case "SET_MODALS_LOADING":
            return Object.assign({}, state, {
                modalsLoading: action.status,
            });
        case "SET_SESSION_STATUS":
            return {
                ...state,
                isLogin: action.status,
            };
        case "ENABLE_LOAD_UPLOADER":
            return Object.assign({}, state, {
                loadUploader: true,
            });
        case "REFRESH_STORAGE":
            return Object.assign({}, state, {
                storageRefresh: !state.storageRefresh,
            });
        case "SEARCH_MY_FILE":
            return Object.assign({}, state, {
                contextOpen: false,
                navigatorError: false,
                navigatorLoading: true,
            });
        case "CHANGE_CONTEXT_MENU":
            if (state.contextOpen && action.open) {
                return Object.assign({}, state);
            }
            return Object.assign({}, state, {
                contextOpen: action.open,
                contextType: action.menuType,
            });
        case "SET_SUBTITLE":
            return Object.assign({}, state, {
                subTitle: action.title,
            });
        case "SET_SORT_METHOD":
            return {
                ...state,
                sortMethod: action.method,
            };
        case "SET_NAVIGATOR":
            return {
                ...state,
                contextOpen: false,
                navigatorError: false,
                navigatorLoading: action.navigatorLoading,
            };
        default:
            return state;
    }
};

export default viewUpdate;
