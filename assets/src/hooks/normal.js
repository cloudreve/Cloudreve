import React, {
    forwardRef,
    useCallback,
    useEffect,
    useRef,
    useState,
} from "react";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../actions";
import API from "../middleware/Api";
import { FormControl, Input, InputLabel } from "@material-ui/core";
import Placeholder from "../component/Placeholder/Captcha";
import { defaultValidate, useStyle } from "./useCaptcha";

const NormalCaptcha = forwardRef(function NormalCaptcha(
    { captchaRef, setLoading },
    ref
) {
    const classes = useStyle();

    const [captcha, setCaptcha] = useState("");
    const [captchaData, setCaptchaData] = useState(null);

    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const refreshCaptcha = () => {
        API.get("/site/captcha")
            .then((response) => {
                setCaptchaData(response.data);
                setLoading(false);
            })
            .catch((error) => {
                ToggleSnackbar(
                    "top",
                    "right",
                    "无法加载验证码：" + error.message,
                    "error"
                );
            });
    };

    useEffect(() => {
        ref.current = refreshCaptcha;
        refreshCaptcha();
    }, []);

    useEffect(() => {
        captchaRef.current.captchaCode = captcha;
    }, [captcha]);

    return (
        <div className={classes.captchaContainer}>
            <FormControl margin="normal" required fullWidth>
                <InputLabel htmlFor="captcha">验证码</InputLabel>
                <Input
                    name="captcha"
                    onChange={(e) => setCaptcha(e.target.value)}
                    type="text"
                    id="captcha"
                    value={captcha}
                    autoComplete
                />
            </FormControl>{" "}
            <div>
                {captchaData === null && (
                    <div className={classes.captchaPlaceholder}>
                        <Placeholder />
                    </div>
                )}
                {captchaData !== null && (
                    <img
                        src={captchaData}
                        alt="captcha"
                        onClick={refreshCaptcha}
                    />
                )}
            </div>
        </div>
    );
});

export default function useNormalCaptcha(captchaRefreshRef, setLoading) {
    const isValidate = useRef({
        isValidate: true,
    });

    const captchaParamsRef = useRef({
        captchaCode: "",
    });

    const CaptchaRender = useCallback(
        function Normal() {
            return (
                <NormalCaptcha
                    captchaRef={captchaParamsRef}
                    ref={captchaRefreshRef}
                    setLoading={setLoading}
                />
            );
        },
        [captchaParamsRef, captchaRefreshRef, setLoading]
    );

    return {
        isValidate,
        validate: defaultValidate,
        captchaParamsRef,
        CaptchaRender,
    };
}
