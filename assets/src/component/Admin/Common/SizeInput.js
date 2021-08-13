import FormControl from "@material-ui/core/FormControl";
import Input from "@material-ui/core/Input";
import InputAdornment from "@material-ui/core/InputAdornment";
import InputLabel from "@material-ui/core/InputLabel";
import MenuItem from "@material-ui/core/MenuItem";
import Select from "@material-ui/core/Select";
import React, { useCallback, useState } from "react";
import { useDispatch } from "react-redux";
import { toggleSnackbar } from "../../../actions";

const unitTransform = (v) => {
    if (v < 1024) {
        return [Math.round(v), 1];
    }
    if (v < 1024 * 1024) {
        return [Math.round(v / 1024), 1024];
    }
    if (v < 1024 * 1024 * 1024) {
        return [Math.round(v / (1024 * 1024)), 1024 * 1024];
    }
    if (v < 1024 * 1024 * 1024 * 1024) {
        return [Math.round(v / (1024 * 1024 * 1024)), 1024 * 1024 * 1024];
    }
    return [
        Math.round(v / (1024 * 1024 * 1024 * 1024)),
        1024 * 1024 * 1024 * 1024,
    ];
};

export default function SizeInput({
    onChange,
    min,
    value,
    required,
    label,
    max,
    suffix,
}) {
    const dispatch = useDispatch();
    const ToggleSnackbar = useCallback(
        (vertical, horizontal, msg, color) =>
            dispatch(toggleSnackbar(vertical, horizontal, msg, color)),
        [dispatch]
    );

    const [unit, setUnit] = useState(1);
    let first = true;

    const transform = useCallback(() => {
        const res = unitTransform(value);
        if (first && value !== 0) {
            setUnit(res[1]);
            first = false;
        }
        return res;
    }, [value]);

    return (
        <FormControl>
            <InputLabel htmlFor="component-helper">{label}</InputLabel>
            <Input
                style={{ width: 200 }}
                value={transform()[0]}
                type={"number"}
                inputProps={{ min: min, step: 1 }}
                onChange={(e) => {
                    if (e.target.value * unit < max) {
                        onChange({
                            target: {
                                value: (e.target.value * unit).toString(),
                            },
                        });
                    } else {
                        ToggleSnackbar(
                            "top",
                            "right",
                            "超出最大尺寸限制",
                            "warning"
                        );
                    }
                }}
                required={required}
                endAdornment={
                    <InputAdornment position="end">
                        <Select
                            labelId="demo-simple-select-label"
                            id="demo-simple-select"
                            value={unit}
                            onChange={(e) => {
                                if (transform()[0] * e.target.value < max) {
                                    onChange({
                                        target: {
                                            value: (
                                                transform()[0] * e.target.value
                                            ).toString(),
                                        },
                                    });
                                    setUnit(e.target.value);
                                } else {
                                    ToggleSnackbar(
                                        "top",
                                        "right",
                                        "超出最大尺寸限制",
                                        "warning"
                                    );
                                }
                            }}
                        >
                            <MenuItem value={1}>B{suffix && suffix}</MenuItem>
                            <MenuItem value={1024}>
                                KB{suffix && suffix}
                            </MenuItem>
                            <MenuItem value={1024 * 1024}>
                                MB{suffix && suffix}
                            </MenuItem>
                            <MenuItem value={1024 * 1024 * 1024}>
                                GB{suffix && suffix}
                            </MenuItem>
                            <MenuItem value={1024 * 1024 * 1024 * 1024}>
                                TB{suffix && suffix}
                            </MenuItem>
                        </Select>
                    </InputAdornment>
                }
            />
        </FormControl>
    );
}
