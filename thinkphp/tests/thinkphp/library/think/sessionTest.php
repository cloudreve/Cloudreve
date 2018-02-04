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
 * Session测试
 * @author    大漠 <zhylninc@gmail.com>
 */

namespace tests\thinkphp\library\think;

use think\Session;

class sessionTest extends \PHPUnit_Framework_TestCase
{

    /**
     *
     * @var \think\Session
     */
    protected $object;

    /**
     * Sets up the fixture, for example, opens a network connection.
     * This method is called before a test is executed.
     */
    protected function setUp()
    {
        // $this->object = new Session ();
        // register_shutdown_function ( function () {
        // } ); // 此功能无法取消，需要回调函数配合。
        set_exception_handler(function () {});
        set_error_handler(function () {});
    }

    /**
     * Tears down the fixture, for example, closes a network connection.
     * This method is called after a test is executed.
     */
    protected function tearDown()
    {
        register_shutdown_function('think\Error::appShutdown');
        set_error_handler('think\Error::appError');
        set_exception_handler('think\Error::appException');
    }

    /**
     * @covers think\Session::prefix
     *
     * @todo Implement testPrefix().
     */
    public function testPrefix()
    {
        Session::prefix(null);
        Session::prefix('think_');

        $this->assertEquals('think_', Session::prefix());
    }

    /**
     * @covers think\Session::init
     *
     * @todo Implement testInit().
     */
    public function testInit()
    {
        Session::prefix(null);
        $config = [
            // cookie 名称前缀
            'prefix'         => 'think_',
            // cookie 保存时间
            'expire'         => 60,
            // cookie 保存路径
            'path'           => '/path/to/test/session/',
            // cookie 有效域名
            'domain'         => '.thinkphp.cn',
            'var_session_id' => 'sessionidtest',
            'id'             => 'sess_8fhgkjuakhatbeg2fa14lo84q1',
            'name'           => 'session_name',
            'use_trans_sid'  => '1',
            'use_cookies'    => '1',
            'cache_limiter'  => '60',
            'cache_expire'   => '60',
            'type'           => '', // memcache
            'namespace'      => '\\think\\session\\driver\\', // ?
            'auto_start'     => '1',
        ];

        $_REQUEST[$config['var_session_id']] = $config['id'];
        Session::init($config);

        // 开始断言
        $this->assertEquals($config['prefix'], Session::prefix());
        $this->assertEquals($config['id'], $_REQUEST[$config['var_session_id']]);
        $this->assertEquals($config['name'], session_name());

        $this->assertEquals($config['path'], session_save_path());
        $this->assertEquals($config['use_cookies'], ini_get('session.use_cookies'));
        $this->assertEquals($config['domain'], ini_get('session.cookie_domain'));
        $this->assertEquals($config['expire'], ini_get('session.gc_maxlifetime'));
        $this->assertEquals($config['expire'], ini_get('session.cookie_lifetime'));

        $this->assertEquals($config['cache_limiter'], session_cache_limiter($config['cache_limiter']));
        $this->assertEquals($config['cache_expire'], session_cache_expire($config['cache_expire']));

        // 检测分支
        $_REQUEST[$config['var_session_id']] = null;
        session_write_close();
        session_destroy();

        Session::init($config);

        // 测试auto_start
        // PHP_SESSION_DISABLED
        // PHP_SESSION_NONE
        // PHP_SESSION_ACTIVE
        // session_status()
        if (strstr(PHP_VERSION, 'hhvm')) {
            $this->assertEquals('', ini_get('session.auto_start'));
        } else {
            $this->assertEquals(0, ini_get('session.auto_start'));
        }

        $this->assertEquals($config['use_trans_sid'], ini_get('session.use_trans_sid'));

        Session::init($config);
        $this->assertEquals($config['id'], session_id());
    }

    /**
     * 单独重现异常
     * @expectedException \think\Exception
     */
    public function testException()
    {
        $config = [
            // cookie 名称前缀
            'prefix'         => 'think_',
            // cookie 保存时间
            'expire'         => 0,
            // cookie 保存路径
            'path'           => '/path/to/test/session/',
            // cookie 有效域名
            'domain'         => '.thinkphp.cn',
            'var_session_id' => 'sessionidtest',
            'id'             => 'sess_8fhgkjuakhatbeg2fa14lo84q1',
            'name'           => 'session_name',
            'use_trans_sid'  => '1',
            'use_cookies'    => '1',
            'cache_limiter'  => '60',
            'cache_expire'   => '60',
            'type'           => '\\think\\session\\driver\\Memcache', //
            'auto_start'     => '1',
        ];

        // 测试session驱动是否存在
        // @expectedException 异常类名
        $this->setExpectedException('\think\exception\ClassNotFoundException', 'error session handler');

        Session::init($config);
    }

