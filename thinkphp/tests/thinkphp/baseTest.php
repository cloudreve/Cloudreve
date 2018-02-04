<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: Haotong Lin <lofanmi@gmail.com>
// +----------------------------------------------------------------------

/**
 * 保证运行环境正常
 */
class baseTest extends \PHPUnit_Framework_TestCase
{
    public function testConstants()
    {
        $this->assertNotEmpty(THINK_START_TIME);
        $this->assertNotEmpty(THINK_START_MEM);
        $this->assertNotEmpty(THINK_VERSION);
        $this->assertNotEmpty(DS);
        $this->assertNotEmpty(THINK_PATH);
        $this->assertNotEmpty(LIB_PATH);
        $this->assertNotEmpty(EXTEND_PATH);
        $this->assertNotEmpty(CORE_PATH);
        $this->assertNotEmpty(TRAIT_PATH);
        $this->assertNotEmpty(APP_PATH);
        $this->assertNotEmpty(RUNTIME_PATH);
        $this->assertNotEmpty(LOG_PATH);
        $this->assertNotEmpty(CACHE_PATH);
        $this->assertNotEmpty(TEMP_PATH);
        $this->assertNotEmpty(VENDOR_PATH);
        $this->assertNotEmpty(EXT);
        $this->assertNotEmpty(ENV_PREFIX);
        $this->assertTrue(!is_null(IS_WIN));
        $this->assertTrue(!is_null(IS_CLI));
    }
}
