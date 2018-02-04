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
 * Lite缓存驱动测试
 * @author    刘志淳 <chun@engineer.com>
 */

namespace tests\thinkphp\library\think\cache\driver;

use think\Cache;

class liteTest extends \PHPUnit_Framework_TestCase
{
    protected function getCacheInstance()
    {
        return Cache::connect(['type' => 'Lite', 'path' => CACHE_PATH]);
    }

    /**
     * 测试缓存读取
     * @return  mixed
     * @access public
     */
    public function testGet()
    {
        $cache = $this->getCacheInstance();
        $this->assertFalse($cache->get('test'));
    }

    /**
     * 测试缓存设置
     * @return  mixed
     * @access public
     */
    public function testSet()
    {
        $cache = $this->getCacheInstance();
        $this->assertNotEmpty($cache->set('test', 'test'));
    }

    /**
     * 删除缓存测试
     * @return  mixed
     * @access public
     */
    public function testRm()
    {
        $cache = $this->getCacheInstance();
        $this->assertTrue($cache->rm('test'));
    }

    /**
     * 清空缓存测试
     * @return  mixed
     * @access public
     */
    public function testClear()
    {
    }
}
