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

use think\exception\TemplateNotFoundException;
use think\template\TagLib;

/**
 * ThinkPHP分离出来的模板引擎
 * 支持XML标签和普通标签的模板解析
 * 编译型模板引擎 支持动态缓存
 */
class Template
{
    // 模板变量
    protected $data = [];
    // 引擎配置
    protected $config = [
        'view_path'          => '', // 模板路径
        'view_base'          => '',
        'view_suffix'        => 'html', // 默认模板文件后缀
        'view_depr'          => DS,
        'cache_suffix'       => 'php', // 默认模板缓存后缀
        'tpl_deny_func_list' => 'echo,exit', // 模板引擎禁用函数
        'tpl_deny_php'       => false, // 默认模板引擎是否禁用PHP原生代码
        'tpl_begin'          => '{', // 模板引擎普通标签开始标记
        'tpl_end'            => '}', // 模板引擎普通标签结束标记
        'strip_space'        => false, // 是否去除模板文件里面的html空格与换行
        'tpl_cache'          => true, // 是否开启模板编译缓存,设为false则每次都会重新编译
        'compile_type'       => 'file', // 模板编译类型
        'cache_prefix'       => '', // 模板缓存前缀标识，可以动态改变
        'cache_time'         => 0, // 模板缓存有效期 0 为永久，(以数字为值，单位:秒)
        'layout_on'          => false, // 布局模板开关
        'layout_name'        => 'layout', // 布局模板入口文件
        'layout_item'        => '{__CONTENT__}', // 布局模板的内容替换标识
        'taglib_begin'       => '{', // 标签库标签开始标记
        'taglib_end'         => '}', // 标签库标签结束标记
        'taglib_load'        => true, // 是否使用内置标签库之外的其它标签库，默认自动检测
        'taglib_build_in'    => 'cx', // 内置标签库名称(标签使用不必指定标签库名称),以逗号分隔 注意解析顺序
        'taglib_pre_load'    => '', // 需要额外加载的标签库(须指定标签库名称)，多个以逗号分隔
        'display_cache'      => false, // 模板渲染缓存
        'cache_id'           => '', // 模板缓存ID
        'tpl_replace_string' => [],
        'tpl_var_identify'   => 'array', // .语法变量识别，array|object|'', 为空时自动识别
    ];

    private $literal     = [];
    private $includeFile = []; // 记录所有模板包含的文件路径及更新时间
    protected $storage;

    /**
     * 构造函数
     * @access public
     * @param array $config
     */
    public function __construct(array $config = [])
    {
        $this->config['cache_path'] = TEMP_PATH;
        $this->config               = array_merge($this->config, $config);

        $this->config['taglib_begin_origin'] = $this->config['taglib_begin'];
        $this->config['taglib_end_origin']   = $this->config['taglib_end'];

        $this->config['taglib_begin'] = preg_quote($this->config['taglib_begin'], '/');
        $this->config['taglib_end']   = preg_quote($this->config['taglib_end'], '/');
        $this->config['tpl_begin']    = preg_quote($this->config['tpl_begin'], '/');
        $this->config['tpl_end']      = preg_quote($this->config['tpl_end'], '/');

        // 初始化模板编译存储器
        $type          = $this->config['compile_type'] ? $this->config['compile_type'] : 'File';
        $class         = false !== strpos($type, '\\') ? $type : '\\think\\template\\driver\\' . ucwords($type);
        $this->storage = new $class();
    }

    /**
     * 模板变量赋值
     * @access public
     * @param mixed $name
     * @param mixed $value
     * @return void
     */
    public function assign($name, $value = '')
    {
        if (is_array($name)) {
            $this->data = array_merge($this->data, $name);
        } else {
            $this->data[$name] = $value;
        }
    }

    /**
     * 模板引擎参数赋值
     * @access public
     * @param mixed $name
     * @param mixed $value
     */
    public function __set($name, $value)
    {
        $this->config[$name] = $value;
    }

    /**
     * 模板引擎配置项
     * @access public
     * @param array|string $config
     * @return string|void|array
     */
    public function config($config)
    {
        if (is_array($config)) {
            $this->config = array_merge($this->config, $config);
        } elseif (isset($this->config[$config])) {
            return $this->config[$config];
        } else {
            return;
        }
    }

    /**
     * 模板变量获取
     * @access public
     * @param  string $name 变量名
     * @return mixed
     */
    public function get($name = '')
    {
        if ('' == $name) {
            return $this->data;
        } else {
            $data = $this->data;
            foreach (explode('.', $name) as $key => $val) {
                if (isset($data[$val])) {
                    $data = $data[$val];
                } else {
                    $data = null;
                    break;
                }
            }
            return $data;
        }
    }

