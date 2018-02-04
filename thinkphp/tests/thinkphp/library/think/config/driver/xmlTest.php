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
 * Xml配置测试
 * @author    7IN0SAN9 <me@7in0.me>
 */

namespace tests\thinkphp\library\think\config\driver;

use think\config;

class xmlTest extends \PHPUnit_Framework_TestCase
{
    public function testParse()
    {
        Config::parse('<?xml version="1.0"?><document><xmlstring>1</xmlstring></document>', 'xml');
        $this->assertEquals(1, Config::get('xmlstring'));
        Config::reset();
        Config::parse(__DIR__ . '/fixtures/config.xml');
        $this->assertTrue(Config::has('xmlfile.istrue'));
        $this->assertEquals(1, Config::get('xmlfile.istrue'));
        Config::reset();
    }
}
