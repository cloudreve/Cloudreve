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
 * Lang测试
 * @author    liu21st <liu21st@gmail.com>
 */

namespace tests\thinkphp\library\think;

use think\Config;
use think\Lang;

class langTest extends \PHPUnit_Framework_TestCase
{

    public function testSetAndGet()
    {
        Lang::set('hello,%s', '欢迎,%s');
        $this->assertEquals('欢迎,ThinkPHP', Lang::get('hello,%s', ['ThinkPHP']));
        Lang::set('hello,%s', '歡迎,%s', 'zh-tw');
        $this->assertEquals('歡迎,ThinkPHP', Lang::get('hello,%s', ['ThinkPHP'], 'zh-tw'));
        Lang::set(['hello' => '欢迎', 'use' => '使用']);
        $this->assertEquals('欢迎', Lang::get('hello'));
        $this->assertEquals('欢迎', Lang::get('HELLO'));
        $this->assertEquals('使用', Lang::get('use'));

        Lang::set('hello,{:name}', '欢迎,{:name}');
        $this->assertEquals('欢迎,liu21st', Lang::get('hello,{:name}', ['name' => 'liu21st']));
    }

    public function testLoad()
    {
        Lang::load(__DIR__ . DS . 'lang' . DS . 'lang.php');
        $this->assertEquals('加载', Lang::get('load'));
        Lang::load(__DIR__ . DS . 'lang' . DS . 'lang.php', 'test');
        $this->assertEquals('加载', Lang::get('load', [], 'test'));
    }

    public function testDetect()
    {

        Config::set('lang_list', ['zh-cn', 'zh-tw']);
        Lang::set('hello', '欢迎', 'zh-cn');
        Lang::set('hello', '歡迎', 'zh-tw');

        Config::set('lang_detect_var', 'lang');
        Config::set('lang_cookie_var', 'think_cookie');

        $_GET['lang'] = 'zh-tw';
        Lang::detect();
        $this->assertEquals('歡迎', Lang::get('hello'));

        $_GET['lang'] = 'zh-cn';
        Lang::detect();
        $this->assertEquals('欢迎', Lang::get('hello'));

    }

    public function testRange()
    {
        $this->assertEquals('zh-cn', Lang::range());
        Lang::set('hello', '欢迎', 'test');
        Lang::range('test');
        $this->assertEquals('test', Lang::range());
        $this->assertEquals('欢迎', Lang::get('hello'));
    }
}