    /**
     * 渲染模板文件
     * @access public
     * @param string    $template 模板文件
     * @param array     $vars 模板变量
     * @param array     $config 模板参数
     * @return void
     */
    public function fetch($template, $vars = [], $config = [])
    {
        if ($vars) {
            $this->data = $vars;
        }
        if ($config) {
            $this->config($config);
        }
        if (!empty($this->config['cache_id']) && $this->config['display_cache']) {
            // 读取渲染缓存
            $cacheContent = Cache::get($this->config['cache_id']);
            if (false !== $cacheContent) {
                echo $cacheContent;
                return;
            }
        }
        $template = $this->parseTemplateFile($template);
        if ($template) {
            $cacheFile = $this->config['cache_path'] . $this->config['cache_prefix'] . md5($this->config['layout_name'] . $template) . '.' . ltrim($this->config['cache_suffix'], '.');
            if (!$this->checkCache($cacheFile)) {
                // 缓存无效 重新模板编译
                $content = file_get_contents($template);
                $this->compiler($content, $cacheFile);
            }
            // 页面缓存
            ob_start();
            ob_implicit_flush(0);
            // 读取编译存储
            $this->storage->read($cacheFile, $this->data);
            // 获取并清空缓存
            $content = ob_get_clean();
            if (!empty($this->config['cache_id']) && $this->config['display_cache']) {
                // 缓存页面输出
                Cache::set($this->config['cache_id'], $content, $this->config['cache_time']);
            }
            echo $content;
        }
    }

    /**
     * 渲染模板内容
     * @access public
     * @param string    $content 模板内容
     * @param array     $vars 模板变量
     * @param array     $config 模板参数
     * @return void
     */
    public function display($content, $vars = [], $config = [])
    {
        if ($vars) {
            $this->data = $vars;
        }
        if ($config) {
            $this->config($config);
        }
        $cacheFile = $this->config['cache_path'] . $this->config['cache_prefix'] . md5($content) . '.' . ltrim($this->config['cache_suffix'], '.');
        if (!$this->checkCache($cacheFile)) {
            // 缓存无效 模板编译
            $this->compiler($content, $cacheFile);
        }
        // 读取编译存储
        $this->storage->read($cacheFile, $this->data);
    }

    /**
     * 设置布局
     * @access public
     * @param mixed     $name 布局模板名称 false 则关闭布局
     * @param string    $replace 布局模板内容替换标识
     * @return Template
     */
    public function layout($name, $replace = '')
    {
        if (false === $name) {
            // 关闭布局
            $this->config['layout_on'] = false;
        } else {
            // 开启布局
            $this->config['layout_on'] = true;
            // 名称必须为字符串
            if (is_string($name)) {
                $this->config['layout_name'] = $name;
            }
            if (!empty($replace)) {
                $this->config['layout_item'] = $replace;
            }
        }
        return $this;
    }

    /**
     * 检查编译缓存是否有效
     * 如果无效则需要重新编译
     * @access private
     * @param string $cacheFile 缓存文件名
     * @return boolean
     */
    private function checkCache($cacheFile)
    {
        // 未开启缓存功能
        if (!$this->config['tpl_cache']) {
            return false;
        }
        // 缓存文件不存在
        if (!is_file($cacheFile)) {
            return false;
        }
        // 读取缓存文件失败
        if (!$handle = @fopen($cacheFile, "r")) {
            return false;
        }
        // 读取第一行
        preg_match('/\/\*(.+?)\*\//', fgets($handle), $matches);
        if (!isset($matches[1])) {
            return false;
        }
        $includeFile = unserialize($matches[1]);
        if (!is_array($includeFile)) {
            return false;
        }
        // 检查模板文件是否有更新
        foreach ($includeFile as $path => $time) {
            if (is_file($path) && filemtime($path) > $time) {
                // 模板文件如果有更新则缓存需要更新
                return false;
            }
        }
        // 检查编译存储是否有效
        return $this->storage->check($cacheFile, $this->config['cache_time']);
    }

    /**
     * 检查编译缓存是否存在
     * @access public
     * @param string $cacheId 缓存的id
     * @return boolean
     */
    public function isCache($cacheId)
    {
        if ($cacheId && $this->config['display_cache']) {
            // 缓存页面输出
            return Cache::has($cacheId);
        }
        return false;
    }

