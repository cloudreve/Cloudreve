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
 * Validate类测试
 */

namespace tests\thinkphp\library\think;

use think\File;
use think\Validate;

class validateTest extends \PHPUnit_Framework_TestCase
{

    public function testCheck()
    {
        $rule = [
            'name'  => 'require|max:25',
            'age'   => 'number|between:1,120',
            'email' => 'email',
        ];
        $msg = [
            'name.require' => '名称必须',
            'name.max'     => '名称最多不能超过25个字符',
            'age.number'   => '年龄必须是数字',
            'age.between'  => '年龄只能在1-120之间',
            'email'        => '邮箱格式错误',
        ];
        $data = [
            'name'  => 'thinkphp',
            'age'   => 10,
            'email' => 'thinkphp@qq.com',
        ];
        $validate = new Validate($rule, $msg);
        $result   = $validate->check($data);
        $this->assertEquals(true, $result);
    }

    public function testRule()
    {
        $rule = [
            'name'       => 'require|method:get|alphaNum|max:25|expire:2016-1-1,2026-1-1',
            'account'    => 'requireIf:name,thinkphp|alphaDash|min:4|length:4,30',
            'age'        => 'number|between:1,120',
            'email'      => 'requireWith:name|email',
            'host'       => 'activeUrl|activeUrl:A',
            'url'        => 'url',
            'ip'         => 'ip|ip:ipv4',
            'score'      => 'float|gt:60|notBetween:90,100|notIn:70,80|lt:100|elt:100|egt:60',
            'status'     => 'integer|in:0,1,2',
            'begin_time' => 'after:2016-3-18',
            'end_time'   => 'before:2016-10-01',
            'info'       => 'require|array|length:4|max:5|min:2',
            'info.name'  => 'require|length:8|alpha|same:thinkphp',
            'value'      => 'same:100|different:status',
            'bool'       => 'boolean',
            'title'      => 'chsAlpha',
            'city'       => 'chs',
            'nickname'   => 'chsDash',
            'aliasname'  => 'chsAlphaNum',
            'file'       => 'file|fileSize:20480',
            'image'      => 'image|fileMime:image/png|image:80,80,png',
            'test'       => 'test',
        ];
        $data = [
            'name'       => 'thinkphp',
            'account'    => 'liuchen',
            'age'        => 10,
            'email'      => 'thinkphp@qq.com',
            'host'       => 'thinkphp.cn',
            'url'        => 'http://thinkphp.cn/topic',
            'ip'         => '114.34.54.5',
            'score'      => '89.15',
            'status'     => 1,
            'begin_time' => '2016-3-20',
            'end_time'   => '2016-5-1',
            'info'       => [1, 2, 3, 'name' => 'thinkphp'],
            'zip'        => '200000',
            'date'       => '16-3-8',
            'ok'         => 'yes',
            'value'      => 100,
            'bool'       => true,
            'title'      => '流年ThinkPHP',
            'city'       => '上海',
            'nickname'   => '流年ThinkPHP_2016',
            'aliasname'  => '流年Think2016',
            'file'       => new File(THINK_PATH . 'base.php'),
            'image'      => new File(THINK_PATH . 'logo.png'),
            'test'       => 'test',
        ];
        $validate = new Validate($rule);
        $validate->extend('test', function ($value) {return 'test' == $value ? true : false;});
        $validate->rule('zip', '/^\d{6}$/');
        $validate->rule([
            'ok'   => 'require|accepted',
            'date' => 'date|dateFormat:y-m-d',
        ]);
        $result = $validate->batch()->check($data);
        $this->assertEquals(true, $result);
    }

    public function testMsg()
    {
        $validate = new Validate();
        $validate->message('name.require', '名称必须');
        $validate->message([
            'name.require' => '名称必须',
            'name.max'     => '名称最多不能超过25个字符',
            'age.number'   => '年龄必须是数字',
            'age.between'  => '年龄只能在1-120之间',
            'email'        => '邮箱格式错误',
        ]);
    }

    public function testMake()
    {
        $rule = [
            'name'  => 'require|max:25',
            'age'   => 'number|between:1,120',
            'email' => 'email',
        ];
        $msg = [
            'name.require' => '名称必须',
            'name.max'     => '名称最多不能超过25个字符',
            'age.number'   => '年龄必须是数字',
            'age.between'  => '年龄只能在1-120之间',
            'email'        => '邮箱格式错误',
        ];
        $validate = Validate::make($rule, $msg);
    }

    public function testExtend()
    {
        $validate = new Validate(['name' => 'check:1']);
        $validate->extend('check', function ($value, $rule) {return $rule == $value ? true : false;});
        $validate->extend(['check' => function ($value, $rule) {return $rule == $value ? true : false;}]);
        $data   = ['name' => 1];
        $result = $validate->check($data);
        $this->assertEquals(true, $result);
    }

    public function testScene()
    {
        $rule = [
            'name'  => 'require|max:25',
            'age'   => 'number|between:1,120',
            'email' => 'email',
        ];
        $msg = [
            'name.require' => '名称必须',
            'name.max'     => '名称最多不能超过25个字符',
            'age.number'   => '年龄必须是数字',
            'age.between'  => '年龄只能在1-120之间',
            'email'        => '邮箱格式错误',
        ];
        $data = [
            'name'  => 'thinkphp',
            'age'   => 10,
            'email' => 'thinkphp@qq.com',
        ];
        $validate = new Validate($rule);
        $validate->scene(['edit' => ['name', 'age']]);
        $validate->scene('edit', ['name', 'age']);
        $validate->scene('edit');
        $result = $validate->check($data);
        $this->assertEquals(true, $result);
    }

    public function testSetTypeMsg()
    {
        $rule = [
            'name|名称' => 'require|max:25',
            'age'     => 'number|between:1,120',
            'email'   => 'email',
            ['sex', 'in:1,2', '性别错误'],
        ];
        $data = [
            'name'  => '',
            'age'   => 10,
            'email' => 'thinkphp@qq.com',
            'sex'   => '3',
        ];
        $validate = new Validate($rule);
        $validate->setTypeMsg('require', ':attribute必须');
        $validate->setTypeMsg(['require' => ':attribute必须']);
        $result = $validate->batch()->check($data);
        $this->assertFalse($result);
        $this->assertEquals(['name' => '名称必须', 'sex' => '性别错误'], $validate->getError());
    }

}
