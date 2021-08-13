import { AnyAction } from "redux";
import { ThunkAction } from "redux-thunk";
import { CloudreveFile, SortMethod } from "./../../types/index";
import { closeContextMenu } from "../viewUpdate/action";

export interface ActionSetFileList extends AnyAction {
    type: "SET_FILE_LIST";
    list: CloudreveFile[];
}
export const setFileList = (list: CloudreveFile[]): ActionSetFileList => {
    return {
        type: "SET_FILE_LIST",
        list,
    };
};

export interface ActionSetDirList extends AnyAction {
    type: "SET_DIR_LIST";
    list: CloudreveFile[];
}
export const setDirList = (list: CloudreveFile[]): ActionSetDirList => {
    return {
        type: "SET_DIR_LIST",
        list,
    };
};

export interface ActionSetSortMethod extends AnyAction {
    type: "SET_SORT_METHOD";
    method: SortMethod;
}
export const setSortMethod = (method: SortMethod): ActionSetSortMethod => {
    return {
        type: "SET_SORT_METHOD",
        method,
    };
};

export const setSideBar = (open: boolean) => {
    return {
        type: "SET_SIDE_BAR",
        open,
    };
};

type SortFunc = (a: CloudreveFile, b: CloudreveFile) => number;
const sortMethodFuncs: Record<SortMethod, SortFunc> = {
    sizePos: (a: CloudreveFile, b: CloudreveFile) => {
        return a.size - b.size;
    },
    sizeRes: (a: CloudreveFile, b: CloudreveFile) => {
        return b.size - a.size;
    },
    namePos: (a: CloudreveFile, b: CloudreveFile) => {
        return a.name.localeCompare(
            b.name,
            navigator.languages[0] || navigator.language,
            { numeric: true, ignorePunctuation: true }
        );
    },
    nameRev: (a: CloudreveFile, b: CloudreveFile) => {
        return b.name.localeCompare(
            a.name,
            navigator.languages[0] || navigator.language,
            { numeric: true, ignorePunctuation: true }
        );
    },
    timePos: (a: CloudreveFile, b: CloudreveFile) => {
        return Date.parse(a.date) - Date.parse(b.date);
    },
    timeRev: (a: CloudreveFile, b: CloudreveFile) => {
        return Date.parse(b.date) - Date.parse(a.date);
    },
};

export const updateFileList = (
    list: CloudreveFile[]
): ThunkAction<any, any, any, any> => {
    return (dispatch, getState): void => {
        const state = getState();
        // TODO: define state type
        const { sortMethod } = state.viewUpdate;
        const dirList = list.filter((x) => {
            return x.type === "dir";
        });
        const fileList = list.filter((x) => {
            return x.type === "file";
        });
        const sortFunc = sortMethodFuncs[sortMethod as SortMethod];
        dispatch(setDirList(dirList.sort(sortFunc)));
        dispatch(setFileList(fileList.sort(sortFunc)));
    };
};

export const changeSortMethod = (
    method: SortMethod
): ThunkAction<any, any, any, any> => {
    return (dispatch, getState): void => {
        const state = getState();
        const { fileList, dirList } = state.explorer;
        const sortFunc = sortMethodFuncs[method];
        dispatch(setSortMethod(method));
        dispatch(setDirList(dirList.sort(sortFunc)));
        dispatch(setFileList(fileList.sort(sortFunc)));
    };
};

export const toggleObjectInfoSidebar = (
    open: boolean
): ThunkAction<any, any, any, any> => {
    return (dispatch, getState): void => {
        const state = getState();
        if (open) {
            dispatch(closeContextMenu());
        }
        dispatch(setSideBar(true));
    };
};