    /**
     * 编译模板文件内容
     * @access private
     * @param string    $content 模板内容
     * @param string    $cacheFile 缓存文件名
     * @return void
     */
    private function compiler(&$content, $cacheFile)
    {
        // 判断是否启用布局
        if ($this->config['layout_on']) {
            if (false !== strpos($content, '{__NOLAYOUT__}')) {
                // 可以单独定义不使用布局
                $content = str_replace('{__NOLAYOUT__}', '', $content);
            } else {
                // 读取布局模板
                $layoutFile = $this->parseTemplateFile($this->config['layout_name']);
                if ($layoutFile) {
                    // 替换布局的主体内容
                    $content = str_replace($this->config['layout_item'], $content, file_get_contents($layoutFile));
                }
            }
        } else {
            $content = str_replace('{__NOLAYOUT__}', '', $content);
        }

        // 模板解析
        $this->parse($content);
        if ($this->config['strip_space']) {
            /* 去除html空格与换行 */
            $find    = ['~>\s+<~', '~>(\s+\n|\r)~'];
            $replace = ['><', '>'];
            $content = preg_replace($find, $replace, $content);
        }
        // 优化生成的php代码
        $content = preg_replace('/\?>\s*<\?php\s(?!echo\b)/s', '', $content);
        // 模板过滤输出
        $replace = $this->config['tpl_replace_string'];
        $content = str_replace(array_keys($replace), array_values($replace), $content);
        // 添加安全代码及模板引用记录
        $content = '<?php if (!defined(\'THINK_PATH\')) exit(); /*' . serialize($this->includeFile) . '*/ ?>' . "\n" . $content;
        // 编译存储
        $this->storage->write($cacheFile, $content);
        $this->includeFile = [];
        return;
    }

    /**
     * 模板解析入口
     * 支持普通标签和TagLib解析 支持自定义标签库
     * @access public
     * @param string $content 要解析的模板内容
     * @return void
     */
    public function parse(&$content)
    {
        // 内容为空不解析
        if (empty($content)) {
            return;
        }
        // 替换literal标签内容
        $this->parseLiteral($content);
        // 解析继承
        $this->parseExtend($content);
        // 解析布局
        $this->parseLayout($content);
        // 检查include语法
        $this->parseInclude($content);
        // 替换包含文件中literal标签内容
        $this->parseLiteral($content);
        // 检查PHP语法
        $this->parsePhp($content);

        // 获取需要引入的标签库列表
        // 标签库只需要定义一次，允许引入多个一次
        // 一般放在文件的最前面
        // 格式：<taglib name="html,mytag..." />
        // 当TAGLIB_LOAD配置为true时才会进行检测
        if ($this->config['taglib_load']) {
            $tagLibs = $this->getIncludeTagLib($content);
            if (!empty($tagLibs)) {
                // 对导入的TagLib进行解析
                foreach ($tagLibs as $tagLibName) {
                    $this->parseTagLib($tagLibName, $content);
                }
            }
        }
        // 预先加载的标签库 无需在每个模板中使用taglib标签加载 但必须使用标签库XML前缀
        if ($this->config['taglib_pre_load']) {
            $tagLibs = explode(',', $this->config['taglib_pre_load']);
            foreach ($tagLibs as $tag) {
                $this->parseTagLib($tag, $content);
            }
        }
        // 内置标签库 无需使用taglib标签导入就可以使用 并且不需使用标签库XML前缀
        $tagLibs = explode(',', $this->config['taglib_build_in']);
        foreach ($tagLibs as $tag) {
            $this->parseTagLib($tag, $content, true);
        }
        // 解析普通模板标签 {$tagName}
        $this->parseTag($content);

        // 还原被替换的Literal标签
        $this->parseLiteral($content, true);
        return;
    }

    /**
     * 检查PHP语法
     * @access private
     * @param string $content 要解析的模板内容
     * @return void
     * @throws \think\Exception
     */
    private function parsePhp(&$content)
    {
        // 短标签的情况要将<?标签用echo方式输出 否则无法正常输出xml标识
        $content = preg_replace('/(<\?(?!php|=|$))/i', '<?php echo \'\\1\'; ?>' . "\n", $content);
        // PHP语法检查
        if ($this->config['tpl_deny_php'] && false !== strpos($content, '<?php')) {
            throw new Exception('not allow php tag', 11600);
        }
        return;
    }

    /**
     * 解析模板中的布局标签
     * @access private
     * @param string $content 要解析的模板内容
     * @return void
     */
    private function parseLayout(&$content)
    {
        // 读取模板中的布局标签
        if (preg_match($this->getRegex('layout'), $content, $matches)) {
            // 替换Layout标签
            $content = str_replace($matches[0], '', $content);
            // 解析Layout标签
            $array = $this->parseAttr($matches[0]);
            if (!$this->config['layout_on'] || $this->config['layout_name'] != $array['name']) {
                // 读取布局模板
                $layoutFile = $this->parseTemplateFile($array['name']);
                if ($layoutFile) {
                    $replace = isset($array['replace']) ? $array['replace'] : $this->config['layout_item'];
                    // 替换布局的主体内容
                    $content = str_replace($replace, $content, file_get_contents($layoutFile));
                }
            }
        } else {
            $content = str_replace('{__NOLAYOUT__}', '', $content);
        }
        return;
    }

