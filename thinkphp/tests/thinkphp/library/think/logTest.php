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

/**
 * Log测试
 * @author    liu21st <liu21st@gmail.com>
 */
namespace tests\thinkphp\library\think;

use think\Log;

class logTest extends \PHPUnit_Framework_TestCase
{

    public function testSave()
    {
        Log::init(['type' => 'test']);
        Log::clear();
        Log::record('test');
        Log::record([1, 2, 3]);
        $this->assertTrue(Log::save());
    }

    public function testWrite()
    {
        Log::init(['type' => 'test']);
        Log::clear();
        $this->assertTrue(Log::write('hello', 'info'));
        $this->assertTrue(Log::write([1, 2, 3], 'log'));
    }
}