    /**
     * @covers think\Session::set
     *
     * @todo Implement testSet().
     */
    public function testSet()
    {
        Session::prefix(null);
        Session::set('sessionname', 'sessionvalue');
        $this->assertEquals('sessionvalue', $_SESSION['sessionname']);

        Session::set('sessionnamearr.subname', 'sessionvalue');
        $this->assertEquals('sessionvalue', $_SESSION['sessionnamearr']['subname']);

        Session::set('sessionnameper', 'sessionvalue', 'think_');
        $this->assertEquals('sessionvalue', $_SESSION['think_']['sessionnameper']);

        Session::set('sessionnamearrper.subname', 'sessionvalue', 'think_');
        $this->assertEquals('sessionvalue', $_SESSION['think_']['sessionnamearrper']['subname']);
    }

    /**
     * @covers think\Session::get
     *
     * @todo Implement testGet().
     */
    public function testGet()
    {
        Session::prefix(null);

        Session::set('sessionnameget', 'sessionvalue');
        $this->assertEquals(Session::get('sessionnameget'), $_SESSION['sessionnameget']);

        Session::set('sessionnamegetarr.subname', 'sessionvalue');
        $this->assertEquals(Session::get('sessionnamegetarr.subname'), $_SESSION['sessionnamegetarr']['subname']);

        Session::set('sessionnamegetarrperall', 'sessionvalue', 'think_');
        $this->assertEquals(Session::get('', 'think_')['sessionnamegetarrperall'], $_SESSION['think_']['sessionnamegetarrperall']);

        Session::set('sessionnamegetper', 'sessionvalue', 'think_');
        $this->assertEquals(Session::get('sessionnamegetper', 'think_'), $_SESSION['think_']['sessionnamegetper']);

        Session::set('sessionnamegetarrper.subname', 'sessionvalue', 'think_');
        $this->assertEquals(Session::get('sessionnamegetarrper.subname', 'think_'), $_SESSION['think_']['sessionnamegetarrper']['subname']);
    }

    public function testPull()
    {
        Session::prefix(null);
        Session::set('sessionnamedel', 'sessionvalue');
        $this->assertEquals('sessionvalue', Session::pull('sessionnameget'));
        $this->assertNull(Session::get('sessionnameget'));
    }

    /**
     * @covers think\Session::delete
     *
     * @todo Implement testDelete().
     */
    public function testDelete()
    {
        Session::prefix(null);
        Session::set('sessionnamedel', 'sessionvalue');
        Session::delete('sessionnamedel');
        $this->assertEmpty($_SESSION['sessionnamedel']);

        Session::set('sessionnamedelarr.subname', 'sessionvalue');
        Session::delete('sessionnamedelarr.subname');
        $this->assertEmpty($_SESSION['sessionnamedelarr']['subname']);

        Session::set('sessionnamedelper', 'sessionvalue', 'think_');
        Session::delete('sessionnamedelper', 'think_');
        $this->assertEmpty($_SESSION['think_']['sessionnamedelper']);

        Session::set('sessionnamedelperarr.subname', 'sessionvalue', 'think_');
        Session::delete('sessionnamedelperarr.subname', 'think_');
        $this->assertEmpty($_SESSION['think_']['sessionnamedelperarr']['subname']);
    }

    /**
     * @covers think\Session::clear
     *
     * @todo Implement testClear().
     */
    public function testClear()
    {
        Session::prefix(null);

        Session::set('sessionnameclsper', 'sessionvalue1', 'think_');
        Session::clear('think_');
        $this->assertNull($_SESSION['think_']);

        Session::set('sessionnameclsper', 'sessionvalue1', 'think_');
        Session::clear();
        $this->assertEmpty($_SESSION);
    }

    /**
     * @covers think\Session::has
     *
     * @todo Implement testHas().
     */
    public function testHas()
    {
        Session::prefix(null);
        Session::set('sessionnamehas', 'sessionvalue');
        $this->assertTrue(Session::has('sessionnamehas'));

        Session::set('sessionnamehasarr.subname', 'sessionvalue');
        $this->assertTrue(Session::has('sessionnamehasarr.subname'));

        Session::set('sessionnamehasper', 'sessionvalue', 'think_');
        $this->assertTrue(Session::has('sessionnamehasper', 'think_'));

        Session::set('sessionnamehasarrper.subname', 'sessionvalue', 'think_');
        $this->assertTrue(Session::has('sessionnamehasarrper.subname', 'think_'));
    }

    /**
     * @covers think\Session::pause
     *
     * @todo Implement testPause().
     */
    public function testPause()
    {
        Session::pause();
    }

    /**
     * @covers think\Session::start
     *
     * @todo Implement testStart().
     */
    public function testStart()
    {
        Session::start();
    }

    /**
     * @covers think\Session::destroy
     *
     * @todo Implement testDestroy().
     */
    public function testDestroy()
    {
        Session::set('sessionnamedestroy', 'sessionvalue');
        Session::destroy();
        $this->assertEmpty($_SESSION['sessionnamedestroy']);
    }
}
