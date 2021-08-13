/* eslint-disable no-case-declarations */
import { InitSiteConfig } from "../middleware/Init";
import { combineReducers } from "../redux/combineReducers";
import viewUpdate from "../redux/viewUpdate/reducer";
import explorer from "../redux/explorer/reducer";
import { connectRouter } from "connected-react-router";

const doNavigate = (path, state) => {
    window.currntPath = path;
    return Object.assign({}, state, {
        navigator: Object.assign({}, state.navigator, {
            path: path,
        }),
    });
};

export const initState = {
    siteConfig: {
        title: window.subTitle,
        siteICPId: "",
        loginCaptcha: false,
        regCaptcha: false,
        forgetCaptcha: false,
        emailActive: false,
        QQLogin: false,
        themes: null,
        authn: false,
        theme: {
            palette: {
                common: { black: "#000", white: "#fff" },
                background: { paper: "#fff", default: "#fafafa" },
                primary: {
                    light: "#7986cb",
                    main: "#3f51b5",
                    dark: "#303f9f",
                    contrastText: "#fff",
                },
                secondary: {
                    light: "#ff4081",
                    main: "#f50057",
                    dark: "#c51162",
                    contrastText: "#fff",
                },
                error: {
                    light: "#e57373",
                    main: "#f44336",
                    dark: "#d32f2f",
                    contrastText: "#fff",
                },
                text: {
                    primary: "rgba(0, 0, 0, 0.87)",
                    secondary: "rgba(0, 0, 0, 0.54)",
                    disabled: "rgba(0, 0, 0, 0.38)",
                    hint: "rgba(0, 0, 0, 0.38)",
                },
                explorer: {
                    filename: "#474849",
                    icon: "#8f8f8f",
                    bgSelected: "#D5DAF0",
                    emptyIcon: "#e8e8e8",
                },
            },
        },
        captcha_ReCaptchaKey: "defaultKey",
        captcha_type: "normal",
        tcaptcha_captcha_app_id: "",
    },
    navigator: {
        path: "/",
        refresh: true,
    },
};

const defaultStatus = InitSiteConfig(initState);

// TODO: 将cloureveApp切分成小的reducer
const cloudreveApp = (state = defaultStatus, action) => {
    switch (action.type) {
        case "SET_NAVIGATOR":
            return doNavigate(action.path, state);
        case "TOGGLE_DAYLIGHT_MODE": {
            const copy = Object.assign({}, state);
            if (
                copy.siteConfig.theme.palette.type === undefined ||
                copy.siteConfig.theme.palette.type === "light"
            ) {
                return {
                    ...state,
                    siteConfig: {
                        ...state.siteConfig,
                        theme: {
                            ...state.siteConfig.theme,
                            palette: {
                                ...state.siteConfig.theme.palette,
                                type: "dark",
                            },
                        },
                    },
                };
            }
            return {
                ...state,
                siteConfig: {
                    ...state.siteConfig,
                    theme: {
                        ...state.siteConfig.theme,
                        palette: {
                            ...state.siteConfig.theme.palette,
                            type: "light",
                        },
                    },
                },
            };
        }
        case "APPLY_THEME":
            if (state.siteConfig.themes !== null) {
                const themes = JSON.parse(state.siteConfig.themes);
                if (themes[action.theme] === undefined) {
                    return state;
                }
                return Object.assign({}, state, {
                    siteConfig: Object.assign({}, state.siteConfig, {
                        theme: themes[action.theme],
                    }),
                });
            }
            break;
        case "NAVIGATOR_UP":
            return doNavigate(action.path, state);
        case "SET_SITE_CONFIG":
            return Object.assign({}, state, {
                siteConfig: action.config,
            });
        case "REFRESH_FILE_LIST":
            return Object.assign({}, state, {
                navigator: Object.assign({}, state.navigator, {
                    refresh: !state.navigator.refresh,
                }),
            });
        case "SEARCH_MY_FILE":
            return Object.assign({}, state, {
                navigator: Object.assign({}, state.navigator, {
                    path: "/搜索结果",
                    refresh:
                        state.explorer.keywords === ""
                            ? state.navigator.refresh
                            : !state.navigator.refresh,
                }),
            });
        default:
            return state;
    }
};

export default (history) => (state, action) => {
    const { viewUpdate: viewUpdateState, explorer: explorerState } =
        state || {};
    const appState = cloudreveApp(state, action);
    const combinedState = combineReducers({
        viewUpdate,
        explorer,
        router: connectRouter(history),
    })({ viewUpdate: viewUpdateState, explorer: explorerState }, action);
    return {
        ...appState,
        ...combinedState,
    };
};
