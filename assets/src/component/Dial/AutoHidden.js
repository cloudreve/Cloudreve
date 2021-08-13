import React, { useState, useEffect } from "react";
import Zoom from "@material-ui/core/Zoom";

function AutoHidden({ children, enable }) {
    const [hidden, setHidden] = useState(false);

    let prev = window.scrollY;
    let lastUpdate = window.scrollY;
    const show = 50;

    useEffect(() => {
        const handleNavigation = (e) => {
            const window = e.currentTarget;

            if (prev > window.scrollY) {
                if (lastUpdate - window.scrollY > show) {
                    lastUpdate = window.scrollY;
                    setHidden(false);
                }
            } else if (prev < window.scrollY) {
                if (window.scrollY - lastUpdate > show) {
                    lastUpdate = window.scrollY;
                    setHidden(true);
                }
            }
            prev = window.scrollY;
        };
        if (enable) {
            window.addEventListener("scroll", (e) => handleNavigation(e));
        }
        // eslint-disable-next-line
    }, [enable]);

    return <Zoom in={!hidden}>{children}</Zoom>;
}

export default AutoHidden;