    /**
     * 解析模板中的include标签
     * @access private
     * @param  string $content 要解析的模板内容
     * @return void
     */
    private function parseInclude(&$content)
    {
        $regex = $this->getRegex('include');
        $func  = function ($template) use (&$func, &$regex, &$content) {
            if (preg_match_all($regex, $template, $matches, PREG_SET_ORDER)) {
                foreach ($matches as $match) {
                    $array = $this->parseAttr($match[0]);
                    $file  = $array['file'];
                    unset($array['file']);
                    // 分析模板文件名并读取内容
                    $parseStr = $this->parseTemplateName($file);
                    foreach ($array as $k => $v) {
                        // 以$开头字符串转换成模板变量
                        if (0 === strpos($v, '$')) {
                            $v = $this->get(substr($v, 1));
                        }
                        $parseStr = str_replace('[' . $k . ']', $v, $parseStr);
                    }
                    $content = str_replace($match[0], $parseStr, $content);
                    // 再次对包含文件进行模板分析
                    $func($parseStr);
                }
                unset($matches);
            }
        };
        // 替换模板中的include标签
        $func($content);
        return;
    }

    /**
     * 解析模板中的extend标签
     * @access private
     * @param  string $content 要解析的模板内容
     * @return void
     */
    private function parseExtend(&$content)
    {
        $regex  = $this->getRegex('extend');
        $array  = $blocks  = $baseBlocks  = [];
        $extend = '';
        $func   = function ($template) use (&$func, &$regex, &$array, &$extend, &$blocks, &$baseBlocks) {
            if (preg_match($regex, $template, $matches)) {
                if (!isset($array[$matches['name']])) {
                    $array[$matches['name']] = 1;
                    // 读取继承模板
                    $extend = $this->parseTemplateName($matches['name']);
                    // 递归检查继承
                    $func($extend);
                    // 取得block标签内容
                    $blocks = array_merge($blocks, $this->parseBlock($template));
                    return;
                }
            } else {
                // 取得顶层模板block标签内容
                $baseBlocks = $this->parseBlock($template, true);
                if (empty($extend)) {
                    // 无extend标签但有block标签的情况
                    $extend = $template;
                }
            }
        };

        $func($content);
        if (!empty($extend)) {
            if ($baseBlocks) {
                $children = [];
                foreach ($baseBlocks as $name => $val) {
                    $replace = $val['content'];
                    if (!empty($children[$name])) {
                        // 如果包含有子block标签
                        foreach ($children[$name] as $key) {
                            $replace = str_replace($baseBlocks[$key]['begin'] . $baseBlocks[$key]['content'] . $baseBlocks[$key]['end'], $blocks[$key]['content'], $replace);
                        }
                    }
                    if (isset($blocks[$name])) {
                        // 带有{__block__}表示与所继承模板的相应标签合并，而不是覆盖
                        $replace = str_replace(['{__BLOCK__}', '{__block__}'], $replace, $blocks[$name]['content']);
                        if (!empty($val['parent'])) {
                            // 如果不是最顶层的block标签
                            $parent = $val['parent'];
                            if (isset($blocks[$parent])) {
                                $blocks[$parent]['content'] = str_replace($blocks[$name]['begin'] . $blocks[$name]['content'] . $blocks[$name]['end'], $replace, $blocks[$parent]['content']);
                            }
                            $blocks[$name]['content'] = $replace;
                            $children[$parent][]      = $name;
                            continue;
                        }
                    } elseif (!empty($val['parent'])) {
                        // 如果子标签没有被继承则用原值
                        $children[$val['parent']][] = $name;
                        $blocks[$name]              = $val;
                    }
                    if (!$val['parent']) {
                        // 替换模板中的顶级block标签
                        $extend = str_replace($val['begin'] . $val['content'] . $val['end'], $replace, $extend);
                    }
                }
            }
            $content = $extend;
            unset($blocks, $baseBlocks);
        }
        return;
    }

    /**
     * 替换页面中的literal标签
     * @access private
     * @param  string   $content 模板内容
     * @param  boolean  $restore 是否为还原
     * @return void
     */
    private function parseLiteral(&$content, $restore = false)
    {
        $regex = $this->getRegex($restore ? 'restoreliteral' : 'literal');
        if (preg_match_all($regex, $content, $matches, PREG_SET_ORDER)) {
            if (!$restore) {
                $count = count($this->literal);
                // 替换literal标签
                foreach ($matches as $match) {
                    $this->literal[] = substr($match[0], strlen($match[1]), -strlen($match[2]));
                    $content         = str_replace($match[0], "<!--###literal{$count}###-->", $content);
                    $count++;
                }
            } else {
                // 还原literal标签
                foreach ($matches as $match) {
                    $content = str_replace($match[0], $this->literal[$match[1]], $content);
                }
                // 清空literal记录
                $this->literal = [];
            }
            unset($matches);
        }
        return;
    }

