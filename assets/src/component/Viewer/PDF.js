import React, { useCallback, useEffect, useState } from "react";
import { Document, Page, pdfjs } from "react-pdf";
import { Paper } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import { useLocation, useParams, useRouteMatch } from "react-router";
import { getBaseURL } from "../../middleware/Api";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../actions";
import { changeSubTitle } from "../../redux/viewUpdate/action";
import pathHelper from "../../utils/page";
import TextLoading from "../Placeholder/TextLoading";
pdfjs.GlobalWorkerOptions.workerSrc = `//cdnjs.cloudflare.com/ajax/libs/pdf.js/${pdfjs.version}/pdf.worker.js`;

const useStyles = makeStyles((theme) => ({
    layout: {
        marginTop: "30px",
        marginLeft: theme.spacing(3),
        marginRight: theme.spacing(3),
        [theme.breakpoints.up(1100 + theme.spacing(3) * 2)]: {
            maxWidth: 900,
            marginLeft: "auto",
            marginRight: "auto",
        },
        marginBottom: 50,
    },
    "@global": {
        canvas: {
            width: "100% !important",
            height: "auto !important",
            borderRadius: 4,
        },
    },
    paper: {
        marginBottom: theme.spacing(3),
    },
}));

function useQuery() {
    return new URLSearchParams(useLocation().search);
}

export default function PDFViewer() {
    const math = useRouteMatch();
    const location = useLocation();
    const query = useQuery();
    const { id } = useParams();

    const [pageNumber, setPageNumber] = useState(1);

    const dispatch = useDispatch();
    const SetSubTitle = useCallback(
        (title) => dispatch(changeSubTitle(title)),
        [dispatch]
    );
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    useEffect(() => {
        if (!pathHelper.isSharePage(location.pathname)) {
            const path = query.get("p").split("/");
            SetSubTitle(path[path.length - 1]);
        } else {
            SetSubTitle(query.get("name"));
        }
        // eslint-disable-next-line
    }, [math.params[0], location]);

    const removeTextLayerOffset = () => {
        const textLayers = document.querySelectorAll(
            ".react-pdf__Page__textContent"
        );
        textLayers.forEach((layer) => {
            const { style } = layer;
            style.display = "none";
        });
    };

    const classes = useStyles();
    return (
        <div className={classes.layout}>
            <Document
                onLoadSuccess={({ numPages }) => setPageNumber(numPages)}
                onLoadError={(error) => {
                    ToggleSnackbar(
                        "top",
                        "right",
                        "PDF 加载失败，" + error.message,
                        "error"
                    );
                }}
                loading={
                    <Paper className={classes.paper} elevation={1}>
                        <TextLoading />
                    </Paper>
                }
                file={
                    getBaseURL() +
                    (pathHelper.isSharePage(location.pathname)
                        ? "/share/preview/" +
                          id +
                          (query.get("share_path") !== ""
                              ? "?path=" +
                                encodeURIComponent(query.get("share_path"))
                              : "")
                        : "/file/preview/" + query.get("id"))
                }
            >
                {Array.from(new Array(pageNumber), (el, index) => (
                    <Paper className={classes.paper} elevation={1}>
                        <Page
                            width={900}
                            onLoadSuccess={removeTextLayerOffset}
                            key={`page_${index + 1}`}
                            pageNumber={index + 1}
                            renderAnnotationLayer={false}
                        />
                    </Paper>
                ))}
            </Document>
        </div>
    );
}
