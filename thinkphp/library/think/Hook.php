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

class Hook
{

    private static $tags = [];

    /**
     * 动态添加行为扩展到某个标签
     * @param string    $tag 标签名称
     * @param mixed     $behavior 行为名称
     * @param bool      $first 是否放到开头执行
     * @return void
     */
    public static function add($tag, $behavior, $first = false)
    {
        isset(self::$tags[$tag]) || self::$tags[$tag] = [];
        if (is_array($behavior) && !is_callable($behavior)) {
            if (!array_key_exists('_overlay', $behavior) || !$behavior['_overlay']) {
                unset($behavior['_overlay']);
                self::$tags[$tag] = array_merge(self::$tags[$tag], $behavior);
            } else {
                unset($behavior['_overlay']);
                self::$tags[$tag] = $behavior;
            }
        } elseif ($first) {
            array_unshift(self::$tags[$tag], $behavior);
        } else {
            self::$tags[$tag][] = $behavior;
        }
    }

    /**
     * 批量导入插件
     * @param array        $tags 插件信息
     * @param boolean     $recursive 是否递归合并
     */
    public static function import(array $tags, $recursive = true)
    {
        if ($recursive) {
            foreach ($tags as $tag => $behavior) {
                self::add($tag, $behavior);
            }
        } else {
            self::$tags = $tags + self::$tags;
        }
    }

    /**
     * 获取插件信息
     * @param string $tag 插件位置 留空获取全部
     * @return array
     */
    public static function get($tag = '')
    {
        if (empty($tag)) {
            //获取全部的插件信息
            return self::$tags;
        } else {
            return array_key_exists($tag, self::$tags) ? self::$tags[$tag] : [];
        }
    }

    /**
     * 监听标签的行为
     * @param string $tag    标签名称
     * @param mixed  $params 传入参数
     * @param mixed  $extra  额外参数
     * @param bool   $once   只获取一个有效返回值
     * @return mixed
     */
    public static function listen($tag, &$params = null, $extra = null, $once = false)
    {
        $results = [];
        $tags    = static::get($tag);
        foreach ($tags as $key => $name) {
            $results[$key] = self::exec($name, $tag, $params, $extra);
            if (false === $results[$key]) {
                // 如果返回false 则中断行为执行
                break;
            } elseif (!is_null($results[$key]) && $once) {
                break;
            }
        }
        return $once ? end($results) : $results;
    }

    /**
     * 执行某个行为
     * @param mixed     $class 要执行的行为
     * @param string    $tag 方法名（标签名）
     * @param Mixed     $params 传人的参数
     * @param mixed     $extra 额外参数
     * @return mixed
     */
    public static function exec($class, $tag = '', &$params = null, $extra = null)
    {
        App::$debug && Debug::remark('behavior_start', 'time');
        $method = Loader::parseName($tag, 1, false);
        if ($class instanceof \Closure) {
            $result = call_user_func_array($class, [ & $params, $extra]);
            $class  = 'Closure';
        } elseif (is_array($class)) {
            list($class, $method) = $class;

            $result = (new $class())->$method($params, $extra);
            $class  = $class . '->' . $method;
        } elseif (is_object($class)) {
            $result = $class->$method($params, $extra);
            $class  = get_class($class);
        } elseif (strpos($class, '::')) {
            $result = call_user_func_array($class, [ & $params, $extra]);
        } else {
            $obj    = new $class();
            $method = ($tag && is_callable([$obj, $method])) ? $method : 'run';
            $result = $obj->$method($params, $extra);
        }
        if (App::$debug) {
            Debug::remark('behavior_end', 'time');
            Log::record('[ BEHAVIOR ] Run ' . $class . ' @' . $tag . ' [ RunTime:' . Debug::getRangeTime('behavior_start', 'behavior_end') . 's ]', 'info');
        }
        return $result;
    }

}
