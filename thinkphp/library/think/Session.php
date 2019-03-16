<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2018 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------

namespace think;

use think\exception\ClassNotFoundException;

class Session
{
    protected static $prefix = '';
    protected static $init   = null;

    /**
     * 设置或者获取session作用域（前缀）
     * @param string $prefix
     * @return string|void
     */
    public static function prefix($prefix = '')
    {
        empty(self::$init) && self::boot();
        if (empty($prefix) && null !== $prefix) {
            return self::$prefix;
        } else {
            self::$prefix = $prefix;
        }
    }

    /**
     * session初始化
     * @param array $config
     * @return void
     * @throws \think\Exception
     */
    public static function init(array $config = [])
    {
        if (empty($config)) {
            $config = Config::get('session');
        }
        // 记录初始化信息
        App::$debug && Log::record('[ SESSION ] INIT ' . var_export($config, true), 'info');
        $isDoStart = false;
        if (isset($config['use_trans_sid'])) {
            ini_set('session.use_trans_sid', $config['use_trans_sid'] ? 1 : 0);
        }

        // 启动session
        if (!empty($config['auto_start']) && PHP_SESSION_ACTIVE != session_status()) {
            ini_set('session.auto_start', 0);
            $isDoStart = true;
        }

        if (isset($config['prefix']) && ('' === self::$prefix || null === self::$prefix)) {
            self::$prefix = $config['prefix'];
        }
        if (isset($config['var_session_id']) && isset($_REQUEST[$config['var_session_id']])) {
            session_id($_REQUEST[$config['var_session_id']]);
        } elseif (isset($config['id']) && !empty($config['id'])) {
            session_id($config['id']);
        }
        if (isset($config['name'])) {
            session_name($config['name']);
        }
        if (isset($config['path'])) {
            session_save_path($config['path']);
        }
        if (isset($config['domain'])) {
            ini_set('session.cookie_domain', $config['domain']);
        }
        if (isset($config['expire'])) {
            ini_set('session.gc_maxlifetime', $config['expire']);
            ini_set('session.cookie_lifetime', $config['expire']);
        }
        if (isset($config['secure'])) {
            ini_set('session.cookie_secure', $config['secure']);
        }
        if (isset($config['httponly'])) {
            ini_set('session.cookie_httponly', $config['httponly']);
        }
        if (isset($config['use_cookies'])) {
            ini_set('session.use_cookies', $config['use_cookies'] ? 1 : 0);
        }
        if (isset($config['cache_limiter'])) {
            session_cache_limiter($config['cache_limiter']);
        }
        if (isset($config['cache_expire'])) {
            session_cache_expire($config['cache_expire']);
        }
        if (!empty($config['type'])) {
            // 读取session驱动
            $class = false !== strpos($config['type'], '\\') ? $config['type'] : '\\think\\session\\driver\\' . ucwords($config['type']);

            // 检查驱动类
            if (!class_exists($class) || !session_set_save_handler(new $class($config))) {
                throw new ClassNotFoundException('error session handler:' . $class, $class);
            }
        }
        if ($isDoStart) {
            session_start();
            self::$init = true;
        } else {
            self::$init = false;
        }
    }

    /**
     * session自动启动或者初始化
     * @return void
     */
    public static function boot()
    {
        if (is_null(self::$init)) {
            self::init();
        } elseif (false === self::$init) {
            if (PHP_SESSION_ACTIVE != session_status()) {
                session_start();
            }
            self::$init = true;
        }
    }

    /**
     * session设置
     * @param string        $name session名称
     * @param mixed         $value session值
     * @param string|null   $prefix 作用域（前缀）
     * @return void
     */
    public static function set($name, $value = '', $prefix = null)
    {
        empty(self::$init) && self::boot();

        $prefix = !is_null($prefix) ? $prefix : self::$prefix;
        if (strpos($name, '.')) {
            // 二维数组赋值
            list($name1, $name2) = explode('.', $name);
            if ($prefix) {
                $_SESSION[$prefix][$name1][$name2] = $value;
            } else {
                $_SESSION[$name1][$name2] = $value;
            }
        } elseif ($prefix) {
            $_SESSION[$prefix][$name] = $value;
        } else {
            $_SESSION[$name] = $value;
        }
    }