    /**
     * 获取模板中的block标签
     * @access private
     * @param  string   $content 模板内容
     * @param  boolean  $sort 是否排序
     * @return array
     */
    private function parseBlock(&$content, $sort = false)
    {
        $regex  = $this->getRegex('block');
        $result = [];
        if (preg_match_all($regex, $content, $matches, PREG_SET_ORDER | PREG_OFFSET_CAPTURE)) {
            $right = $keys = [];
            foreach ($matches as $match) {
                if (empty($match['name'][0])) {
                    if (count($right) > 0) {
                        $tag                  = array_pop($right);
                        $start                = $tag['offset'] + strlen($tag['tag']);
                        $length               = $match[0][1] - $start;
                        $result[$tag['name']] = [
                            'begin'   => $tag['tag'],
                            'content' => substr($content, $start, $length),
                            'end'     => $match[0][0],
                            'parent'  => count($right) ? end($right)['name'] : '',
                        ];
                        $keys[$tag['name']] = $match[0][1];
                    }
                } else {
                    // 标签头压入栈
                    $right[] = [
                        'name'   => $match[2][0],
                        'offset' => $match[0][1],
                        'tag'    => $match[0][0],
                    ];
                }
            }
            unset($right, $matches);
            if ($sort) {
                // 按block标签结束符在模板中的位置排序
                array_multisort($keys, $result);
            }
        }
        return $result;
    }

    /**
     * 搜索模板页面中包含的TagLib库
     * 并返回列表
     * @access private
     * @param  string $content 模板内容
     * @return array|null
     */
    private function getIncludeTagLib(&$content)
    {
        // 搜索是否有TagLib标签
        if (preg_match($this->getRegex('taglib'), $content, $matches)) {
            // 替换TagLib标签
            $content = str_replace($matches[0], '', $content);
            return explode(',', $matches['name']);
        }
        return;
    }

    /**
     * TagLib库解析
     * @access public
     * @param  string   $tagLib 要解析的标签库
     * @param  string   $content 要解析的模板内容
     * @param  boolean  $hide 是否隐藏标签库前缀
     * @return void
     */
    public function parseTagLib($tagLib, &$content, $hide = false)
    {
        if (false !== strpos($tagLib, '\\')) {
            // 支持指定标签库的命名空间
            $className = $tagLib;
            $tagLib    = substr($tagLib, strrpos($tagLib, '\\') + 1);
        } else {
            $className = '\\think\\template\\taglib\\' . ucwords($tagLib);
        }
        /** @var Taglib $tLib */
        $tLib = new $className($this);
        $tLib->parseTag($content, $hide ? '' : $tagLib);
        return;
    }

    /**
     * 分析标签属性
     * @access public
     * @param  string   $str 属性字符串
     * @param  string   $name 不为空时返回指定的属性名
     * @return array
     */
    public function parseAttr($str, $name = null)
    {
        $regex = '/\s+(?>(?P<name>[\w-]+)\s*)=(?>\s*)([\"\'])(?P<value>(?:(?!\\2).)*)\\2/is';
        $array = [];
        if (preg_match_all($regex, $str, $matches, PREG_SET_ORDER)) {
            foreach ($matches as $match) {
                $array[$match['name']] = $match['value'];
            }
            unset($matches);
        }
        if (!empty($name) && isset($array[$name])) {
            return $array[$name];
        } else {
            return $array;
        }
    }

