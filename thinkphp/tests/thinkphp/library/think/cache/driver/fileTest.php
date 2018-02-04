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
 * File缓存驱动测试
 * @author    刘志淳 <chun@engineer.com>
 */

namespace tests\thinkphp\library\think\cache\driver;

class fileTest extends cacheTestCase
{
    private $_cacheInstance = null;

    /**
     * 基境缓存类型
     */
    protected function setUp()
    {
        \think\Cache::connect(['type' => 'File', 'path' => CACHE_PATH]);
    }

    /**
     * @return FileCache
     */
    protected function getCacheInstance()
    {
        if (null === $this->_cacheInstance) {
            $this->_cacheInstance = new \think\cache\driver\File();
        }
        return $this->_cacheInstance;
    }

    // skip testExpire
    public function testExpire()
    {
    }
}