    /**
     * session获取
     * @param string        $name session名称
     * @param string|null   $prefix 作用域（前缀）
     * @return mixed
     */
    public static function get($name = '', $prefix = null)
    {
        empty(self::$init) && self::boot();
        $prefix = !is_null($prefix) ? $prefix : self::$prefix;
        if ('' == $name) {
            // 获取全部的session
            $value = $prefix ? (!empty($_SESSION[$prefix]) ? $_SESSION[$prefix] : []) : $_SESSION;
        } elseif ($prefix) {
            // 获取session
            if (strpos($name, '.')) {
                list($name1, $name2) = explode('.', $name);
                $value               = isset($_SESSION[$prefix][$name1][$name2]) ? $_SESSION[$prefix][$name1][$name2] : null;
            } else {
                $value = isset($_SESSION[$prefix][$name]) ? $_SESSION[$prefix][$name] : null;
            }
        } else {
            if (strpos($name, '.')) {
                list($name1, $name2) = explode('.', $name);
                $value               = isset($_SESSION[$name1][$name2]) ? $_SESSION[$name1][$name2] : null;
            } else {
                $value = isset($_SESSION[$name]) ? $_SESSION[$name] : null;
            }
        }
        return $value;
    }

    /**
     * session获取并删除
     * @param string        $name session名称
     * @param string|null   $prefix 作用域（前缀）
     * @return mixed
     */
    public static function pull($name, $prefix = null)
    {
        $result = self::get($name, $prefix);
        if ($result) {
            self::delete($name, $prefix);
            return $result;
        } else {
            return;
        }
    }

    /**
     * session设置 下一次请求有效
     * @param string        $name session名称
     * @param mixed         $value session值
     * @param string|null   $prefix 作用域（前缀）
     * @return void
     */
    public static function flash($name, $value)
    {
        self::set($name, $value);
        if (!self::has('__flash__.__time__')) {
            self::set('__flash__.__time__', $_SERVER['REQUEST_TIME_FLOAT']);
        }
        self::push('__flash__', $name);
    }

    /**
     * 清空当前请求的session数据
     * @return void
     */
    public static function flush()
    {
        if (self::$init) {
            $item = self::get('__flash__');

            if (!empty($item)) {
                $time = $item['__time__'];
                if ($_SERVER['REQUEST_TIME_FLOAT'] > $time) {
                    unset($item['__time__']);
                    self::delete($item);
                    self::set('__flash__', []);
                }
            }
        }
    }

    /**
     * 删除session数据
     * @param string|array  $name session名称
     * @param string|null   $prefix 作用域（前缀）
     * @return void
     */
    public static function delete($name, $prefix = null)
    {
        empty(self::$init) && self::boot();
        $prefix = !is_null($prefix) ? $prefix : self::$prefix;
        if (is_array($name)) {
            foreach ($name as $key) {
                self::delete($key, $prefix);
            }
        } elseif (strpos($name, '.')) {
            list($name1, $name2) = explode('.', $name);
            if ($prefix) {
                unset($_SESSION[$prefix][$name1][$name2]);
            } else {
                unset($_SESSION[$name1][$name2]);
            }
        } else {
            if ($prefix) {
                unset($_SESSION[$prefix][$name]);
            } else {
                unset($_SESSION[$name]);
            }
        }
    }

    /**
     * 清空session数据
     * @param string|null   $prefix 作用域（前缀）
     * @return void
     */
    public static function clear($prefix = null)
    {
        empty(self::$init) && self::boot();
        $prefix = !is_null($prefix) ? $prefix : self::$prefix;
        if ($prefix) {
            unset($_SESSION[$prefix]);
        } else {
            $_SESSION = [];
        }
    }

    /**
     * 判断session数据
     * @param string        $name session名称
     * @param string|null   $prefix
     * @return bool
     */
    public static function has($name, $prefix = null)
    {
        empty(self::$init) && self::boot();
        $prefix = !is_null($prefix) ? $prefix : self::$prefix;
        if (strpos($name, '.')) {
            // 支持数组
            list($name1, $name2) = explode('.', $name);
            return $prefix ? isset($_SESSION[$prefix][$name1][$name2]) : isset($_SESSION[$name1][$name2]);
        } else {
            return $prefix ? isset($_SESSION[$prefix][$name]) : isset($_SESSION[$name]);
        }
    }

    /**
     * 添加数据到一个session数组
     * @param  string  $key
     * @param  mixed   $value
     * @return void
     */
    public static function push($key, $value)
    {
        $array = self::get($key);
        if (is_null($array)) {
            $array = [];
        }
        $array[] = $value;
        self::set($key, $array);
    }

    /**
     * 启动session
     * @return void
     */
    public static function start()
    {
        session_start();
        self::$init = true;
    }

    /**
     * 销毁session
     * @return void
     */
    public static function destroy()
    {
        if (!empty($_SESSION)) {
            $_SESSION = [];
        }
        session_unset();
        session_destroy();
        self::$init = null;
    }

    /**
     * 重新生成session_id
     * @param bool $delete 是否删除关联会话文件
     * @return void
     */
    public static function regenerate($delete = false)
    {
        session_regenerate_id($delete);
    }

    /**
     * 暂停session
     * @return void
     */
    public static function pause()
    {
        // 暂停session
        session_write_close();
        self::$init = false;
    }
}
