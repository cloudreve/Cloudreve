import React, { forwardRef, useCallback, useRef } from "react";
import Script from "react-load-script";
import { useSelector } from "react-redux";

const TCaptcha = forwardRef(function TCaptcha(
    { captchaRef, setLoading, isValidateRef, submitRef },
    ref
) {
    const appid = useSelector(
        (state) => state.siteConfig.tcaptcha_captcha_app_id
    );
    const onLoad = () => {
        try {
            ref.current = new window.TencentCaptcha(appid, function (res) {
                if (res.ret === 0) {
                    captchaRef.current.ticket = res.ticket;
                    captchaRef.current.randstr = res.randstr;

                    isValidateRef.current.isValidate = true;
                    submitRef.current.submit();
                    console.log(submitRef);
                } else {
                    submitRef.current.setLoading(false);
                }
            });
        } catch (e) {
            console.error(e);
        }
        setLoading(false);
    };
    return (
        <Script
            url={"https://ssl.captcha.qq.com/TCaptcha.js"}
            onLoad={onLoad}
        />
    );
});

export default function useTCaptcha(setLoading) {
    const isValidate = useRef({
        isValidate: false,
    });
    const captchaParamsRef = useRef({
        ticket: "",
        randstr: "",
    });
    const submitRef = useRef({
        // eslint-disable-next-line @typescript-eslint/no-empty-function
        submit: () => {},
        // eslint-disable-next-line @typescript-eslint/no-empty-function
        setLoading: () => {},
    });

    const captchaRef = useRef();

    const CaptchaRender = useCallback(
        function TCaptchaRender() {
            return (
                <TCaptcha
                    captchaRef={captchaParamsRef}
                    setLoading={setLoading}
                    isValidateRef={isValidate}
                    submitRef={submitRef}
                    ref={captchaRef}
                />
            );
        },
        [captchaParamsRef, setLoading, isValidate, submitRef, captchaRef]
    );

    return {
        isValidate: isValidate,
        validate: (submit, setLoading) => {
            submitRef.current.submit = submit;
            submitRef.current.setLoading = setLoading;
            captchaRef.current.show();
        },
        captchaParamsRef,
        CaptchaRender,
    };
}
