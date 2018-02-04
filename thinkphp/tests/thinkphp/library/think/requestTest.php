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
 * Db类测试
 */

namespace tests\thinkphp\library\think;

use think\Config;
use think\Request;

class requestTest extends \PHPUnit_Framework_TestCase
{
    protected $request;

    public function setUp()
    {
        //$request = Request::create('http://www.domain.com/index/index/hello/?name=thinkphp');

    }

    public function testCreate()
    {
        $request = Request::create('http://www.thinkphp.cn/index/index/hello.html?name=thinkphp');
        $this->assertEquals('http://www.thinkphp.cn', $request->domain());
        $this->assertEquals('/index/index/hello.html?name=thinkphp', $request->url());
        $this->assertEquals('/index/index/hello.html', $request->baseurl());
        $this->assertEquals('index/index/hello.html', $request->pathinfo());
        $this->assertEquals('index/index/hello', $request->path());
        $this->assertEquals('html', $request->ext());
        $this->assertEquals('name=thinkphp', $request->query());
        $this->assertEquals('www.thinkphp.cn', $request->host());
        $this->assertEquals(80, $request->port());
        $this->assertEquals($_SERVER['REQUEST_TIME'], $request->time());
        $this->assertEquals($_SERVER['REQUEST_TIME_FLOAT'], $request->time(true));
        $this->assertEquals('GET', $request->method());
        $this->assertEquals(['name' => 'thinkphp'], $request->param());
        $this->assertFalse($request->isSsl());
        $this->assertEquals('http', $request->scheme());
    }

    public function testDomain()
    {
        $request = Request::instance();
        $request->domain('http://thinkphp.cn');
        $this->assertEquals('http://thinkphp.cn', $request->domain());
    }

    public function testUrl()
    {
        $request = Request::instance();
        $request->url('/index.php/index/hello?name=thinkphp');
        $this->assertEquals('/index.php/index/hello?name=thinkphp', $request->url());
        $this->assertEquals('http://thinkphp.cn/index.php/index/hello?name=thinkphp', $request->url(true));
    }

    public function testBaseUrl()
    {
        $request = Request::instance();
        $request->baseurl('/index.php/index/hello');
        $this->assertEquals('/index.php/index/hello', $request->baseurl());
        $this->assertEquals('http://thinkphp.cn/index.php/index/hello', $request->baseurl(true));
    }

    public function testbaseFile()
    {
        $request = Request::instance();
        $request->basefile('/index.php');
        $this->assertEquals('/index.php', $request->basefile());
        $this->assertEquals('http://thinkphp.cn/index.php', $request->basefile(true));
    }

    public function testroot()
    {
        $request = Request::instance();
        $request->root('/index.php');
        $this->assertEquals('/index.php', $request->root());
        $this->assertEquals('http://thinkphp.cn/index.php', $request->root(true));
    }

    public function testType()
    {
        $request = Request::instance();
        $request->server(['HTTP_ACCEPT' => 'application/json']);

        $this->assertEquals('json', $request->type());
        $request->mimeType('test', 'application/test');
        $request->mimeType(['test' => 'application/test']);
        $request->server(['HTTP_ACCEPT' => 'application/test']);

        $this->assertEquals('test', $request->type());
    }

    public function testmethod()
    {
        $_SERVER['HTTP_X_HTTP_METHOD_OVERRIDE'] = 'DELETE';

        $request = Request::create('', '');
        $this->assertEquals('DELETE', $request->method());
        $this->assertEquals('GET', $request->method(true));

        Config::set('var_method', '_method');
        $_POST['_method'] = 'POST';
        $request          = Request::create('', '');
        $this->assertEquals('POST', $request->method());
        $this->assertEquals('GET', $request->method(true));
        $this->assertTrue($request->isPost());
        $this->assertFalse($request->isGet());
        $this->assertFalse($request->isPut());
        $this->assertFalse($request->isDelete());
        $this->assertFalse($request->isHead());
        $this->assertFalse($request->isPatch());
        $this->assertFalse($request->isOptions());
    }

    public function testCli()
    {
        $request = Request::instance();
        $this->assertTrue($request->isCli());
    }

    public function testVar()
    {
        Config::set('app_multi_module', true);
        $request = Request::create('');
        $request->route(['name' => 'thinkphp', 'id' => 6]);
        $request->get(['id' => 10]);
        $request->post(['id' => 8]);
        $request->put(['id' => 7]);
        $request->request(['test' => 'value']);
        $this->assertEquals(['name' => 'thinkphp', 'id' => 6], $request->route());
        //$this->assertEquals(['id' => 10], $request->get());
        $this->assertEquals('thinkphp', $request->route('name'));
        $this->assertEquals('default', $request->route('test', 'default'));
        $this->assertEquals(10, $request->get('id'));
        $this->assertEquals(0, $request->get('ids', 0));
        $this->assertEquals(8, $request->post('id'));
        $this->assertEquals(7, $request->put('id'));
        $this->assertEquals('value', $request->request('test'));
        $this->assertEquals('thinkphp', $request->param('name'));
        $this->assertEquals(6, $request->param('id'));
        $this->assertFalse($request->has('user_id'));
        $this->assertTrue($request->has('test', 'request'));
        $this->assertEquals(['id' => 6], $request->only('id'));
        $this->assertEquals(['name' => 'thinkphp', 'lang' => 'zh-cn'], $request->except('id'));
        $this->assertEquals('THINKPHP', $request->param('name', '', 'strtoupper'));
    }

    public function testIsAjax()
    {
        $request                          = Request::create('');
        $_SERVER['HTTP_X_REQUESTED_WITH'] = 'xmlhttprequest';
        $this->assertTrue($request->isAjax());
    }

    public function testIsPjax()
    {
        $request                = Request::create('');
        $_SERVER['HTTP_X_PJAX'] = true;
        $this->assertTrue($request->isPjax());
    }

    public function testIsMobile()
    {
        $request             = Request::create('');
        $_SERVER['HTTP_VIA'] = 'wap';
        $this->assertTrue($request->isMobile());
    }

    public function testBind()
    {
        $request       = Request::create('');
        $request->user = 'User1';
        $request->bind(['user' => 'User2']);
        $this->assertEquals('User2', $request->user);
    }

}
