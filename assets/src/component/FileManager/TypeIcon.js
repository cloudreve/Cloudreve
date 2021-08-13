import React from "react";
import { mediaType } from "../../config";
import ImageIcon from "@material-ui/icons/PhotoSizeSelectActual";
import VideoIcon from "@material-ui/icons/Videocam";
import AudioIcon from "@material-ui/icons/Audiotrack";
import PdfIcon from "@material-ui/icons/PictureAsPdf";
import {
    Android,
    FileExcelBox,
    FilePowerpointBox,
    FileWordBox,
    LanguageC,
    LanguageCpp,
    LanguageGo,
    LanguageJavascript,
    LanguagePhp,
    LanguagePython,
    MagnetOn,
    ScriptText,
    WindowRestore,
    ZipBox,
} from "mdi-material-ui";
import FileShowIcon from "@material-ui/icons/InsertDriveFile";
import { lighten } from "@material-ui/core/styles";
import useTheme from "@material-ui/core/styles/useTheme";
import { Avatar } from "@material-ui/core";

const icons = {
    audio: {
        color: "#651fff",
        icon: AudioIcon,
    },
    video: {
        color: "#d50000",
        icon: VideoIcon,
    },
    image: {
        color: "#d32f2f",
        icon: ImageIcon,
    },
    pdf: {
        color: "#f44336",
        icon: PdfIcon,
    },
    word: {
        color: "#538ce5",
        icon: FileWordBox,
    },
    ppt: {
        color: "rgb(239, 99, 63)",
        icon: FilePowerpointBox,
    },
    excel: {
        color: "#4caf50",
        icon: FileExcelBox,
    },
    text: {
        color: "#607d8b",
        icon: ScriptText,
    },
    torrent: {
        color: "#5c6bc0",
        icon: MagnetOn,
    },
    zip: {
        color: "#f9a825",
        icon: ZipBox,
    },
    excute: {
        color: "#1a237e",
        icon: WindowRestore,
    },
    android: {
        color: "#8bc34a",
        icon: Android,
    },
    file: {
        color: "#607d8b",
        icon: FileShowIcon,
    },
    php: {
        color: "#777bb3",
        icon: LanguagePhp,
    },
    go: {
        color: "#16b3da",
        icon: LanguageGo,
    },
    python: {
        color: "#3776ab",
        icon: LanguagePython,
    },
    c: {
        color: "#a8b9cc",
        icon: LanguageC,
    },
    cpp: {
        color: "#004482",
        icon: LanguageCpp,
    },
    js: {
        color: "#f4d003",
        icon: LanguageJavascript,
    },
};

const getColor = (theme, color) =>
    theme.palette.type === "light" ? color : lighten(color, 0.2);

let color;

const TypeIcon = (props) => {
    const theme = useTheme();

    const fileSuffix = props.fileName.split(".").pop().toLowerCase();
    let fileType = "file";
    Object.keys(mediaType).forEach((k) => {
        if (mediaType[k].indexOf(fileSuffix) !== -1) {
            fileType = k;
        }
    });
    const IconComponent = icons[fileType].icon;
    color = getColor(theme, icons[fileType].color);
    if (props.getColorValue) {
        props.getColorValue(color);
    }

    return (
        <>
            {props.isUpload && (
                <Avatar
                    className={props.className}
                    style={{
                        backgroundColor: color,
                    }}
                >
                    <IconComponent
                        className={props.iconClassName}
                        style={{
                            color: theme.palette.background.paper,
                        }}
                    />
                </Avatar>
            )}
            {!props.isUpload && (
                <IconComponent
                    className={props.className}
                    style={{
                        color: color,
                    }}
                />
            )}
        </>
    );
};

export default TypeIcon;
