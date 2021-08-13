import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogTitle from "@material-ui/core/DialogTitle";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import React from "react";

export default function MagicVar({ isFile, open, onClose, isSlave }) {
    return (
        <Dialog
            open={open}
            onClose={onClose}
            aria-labelledby="alert-dialog-title"
            aria-describedby="alert-dialog-description"
        >
            <DialogTitle id="alert-dialog-title">
                {isFile ? "文件名魔法变量" : "路径魔法变量"}
            </DialogTitle>
            <DialogContent>
                <TableContainer>
                    <Table size="small" aria-label="a dense table">
                        <TableHead>
                            <TableRow>
                                <TableCell>魔法变量</TableCell>
                                <TableCell>描述</TableCell>
                                <TableCell>示例</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{randomkey16}"}
                                </TableCell>
                                <TableCell>16位随机字符</TableCell>
                                <TableCell>N6IimT5XZP324ACK</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{randomkey8}"}
                                </TableCell>
                                <TableCell>8位随机字符</TableCell>
                                <TableCell>gWz78q30</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{timestamp}"}
                                </TableCell>
                                <TableCell>秒级时间戳</TableCell>
                                <TableCell>1582692933</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{timestamp_nano}"}
                                </TableCell>
                                <TableCell>纳秒级时间戳</TableCell>
                                <TableCell>1582692933231834600</TableCell>
                            </TableRow>
                            {!isSlave && (
                                <TableRow>
                                    <TableCell component="th" scope="row">
                                        {"{uid}"}
                                    </TableCell>
                                    <TableCell>用户ID</TableCell>
                                    <TableCell>1</TableCell>
                                </TableRow>
                            )}
                            {isFile && (
                                <TableRow>
                                    <TableCell component="th" scope="row">
                                        {"{originname}"}
                                    </TableCell>
                                    <TableCell>原始文件名</TableCell>
                                    <TableCell>MyPico.mp4</TableCell>
                                </TableRow>
                            )}
                            {!isFile && !isSlave && (
                                <TableRow>
                                    <TableCell component="th" scope="row">
                                        {"{path}"}
                                    </TableCell>
                                    <TableCell>用户上传路径</TableCell>
                                    <TableCell>/我的文件/学习资料/</TableCell>
                                </TableRow>
                            )}
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{date}"}
                                </TableCell>
                                <TableCell>日期</TableCell>
                                <TableCell>20060102</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{datetime}"}
                                </TableCell>
                                <TableCell>日期时间</TableCell>
                                <TableCell>20060102150405</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{year}"}
                                </TableCell>
                                <TableCell>年份</TableCell>
                                <TableCell>2006</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{month}"}
                                </TableCell>
                                <TableCell>月份</TableCell>
                                <TableCell>01</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{day}"}
                                </TableCell>
                                <TableCell>日</TableCell>
                                <TableCell>02</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{hour}"}
                                </TableCell>
                                <TableCell>小时</TableCell>
                                <TableCell>15</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{minute}"}
                                </TableCell>
                                <TableCell>分钟</TableCell>
                                <TableCell>04</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell component="th" scope="row">
                                    {"{second}"}
                                </TableCell>
                                <TableCell>秒</TableCell>
                                <TableCell>05</TableCell>
                            </TableRow>
                        </TableBody>
                    </Table>
                </TableContainer>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose} color="primary">
                    关闭
                </Button>
            </DialogActions>
        </Dialog>
    );
}
