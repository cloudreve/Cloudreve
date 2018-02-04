<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2017 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------

namespace think;

class Lang
{
    // 语言数据
    private static $lang = [];
    // 语言作用域
    private static $range = 'zh-cn';
    // 语言自动侦测的变量
    protected static $langDetectVar = 'lang';
    // 语言Cookie变量
    protected static $langCookieVar = 'think_var';
    // 语言Cookie的过期时间
    protected static $langCookieExpire = 3600;
    // 允许语言列表
    protected static $allowLangList = [];
    // Accept-Language转义为对应语言包名称 系统默认配置
    protected static $acceptLanguage = [
        'zh-hans-cn' => 'zh-cn',
    ];

    // 设定当前的语言
    public static function range($range = '')
    {
        if ('' == $range) {
            return self::$range;
        } else {
            self::$range = $range;
        }
        return self::$range;
    }

    /**
     * 设置语言定义(不区分大小写)
     * @param string|array  $name 语言变量
     * @param string        $value 语言值
     * @param string        $range 语言作用域
     * @return mixed
     */
    public static function set($name, $value = null, $range = '')
    {
        $range = $range ?: self::$range;
        // 批量定义
        if (!isset(self::$lang[$range])) {
            self::$lang[$range] = [];
        }
        if (is_array($name)) {
            return self::$lang[$range] = array_change_key_case($name) + self::$lang[$range];
        } else {
            return self::$lang[$range][strtolower($name)] = $value;
        }
    }

    /**
     * 加载语言定义(不区分大小写)
     * @param string $file 语言文件
     * @param string $range 语言作用域
     * @return mixed
     */
    public static function load($file, $range = '')
    {
        $range = $range ?: self::$range;
        if (!isset(self::$lang[$range])) {
            self::$lang[$range] = [];
        }
        // 批量定义
        if (is_string($file)) {
            $file = [$file];
        }
        $lang = [];
        foreach ($file as $_file) {
            if (is_file($_file)) {
                // 记录加载信息
                App::$debug && Log::record('[ LANG ] ' . $_file, 'info');
                $_lang = include $_file;
                if (is_array($_lang)) {
                    $lang = array_change_key_case($_lang) + $lang;
                }
            }
        }
        if (!empty($lang)) {
            self::$lang[$range] = $lang + self::$lang[$range];
        }
        return self::$lang[$range];
    }

    /**
     * 获取语言定义(不区分大小写)
     * @param string|null   $name 语言变量
     * @param string        $range 语言作用域
     * @return mixed
     */
    public static function has($name, $range = '')
    {
        $range = $range ?: self::$range;
        return isset(self::$lang[$range][strtolower($name)]);
    }

    /**
     * 获取语言定义(不区分大小写)
     * @param string|null   $name 语言变量
     * @param array         $vars 变量替换
     * @param string        $range 语言作用域
     * @return mixed
     */
    public static function get($name = null, $vars = [], $range = '')
    {
        $range = $range ?: self::$range;
        // 空参数返回所有定义
        if (empty($name)) {
            return self::$lang[$range];
        }
        $key   = strtolower($name);
        $value = isset(self::$lang[$range][$key]) ? self::$lang[$range][$key] : $name;

        // 变量解析
        if (!empty($vars) && is_array($vars)) {
            /**
             * Notes:
             * 为了检测的方便，数字索引的判断仅仅是参数数组的第一个元素的key为数字0
             * 数字索引采用的是系统的 sprintf 函数替换，用法请参考 sprintf 函数
             */
            if (key($vars) === 0) {
                // 数字索引解析
                array_unshift($vars, $value);
                $value = call_user_func_array('sprintf', $vars);
            } else {
                // 关联索引解析
                $replace = array_keys($vars);
                foreach ($replace as &$v) {
                    $v = "{:{$v}}";
                }
                $value = str_replace($replace, $vars, $value);
            }

        }
        return $value;
    }

    /**
     * 自动侦测设置获取语言选择
     * @return string
     */
    public static function detect()
    {
        // 自动侦测设置获取语言选择
        $langSet = '';

        if (isset($_GET[self::$langDetectVar])) {
            // url中设置了语言变量
            $langSet = strtolower($_GET[self::$langDetectVar]);
        } elseif (isset($_SERVER['HTTP_ACCEPT_LANGUAGE'])) {
            // 自动侦测浏览器语言
            preg_match('/^([a-z\d\-]+)/i', $_SERVER['HTTP_ACCEPT_LANGUAGE'], $matches);
            $langSet     = strtolower($matches[1]);
            $acceptLangs = Config::get('header_accept_lang');
            if (isset($acceptLangs[$langSet])) {
                $langSet = $acceptLangs[$langSet];
            } elseif (isset(self::$acceptLanguage[$langSet])) {
                $langSet = self::$acceptLanguage[$langSet];
            }
        }
        if (empty(self::$allowLangList) || in_array($langSet, self::$allowLangList)) {
            // 合法的语言
            self::$range = $langSet ?: self::$range;
        }
        return self::$range;
    }

    /**
     * 设置语言自动侦测的变量
     * @param string $var 变量名称
     * @return void
     */
    public static function setLangDetectVar($var)
    {
        self::$langDetectVar = $var;
    }

    /**
     * 设置语言的cookie保存变量
     * @param string $var 变量名称
     * @return void
     */
    public static function setLangCookieVar($var)
    {
        self::$langCookieVar = $var;
    }

    /**
     * 设置语言的cookie的过期时间
     * @param string $expire 过期时间
     * @return void
     */
    public static function setLangCookieExpire($expire)
    {
        self::$langCookieExpire = $expire;
    }

    /**
     * 设置允许的语言列表
     * @param array $list 语言列表
     * @return void
     */
    public static function setAllowLangList($list)
    {
        self::$allowLangList = $list;
    }
}
