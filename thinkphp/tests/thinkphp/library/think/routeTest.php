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
 * Route测试
 * @author    liu21st <liu21st@gmail.com>
 */

namespace tests\thinkphp\library\think;

use think\Config;
use think\Request;
use think\Route;

class routeTest extends \PHPUnit_Framework_TestCase
{

    protected function setUp()
    {
        Config::set('app_multi_module', true);
    }

    public function testRegister()
    {
        $request = Request::instance();
        Route::get('hello/:name', 'index/hello');
        Route::get(['hello/:name' => 'index/hello']);
        Route::post('hello/:name', 'index/post');
        Route::put('hello/:name', 'index/put');
        Route::delete('hello/:name', 'index/delete');
        Route::patch('hello/:name', 'index/patch');
        Route::any('user/:id', 'index/user');
        $result = Route::check($request, 'hello/thinkphp');
        $this->assertEquals([null, 'index', 'hello'], $result['module']);
        $this->assertEquals(['hello' => true, 'user/:id' => true, 'hello/:name' => ['rule' => 'hello/:name', 'route' => 'index/hello', 'var' => ['name' => 1], 'option' => [], 'pattern' => []]], Route::rules('GET'));
        Route::rule('type1/:name', 'index/type', 'PUT|POST');
        Route::rule(['type2/:name' => 'index/type1']);
        Route::rule([['type3/:name', 'index/type2', ['method' => 'POST']]]);
        Route::rule(['name', 'type4/:name'], 'index/type4');
    }

    public function testImport()
    {
        $rule = [
            '__domain__' => ['subdomain2.thinkphp.cn' => 'blog1'],
            '__alias__'  => ['blog1' => 'blog1'],
            '__rest__'   => ['res' => ['index/blog']],
            'bbb'        => ['index/blog1', ['method' => 'POST']],
            'ddd'        => '',
            ['hello1/:ddd', 'index/hello1', ['method' => 'POST']],
        ];
        Route::import($rule);
    }

    public function testResource()
    {
        $request = Request::instance();
        Route::resource('res', 'index/blog');
        Route::resource(['res' => ['index/blog']]);
        $result = Route::check($request, 'res');
        $this->assertEquals(['index', 'blog', 'index'], $result['module']);
        $result = Route::check($request, 'res/create');
        $this->assertEquals(['index', 'blog', 'create'], $result['module']);
        $result = Route::check($request, 'res/8');
        $this->assertEquals(['index', 'blog', 'read'], $result['module']);
        $result = Route::check($request, 'res/8/edit');
        $this->assertEquals(['index', 'blog', 'edit'], $result['module']);

        Route::resource('blog.comment', 'index/comment');
        $result = Route::check($request, 'blog/8/comment/10');
        $this->assertEquals(['index', 'comment', 'read'], $result['module']);
        $result = Route::check($request, 'blog/8/comment/10/edit');
        $this->assertEquals(['index', 'comment', 'edit'], $result['module']);

    }

    public function testRest()
    {
        $request = Request::instance();
        Route::rest('read', ['GET', '/:id', 'look']);
        Route::rest('create', ['GET', '/create', 'add']);
        Route::rest(['read' => ['GET', '/:id', 'look'], 'create' => ['GET', '/create', 'add']]);
        Route::resource('res', 'index/blog');
        $result = Route::check($request, 'res/create');
        $this->assertEquals(['index', 'blog', 'add'], $result['module']);
        $result = Route::check($request, 'res/8');
        $this->assertEquals(['index', 'blog', 'look'], $result['module']);

    }

    public function testMixVar()
    {
        $request = Request::instance();
        Route::get('hello-<name>', 'index/hello', [], ['name' => '\w+']);
        $result = Route::check($request, 'hello-thinkphp');
        $this->assertEquals([null, 'index', 'hello'], $result['module']);
        Route::get('hello-<name><id?>', 'index/hello', [], ['name' => '\w+', 'id' => '\d+']);
        $result = Route::check($request, 'hello-thinkphp2016');
        $this->assertEquals([null, 'index', 'hello'], $result['module']);
        Route::get('hello-<name>/[:id]', 'index/hello', [], ['name' => '\w+', 'id' => '\d+']);
        $result = Route::check($request, 'hello-thinkphp/2016');
        $this->assertEquals([null, 'index', 'hello'], $result['module']);
    }

