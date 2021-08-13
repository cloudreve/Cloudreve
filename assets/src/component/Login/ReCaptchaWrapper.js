import React from "react";
import PropTypes from "prop-types";

export default class ReCAPTCHA extends React.Component {
    constructor() {
        super();
        this.handleExpired = this.handleExpired.bind(this);
        this.handleErrored = this.handleErrored.bind(this);
        this.handleChange = this.handleChange.bind(this);
        this.handleRecaptchaRef = this.handleRecaptchaRef.bind(this);
    }

    getValue() {
        if (this.props.grecaptcha && this._widgetId !== undefined) {
            return this.props.grecaptcha.getResponse(this._widgetId);
        }
        return null;
    }

    getWidgetId() {
        if (this.props.grecaptcha && this._widgetId !== undefined) {
            return this._widgetId;
        }
        return null;
    }

    execute() {
        const { grecaptcha } = this.props;

        if (grecaptcha && this._widgetId !== undefined) {
            return grecaptcha.execute(this._widgetId);
        } else {
            this._executeRequested = true;
        }
    }

    reset() {
        if (this.props.grecaptcha && this._widgetId !== undefined) {
            this.props.grecaptcha.reset(this._widgetId);
        }
    }

    handleExpired() {
        if (this.props.onExpired) {
            this.props.onExpired();
        } else {
            this.handleChange(null);
        }
    }

    handleErrored() {
        if (this.props.onErrored) this.props.onErrored();
    }

    handleChange(token) {
        if (this.props.onChange) this.props.onChange(token);
    }

    explicitRender() {
        if (
            this.props.grecaptcha &&
            this.props.grecaptcha.render &&
            this._widgetId === undefined
        ) {
            const wrapper = document.createElement("div");
            this._widgetId = this.props.grecaptcha.render(wrapper, {
                sitekey: this.props.sitekey,
                callback: this.handleChange,
                theme: this.props.theme,
                type: this.props.type,
                tabindex: this.props.tabindex,
                "expired-callback": this.handleExpired,
                "error-callback": this.handleErrored,
                size: this.props.size,
                stoken: this.props.stoken,
                hl: this.props.hl,
                badge: this.props.badge,
            });
            this.captcha.appendChild(wrapper);
        }
        if (
            this._executeRequested &&
            this.props.grecaptcha &&
            this._widgetId !== undefined
        ) {
            this._executeRequested = false;
            this.execute();
        }
    }

    componentDidMount() {
        this.explicitRender();
    }

    componentDidUpdate() {
        this.explicitRender();
    }

    componentWillUnmount() {
        if (this._widgetId !== undefined) {
            this.delayOfCaptchaIframeRemoving();
            this.reset();
        }
    }

    delayOfCaptchaIframeRemoving() {
        const temporaryNode = document.createElement("div");
        document.body.appendChild(temporaryNode);
        temporaryNode.style.display = "none";

        // move of the recaptcha to a temporary node
        while (this.captcha.firstChild) {
            temporaryNode.appendChild(this.captcha.firstChild);
        }

        // delete the temporary node after reset will be done
        setTimeout(() => {
            document.body.removeChild(temporaryNode);
        }, 5000);
    }

    handleRecaptchaRef(elem) {
        this.captcha = elem;
    }

    render() {
        // consume properties owned by the reCATPCHA, pass the rest to the div so the user can style it.
        /* eslint-disable no-unused-vars */
        /* eslint-disable @typescript-eslint/no-unused-vars */
        const {
            sitekey,
            onChange,
            theme,
            type,
            tabindex,
            onExpired,
            onErrored,
            size,
            stoken,
            grecaptcha,
            badge,
            hl,
            ...childProps
        } = this.props;
        /* eslint-enable no-unused-vars */
        return <div {...childProps} ref={this.handleRecaptchaRef} />;
    }
}

ReCAPTCHA.displayName = "ReCAPTCHA";
ReCAPTCHA.propTypes = {
    sitekey: PropTypes.string.isRequired,
    onChange: PropTypes.func,
    grecaptcha: PropTypes.object,
    theme: PropTypes.oneOf(["dark", "light"]),
    type: PropTypes.oneOf(["image", "audio"]),
    tabindex: PropTypes.number,
    onExpired: PropTypes.func,
    onErrored: PropTypes.func,
    size: PropTypes.oneOf(["compact", "normal", "invisible"]),
    stoken: PropTypes.string,
    hl: PropTypes.string,
    badge: PropTypes.oneOf(["bottomright", "bottomleft", "inline"]),
};
ReCAPTCHA.defaultProps = {
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    onChange: () => {},
    theme: "light",
    type: "image",
    tabindex: 0,
    size: "normal",
    badge: "bottomright",
};
