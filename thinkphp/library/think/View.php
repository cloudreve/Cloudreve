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

class View
{
    // 视图实例
    protected static $instance;
    // 模板引擎实例
    public $engine;
    // 模板变量
    protected $data = [];
    // 用于静态赋值的模板变量
    protected static $var = [];
    // 视图输出替换
    protected $replace = [];

    /**
     * 构造函数
     * @access public
     * @param array $engine  模板引擎参数
     * @param array $replace  字符串替换参数
     */
    public function __construct($engine = [], $replace = [])
    {
        // 初始化模板引擎
        $this->engine($engine);
        // 基础替换字符串
        $request = Request::instance();
        $base    = $request->root();
        $root    = strpos($base, '.') ? ltrim(dirname($base), DS) : $base;
        if ('' != $root) {
            $root = '/' . ltrim($root, '/');
        }
        $baseReplace = [
            '__ROOT__'   => $root,
            '__URL__'    => $base . '/' . $request->module() . '/' . Loader::parseName($request->controller()),
            '__STATIC__' => $root . '/static',
            '__CSS__'    => $root . '/static/css',
            '__JS__'     => $root . '/static/js',
        ];
        $this->replace = array_merge($baseReplace, (array) $replace);
    }

    /**
     * 初始化视图
     * @access public
     * @param array $engine  模板引擎参数
     * @param array $replace  字符串替换参数
     * @return object
     */
    public static function instance($engine = [], $replace = [])
    {
        if (is_null(self::$instance)) {
            self::$instance = new self($engine, $replace);
        }
        return self::$instance;
    }

    /**
     * 模板变量静态赋值
     * @access public
     * @param mixed $name  变量名
     * @param mixed $value 变量值
     * @return void
     */
    public static function share($name, $value = '')
    {
        if (is_array($name)) {
            self::$var = array_merge(self::$var, $name);
        } else {
            self::$var[$name] = $value;
        }
    }

    /**
     * 模板变量赋值
     * @access public
     * @param mixed $name  变量名
     * @param mixed $value 变量值
     * @return $this
     */
    public function assign($name, $value = '')
    {
        if (is_array($name)) {
            $this->data = array_merge($this->data, $name);
        } else {
            $this->data[$name] = $value;
        }
        return $this;
    }

    /**
     * 设置当前模板解析的引擎
     * @access public
     * @param array|string $options 引擎参数
     * @return $this
     */
    public function engine($options = [])
    {
        if (is_string($options)) {
            $type    = $options;
            $options = [];
        } else {
            $type = !empty($options['type']) ? $options['type'] : 'Think';
        }

        $class = false !== strpos($type, '\\') ? $type : '\\think\\view\\driver\\' . ucfirst($type);
        if (isset($options['type'])) {
            unset($options['type']);
        }
        $this->engine = new $class($options);
        return $this;
    }

    /**
     * 配置模板引擎
     * @access private
     * @param string|array  $name 参数名
     * @param mixed         $value 参数值
     * @return $this
     */
    public function config($name, $value = null)
    {
        $this->engine->config($name, $value);
        return $this;
    }

    /**
     * 解析和获取模板内容 用于输出
     * @param string    $template 模板文件名或者内容
     * @param array     $vars     模板输出变量
     * @param array     $replace 替换内容
     * @param array     $config     模板参数
     * @param bool      $renderContent     是否渲染内容
     * @return string
     * @throws Exception
     */
    public function fetch($template = '', $vars = [], $replace = [], $config = [], $renderContent = false)
    {
        // 模板变量
        $vars = array_merge(self::$var, $this->data, $vars);

        // 页面缓存
        ob_start();
        ob_implicit_flush(0);

        // 渲染输出
        try {
            $method = $renderContent ? 'display' : 'fetch';
            // 允许用户自定义模板的字符串替换
            $replace = array_merge($this->replace, $replace, (array) $this->engine->config('tpl_replace_string'));
            $this->engine->config('tpl_replace_string', $replace);
            $this->engine->$method($template, $vars, $config);
        } catch (\Exception $e) {
            ob_end_clean();
            throw $e;
        }

        // 获取并清空缓存
        $content = ob_get_clean();
        // 内容过滤标签
        Hook::listen('view_filter', $content);
        return $content;
    }

    /**
     * 视图内容替换
     * @access public
     * @param string|array  $content 被替换内容（支持批量替换）
     * @param string        $replace    替换内容
     * @return $this
     */
    public function replace($content, $replace = '')
    {
        if (is_array($content)) {
            $this->replace = array_merge($this->replace, $content);
        } else {
            $this->replace[$content] = $replace;
        }
        return $this;
    }

    /**
     * 渲染内容输出
     * @access public
     * @param string $content 内容
     * @param array  $vars    模板输出变量
     * @param array  $replace 替换内容
     * @param array  $config     模板参数
     * @return mixed
     */
    public function display($content, $vars = [], $replace = [], $config = [])
    {
        return $this->fetch($content, $vars, $replace, $config, true);
    }

    /**
     * 模板变量赋值
     * @access public
     * @param string    $name  变量名
     * @param mixed     $value 变量值
     */
    public function __set($name, $value)
    {
        $this->data[$name] = $value;
    }

    /**
     * 取得模板显示变量的值
     * @access protected
     * @param string $name 模板变量
     * @return mixed
     */
    public function __get($name)
    {
        return $this->data[$name];
    }

    /**
     * 检测模板变量是否设置
     * @access public
     * @param string $name 模板变量名
     * @return bool
     */
    public function __isset($name)
    {
        return isset($this->data[$name]);
    }
}