    /**
     * 模板标签解析
     * 格式： {TagName:args [|content] }
     * @access private
     * @param  string $content 要解析的模板内容
     * @return void
     */
    private function parseTag(&$content)
    {
        $regex = $this->getRegex('tag');
        if (preg_match_all($regex, $content, $matches, PREG_SET_ORDER)) {
            foreach ($matches as $match) {
                $str  = stripslashes($match[1]);
                $flag = substr($str, 0, 1);
                switch ($flag) {
                    case '$':
                        // 解析模板变量 格式 {$varName}
                        // 是否带有?号
                        if (false !== $pos = strpos($str, '?')) {
                            $array = preg_split('/([!=]={1,2}|(?<!-)[><]={0,1})/', substr($str, 0, $pos), 2, PREG_SPLIT_DELIM_CAPTURE);
                            $name  = $array[0];
                            $this->parseVar($name);
                            $this->parseVarFunction($name);

                            $str = trim(substr($str, $pos + 1));
                            $this->parseVar($str);
                            $first = substr($str, 0, 1);
                            if (strpos($name, ')')) {
                                // $name为对象或是自动识别，或者含有函数
                                if (isset($array[1])) {
                                    $this->parseVar($array[2]);
                                    $name .= $array[1] . $array[2];
                                }
                                switch ($first) {
                                    case '?':
                                        $str = '<?php echo (' . $name . ') ? ' . $name . ' : ' . substr($str, 1) . '; ?>';
                                        break;
                                    case '=':
                                        $str = '<?php if(' . $name . ') echo ' . substr($str, 1) . '; ?>';
                                        break;
                                    default:
                                        $str = '<?php echo ' . $name . '?' . $str . '; ?>';
                                }
                            } else {
                                if (isset($array[1])) {
                                    $this->parseVar($array[2]);
                                    $express = $name . $array[1] . $array[2];
                                } else {
                                    $express = false;
                                }
                                // $name为数组
                                switch ($first) {
                                    case '?':
                                        // {$varname??'xxx'} $varname有定义则输出$varname,否则输出xxx
                                        $str = '<?php echo ' . ($express ?: 'isset(' . $name . ')') . '?' . $name . ':' . substr($str, 1) . '; ?>';
                                        break;
                                    case '=':
                                        // {$varname?='xxx'} $varname为真时才输出xxx
                                        $str = '<?php if(' . ($express ?: '!empty(' . $name . ')') . ') echo ' . substr($str, 1) . '; ?>';
                                        break;
                                    case ':':
                                        // {$varname?:'xxx'} $varname为真时输出$varname,否则输出xxx
                                        $str = '<?php echo ' . ($express ?: '!empty(' . $name . ')') . '?' . $name . $str . '; ?>';
                                        break;
                                    default:
                                        $str = '<?php echo ' . ($express ?: '!empty(' . $name . ')') . '?' . $str . '; ?>';
                                }
                            }
                        } else {
                            $this->parseVar($str);
                            $this->parseVarFunction($str);
                            $str = '<?php echo ' . $str . '; ?>';
                        }
                        break;
                    case ':':
                        // 输出某个函数的结果
                        $str = substr($str, 1);
                        $this->parseVar($str);
                        $str = '<?php echo ' . $str . '; ?>';
                        break;
                    case '~':
                        // 执行某个函数
                        $str = substr($str, 1);
                        $this->parseVar($str);
                        $str = '<?php ' . $str . '; ?>';
                        break;
                    case '-':
                    case '+':
                        // 输出计算
                        $this->parseVar($str);
                        $str = '<?php echo ' . $str . '; ?>';
                        break;
                    case '/':
                        // 注释标签
                        $flag2 = substr($str, 1, 1);
                        if ('/' == $flag2 || ('*' == $flag2 && substr(rtrim($str), -2) == '*/')) {
                            $str = '';
                        }
                        break;
                    default:
                        // 未识别的标签直接返回
                        $str = $this->config['tpl_begin'] . $str . $this->config['tpl_end'];
                        break;
                }
                $content = str_replace($match[0], $str, $content);
            }
            unset($matches);
        }
        return;
    }

    /**
     * 模板变量解析,支持使用函数
     * 格式： {$varname|function1|function2=arg1,arg2}
     * @access public
     * @param  string $varStr 变量数据
     * @return void
     */
    public function parseVar(&$varStr)
    {
        $varStr = trim($varStr);
        if (preg_match_all('/\$[a-zA-Z_](?>\w*)(?:[:\.][0-9a-zA-Z_](?>\w*))+/', $varStr, $matches, PREG_OFFSET_CAPTURE)) {
            static $_varParseList = [];
            while ($matches[0]) {
                $match = array_pop($matches[0]);
                //如果已经解析过该变量字串，则直接返回变量值
                if (isset($_varParseList[$match[0]])) {
                    $parseStr = $_varParseList[$match[0]];
                } else {
                    if (strpos($match[0], '.')) {
                        $vars  = explode('.', $match[0]);
                        $first = array_shift($vars);
                        if ('$Think' == $first) {
                            // 所有以Think.打头的以特殊变量对待 无需模板赋值就可以输出
                            $parseStr = $this->parseThinkVar($vars);
                        } elseif ('$Request' == $first) {
                            // 获取Request请求对象参数
                            $method = array_shift($vars);
                            if (!empty($vars)) {
                                $params = implode('.', $vars);
                                if ('true' != $params) {
                                    $params = '\'' . $params . '\'';
                                }
                            } else {
                                $params = '';
                            }
                            $parseStr = '\think\Request::instance()->' . $method . '(' . $params . ')';
                        } else {
                            switch ($this->config['tpl_var_identify']) {
                                case 'array': // 识别为数组
                                    $parseStr = $first . '[\'' . implode('\'][\'', $vars) . '\']';
                                    break;
                                case 'obj': // 识别为对象
                                    $parseStr = $first . '->' . implode('->', $vars);
                                    break;
                                default: // 自动判断数组或对象
                                    $parseStr = '(is_array(' . $first . ')?' . $first . '[\'' . implode('\'][\'', $vars) . '\']:' . $first . '->' . implode('->', $vars) . ')';
                            }
                        }
                    } else {
                        $parseStr = str_replace(':', '->', $match[0]);
                    }
                    $_varParseList[$match[0]] = $parseStr;
                }
                $varStr = substr_replace($varStr, $parseStr, $match[1], strlen($match[0]));
            }
            unset($matches);
        }
        return;
    }

