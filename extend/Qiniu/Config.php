<?php
namespace Qiniu;

use Qiniu\Zone;

final class Config
{
    const SDK_VER = '7.1.3';

    const BLOCK_SIZE = 4194304; //4*1024*1024 分块上传块大小，该参数为接口规格，不能修改

    const RS_HOST  = 'http://rs.qbox.me';               // 文件元信息管理操作Host
    const RSF_HOST = 'http://rsf.qbox.me';              // 列举操作Host
    const API_HOST = 'http://api.qiniu.com';            // 数据处理操作Host
    const UC_HOST  = 'http://uc.qbox.me';              // Host

    public $zone;

    public function __construct(Zone $z = null)         // 构造函数，默认为zone0
    {
        // if ($z === null) {
            $this->zone = new Zone();
        // }
    }
}
