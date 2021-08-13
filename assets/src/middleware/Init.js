import {
    setSiteConfig,
    toggleSnackbar,
    enableLoadUploader,
    changeViewMethod,
} from "../actions/index";
import API from "./Api";
import Auth from "./Auth";
import pathHelper from "../utils/page";

const initUserConfig = (siteConfig) => {
    if (siteConfig.user !== undefined && !siteConfig.user.anonymous) {
        const themes = JSON.parse(siteConfig.themes);
        const user = siteConfig.user;
        delete siteConfig.user;

        //更换用户自定配色
        if (
            user["preferred_theme"] !== "" &&
            themes[user["preferred_theme"]] !== undefined
        ) {
            siteConfig.theme = themes[user["preferred_theme"]];
        }

        // 更新登录态
        Auth.authenticate(user);
    }
    if (siteConfig.user !== undefined && siteConfig.user.anonymous) {
        Auth.SetUser(siteConfig.user);
    }
    return siteConfig;
};

export const InitSiteConfig = (rawStore) => {
    // 从缓存获取默认配置
    const configCache = JSON.parse(localStorage.getItem("siteConfigCache"));
    if (configCache != null) {
        rawStore.siteConfig = configCache;
    }
    // 检查是否有path参数
    const url = new URL(window.location.href);
    const c = url.searchParams.get("path");
    rawStore.navigator.path = c === null ? "/" : c;
    // 初始化用户个性配置
    rawStore.siteConfig = initUserConfig(rawStore.siteConfig);

    // 更改站点标题
    document.title = rawStore.siteConfig.title;
    return rawStore;
};

export function enableUploaderLoad() {
    // 开启上传组件加载
    const user = Auth.GetUser();
    window.policyType = user !== null ? user.policy.saveType : "local";
    window.uploadConfig = user !== null ? user.policy : {};
    window.pathCache = [];
}

export async function UpdateSiteConfig(store) {
    API.get("/site/config")
        .then(function (response) {
            const themes = JSON.parse(response.data.themes);
            response.data.theme = themes[response.data.defaultTheme];
            response.data = initUserConfig(response.data);
            store.dispatch(setSiteConfig(response.data));
            localStorage.setItem(
                "siteConfigCache",
                JSON.stringify(response.data)
            );

            // 偏爱的列表样式
            const preferListMethod = Auth.GetPreference("view_method");
            if (preferListMethod) {
                store.dispatch(changeViewMethod(preferListMethod));
            } else {
                if (pathHelper.isSharePage(window.location.pathname)) {
                    store.dispatch(
                        changeViewMethod(response.data.share_view_method)
                    );
                } else {
                    store.dispatch(
                        changeViewMethod(response.data.home_view_method)
                    );
                }
            }
        })
        .catch(function (error) {
            store.dispatch(
                toggleSnackbar(
                    "top",
                    "right",
                    "无法加载站点配置：" + error.message,
                    "error"
                )
            );
        })
        .then(function () {
            enableUploaderLoad(store);
            store.dispatch(enableLoadUploader());
        });
}