    /**
     * 对模板中使用了函数的变量进行解析
     * 格式 {$varname|function1|function2=arg1,arg2}
     * @access public
     * @param  string $varStr 变量字符串
     * @return void
     */
    public function parseVarFunction(&$varStr)
    {
        if (false == strpos($varStr, '|')) {
            return;
        }
        static $_varFunctionList = [];
        $_key                    = md5($varStr);
        //如果已经解析过该变量字串，则直接返回变量值
        if (isset($_varFunctionList[$_key])) {
            $varStr = $_varFunctionList[$_key];
        } else {
            $varArray = explode('|', $varStr);
            // 取得变量名称
            $name = array_shift($varArray);
            // 对变量使用函数
            $length = count($varArray);
            // 取得模板禁止使用函数列表
            $template_deny_funs = explode(',', $this->config['tpl_deny_func_list']);
            for ($i = 0; $i < $length; $i++) {
                $args = explode('=', $varArray[$i], 2);
                // 模板函数过滤
                $fun = trim($args[0]);
                switch ($fun) {
                    case 'default': // 特殊模板函数
                        if (false === strpos($name, '(')) {
                            $name = '(isset(' . $name . ') && (' . $name . ' !== \'\')?' . $name . ':' . $args[1] . ')';
                        } else {
                            $name = '(' . $name . ' ?: ' . $args[1] . ')';
                        }
                        break;
                    default: // 通用模板函数
                        if (!in_array($fun, $template_deny_funs)) {
                            if (isset($args[1])) {
                                if (strstr($args[1], '###')) {
                                    $args[1] = str_replace('###', $name, $args[1]);
                                    $name    = "$fun($args[1])";
                                } else {
                                    $name = "$fun($name,$args[1])";
                                }
                            } else {
                                if (!empty($args[0])) {
                                    $name = "$fun($name)";
                                }
                            }
                        }
                }
            }
            $_varFunctionList[$_key] = $name;
            $varStr                  = $name;
        }
        return;
    }

    /**
     * 特殊模板变量解析
     * 格式 以 $Think. 打头的变量属于特殊模板变量
     * @access public
     * @param  array $vars 变量数组
     * @return string
     */
    public function parseThinkVar($vars)
    {
        $type  = strtoupper(trim(array_shift($vars)));
        $param = implode('.', $vars);
        if ($vars) {
            switch ($type) {
                case 'SERVER':
                    $parseStr = '\\think\\Request::instance()->server(\'' . $param . '\')';
                    break;
                case 'GET':
                    $parseStr = '\\think\\Request::instance()->get(\'' . $param . '\')';
                    break;
                case 'POST':
                    $parseStr = '\\think\\Request::instance()->post(\'' . $param . '\')';
                    break;
                case 'COOKIE':
                    $parseStr = '\\think\\Cookie::get(\'' . $param . '\')';
                    break;
                case 'SESSION':
                    $parseStr = '\\think\\Session::get(\'' . $param . '\')';
                    break;
                case 'ENV':
                    $parseStr = '\\think\\Request::instance()->env(\'' . $param . '\')';
                    break;
                case 'REQUEST':
                    $parseStr = '\\think\\Request::instance()->request(\'' . $param . '\')';
                    break;
                case 'CONST':
                    $parseStr = strtoupper($param);
                    break;
                case 'LANG':
                    $parseStr = '\\think\\Lang::get(\'' . $param . '\')';
                    break;
                case 'CONFIG':
                    $parseStr = '\\think\\Config::get(\'' . $param . '\')';
                    break;
                default:
                    $parseStr = '\'\'';
                    break;
            }
        } else {
            switch ($type) {
                case 'NOW':
                    $parseStr = "date('Y-m-d g:i a',time())";
                    break;
                case 'VERSION':
                    $parseStr = 'THINK_VERSION';
                    break;
                case 'LDELIM':
                    $parseStr = '\'' . ltrim($this->config['tpl_begin'], '\\') . '\'';
                    break;
                case 'RDELIM':
                    $parseStr = '\'' . ltrim($this->config['tpl_end'], '\\') . '\'';
                    break;
                default:
                    if (defined($type)) {
                        $parseStr = $type;
                    } else {
                        $parseStr = '';
                    }
            }
        }
        return $parseStr;
    }

    /**
     * 分析加载的模板文件并读取内容 支持多个模板文件读取
     * @access private
     * @param  string $templateName 模板文件名
     * @return string
     */
    private function parseTemplateName($templateName)
    {
        $array    = explode(',', $templateName);
        $parseStr = '';
        foreach ($array as $templateName) {
            if (empty($templateName)) {
                continue;
            }
            if (0 === strpos($templateName, '$')) {
                //支持加载变量文件名
                $templateName = $this->get(substr($templateName, 1));
            }
            $template = $this->parseTemplateFile($templateName);
            if ($template) {
                // 获取模板文件内容
                $parseStr .= file_get_contents($template);
            }
        }
        return $parseStr;
    }

