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

use SplFileObject;

class File extends SplFileObject
{
    /**
     * @var string 错误信息
     */
    private $error = '';

    /**
     * @var string 当前完整文件名
     */
    protected $filename;

    /**
     * @var string 上传文件名
     */
    protected $saveName;

    /**
     * @var string 文件上传命名规则
     */
    protected $rule = 'date';

    /**
     * @var array 文件上传验证规则
     */
    protected $validate = [];

    /**
     * @var bool 单元测试
     */
    protected $isTest;

    /**
     * @var array 上传文件信息
     */
    protected $info;

    /**
     * @var array 文件 hash 信息
     */
    protected $hash = [];

    /**
     * File constructor.
     * @access public
     * @param  string $filename 文件名称
     * @param  string $mode     访问模式
     */
    public function __construct($filename, $mode = 'r')
    {
        parent::__construct($filename, $mode);
        $this->filename = $this->getRealPath() ?: $this->getPathname();
    }

    /**
     * 设置是否是单元测试
     * @access public
     * @param  bool $test 是否是测试
     * @return $this
     */
    public function isTest($test = false)
    {
        $this->isTest = $test;

        return $this;
    }

    /**
     * 设置上传信息
     * @access public
     * @param  array $info 上传文件信息
     * @return $this
     */
    public function setUploadInfo($info)
    {
        $this->info = $info;

        return $this;
    }

    /**
     * 获取上传文件的信息
     * @access public
     * @param  string $name 信息名称
     * @return array|string
     */
    public function getInfo($name = '')
    {
        return isset($this->info[$name]) ? $this->info[$name] : $this->info;
    }

    /**
     * 获取上传文件的文件名
     * @access public
     * @return string
     */
    public function getSaveName()
    {
        return $this->saveName;
    }

    /**
     * 设置上传文件的保存文件名
     * @access public
     * @param  string $saveName 保存名称
     * @return $this
     */
    public function setSaveName($saveName)
    {
        $this->saveName = $saveName;

        return $this;
    }

    /**
     * 获取文件的哈希散列值
     * @access public
     * @param  string $type 类型
     * @return string
     */
    public function hash($type = 'sha1')
    {
        if (!isset($this->hash[$type])) {
            $this->hash[$type] = hash_file($type, $this->filename);
        }

        return $this->hash[$type];
    }

    /**
     * 检查目录是否可写
     * @access protected
     * @param  string $path 目录
     * @return boolean
     */
    protected function checkPath($path)
    {
        if (is_dir($path) || mkdir($path, 0755, true)) {
            return true;
        }

        $this->error = ['directory {:path} creation failed', ['path' => $path]];

        return false;
    }

    /**
     * 获取文件类型信息
     * @access public
     * @return string
     */
    public function getMime()
    {
        $finfo = finfo_open(FILEINFO_MIME_TYPE);

        return finfo_file($finfo, $this->filename);
    }

    /**
     * 设置文件的命名规则
     * @access public
     * @param  string $rule 文件命名规则
     * @return $this
     */
    public function rule($rule)
    {
        $this->rule = $rule;

        return $this;
    }

    /**
     * 设置上传文件的验证规则
     * @access public
     * @param  array $rule 验证规则
     * @return $this
     */
    public function validate(array $rule = [])
    {
        $this->validate = $rule;

        return $this;
    }

    /**
     * 检测是否合法的上传文件
     * @access public
     * @return bool
     */
    public function isValid()
    {
        return $this->isTest ? is_file($this->filename) : is_uploaded_file($this->filename);
    }

    /**
     * 检测上传文件
     * @access public
     * @param  array $rule 验证规则
     * @return bool
     */
    public function check($rule = [])
    {
        $rule = $rule ?: $this->validate;

        /* 检查文件大小 */
        if (isset($rule['size']) && !$this->checkSize($rule['size'])) {
            $this->error = 'filesize not match';
            return false;
        }

        /* 检查文件 Mime 类型 */
        if (isset($rule['type']) && !$this->checkMime($rule['type'])) {
            $this->error = 'mimetype to upload is not allowed';
            return false;
        }

        /* 检查文件后缀 */
        if (isset($rule['ext']) && !$this->checkExt($rule['ext'])) {
            $this->error = 'extensions to upload is not allowed';
            return false;
        }

        /* 检查图像文件 */
        if (!$this->checkImg()) {
            $this->error = 'illegal image files';
            return false;
        }

        return true;
    }

    /**
     * 检测上传文件后缀
     * @access public
     * @param  array|string $ext 允许后缀
     * @return bool
     */
    public function checkExt($ext)
    {
        if (is_string($ext)) {
            $ext = explode(',', $ext);
        }

        $extension = strtolower(pathinfo($this->getInfo('name'), PATHINFO_EXTENSION));

        return in_array($extension, $ext);
    }

