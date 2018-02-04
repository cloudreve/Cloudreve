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
 * Debug测试
 * @author    大漠 <zhylninc@gmail.com>
 */

namespace tests\thinkphp\library\think;

use think\Debug;

class debugTest extends \PHPUnit_Framework_TestCase
{

    /**
     *
     * @var Debug
     */
    protected $object;

    /**
     * Sets up the fixture, for example, opens a network connection.
     * This method is called before a test is executed.
     */
    protected function setUp()
    {
        $this->object = new Debug();
    }

    /**
     * Tears down the fixture, for example, closes a network connection.
     * This method is called after a test is executed.
     */
    protected function tearDown()
    {}

    /**
     * @covers think\Debug::remark
     * @todo Implement testRemark().
     */
    public function testRemark()
    {
        $name = "testremarkkey";
        Debug::remark($name);
    }

    /**
     * @covers think\Debug::getRangeTime
     * @todo Implement testGetRangeTime().
     */
    public function testGetRangeTime()
    {
        $start = "testGetRangeTimeStart";
        $end   = "testGetRangeTimeEnd";
        Debug::remark($start);
        usleep(20000);
        // \think\Debug::remark($end);

        $time = Debug::getRangeTime($start, $end);
        $this->assertLessThan(0.03, $time);
        //$this->assertEquals(0.03, ceil($time));
    }

    /**
     * @covers think\Debug::getUseTime
     * @todo Implement testGetUseTime().
     */
    public function testGetUseTime()
    {
        $time = Debug::getUseTime();
        $this->assertLessThan(20, $time);
    }

    /**
     * @covers think\Debug::getThroughputRate
     * @todo Implement testGetThroughputRate().
     */
    public function testGetThroughputRate()
    {
        usleep(100000);
        $throughputRate = Debug::getThroughputRate();
        $this->assertLessThan(10, $throughputRate);
    }

    /**
     * @covers think\Debug::getRangeMem
     * @todo Implement testGetRangeMem().
     */
    public function testGetRangeMem()
    {
        $start = "testGetRangeMemStart";
        $end   = "testGetRangeMemEnd";
        Debug::remark($start);
        $str = "";
        for ($i = 0; $i < 10000; $i++) {
            $str .= "mem";
        }

        $rangeMem = Debug::getRangeMem($start, $end);

        $this->assertLessThan(33, explode(" ", $rangeMem)[0]);
    }

    /**
     * @covers think\Debug::getUseMem
     * @todo Implement testGetUseMem().
     */
    public function testGetUseMem()
    {
        $useMem = Debug::getUseMem();

        $this->assertLessThan(35, explode(" ", $useMem)[0]);
    }

    /**
     * @covers think\Debug::getMemPeak
     * @todo Implement testGetMemPeak().
     */
    public function testGetMemPeak()
    {
        $start = "testGetMemPeakStart";
        $end   = "testGetMemPeakEnd";
        Debug::remark($start);
        $str = "";
        for ($i = 0; $i < 100000; $i++) {
            $str .= "mem";
        }
        $memPeak = Debug::getMemPeak($start, $end);
        $this->assertLessThan(500, explode(" ", $memPeak)[0]);
    }

    /**
     * @covers think\Debug::getFile
     * @todo Implement testGetFile().
     */
    public function testGetFile()
    {
        $count = Debug::getFile();

        $this->assertEquals(count(get_included_files()), $count);

        $info = Debug::getFile(true);
        $this->assertEquals(count(get_included_files()), count($info));

        $this->assertContains("KB", $info[0]);
    }

    /**
     * @covers think\Debug::dump
     * @todo Implement testDump().
     */
    public function testDump()
    {
        if (strstr(PHP_VERSION, 'hhvm')) {
            return;
        }

        $var        = [];
        $var["key"] = "val";
        $output     = Debug::dump($var, false, $label = "label");
        $array      = explode("array", json_encode($output));
        if (IS_WIN) {
            $this->assertEquals("(1) {\\n  [\\\"key\\\"] => string(3) \\\"val\\\"\\n}\\n\\r\\n\"", end($array));
        } elseif (strstr(PHP_OS, 'Darwin')) {
            $this->assertEquals("(1) {\\n  [\\\"key\\\"] => string(3) \\\"val\\\"\\n}\\n\\n\"", end($array));
        } else {
            $this->assertEquals("(1) {\\n  'key' =>\\n  string(3) \\\"val\\\"\\n}\\n\\n\"", end($array));
        }
    }
}
