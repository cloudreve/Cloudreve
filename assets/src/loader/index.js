import React, { Component } from "react";
import PropTypes from "prop-types";
import hoistStatics from "hoist-non-react-statics";
import { newScript, series, noop } from "./utils";

const loadedScript = [];
const pendingScripts = {};
let failedScript = [];

const addCache = (entry) => {
    if (loadedScript.indexOf(entry) < 0) {
        loadedScript.push(entry);
    }
};

const removeFailedScript = () => {
    if (failedScript.length > 0) {
        failedScript.forEach((script) => {
            const node = document.querySelector(`script[src='${script}']`);
            if (node != null) {
                node.parentNode.removeChild(node);
            }
        });

        failedScript = [];
    }
};

export function startLoadingScripts(scripts, onComplete = noop) {
    // sequence load
    const loadNewScript = (script) => {
        const src = typeof script === "object" ? script.src : script;
        if (loadedScript.indexOf(src) < 0) {
            return (taskComplete) => {
                const callbacks = pendingScripts[src] || [];
                callbacks.push(taskComplete);
                pendingScripts[src] = callbacks;
                if (callbacks.length === 1) {
                    return newScript(script)((err) => {
                        pendingScripts[src].forEach((cb) => cb(err, src));
                        delete pendingScripts[src];
                    });
                }
            };
        }
    };
    const tasks = scripts.map((src) => {
        if (Array.isArray(src)) {
            return src.map(loadNewScript);
        } else return loadNewScript(src);
    });

    series(...tasks)((err, src) => {
        if (err) {
            failedScript.push(src);
        } else {
            if (Array.isArray(src)) {
                src.forEach(addCache);
            } else addCache(src);
        }
    })((err) => {
        removeFailedScript();
        onComplete(err);
    });
}

const uploaderLoader = () => (WrappedComponent) => {
    class ScriptLoader extends Component {
        static propTypes = {
            onScriptLoaded: PropTypes.func,
        };

        static defaultProps = {
            onScriptLoaded: noop,
        };

        constructor(props, context) {
            super(props, context);

            this.state = {
                isScriptLoaded: false,
                isScriptLoadSucceed: false,
            };

            this._isMounted = false;
        }

        componentDidMount() {
            this._isMounted = true;
            const scripts = [
                ["/static/js/uploader/moxie.js"],
                ["/static/js/uploader/plupload.dev.js"],
                ["/static/js/uploader/i18n/zh_CN.js"],
                ["/static/js/uploader/ui.js"],
                ["/static/js/uploader/uploader_" + window.policyType + ".js"],
            ];
            startLoadingScripts(scripts, (err) => {
                if (this._isMounted) {
                    this.setState(
                        {
                            isScriptLoaded: true,
                            isScriptLoadSucceed: !err,
                        },
                        () => {
                            if (!err) {
                                this.props.onScriptLoaded();
                            }
                        }
                    );
                }
            });
        }

        componentWillUnmount() {
            this._isMounted = false;
        }

        // getWrappedInstance() {
        //     return this.refs.wrappedInstance;
        // }

        render() {
            const props = {
                ...this.props,
                ...this.state,
                ref: "wrappedInstance",
            };

            return <WrappedComponent {...props} />;
        }
    }

    return hoistStatics(ScriptLoader, WrappedComponent);
};

export default uploaderLoader;
