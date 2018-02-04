<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------
// $Id$

return [
    'url_route_on' => true,
    'log'          => [
        'type' => 'file', // 支持 socket trace file
    ],
    'view'         => [
        // 模板引擎
        'engine_type'   => 'think',
        // 模板引擎配置
        'engine_config' => [
            // 模板路径
            'view_path'   => '',
            // 模板后缀
            'view_suffix' => '.html',
            // 模板文件名分隔符
            'view_depr'   => DS,
        ],
        // 输出字符串替换
        'parse_str'     => [],
    ],
];
