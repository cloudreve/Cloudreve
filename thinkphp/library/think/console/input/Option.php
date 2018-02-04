<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2015 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: yunwuxin <448901948@qq.com>
// +----------------------------------------------------------------------

namespace think\console\input;

class Option
{

    const VALUE_NONE     = 1;
    const VALUE_REQUIRED = 2;
    const VALUE_OPTIONAL = 4;
    const VALUE_IS_ARRAY = 8;

    private $name;
    private $shortcut;
    private $mode;
    private $default;
    private $description;

    /**
     * 构造方法
     * @param string       $name        选项名
     * @param string|array $shortcut    短名称,多个用|隔开或者使用数组
     * @param int          $mode        选项类型(可选类型为 self::VALUE_*)
     * @param string       $description 描述
     * @param mixed        $default     默认值 (类型为 self::VALUE_REQUIRED 或者 self::VALUE_NONE 的时候必须为null)
     * @throws \InvalidArgumentException
     */
    public function __construct($name, $shortcut = null, $mode = null, $description = '', $default = null)
    {
        if (0 === strpos($name, '--')) {
            $name = substr($name, 2);
        }

        if (empty($name)) {
            throw new \InvalidArgumentException('An option name cannot be empty.');
        }

        if (empty($shortcut)) {
            $shortcut = null;
        }

        if (null !== $shortcut) {
            if (is_array($shortcut)) {
                $shortcut = implode('|', $shortcut);
            }
            $shortcuts = preg_split('{(\|)-?}', ltrim($shortcut, '-'));
            $shortcuts = array_filter($shortcuts);
            $shortcut  = implode('|', $shortcuts);

            if (empty($shortcut)) {
                throw new \InvalidArgumentException('An option shortcut cannot be empty.');
            }
        }

        if (null === $mode) {
            $mode = self::VALUE_NONE;
        } elseif (!is_int($mode) || $mode > 15 || $mode < 1) {
            throw new \InvalidArgumentException(sprintf('Option mode "%s" is not valid.', $mode));
        }

        $this->name        = $name;
        $this->shortcut    = $shortcut;
        $this->mode        = $mode;
        $this->description = $description;

        if ($this->isArray() && !$this->acceptValue()) {
            throw new \InvalidArgumentException('Impossible to have an option mode VALUE_IS_ARRAY if the option does not accept a value.');
        }

        $this->setDefault($default);
    }

    /**
     * 获取短名称
     * @return string
     */
    public function getShortcut()
    {
        return $this->shortcut;
    }

    /**
     * 获取选项名
     * @return string
     */
    public function getName()
    {
        return $this->name;
    }

    /**
     * 是否可以设置值
     * @return bool 类型不是 self::VALUE_NONE 的时候返回true,其他均返回false
     */
    public function acceptValue()
    {
        return $this->isValueRequired() || $this->isValueOptional();
    }

    /**
     * 是否必须
     * @return bool 类型是 self::VALUE_REQUIRED 的时候返回true,其他均返回false
     */
    public function isValueRequired()
    {
        return self::VALUE_REQUIRED === (self::VALUE_REQUIRED & $this->mode);
    }

    /**
     * 是否可选
     * @return bool 类型是 self::VALUE_OPTIONAL 的时候返回true,其他均返回false
     */
    public function isValueOptional()
    {
        return self::VALUE_OPTIONAL === (self::VALUE_OPTIONAL & $this->mode);
    }

    /**
     * 选项值是否接受数组
     * @return bool 类型是 self::VALUE_IS_ARRAY 的时候返回true,其他均返回false
     */
    public function isArray()
    {
        return self::VALUE_IS_ARRAY === (self::VALUE_IS_ARRAY & $this->mode);
    }

    /**
     * 设置默认值
     * @param mixed $default 默认值
     * @throws \LogicException
     */
    public function setDefault($default = null)
    {
        if (self::VALUE_NONE === (self::VALUE_NONE & $this->mode) && null !== $default) {
            throw new \LogicException('Cannot set a default value when using InputOption::VALUE_NONE mode.');
        }

        if ($this->isArray()) {
            if (null === $default) {
                $default = [];
            } elseif (!is_array($default)) {
                throw new \LogicException('A default value for an array option must be an array.');
            }
        }

        $this->default = $this->acceptValue() ? $default : false;
    }

    /**
     * 获取默认值
     * @return mixed
     */
    public function getDefault()
    {
        return $this->default;
    }

    /**
     * 获取描述文字
     * @return string
     */
    public function getDescription()
    {
        return $this->description;
    }

    /**
     * 检查所给选项是否是当前这个
     * @param Option $option
     * @return bool
     */
    public function equals(Option $option)
    {
        return $option->getName() === $this->getName()
        && $option->getShortcut() === $this->getShortcut()
        && $option->getDefault() === $this->getDefault()
        && $option->isArray() === $this->isArray()
        && $option->isValueRequired() === $this->isValueRequired()
        && $option->isValueOptional() === $this->isValueOptional();
    }
}
