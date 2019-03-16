<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2018 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: 麦当苗儿 <zuojiazi@vip.qq.com> <http://zjzit.cn>
// +----------------------------------------------------------------------

namespace think\db\exception;

use think\exception\DbException;

class DataNotFoundException extends DbException
{
    protected $table;

    /**
     * DbException constructor.
     * @param string $message
     * @param string $table
     * @param array $config
     */
    public function __construct($message, $table = '', array $config = [])
    {
        $this->message = $message;
        $this->table   = $table;

        $this->setData('Database Config', $config);
    }

    /**
     * 获取数据表名
     * @access public
     * @return string
     */
    public function getTable()
    {
        return $this->table;
    }
}
