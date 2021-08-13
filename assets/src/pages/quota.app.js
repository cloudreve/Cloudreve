import React, { Component } from "react";
import PropTypes from "prop-types";
import Navbar from "../component/Navbar/Navbar.js";
import AlertBar from "../component/Common/Snackbar";
import Quota from "../component/VAS/Quota";
import BuyQuota from "../component/VAS/BuyQuota";
import { createMuiTheme } from "@material-ui/core/styles";
import { checkGetParameters } from "../utils/index";

import { CssBaseline, withStyles, MuiThemeProvider } from "@material-ui/core";

const theme = createMuiTheme(window.colorTheme);
const styles = (theme) => ({
    root: {
        display: "flex",
    },
    content: {
        flexGrow: 1,
        padding: theme.spacing(0),
        minWidth: 0,
    },
    toolbar: theme.mixins.toolbar,
});

class QuatoApp extends Component {
    render() {
        const { classes } = this.props;
        return (
            <React.Fragment>
                <MuiThemeProvider theme={theme}>
                    <div className={classes.root} id="container">
                        <CssBaseline />
                        <AlertBar />
                        <Navbar />
                        <main className={classes.content}>
                            <div className={classes.toolbar} />
                            {checkGetParameters("buyPack") && (
                                <BuyQuota tabID={0} />
                            )}
                            {checkGetParameters("buyGroup") && (
                                <BuyQuota tabID={1} />
                            )}
                            {checkGetParameters("redeem") && (
                                <BuyQuota tabID={2} />
                            )}
                            {!checkGetParameters("redeem") &&
                                !(
                                    checkGetParameters("buyGroup") ||
                                    checkGetParameters("buyPack")
                                ) && <Quota />}
                        </main>
                    </div>
                </MuiThemeProvider>
            </React.Fragment>
        );
    }
}

QuatoApp.propTypes = {
    classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(QuatoApp);