    /**
     * 解析模板文件名
     * @access private
     * @param  string $template 文件名
     * @return string|false
     */
    private function parseTemplateFile($template)
    {
        if ('' == pathinfo($template, PATHINFO_EXTENSION)) {
            if (strpos($template, '@')) {
                list($module, $template) = explode('@', $template);
            }
            if (0 !== strpos($template, '/')) {
                $template = str_replace(['/', ':'], $this->config['view_depr'], $template);
            } else {
                $template = str_replace(['/', ':'], $this->config['view_depr'], substr($template, 1));
            }
            if ($this->config['view_base']) {
                $module = isset($module) ? $module : Request::instance()->module();
                $path   = $this->config['view_base'] . ($module ? $module . DS : '');
            } else {
                $path = isset($module) ? APP_PATH . $module . DS . basename($this->config['view_path']) . DS : $this->config['view_path'];
            }
            $template = realpath($path . $template . '.' . ltrim($this->config['view_suffix'], '.'));
        }

        if (is_file($template)) {
            // 记录模板文件的更新时间
            $this->includeFile[$template] = filemtime($template);
            return $template;
        } else {
            throw new TemplateNotFoundException('template not exists:' . $template, $template);
        }
    }

    /**
     * 按标签生成正则
     * @access private
     * @param  string $tagName 标签名
     * @return string
     */
    private function getRegex($tagName)
    {
        $regex = '';
        if ('tag' == $tagName) {
            $begin = $this->config['tpl_begin'];
            $end   = $this->config['tpl_end'];
            if (strlen(ltrim($begin, '\\')) == 1 && strlen(ltrim($end, '\\')) == 1) {
                $regex = $begin . '((?:[\$]{1,2}[a-wA-w_]|[\:\~][\$a-wA-w_]|[+]{2}[\$][a-wA-w_]|[-]{2}[\$][a-wA-w_]|\/[\*\/])(?>[^' . $end . ']*))' . $end;
            } else {
                $regex = $begin . '((?:[\$]{1,2}[a-wA-w_]|[\:\~][\$a-wA-w_]|[+]{2}[\$][a-wA-w_]|[-]{2}[\$][a-wA-w_]|\/[\*\/])(?>(?:(?!' . $end . ').)*))' . $end;
            }
        } else {
            $begin  = $this->config['taglib_begin'];
            $end    = $this->config['taglib_end'];
            $single = strlen(ltrim($begin, '\\')) == 1 && strlen(ltrim($end, '\\')) == 1 ? true : false;
            switch ($tagName) {
                case 'block':
                    if ($single) {
                        $regex = $begin . '(?:' . $tagName . '\b(?>(?:(?!name=).)*)\bname=([\'\"])(?P<name>[\$\w\-\/\.]+)\\1(?>[^' . $end . ']*)|\/' . $tagName . ')' . $end;
                    } else {
                        $regex = $begin . '(?:' . $tagName . '\b(?>(?:(?!name=).)*)\bname=([\'\"])(?P<name>[\$\w\-\/\.]+)\\1(?>(?:(?!' . $end . ').)*)|\/' . $tagName . ')' . $end;
                    }
                    break;
                case 'literal':
                    if ($single) {
                        $regex = '(' . $begin . $tagName . '\b(?>[^' . $end . ']*)' . $end . ')';
                        $regex .= '(?:(?>[^' . $begin . ']*)(?>(?!' . $begin . '(?>' . $tagName . '\b[^' . $end . ']*|\/' . $tagName . ')' . $end . ')' . $begin . '[^' . $begin . ']*)*)';
                        $regex .= '(' . $begin . '\/' . $tagName . $end . ')';
                    } else {
                        $regex = '(' . $begin . $tagName . '\b(?>(?:(?!' . $end . ').)*)' . $end . ')';
                        $regex .= '(?:(?>(?:(?!' . $begin . ').)*)(?>(?!' . $begin . '(?>' . $tagName . '\b(?>(?:(?!' . $end . ').)*)|\/' . $tagName . ')' . $end . ')' . $begin . '(?>(?:(?!' . $begin . ').)*))*)';
                        $regex .= '(' . $begin . '\/' . $tagName . $end . ')';
                    }
                    break;
                case 'restoreliteral':
                    $regex = '<!--###literal(\d+)###-->';
                    break;
                case 'include':
                    $name = 'file';
                case 'taglib':
                case 'layout':
                case 'extend':
                    if (empty($name)) {
                        $name = 'name';
                    }
                    if ($single) {
                        $regex = $begin . $tagName . '\b(?>(?:(?!' . $name . '=).)*)\b' . $name . '=([\'\"])(?P<name>[\$\w\-\/\.\:@,\\\\]+)\\1(?>[^' . $end . ']*)' . $end;
                    } else {
                        $regex = $begin . $tagName . '\b(?>(?:(?!' . $name . '=).)*)\b' . $name . '=([\'\"])(?P<name>[\$\w\-\/\.\:@,\\\\]+)\\1(?>(?:(?!' . $end . ').)*)' . $end;
                    }
                    break;
            }
        }
        return '/' . $regex . '/is';
    }
}