    public function testParseUrl()
    {
        $result = Route::parseUrl('hello');
        $this->assertEquals(['hello', null, null], $result['module']);
        $result = Route::parseUrl('index/hello');
        $this->assertEquals(['index', 'hello', null], $result['module']);
        $result = Route::parseUrl('index/hello?name=thinkphp');
        $this->assertEquals(['index', 'hello', null], $result['module']);
        $result = Route::parseUrl('index/user/hello');
        $this->assertEquals(['index', 'user', 'hello'], $result['module']);
        $result = Route::parseUrl('index/user/hello/name/thinkphp');
        $this->assertEquals(['index', 'user', 'hello'], $result['module']);
        $result = Route::parseUrl('index-index-hello', '-');
        $this->assertEquals(['index', 'index', 'hello'], $result['module']);
    }

    public function testCheckRoute()
    {
        Route::get('hello/:name', 'index/hello');
        Route::get('blog/:id', 'blog/read', [], ['id' => '\d+']);
        $request = Request::instance();
        $this->assertEquals(false, Route::check($request, 'test/thinkphp'));
        $this->assertEquals(false, Route::check($request, 'blog/thinkphp'));
        $result = Route::check($request, 'blog/5');
        $this->assertEquals([null, 'blog', 'read'], $result['module']);
        $result = Route::check($request, 'hello/thinkphp/abc/test');
        $this->assertEquals([null, 'index', 'hello'], $result['module']);
    }

    public function testCheckRouteGroup()
    {
        $request = Request::instance();
        Route::pattern(['id' => '\d+']);
        Route::pattern('name', '\w{6,25}');
        Route::group('group', [':id' => 'index/hello', ':name' => 'index/say']);
        $this->assertEquals(false, Route::check($request, 'empty/think'));
        $result = Route::check($request, 'group/think');
        $this->assertEquals(false, $result['module']);
        $result = Route::check($request, 'group/10');
        $this->assertEquals([null, 'index', 'hello'], $result['module']);
        $result = Route::check($request, 'group/thinkphp');
        $this->assertEquals([null, 'index', 'say'], $result['module']);
        Route::group('group2', function () {
            Route::group('group3', [':id' => 'index/hello', ':name' => 'index/say']);
            Route::rule(':name', 'index/hello');
            Route::auto('index');
        });
        $result = Route::check($request, 'group2/thinkphp');
        $this->assertEquals([null, 'index', 'hello'], $result['module']);
        $result = Route::check($request, 'group2/think');
        $this->assertEquals(['index', 'group2', 'think'], $result['module']);
        $result = Route::check($request, 'group2/group3/thinkphp');
        $this->assertEquals([null, 'index', 'say'], $result['module']);
        Route::group('group4', function () {
            Route::group('group3', [':id' => 'index/hello', ':name' => 'index/say']);
            Route::rule(':name', 'index/hello');
            Route::miss('index/__miss__');
        });
        $result = Route::check($request, 'group4/thinkphp');
        $this->assertEquals([null, 'index', 'hello'], $result['module']);
        $result = Route::check($request, 'group4/think');
        $this->assertEquals([null, 'index', '__miss__'], $result['module']);

        Route::group(['prefix' => 'prefix/'], function () {
            Route::rule('hello4/:name', 'hello');
        });
        Route::group(['prefix' => 'prefix/'], [
            'hello4/:name' => 'hello',
        ]);
        $result = Route::check($request, 'hello4/thinkphp');
        $this->assertEquals([null, 'prefix', 'hello'], $result['module']);
        Route::group('group5', [
            [':name', 'hello', ['method' => 'GET|POST']],
            ':id' => 'hello',
        ], ['prefix' => 'prefix/']);
        $result = Route::check($request, 'group5/thinkphp');
        $this->assertEquals([null, 'prefix', 'hello'], $result['module']);
    }

