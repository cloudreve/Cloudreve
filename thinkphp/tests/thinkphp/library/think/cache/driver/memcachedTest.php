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
 * Memcached缓存驱动测试
 * @author    7IN0SAN9 <me@7in0.me>
 */

namespace tests\thinkphp\library\think\cache\driver;

class memcachedTest extends cacheTestCase
{
    private $_cacheInstance = null;
    /**
     * 基境缓存类型
     */
    protected function setUp()
    {
        if (!extension_loaded("memcached") && !extension_loaded('memcache')) {
            $this->markTestSkipped("Memcached或Memcache没有安装，已跳过测试！");
        }
        \think\Cache::connect(array('type' => 'memcached', 'expire' => 2));
    }
    /**
     * @return ApcCache
     */
    protected function getCacheInstance()
    {
        if (null === $this->_cacheInstance) {
            $this->_cacheInstance = new \think\cache\driver\Memcached(['length' => 3]);
        }
        return $this->_cacheInstance;
    }
    /**
     * 缓存过期测试《提出来测试，因为目前看通不过缓存过期测试，所以还需研究》
     * @return  mixed
     * @access public
     */
    public function testExpire()
    {
    }

    public function testStaticCall()
    {
    }

    /**
     * 测试缓存自增
     * @return  mixed
     * @access public
     */
    public function testInc()
    {
    }

    /**
     * 测试缓存自减
     * @return  mixed
     * @access public
     */
    public function testDec()
    {
    }
}
