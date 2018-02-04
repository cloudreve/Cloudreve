<?php

namespace Qiniu;

use Qiniu\Config;

final class Etag
{
    private static function packArray($v, $a)
    {
        return call_user_func_array('pack', array_merge(array($v), (array)$a));
    }

    private static function blockCount($fsize)
    {
        return (($fsize + (Config::BLOCK_SIZE - 1)) / Config::BLOCK_SIZE);
    }

    private static function calcSha1($data)
    {
        $sha1Str = sha1($data, true);
        $err = error_get_last();
        if ($err !== null) {
            return array(null, $err);
        }
        $byteArray = unpack('C*', $sha1Str);
        return array($byteArray, null);
    }


    public static function sum($filename)
    {
        $fhandler = fopen($filename, 'r');
        $err = error_get_last();
        if ($err !== null) {
            return array(null, $err);
        }

        $fstat = fstat($fhandler);
        $fsize = $fstat['size'];
        if ((int) $fsize === 0) {
            fclose($fhandler);
            return array('Fto5o-5ea0sNMlW_75VgGJCv2AcJ', null);
        }
        $blockCnt = self::blockCount($fsize);
        $sha1Buf = array();

        if ($blockCnt <= 1) {
            array_push($sha1Buf, 0x16);
            $fdata = fread($fhandler, Config::BLOCK_SIZE);
            if ($err !== null) {
                fclose($fhandler);
                return array(null, $err);
            }
            list($sha1Code, ) = self::calcSha1($fdata);
            $sha1Buf = array_merge($sha1Buf, $sha1Code);
        } else {
            array_push($sha1Buf, 0x96);
            $sha1BlockBuf = array();
            for ($i=0; $i < $blockCnt; $i++) {
                $fdata = fread($fhandler, Config::BLOCK_SIZE);
                list($sha1Code, $err) = self::calcSha1($fdata);
                if ($err !== null) {
                    fclose($fhandler);
                    return array(null, $err);
                }
                $sha1BlockBuf = array_merge($sha1BlockBuf, $sha1Code);
            }
            $tmpData = self::packArray('C*', $sha1BlockBuf);
            list($sha1Final, ) = self::calcSha1($tmpData);
            $sha1Buf = array_merge($sha1Buf, $sha1Final);
        }
        $etag = \Qiniu\base64_urlSafeEncode(self::packArray('C*', $sha1Buf));
        return array($etag, null);
    }
}