    public function testControllerRoute()
    {
        $request = Request::instance();
        Route::controller('controller', 'index/Blog');
        $result = Route::check($request, 'controller/info');
        $this->assertEquals(['index', 'Blog', 'getinfo'], $result['module']);
        Route::setMethodPrefix('GET', 'read');
        Route::setMethodPrefix(['get' => 'read']);
        Route::controller('controller', 'index/Blog');
        $result = Route::check($request, 'controller/phone');
        $this->assertEquals(['index', 'Blog', 'readphone'], $result['module']);
    }

    public function testAliasRoute()
    {
        $request = Request::instance();
        Route::alias('alias', 'index/Alias');
        $result = Route::check($request, 'alias/info');
        $this->assertEquals('index/Alias/info', $result['module']);
    }

    public function testRouteToModule()
    {
        $request = Request::instance();
        Route::get('hello/:name', 'index/hello');
        Route::get('blog/:id', 'blog/read', [], ['id' => '\d+']);
        $this->assertEquals(false, Route::check($request, 'test/thinkphp'));
        $this->assertEquals(false, Route::check($request, 'blog/thinkphp'));
        $result = Route::check($request, 'hello/thinkphp');
        $this->assertEquals([null, 'index', 'hello'], $result['module']);
        $result = Route::check($request, 'blog/5');
        $this->assertEquals([null, 'blog', 'read'], $result['module']);
    }

    public function testRouteToController()
    {
        $request = Request::instance();
        Route::get('say/:name', '@index/hello');
        $this->assertEquals(['type' => 'controller', 'controller' => 'index/hello', 'var' => []], Route::check($request, 'say/thinkphp'));
    }

    public function testRouteToMethod()
    {
        $request = Request::instance();
        Route::get('user/:name', '\app\index\service\User::get', [], ['name' => '\w+']);
        Route::get('info/:name', '\app\index\model\Info@getInfo', [], ['name' => '\w+']);
        $this->assertEquals(['type' => 'method', 'method' => '\app\index\service\User::get', 'var' => []], Route::check($request, 'user/thinkphp'));
        $this->assertEquals(['type' => 'method', 'method' => ['\app\index\model\Info', 'getInfo'], 'var' => []], Route::check($request, 'info/thinkphp'));
    }

    public function testRouteToRedirect()
    {
        $request = Request::instance();
        Route::get('art/:id', '/article/read/id/:id', [], ['id' => '\d+']);
        $this->assertEquals(['type' => 'redirect', 'url' => '/article/read/id/8', 'status' => 301], Route::check($request, 'art/8'));
    }

    public function testBind()
    {
        $request = Request::instance();
        Route::bind('index/blog');
        Route::get('blog/:id', 'index/blog/read');
        $result = Route::check($request, 'blog/10');
        $this->assertEquals(['index', 'blog', 'read'], $result['module']);
        $result = Route::parseUrl('test');
        $this->assertEquals(['index', 'blog', 'test'], $result['module']);

        Route::bind('\app\index\controller', 'namespace');
        $this->assertEquals(['type' => 'method', 'method' => ['\app\index\controller\Blog', 'read'], 'var' => []], Route::check($request, 'blog/read'));

        Route::bind('\app\index\controller\Blog', 'class');
        $this->assertEquals(['type' => 'method', 'method' => ['\app\index\controller\Blog', 'read'], 'var' => []], Route::check($request, 'read'));
    }

    public function testDomain()
    {
        $request = Request::create('http://subdomain.thinkphp.cn');
        Route::domain('subdomain.thinkphp.cn', 'sub?abc=test&status=1');
        $rules = Route::rules('GET');
        Route::checkDomain($request, $rules);
        $this->assertEquals('sub', Route::getbind('module'));
        $this->assertEquals('test', $_GET['abc']);
        $this->assertEquals(1, $_GET['status']);

        Route::domain('subdomain.thinkphp.cn', '\app\index\controller');
        $rules = Route::rules('GET');
        Route::checkDomain($request, $rules);
        $this->assertEquals('\app\index\controller', Route::getbind('namespace'));

        Route::domain(['subdomain.thinkphp.cn' => '@\app\index\controller\blog']);
        $rules = Route::rules('GET');
        Route::checkDomain($request, $rules);
        $this->assertEquals('\app\index\controller\blog', Route::getbind('class'));

    }
}
