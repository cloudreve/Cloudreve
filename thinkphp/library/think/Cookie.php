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

class Cookie
{
    /**
     * @var array cookie 设置参数
     */
    protected static $config = [
        'prefix'    => '', // cookie 名称前缀
        'expire'    => 0, // cookie 保存时间
        'path'      => '/', // cookie 保存路径
        'domain'    => '', // cookie 有效域名
        'secure'    => false, //  cookie 启用安全传输
        'httponly'  => false, // httponly 设置
        'setcookie' => true, // 是否使用 setcookie
    ];

    /**
     * @var bool 是否完成初始化了
     */
    protected static $init;

    /**
     * Cookie初始化
     * @access public
     * @param  array $config 配置参数
     * @return void
     */
    public static function init(array $config = [])
    {
        if (empty($config)) {
            $config = Config::get('cookie');
        }

        self::$config = array_merge(self::$config, array_change_key_case($config));

        if (!empty(self::$config['httponly'])) {
            ini_set('session.cookie_httponly', 1);
        }

        self::$init = true;
    }

    /**
     * 设置或者获取 cookie 作用域（前缀）
     * @access public
     * @param  string $prefix 前缀
     * @return string|
     */
    public static function prefix($prefix = '')
    {
        if (empty($prefix)) {
            return self::$config['prefix'];
        }

        return self::$config['prefix'] = $prefix;
    }

    /**
     * Cookie 设置、获取、删除
     * @access public
     * @param  string $name   cookie 名称
     * @param  mixed  $value  cookie 值
     * @param  mixed  $option 可选参数 可能会是 null|integer|string
     * @return void
     */
    public static function set($name, $value = '', $option = null)
    {
        !isset(self::$init) && self::init();

        // 参数设置(会覆盖黙认设置)
        if (!is_null($option)) {
            if (is_numeric($option)) {
                $option = ['expire' => $option];
            } elseif (is_string($option)) {
                parse_str($option, $option);
            }

            $config = array_merge(self::$config, array_change_key_case($option));
        } else {
            $config = self::$config;
        }

        $name = $config['prefix'] . $name;

        // 设置 cookie
        if (is_array($value)) {
            array_walk_recursive($value, 'self::jsonFormatProtect', 'encode');
            $value = 'think:' . json_encode($value);
        }

        $expire = !empty($config['expire']) ?
        $_SERVER['REQUEST_TIME'] + intval($config['expire']) :
        0;

        if ($config['setcookie']) {
            setcookie(
                $name, $value, $expire, $config['path'], $config['domain'],
                $config['secure'], $config['httponly']
            );
        }

        $_COOKIE[$name] = $value;
    }

    /**
     * 永久保存 Cookie 数据
     * @access public
     * @param  string $name   cookie 名称
     * @param  mixed  $value  cookie 值
     * @param  mixed  $option 可选参数 可能会是 null|integer|string
     * @return void
     */
    public static function forever($name, $value = '', $option = null)
    {
        if (is_null($option) || is_numeric($option)) {
            $option = [];
        }

        $option['expire'] = 315360000;

        self::set($name, $value, $option);
    }

    /**
     * 判断是否有 Cookie 数据
     * @access public
     * @param  string      $name   cookie 名称
     * @param  string|null $prefix cookie 前缀
     * @return bool
     */
    public static function has($name, $prefix = null)
    {
        !isset(self::$init) && self::init();

        $prefix = !is_null($prefix) ? $prefix : self::$config['prefix'];

        return isset($_COOKIE[$prefix . $name]);
    }

    /**
     * 获取 Cookie 的值
     * @access public
     * @param string      $name   cookie 名称
     * @param string|null $prefix cookie 前缀
     * @return mixed
     */
    public static function get($name = '', $prefix = null)
    {
        !isset(self::$init) && self::init();

        $prefix = !is_null($prefix) ? $prefix : self::$config['prefix'];
        $key    = $prefix . $name;

        if ('' == $name) {
            // 获取全部
            if ($prefix) {
                $value = [];

                foreach ($_COOKIE as $k => $val) {
                    if (0 === strpos($k, $prefix)) {
                        $value[$k] = $val;
                    }

                }
            } else {
                $value = $_COOKIE;
            }
        } elseif (isset($_COOKIE[$key])) {
            $value = $_COOKIE[$key];

            if (0 === strpos($value, 'think:')) {
                $value = json_decode(substr($value, 6), true);
                array_walk_recursive($value, 'self::jsonFormatProtect', 'decode');
            }
        } else {
            $value = null;
        }

        return $value;
    }

    /**
     * 删除 Cookie
     * @access public
     * @param  string      $name   cookie 名称
     * @param  string|null $prefix cookie 前缀
     * @return void
     */
    public static function delete($name, $prefix = null)
    {
        !isset(self::$init) && self::init();

        $config = self::$config;
        $prefix = !is_null($prefix) ? $prefix : $config['prefix'];
        $name   = $prefix . $name;

        if ($config['setcookie']) {
            setcookie(
                $name, '', $_SERVER['REQUEST_TIME'] - 3600, $config['path'],
                $config['domain'], $config['secure'], $config['httponly']
            );
        }

        // 删除指定 cookie
        unset($_COOKIE[$name]);
    }

    /**
     * 清除指定前缀的所有 cookie
     * @access public
     * @param  string|null $prefix cookie 前缀
     * @return void
     */
    public static function clear($prefix = null)
    {
        if (empty($_COOKIE)) {
            return;
        }

        !isset(self::$init) && self::init();

        // 要删除的 cookie 前缀，不指定则删除 config 设置的指定前缀
        $config = self::$config;
        $prefix = !is_null($prefix) ? $prefix : $config['prefix'];

        if ($prefix) {
            foreach ($_COOKIE as $key => $val) {
                if (0 === strpos($key, $prefix)) {
                    if ($config['setcookie']) {
                        setcookie(
                            $key, '', $_SERVER['REQUEST_TIME'] - 3600, $config['path'],
                            $config['domain'], $config['secure'], $config['httponly']
                        );
                    }

                    unset($_COOKIE[$key]);
                }
            }
        }
    }

    /**
     * json 转换时的格式保护
     * @access protected
     * @param  mixed  $val  要转换的值
     * @param  string $key  键名
     * @param  string $type 转换类别
     * @return void
     */
    protected static function jsonFormatProtect(&$val, $key, $type = 'encode')
    {
        if (!empty($val) && true !== $val) {
            $val = 'decode' == $type ? urldecode($val) : urlencode($val);
        }
    }
}
