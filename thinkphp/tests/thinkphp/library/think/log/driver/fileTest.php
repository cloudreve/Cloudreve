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
 * Test File Log
 */
namespace tests\thinkphp\library\think\log\driver;

use think\Log;

class fileTest extends \PHPUnit_Framework_TestCase
{
    protected function setUp()
    {
        Log::init(['type' => 'file']);
    }

    public function testRecord()
    {
        $record_msg = 'record';
        Log::record($record_msg, 'notice');
        $logs = Log::getLog();

        $this->assertEquals([], $logs);
    }
}
