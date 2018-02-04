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
 * 控制器测试
 * @author    Haotong Lin <lofanmi@gmail.com>
 */

namespace tests\thinkphp\library\think;

use ReflectionClass;
use think\Controller;
use think\Request;
use think\View;

require_once CORE_PATH . '../../helper.php';

class Foo extends Controller
{
    public $test = 'test';

    public function _initialize()
    {
        $this->test = 'abcd';
    }

    public function assignTest()
    {
        $this->assign('abcd', 'dcba');
        $this->assign(['key1' => 'value1', 'key2' => 'value2']);
    }

    public function fetchTest()
    {
        $template = dirname(__FILE__) . '/display.html';
        return $this->fetch($template, ['name' => 'ThinkPHP']);
    }

    public function displayTest()
    {
        $template = dirname(__FILE__) . '/display.html';
        return $this->display($template, ['name' => 'ThinkPHP']);
    }
    public function test()
    {
        $data       = [
            'username'   => 'username',
            'nickname'   => 'nickname',
            'password'   => '123456',
            'repassword' => '123456',
            'email'      => 'abc@abc.com',
            'sex'        => '0',
            'age'        => '20',
            'code'       => '1234',
        ];

        $validate = [
            ['username', 'length:5,15', '用户名长度为5到15个字符'],
            ['nickname', 'require', '请填昵称'],
            ['password', '[\w-]{6,15}', '密码长度为6到15个字符'],
            ['repassword', 'confirm:password', '两次密码不一到致'],
            ['email', 'filter:validate_email', '邮箱格式错误'],
            ['sex', 'in:0,1', '性别只能为为男或女'],
            ['age', 'between:1,80', '年龄只能在10-80之间'],
        ];
        return $this->validate($data, $validate);
    }
}

class Bar extends Controller
{
    public $test = 1;

    public $beforeActionList = ['action1', 'action2'];

    public function action1()
    {
        $this->test += 2;
        return 'action1';
    }

    public function action2()
    {
        $this->test += 4;
        return 'action2';
    }
}

class Baz extends Controller
{
    public $test = 1;

    public $beforeActionList = [
        'action1' => ['only' => 'index'],
        'action2' => ['except' => 'index'],
        'action3' => ['only' => 'abcd'],
        'action4' => ['except' => 'abcd'],
    ];

    public function action1()
    {
        $this->test += 2;
        return 'action1';
    }

    public function action2()
    {
        $this->test += 4;
        return 'action2';
    }

    public function action3()
    {
        $this->test += 8;
        return 'action2';
    }

    public function action4()
    {
        $this->test += 16;
        return 'action2';
    }
}

class controllerTest extends \PHPUnit_Framework_TestCase
{
    public function testInitialize()
    {
        $foo = new Foo(Request::instance());
        $this->assertEquals('abcd', $foo->test);
    }

    public function testBeforeAction()
    {
        $obj = new Bar(Request::instance());
        $this->assertEquals(7, $obj->test);

        $obj = new Baz(Request::instance());
        $this->assertEquals(19, $obj->test);
    }

    private function getView($controller)
    {
        $view     = new View();
        $rc       = new ReflectionClass(get_class($controller));
        $property = $rc->getProperty('view');
        $property->setAccessible(true);
        $property->setValue($controller, $view);
        return $view;
    }

    public function testFetch()
    {
        $controller      = new Foo(Request::instance());
        $view            = $this->getView($controller);
        $template        = dirname(__FILE__) . '/display.html';
        $viewFetch       = $view->fetch($template, ['name' => 'ThinkPHP']);
        $this->assertEquals($controller->fetchTest(), $viewFetch);
    }

    public function testDisplay()
    {
        $controller      = new Foo;
        $view            = $this->getView($controller);
        $template        = dirname(__FILE__) . '/display.html';
        $viewFetch       = $view->display($template, ['name' => 'ThinkPHP']);

        $this->assertEquals($controller->displayTest(), $viewFetch);
    }

    public function testAssign()
    {
        $controller = new Foo(Request::instance());
        $view       = $this->getView($controller);
        $controller->assignTest();
        $expect = ['abcd' => 'dcba', 'key1' => 'value1', 'key2' => 'value2'];
        $this->assertAttributeEquals($expect, 'data', $view);
    }

    public function testValidate()
    {
        $controller = new Foo(Request::instance());
        $result = $controller->test();
        $this->assertTrue($result);
    }
}
