import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogTitle from "@material-ui/core/DialogTitle";
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import FormHelperText from "@material-ui/core/FormHelperText";
import Input from "@material-ui/core/Input";
import InputLabel from "@material-ui/core/InputLabel";
import MenuItem from "@material-ui/core/MenuItem";
import Select from "@material-ui/core/Select";
import { makeStyles } from "@material-ui/core/styles";
import Switch from "@material-ui/core/Switch";
import React, { useEffect, useState } from "react";
import API from "../../../middleware/Api";

const useStyles = makeStyles(() => ({
    formContainer: {
        margin: "8px 0 8px 0",
    },
}));

export default function AddGroup({ open, onClose, onSubmit }) {
    const classes = useStyles();
    const [groups, setGroups] = useState([]);
    const [group, setGroup] = useState({
        name: "",
        group_id: 2,
        time: "",
        price: "",
        score: "",
        des: "",
        highlight: false,
    });

    useEffect(() => {
        if (open && groups.length === 0) {
            API.get("/admin/groups")
                .then((response) => {
                    setGroups(response.data);
                })
                // eslint-disable-next-line @typescript-eslint/no-empty-function
                .catch(() => {});
        }
        // eslint-disable-next-line
    }, [open]);

    const handleChange = (name) => (event) => {
        setGroup({
            ...group,
            [name]: event.target.value,
        });
    };

    const handleCheckChange = (name) => (event) => {
        setGroup({
            ...group,
            [name]: event.target.checked,
        });
    };

    const submit = (e) => {
        e.preventDefault();
        const groupCopy = { ...group };
        groupCopy.time = parseInt(groupCopy.time) * 86400;
        groupCopy.price = parseInt(groupCopy.price) * 100;
        groupCopy.score = parseInt(groupCopy.score);
        groupCopy.id = new Date().valueOf();
        groupCopy.des = groupCopy.des.split("\n");
        onSubmit(groupCopy);
    };

    return (
        <Dialog
            open={open}
            onClose={onClose}
            aria-labelledby="alert-dialog-title"
            aria-describedby="alert-dialog-description"
            maxWidth={"xs"}
            scroll={"paper"}
        >
            <form onSubmit={submit}>
                <DialogTitle id="alert-dialog-title">
                    添加可购用户组
                </DialogTitle>
                <DialogContent>
                    <DialogContentText id="alert-dialog-description">
                        <div className={classes.formContainer}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    名称
                                </InputLabel>
                                <Input
                                    value={group.name}
                                    onChange={handleChange("name")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    商品展示名称
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.formContainer}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    用户组
                                </InputLabel>
                                <Select
                                    value={group.group_id}
                                    onChange={handleChange("group_id")}
                                    required
                                >
                                    {groups.map((v) => {
                                        if (v.ID !== 3) {
                                            return (
                                                <MenuItem value={v.ID}>
                                                    {v.Name}
                                                </MenuItem>
                                            );
                                        }
                                        return null;
                                    })}
                                </Select>
                                <FormHelperText id="component-helper-text">
                                    购买后升级的用户组
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.formContainer}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    有效期 (天)
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        min: 1,
                                        step: 1,
                                    }}
                                    value={group.time}
                                    onChange={handleChange("time")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    单位购买时间的有效期
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.formContainer}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    单价 (元)
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        min: 0.01,
                                        step: 0.01,
                                    }}
                                    value={group.price}
                                    onChange={handleChange("price")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    用户组的单价
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.formContainer}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    单价 (积分)
                                </InputLabel>
                                <Input
                                    type={"number"}
                                    inputProps={{
                                        min: 0,
                                        step: 1,
                                    }}
                                    value={group.score}
                                    onChange={handleChange("score")}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    使用积分购买时的价格，填写为 0
                                    表示不能使用积分购买
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.formContainer}>
                            <FormControl fullWidth>
                                <InputLabel htmlFor="component-helper">
                                    商品描述 (一行一个)
                                </InputLabel>
                                <Input
                                    value={group.des}
                                    onChange={handleChange("des")}
                                    multiline
                                    rowsMax={10}
                                    required
                                />
                                <FormHelperText id="component-helper-text">
                                    购买页面展示的商品描述
                                </FormHelperText>
                            </FormControl>
                        </div>

                        <div className={classes.formContainer}>
                            <FormControl fullWidth>
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={group.highlight}
                                            onChange={handleCheckChange(
                                                "highlight"
                                            )}
                                        />
                                    }
                                    label="突出展示"
                                />
                                <FormHelperText id="component-helper-text">
                                    开启后，在商品选择页面会被突出展示
                                </FormHelperText>
                            </FormControl>
                        </div>
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={onClose} color="default">
                        取消
                    </Button>
                    <Button type={"submit"} color="primary">
                        确定
                    </Button>
                </DialogActions>
            </form>
        </Dialog>
    );
}
