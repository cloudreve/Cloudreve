import React, { useCallback, useEffect, useRef, useState } from "react";
import { useSelector } from "react-redux";
import { FormControl } from "@material-ui/core";
import ReCaptcha from "../component/Login/ReCaptcha";
import { defaultValidate, useStyle } from "./useCaptcha";

const Recaptcha = ({ captchaRef, setLoading }) => {
    const classes = useStyle();

    const [captcha, setCaptcha] = useState("");

    const reCaptchaKey = useSelector(
        (state) => state.siteConfig.captcha_ReCaptchaKey
    );

    useEffect(() => {
        captchaRef.current.captchaCode = captcha;
    }, [captcha]);

    useEffect(() => setLoading(false), []);

    return (
        <div className={classes.captchaContainer}>
            <FormControl margin="normal" required fullWidth>
                <div>
                    <ReCaptcha
                        style={{
                            display: "inline-block",
                        }}
                        sitekey={reCaptchaKey}
                        onChange={(value) => setCaptcha(value)}
                        id="captcha"
                        name="captcha"
                    />
                </div>
            </FormControl>{" "}
        </div>
    );
};

export default function useRecaptcha(setLoading) {
    const isValidate = useRef({
        isValidate: true,
    });

    const captchaParamsRef = useRef({
        captchaCode: "",
    });

    const CaptchaRender = useCallback(
        function RecaptchaRender() {
            return (
                <Recaptcha
                    captchaRef={captchaParamsRef}
                    setLoading={setLoading}
                />
            );
        },
        [captchaParamsRef, setLoading]
    );

    return {
        isValidate,
        validate: defaultValidate,
        captchaParamsRef,
        CaptchaRender,
    };
}
