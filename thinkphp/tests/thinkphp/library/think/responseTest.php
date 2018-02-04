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
 * Response测试
 * @author    大漠 <zhylninc@gmail.com>
 */

namespace tests\thinkphp\library\think;

use think\Config;
use think\Request;
use think\Response;

class responseTest extends \PHPUnit_Framework_TestCase
{

    /**
     *
     * @var \think\Response
     */
    protected $object;

    protected $default_return_type;

    protected $default_ajax_return;

    /**
     * Sets up the fixture, for example, opens a network connection.
     * This method is called before a test is executed.
     */
    protected function setUp()
    {
        // 1.
        // restore_error_handler();
        // Warning: Cannot modify header information - headers already sent by (output started at PHPUnit\Util\Printer.php:173)
        // more see in https://www.analysisandsolutions.com/blog/html/writing-phpunit-tests-for-wordpress-plugins-wp-redirect-and-continuing-after-php-errors.htm

        // 2.
        // the Symfony used the HeaderMock.php

        // 3.
        // not run the eclipse will held, and travis-ci.org Searching for coverage reports
        // **> Python coverage not found
        // **> No coverage report found.
        // add the
        // /**
        // * @runInSeparateProcess
        // */
        if (!$this->default_return_type) {
            $this->default_return_type = Config::get('default_return_type');
        }
        if (!$this->default_ajax_return) {
            $this->default_ajax_return = Config::get('default_ajax_return');
        }
    }

    /**
     * Tears down the fixture, for example, closes a network connection.
     * This method is called after a test is executed.
     */
    protected function tearDown()
    {
        Config::set('default_ajax_return', $this->default_ajax_return);
        Config::set('default_return_type', $this->default_return_type);
    }

    /**
     * @covers think\Response::send
     * @todo Implement testSend().
     */
    public function testSend()
    {
        $dataArr        = [];
        $dataArr["key"] = "value";

        $response = Response::create($dataArr, 'json');
        $result   = $response->getContent();
        $this->assertEquals('{"key":"value"}', $result);
        $request = Request::instance();
        $request->get(['callback' => 'callback']);
        $response = Response::create($dataArr, 'jsonp');
        $result   = $response->getContent();
        $this->assertEquals('callback({"key":"value"});', $result);
    }

}
