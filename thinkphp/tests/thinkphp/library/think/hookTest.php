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
 * Hook类测试
 * @author    liu21st <liu21st@gmail.com>
 */

namespace tests\thinkphp\library\think;

use think\Hook;

class hookTest extends \PHPUnit_Framework_TestCase
{

    public function testRun()
    {
        Hook::add('my_pos', '\tests\thinkphp\library\think\behavior\One');
        Hook::add('my_pos', ['\tests\thinkphp\library\think\behavior\Two']);
        Hook::add('my_pos', '\tests\thinkphp\library\think\behavior\Three', true);
        $data['id']   = 0;
        $data['name'] = 'thinkphp';
        Hook::listen('my_pos', $data);
        $this->assertEquals(2, $data['id']);
        $this->assertEquals('thinkphp', $data['name']);
        $this->assertEquals([
            '\tests\thinkphp\library\think\behavior\Three',
            '\tests\thinkphp\library\think\behavior\One',
            '\tests\thinkphp\library\think\behavior\Two'],
            Hook::get('my_pos'));
    }

    public function testImport()
    {
        Hook::import(['my_pos' => [
            '\tests\thinkphp\library\think\behavior\One',
            '\tests\thinkphp\library\think\behavior\Three'],
        ]);
        Hook::import(['my_pos' => ['\tests\thinkphp\library\think\behavior\Two']], false);
        Hook::import(['my_pos' => ['\tests\thinkphp\library\think\behavior\Three', '_overlay' => true]]);
        $data['id']   = 0;
        $data['name'] = 'thinkphp';
        Hook::listen('my_pos', $data);
        $this->assertEquals(3, $data['id']);

    }

    public function testExec()
    {
        $data['id']   = 0;
        $data['name'] = 'thinkphp';
        $this->assertEquals(true, Hook::exec('\tests\thinkphp\library\think\behavior\One'));
        $this->assertEquals(false, Hook::exec('\tests\thinkphp\library\think\behavior\One', 'test', $data));
        $this->assertEquals('test', $data['name']);
        $this->assertEquals('Closure', Hook::exec(function (&$data) {$data['name'] = 'Closure';return 'Closure';}));

    }

}
