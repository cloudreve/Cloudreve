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
 * exception类测试
 * @author    Haotong Lin <lofanmi@gmail.com>
 */

namespace tests\thinkphp\library\think;

use ReflectionMethod;
use think\Exception as ThinkException;
use think\exception\HttpException;

class MyException extends ThinkException
{

}

class exceptionTest extends \PHPUnit_Framework_TestCase
{
    public function testGetHttpStatus()
    {
        try {
            throw new HttpException(404, "Error Processing Request");
        } catch (HttpException $e) {
            $this->assertEquals(404, $e->getStatusCode());
        }
    }

    public function testDebugData()
    {
        $data = ['a' => 'b', 'c' => 'd'];
        try {
            $e      = new MyException("Error Processing Request", 1);
            $method = new ReflectionMethod($e, 'setData');
            $method->setAccessible(true);
            $method->invokeArgs($e, ['test', $data]);
            throw $e;
        } catch (MyException $e) {
            $this->assertEquals(['test' => $data], $e->getData());
        }
    }
}
