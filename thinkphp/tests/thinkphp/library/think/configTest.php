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
 * 配置测试
 * @author    Haotong Lin <lofanmi@gmail.com>
 */

namespace tests\thinkphp\library\think;

use ReflectionClass;
use think\Config;

class configTest extends \PHPUnit_Framework_TestCase
{
    public function testRange()
    {
        $reflectedClass         = new ReflectionClass('\think\Config');
        $reflectedPropertyRange = $reflectedClass->getProperty('range');
        $reflectedPropertyRange->setAccessible(true);
        $reflectedPropertyConfig = $reflectedClass->getProperty('config');
        $reflectedPropertyConfig->setAccessible(true);
        // test default range
        $this->assertEquals('_sys_', $reflectedPropertyRange->getValue());
        $config = $reflectedPropertyConfig->getValue();
        $this->assertTrue(is_array($config));
        // test range initialization
        Config::range('_test_');
        $this->assertEquals('_test_', $reflectedPropertyRange->getValue());
        $config = $reflectedPropertyConfig->getValue();
        $this->assertEquals([], $config['_test_']);
    }

    // public function testParse()
    // {
    //  see \think\config\driver\...Test.php
    // }

    public function testLoad()
    {
        $file   = APP_PATH . 'config' . EXT;
        $config = array_change_key_case(include $file);
        $name   = '_name_';
        $range  = '_test_';

        $reflectedClass          = new ReflectionClass('\think\Config');
        $reflectedPropertyConfig = $reflectedClass->getProperty('config');
        $reflectedPropertyConfig->setAccessible(true);
        $reflectedPropertyConfig->setValue([]);

        $this->assertEquals($config, Config::load($file, $name, $range));
        $this->assertNotEquals(null, Config::load($file, $name, $range));
    }

    public function testHas()
    {
        $range = '_test_';
        $this->assertFalse(Config::has('abcd', $range));
        $reflectedClass          = new ReflectionClass('\think\Config');
        $reflectedPropertyConfig = $reflectedClass->getProperty('config');
        $reflectedPropertyConfig->setAccessible(true);

        // if (!strpos($name, '.')):
        $reflectedPropertyConfig->setValue([
            $range => ['abcd' => 'value'],
        ]);
        $this->assertTrue(Config::has('abcd', $range));

        // else ...
        $this->assertFalse(Config::has('abcd.efg', $range));

        $reflectedPropertyConfig->setValue([
            $range => ['abcd' => ['efg' => 'value']],
        ]);
        $this->assertTrue(Config::has('abcd.efg', $range));
    }

    public function testGet()
    {
        $range                   = '_test_';
        $reflectedClass          = new ReflectionClass('\think\Config');
        $reflectedPropertyConfig = $reflectedClass->getProperty('config');
        $reflectedPropertyConfig->setAccessible(true);
        // test all configurations
        $reflectedPropertyConfig->setValue([$range => []]);
        $this->assertEquals([], Config::get(null, $range));
        $this->assertEquals(null, Config::get(null, 'does_not_exist'));
        $value = 'value';
        // test getting configuration
        $reflectedPropertyConfig->setValue([$range => ['abcd' => 'efg']]);
        $this->assertEquals('efg', Config::get('abcd', $range));
        $this->assertEquals(null, Config::get('does_not_exist', $range));
        $this->assertEquals(null, Config::get('abcd', 'does_not_exist'));
        // test getting configuration with dot syntax
        $reflectedPropertyConfig->setValue([$range => [
            'one' => ['two' => $value],
        ]]);
        $this->assertEquals($value, Config::get('one.two', $range));
        $this->assertEquals(null, Config::get('one.does_not_exist', $range));
        $this->assertEquals(null, Config::get('one.two', 'does_not_exist'));
    }

    public function testSet()
    {
        $range                   = '_test_';
        $reflectedClass          = new ReflectionClass('\think\Config');
        $reflectedPropertyConfig = $reflectedClass->getProperty('config');
        $reflectedPropertyConfig->setAccessible(true);
        $reflectedPropertyConfig->setValue([]);
        // if (is_string($name)):
        // without dot syntax
        $name  = 'name';
        $value = 'value';
        Config::set($name, $value, $range);
        $config = $reflectedPropertyConfig->getValue();
        $this->assertEquals($value, $config[$range][$name]);
        // with dot syntax
        $name  = 'one.two';
        $value = 'dot value';
        Config::set($name, $value, $range);
        $config = $reflectedPropertyConfig->getValue();
        $this->assertEquals($value, $config[$range]['one']['two']);
        // if (is_array($name)):
        // see testLoad()
        // ...
        // test getting all configurations...?
        // return self::$config[$range]; ??
        $value = ['all' => 'configuration'];
        $reflectedPropertyConfig->setValue([$range => $value]);
        $this->assertEquals($value, Config::set(null, null, $range));
        $this->assertNotEquals(null, Config::set(null, null, $range));
    }

    public function testReset()
    {
        $range                   = '_test_';
        $reflectedClass          = new ReflectionClass('\think\Config');
        $reflectedPropertyConfig = $reflectedClass->getProperty('config');
        $reflectedPropertyConfig->setAccessible(true);
        $reflectedPropertyConfig->setValue([$range => ['abcd' => 'efg']]);

        // clear all configurations
        Config::reset(true);
        $config = $reflectedPropertyConfig->getValue();
        $this->assertEquals([], $config);
        // clear the configuration in range of parameter.
        $reflectedPropertyConfig->setValue([
            $range => [
                'abcd' => 'efg',
                'hijk' => 'lmn',
            ],
            'a'    => 'b',
        ]);
        Config::reset($range);
        $config = $reflectedPropertyConfig->getValue();
        $this->assertEquals([
            $range => [],
            'a'    => 'b',
        ], $config);
    }
}
