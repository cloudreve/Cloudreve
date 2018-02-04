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
 * view测试
 * @author    mahuan <mahuan@d1web.top>
 */

namespace tests\thinkphp\library\think;

class viewTest extends \PHPUnit_Framework_TestCase
{

    /**
     * 句柄测试
     * @return  mixed
     * @access public
     */
    public function testGetInstance()
    {
        \think\Cookie::get('a');
        $view_instance = \think\View::instance();
        $this->assertInstanceOf('\think\view', $view_instance, 'instance方法返回错误');
    }

    /**
     * 测试变量赋值
     * @return  mixed
     * @access public
     */
    public function testAssign()
    {
        $view_instance      = \think\View::instance();
        $view_instance->key = 'value';
        $this->assertTrue(isset($view_instance->key));
        $this->assertEquals('value', $view_instance->key);
        $data = $view_instance->assign(array('key' => 'value'));
        $data = $view_instance->assign('key2', 'value2');
        //测试私有属性
        $expect_data = array('key' => 'value', 'key2' => 'value2');
        $this->assertAttributeEquals($expect_data, 'data', $view_instance);
    }

    /**
     *  测试引擎设置
     * @return  mixed
     * @access public
     */
    public function testEngine()
    {
        $view_instance = \think\View::instance();
        $data          = $view_instance->engine('php');
        $data          = $view_instance->engine(['type' => 'php', 'view_path' => '', 'view_suffix' => '.php', 'view_depr' => DS]);
        $php_engine    = new \think\view\driver\Php(['view_path' => '', 'view_suffix' => '.php', 'view_depr' => DS]);
        $this->assertAttributeEquals($php_engine, 'engine', $view_instance);
        //测试模板引擎驱动
        $data         = $view_instance->engine(['type' => 'think', 'view_path' => '', 'view_suffix' => '.html', 'view_depr' => DS]);
        $think_engine = new \think\view\driver\Think(['view_path' => '', 'view_suffix' => '.html', 'view_depr' => DS]);
        $this->assertAttributeEquals($think_engine, 'engine', $view_instance);
    }

    public function testReplace()
    {
        $view_instance = \think\View::instance();
        $view_instance->replace('string', 'replace')->display('string');
    }

}
