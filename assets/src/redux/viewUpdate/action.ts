import { ThunkAction } from "redux-thunk";
import { AnyAction } from "redux";
export interface ActionSetSubtitle extends AnyAction {
    type: "SET_SUBTITLE";
    title: string;
}

export const setSubtitle = (title: string): ActionSetSubtitle => {
    return {
        type: "SET_SUBTITLE",
        title,
    };
};

export const closeContextMenu = () => {
    return {
        type: "CLOSE_CONTEXT_MENU",
    };
};

export const changeSubTitle = (
    title: string
): ThunkAction<any, any, any, any> => {
    return (dispatch, getState) => {
        const state = getState();
        document.title =
            title === null || title === undefined
                ? state.siteConfig.title
                : title + " - " + state.siteConfig.title;
        dispatch(setSubtitle(title));
    };
};
