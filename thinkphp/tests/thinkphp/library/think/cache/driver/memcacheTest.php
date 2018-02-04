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
 * Memcache缓存驱动测试
 * @author    刘志淳 <chun@engineer.com>
 */

namespace tests\thinkphp\library\think\cache\driver;

class memcacheTest extends cacheTestCase
{
    private $_cacheInstance = null;

    /**
     * 基境缓存类型
     */
    protected function setUp()
    {
        if (!extension_loaded('memcache')) {
            $this->markTestSkipped("Memcache没有安装，已跳过测试！");
        }
        \think\Cache::connect(['type' => 'memcache', 'expire' => 2]);
    }

    /**
     * @return ApcCache
     */
    protected function getCacheInstance()
    {
        if (null === $this->_cacheInstance) {
            $this->_cacheInstance = new \think\cache\driver\Memcache(['length' => 3]);
        }
        return $this->_cacheInstance;
    }

    // skip testExpire
    public function testExpire()
    {
    }
}
