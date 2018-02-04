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
 * Redis缓存驱动测试
 * @author    7IN0SAN9 <me@7in0.me>
 */

namespace tests\thinkphp\library\think\cache\driver;

class redisTest extends cacheTestCase
{
    private $_cacheInstance = null;

    protected function setUp()
    {
        if (!extension_loaded("redis")) {
            $this->markTestSkipped("Redis没有安装，已跳过测试！");
        }
        \think\Cache::connect(array('type' => 'redis', 'expire' => 2));
    }

    protected function getCacheInstance()
    {
        if (null === $this->_cacheInstance) {
            $this->_cacheInstance = new \think\cache\driver\Redis(['length' => 3]);
        }
        return $this->_cacheInstance;
    }

    public function testGet()
    {
        $cache = $this->prepare();
        $this->assertEquals('string_test', $cache->get('string_test'));
        $this->assertEquals(11, $cache->get('number_test'));
        $result =  $cache->get('array_test');
        $this->assertEquals('array_test', $result['array_test']);
    }

    public function testStoreSpecialValues()
    {
        $redis = new \think\cache\driver\Redis(['length' => 3]);
        $redis->set('key', 'value');
        $redis->get('key');

        $redis->handler()->setnx('key', 'value');
        $value = $redis->handler()->get('key');
        $this->assertEquals('value', $value);
        
        $redis->handler()->hset('hash', 'key', 'value');
        $value = $redis->handler()->hget('hash', 'key');
        $this->assertEquals('value', $value);
    }

    public function testExpire()
    {
    }
}
