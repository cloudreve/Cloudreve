import ReCAPTCHA from "./ReCaptchaWrapper";
import makeAsyncScriptLoader from "react-async-script";

const callbackName = "onloadcallback";
const globalName = "grecaptcha";

function getURL() {
    const hostname = "recaptcha.net";
    return `https://${hostname}/recaptcha/api.js?onload=${callbackName}&render=explicit`;
}

export default makeAsyncScriptLoader(getURL, {
    callbackName,
    globalName,
})(ReCAPTCHA);