    /**
     * 检测图像文件
     * @access public
     * @return bool
     */
    public function checkImg()
    {
        $extension = strtolower(pathinfo($this->getInfo('name'), PATHINFO_EXTENSION));

        // 如果上传的不是图片，或者是图片而且后缀确实符合图片类型则返回 true
        return !in_array($extension, ['gif', 'jpg', 'jpeg', 'bmp', 'png', 'swf']) || in_array($this->getImageType($this->filename), [1, 2, 3, 4, 6, 13]);
    }

    /**
     * 判断图像类型
     * @access protected
     * @param  string $image 图片名称
     * @return bool|int
     */
    protected function getImageType($image)
    {
        if (function_exists('exif_imagetype')) {
            return exif_imagetype($image);
        }

        try {
            $info = getimagesize($image);
            return $info ? $info[2] : false;
        } catch (\Exception $e) {
            return false;
        }
    }

    /**
     * 检测上传文件大小
     * @access public
     * @param  integer $size 最大大小
     * @return bool
     */
    public function checkSize($size)
    {
        return $this->getSize() <= $size;
    }

    /**
     * 检测上传文件类型
     * @access public
     * @param  array|string $mime 允许类型
     * @return bool
     */
    public function checkMime($mime)
    {
        $mime = is_string($mime) ? explode(',', $mime) : $mime;

        return in_array(strtolower($this->getMime()), $mime);
    }

    /**
     * 移动文件
     * @access public
     * @param  string      $path     保存路径
     * @param  string|bool $savename 保存的文件名 默认自动生成
     * @param  boolean     $replace  同名文件是否覆盖
     * @return false|File
     */
    public function move($path, $savename = true, $replace = true)
    {
        // 文件上传失败，捕获错误代码
        if (!empty($this->info['error'])) {
            $this->error($this->info['error']);
            return false;
        }

        // 检测合法性
        if (!$this->isValid()) {
            $this->error = 'upload illegal files';
            return false;
        }

        // 验证上传
        if (!$this->check()) {
            return false;
        }

        $path = rtrim($path, DS) . DS;
        // 文件保存命名规则
        $saveName = $this->buildSaveName($savename);
        $filename = $path . $saveName;

        // 检测目录
        if (false === $this->checkPath(dirname($filename))) {
            return false;
        }

        // 不覆盖同名文件
        if (!$replace && is_file($filename)) {
            $this->error = ['has the same filename: {:filename}', ['filename' => $filename]];
            return false;
        }

        /* 移动文件 */
        if ($this->isTest) {
            rename($this->filename, $filename);
        } elseif (!move_uploaded_file($this->filename, $filename)) {
            $this->error = 'upload write error';
            return false;
        }

        // 返回 File 对象实例
        $file = new self($filename);
        $file->setSaveName($saveName)->setUploadInfo($this->info);

        return $file;
    }

    /**
     * 获取保存文件名
     * @access protected
     * @param  string|bool $savename 保存的文件名 默认自动生成
     * @return string
     */
    protected function buildSaveName($savename)
    {
        // 自动生成文件名
        if (true === $savename) {
            if ($this->rule instanceof \Closure) {
                $savename = call_user_func_array($this->rule, [$this]);
            } else {
                switch ($this->rule) {
                    case 'date':
                        $savename = date('Ymd') . DS . md5(microtime(true));
                        break;
                    default:
                        if (in_array($this->rule, hash_algos())) {
                            $hash     = $this->hash($this->rule);
                            $savename = substr($hash, 0, 2) . DS . substr($hash, 2);
                        } elseif (is_callable($this->rule)) {
                            $savename = call_user_func($this->rule);
                        } else {
                            $savename = date('Ymd') . DS . md5(microtime(true));
                        }
                }
            }
        } elseif ('' === $savename || false === $savename) {
            $savename = $this->getInfo('name');
        }

        if (!strpos($savename, '.')) {
            $savename .= '.' . pathinfo($this->getInfo('name'), PATHINFO_EXTENSION);
        }

        return $savename;
    }

    /**
     * 获取错误代码信息
     * @access private
     * @param  int $errorNo 错误号
     * @return $this
     */
    private function error($errorNo)
    {
        switch ($errorNo) {
            case 1:
            case 2:
                $this->error = 'upload File size exceeds the maximum value';
                break;
            case 3:
                $this->error = 'only the portion of file is uploaded';
                break;
            case 4:
                $this->error = 'no file to uploaded';
                break;
            case 6:
                $this->error = 'upload temp dir not found';
                break;
            case 7:
                $this->error = 'file write error';
                break;
            default:
                $this->error = 'unknown upload error';
        }

        return $this;
    }

    /**
     * 获取错误信息（支持多语言）
     * @access public
     * @return string
     */
    public function getError()
    {
        if (is_array($this->error)) {
            list($msg, $vars) = $this->error;
        } else {
            $msg  = $this->error;
            $vars = [];
        }

        return Lang::has($msg) ? Lang::get($msg, $vars) : $msg;
    }

    /**
     * 魔法方法，获取文件的 hash 值
     * @access public
     * @param  string $method 方法名
     * @param  mixed  $args   调用参数
     * @return string
     */
    public function __call($method, $args)
    {
        return $this->hash($method);
    }
}
