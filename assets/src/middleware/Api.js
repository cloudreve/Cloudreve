import axios from "axios";
import Auth from "./Auth";

export const baseURL = "/api/v3";

export const getBaseURL = () => {
    return baseURL;
};

const instance = axios.create({
    baseURL: getBaseURL(),
    withCredentials: true,
    crossDomain: true,
});

function AppError(message, code, error) {
    this.code = code;
    this.message = message || "未知错误";
    this.message += error ? " " + error : "";
    this.stack = new Error().stack;
}
AppError.prototype = Object.create(Error.prototype);
AppError.prototype.constructor = AppError;

instance.interceptors.response.use(
    function (response) {
        response.rawData = response.data;
        response.data = response.data.data;
        if (
            response.rawData.code !== undefined &&
            response.rawData.code !== 0 &&
            response.rawData.code !== 203
        ) {
            // 登录过期
            if (response.rawData.code === 401) {
                Auth.signout();
                window.location.href = "/login";
            }

            // 非管理员
            if (response.rawData.code === 40008) {
                window.location.href = "/home";
            }
            throw new AppError(
                response.rawData.msg,
                response.rawData.code,
                response.rawData.error
            );
        }
        return response;
    },
    function (error) {
        return Promise.reject(error);
    }
);

export default instance;
