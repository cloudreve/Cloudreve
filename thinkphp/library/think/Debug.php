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
use think\response\Redirect;

class Debug
{
    /**
     * @var array 区间时间信息
     */
    protected static $info = [];

    /**
     * @var array 区间内存信息
     */
    protected static $mem = [];

    /**
     * 记录时间（微秒）和内存使用情况
     * @access public
     * @param  string $name  标记位置
     * @param  mixed  $value 标记值(留空则取当前 time 表示仅记录时间 否则同时记录时间和内存)
     * @return void
     */
    public static function remark($name, $value = '')
    {
        self::$info[$name] = is_float($value) ? $value : microtime(true);

        if ('time' != $value) {
            self::$mem['mem'][$name]  = is_float($value) ? $value : memory_get_usage();
            self::$mem['peak'][$name] = memory_get_peak_usage();
        }
    }

    /**
     * 统计某个区间的时间（微秒）使用情况 返回值以秒为单位
     * @access public
     * @param  string  $start 开始标签
     * @param  string  $end   结束标签
     * @param  integer $dec   小数位
     * @return string
     */
    public static function getRangeTime($start, $end, $dec = 6)
    {
        if (!isset(self::$info[$end])) {
            self::$info[$end] = microtime(true);
        }

        return number_format((self::$info[$end] - self::$info[$start]), $dec);
    }

    /**
     * 统计从开始到统计时的时间（微秒）使用情况 返回值以秒为单位
     * @access public
     * @param  integer $dec 小数位
     * @return string
     */
    public static function getUseTime($dec = 6)
    {
        return number_format((microtime(true) - THINK_START_TIME), $dec);
    }

    /**
     * 获取当前访问的吞吐率情况
     * @access public
     * @return string
     */
    public static function getThroughputRate()
    {
        return number_format(1 / self::getUseTime(), 2) . 'req/s';
    }

    /**
     * 记录区间的内存使用情况
     * @access public
     * @param  string  $start 开始标签
     * @param  string  $end   结束标签
     * @param  integer $dec   小数位
     * @return string
     */
    public static function getRangeMem($start, $end, $dec = 2)
    {
        if (!isset(self::$mem['mem'][$end])) {
            self::$mem['mem'][$end] = memory_get_usage();
        }

        $size = self::$mem['mem'][$end] - self::$mem['mem'][$start];
        $a    = ['B', 'KB', 'MB', 'GB', 'TB'];
        $pos  = 0;

        while ($size >= 1024) {
            $size /= 1024;
            $pos++;
        }

        return round($size, $dec) . " " . $a[$pos];
    }

    /**
     * 统计从开始到统计时的内存使用情况
     * @access public
     * @param  integer $dec 小数位
     * @return string
     */
    public static function getUseMem($dec = 2)
    {
        $size = memory_get_usage() - THINK_START_MEM;
        $a    = ['B', 'KB', 'MB', 'GB', 'TB'];
        $pos  = 0;

        while ($size >= 1024) {
            $size /= 1024;
            $pos++;
        }

        return round($size, $dec) . " " . $a[$pos];
    }

    /**
     * 统计区间的内存峰值情况
     * @access public
     * @param  string  $start 开始标签
     * @param  string  $end   结束标签
     * @param  integer $dec   小数位
     * @return string
     */
    public static function getMemPeak($start, $end, $dec = 2)
    {
        if (!isset(self::$mem['peak'][$end])) {
            self::$mem['peak'][$end] = memory_get_peak_usage();
        }

        $size = self::$mem['peak'][$end] - self::$mem['peak'][$start];
        $a    = ['B', 'KB', 'MB', 'GB', 'TB'];
        $pos  = 0;

        while ($size >= 1024) {
            $size /= 1024;
            $pos++;
        }

        return round($size, $dec) . " " . $a[$pos];
    }

    /**
     * 获取文件加载信息
     * @access public
     * @param  bool $detail 是否显示详细
     * @return integer|array
     */
    public static function getFile($detail = false)
    {
        $files = get_included_files();

        if ($detail) {
            $info = [];

            foreach ($files as $file) {
                $info[] = $file . ' ( ' . number_format(filesize($file) / 1024, 2) . ' KB )';
            }

            return $info;
        }

        return count($files);
    }

    /**
     * 浏览器友好的变量输出
     * @access public
     * @param  mixed       $var   变量
     * @param  boolean     $echo  是否输出(默认为 true，为 false 则返回输出字符串)
     * @param  string|null $label 标签(默认为空)
     * @param  integer     $flags htmlspecialchars 的标志
     * @return null|string
     */
    public static function dump($var, $echo = true, $label = null, $flags = ENT_SUBSTITUTE)
    {
        $label = (null === $label) ? '' : rtrim($label) . ':';

        ob_start();
        var_dump($var);
        $output = preg_replace('/\]\=\>\n(\s+)/m', '] => ', ob_get_clean());

        if (IS_CLI) {
            $output = PHP_EOL . $label . $output . PHP_EOL;
        } else {
            if (!extension_loaded('xdebug')) {
                $output = htmlspecialchars($output, $flags);
            }

            $output = '<pre>' . $label . $output . '</pre>';
        }

        if ($echo) {
            echo($output);
            return;
        }

        return $output;
    }

    /**
     * 调试信息注入到响应中
     * @access public
     * @param  Response $response 响应实例
     * @param  string   $content  返回的字符串
     * @return void
     */
    public static function inject(Response $response, &$content)
    {
        $config = Config::get('trace');
        $type   = isset($config['type']) ? $config['type'] : 'Html';
        $class  = false !== strpos($type, '\\') ? $type : '\\think\\debug\\' . ucwords($type);

        unset($config['type']);

        if (!class_exists($class)) {
            throw new ClassNotFoundException('class not exists:' . $class, $class);
        }

        /** @var \think\debug\Console|\think\debug\Html $trace */
        $trace = new $class($config);

        if ($response instanceof Redirect) {
            // TODO 记录
        } else {
            $output = $trace->output($response, Log::getLog());

            if (is_string($output)) {
                // trace 调试信息注入
                $pos = strripos($content, '</body>');
                if (false !== $pos) {
                    $content = substr($content, 0, $pos) . $output . substr($content, $pos);
                } else {
                    $content = $content . $output;
                }
            }
        }
    }
}
