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

class Definition
{

    /**
     * @var Argument[]
     */
    private $arguments;

    private $requiredCount;
    private $hasAnArrayArgument = false;
    private $hasOptional;

    /**
     * @var Option[]
     */
    private $options;
    private $shortcuts;

    /**
     * 构造方法
     * @param array $definition
     * @api
     */
    public function __construct(array $definition = [])
    {
        $this->setDefinition($definition);
    }

    /**
     * 设置指令的定义
     * @param array $definition 定义的数组
     */
    public function setDefinition(array $definition)
    {
        $arguments = [];
        $options   = [];
        foreach ($definition as $item) {
            if ($item instanceof Option) {
                $options[] = $item;
            } else {
                $arguments[] = $item;
            }
        }

        $this->setArguments($arguments);
        $this->setOptions($options);
    }

    /**
     * 设置参数
     * @param Argument[] $arguments 参数数组
     */
    public function setArguments($arguments = [])
    {
        $this->arguments          = [];
        $this->requiredCount      = 0;
        $this->hasOptional        = false;
        $this->hasAnArrayArgument = false;
        $this->addArguments($arguments);
    }

    /**
     * 添加参数
     * @param Argument[] $arguments 参数数组
     * @api
     */
    public function addArguments($arguments = [])
    {
        if (null !== $arguments) {
            foreach ($arguments as $argument) {
                $this->addArgument($argument);
            }
        }
    }

    /**
     * 添加一个参数
     * @param Argument $argument 参数
     * @throws \LogicException
     */
    public function addArgument(Argument $argument)
    {
        if (isset($this->arguments[$argument->getName()])) {
            throw new \LogicException(sprintf('An argument with name "%s" already exists.', $argument->getName()));
        }

        if ($this->hasAnArrayArgument) {
            throw new \LogicException('Cannot add an argument after an array argument.');
        }

        if ($argument->isRequired() && $this->hasOptional) {
            throw new \LogicException('Cannot add a required argument after an optional one.');
        }

        if ($argument->isArray()) {
            $this->hasAnArrayArgument = true;
        }

        if ($argument->isRequired()) {
            ++$this->requiredCount;
        } else {
            $this->hasOptional = true;
        }

        $this->arguments[$argument->getName()] = $argument;
    }

    /**
     * 根据名称或者位置获取参数
     * @param string|int $name 参数名或者位置
     * @return Argument 参数
     * @throws \InvalidArgumentException
     */
    public function getArgument($name)
    {
        if (!$this->hasArgument($name)) {
            throw new \InvalidArgumentException(sprintf('The "%s" argument does not exist.', $name));
        }

        $arguments = is_int($name) ? array_values($this->arguments) : $this->arguments;

        return $arguments[$name];
    }

    /**
     * 根据名称或位置检查是否具有某个参数
     * @param string|int $name 参数名或者位置
     * @return bool
     * @api
     */
    public function hasArgument($name)
    {
        $arguments = is_int($name) ? array_values($this->arguments) : $this->arguments;

        return isset($arguments[$name]);
    }

    /**
     * 获取所有的参数
     * @return Argument[] 参数数组
     */
    public function getArguments()
    {
        return $this->arguments;
    }

    /**
     * 获取参数数量
     * @return int
     */
    public function getArgumentCount()
    {
        return $this->hasAnArrayArgument ? PHP_INT_MAX : count($this->arguments);
    }

    /**
     * 获取必填的参数的数量
     * @return int
     */
    public function getArgumentRequiredCount()
    {
        return $this->requiredCount;
    }

    /**
     * 获取参数默认值
     * @return array
     */
    public function getArgumentDefaults()
    {
        $values = [];
        foreach ($this->arguments as $argument) {
            $values[$argument->getName()] = $argument->getDefault();
        }

        return $values;
    }

    /**
     * 设置选项
     * @param Option[] $options 选项数组
     */
    public function setOptions($options = [])
    {
        $this->options   = [];
        $this->shortcuts = [];
        $this->addOptions($options);
    }

    /**
     * 添加选项
     * @param Option[] $options 选项数组
     * @api
     */
    public function addOptions($options = [])
    {
        foreach ($options as $option) {
            $this->addOption($option);
        }
    }

    /**
     * 添加一个选项
     * @param Option $option 选项
     * @throws \LogicException
     * @api
     */
    public function addOption(Option $option)
    {
        if (isset($this->options[$option->getName()]) && !$option->equals($this->options[$option->getName()])) {
            throw new \LogicException(sprintf('An option named "%s" already exists.', $option->getName()));
        }

        if ($option->getShortcut()) {
            foreach (explode('|', $option->getShortcut()) as $shortcut) {
                if (isset($this->shortcuts[$shortcut])
                    && !$option->equals($this->options[$this->shortcuts[$shortcut]])
                ) {
                    throw new \LogicException(sprintf('An option with shortcut "%s" already exists.', $shortcut));
                }
            }
        }

        $this->options[$option->getName()] = $option;
        if ($option->getShortcut()) {
            foreach (explode('|', $option->getShortcut()) as $shortcut) {
                $this->shortcuts[$shortcut] = $option->getName();
            }
        }
    }

    /**
     * 根据名称获取选项
     * @param string $name 选项名
     * @return Option
     * @throws \InvalidArgumentException
     * @api
     */
    public function getOption($name)
    {
        if (!$this->hasOption($name)) {
            throw new \InvalidArgumentException(sprintf('The "--%s" option does not exist.', $name));
        }

        return $this->options[$name];
    }

    /**
     * 根据名称检查是否有这个选项
     * @param string $name 选项名
     * @return bool
     * @api
     */
    public function hasOption($name)
    {
        return isset($this->options[$name]);
    }

    /**
     * 获取所有选项
     * @return Option[]
     * @api
     */
    public function getOptions()
    {
        return $this->options;
    }

    /**
     * 根据名称检查某个选项是否有短名称
     * @param string $name 短名称
     * @return bool
     */
    public function hasShortcut($name)
    {
        return isset($this->shortcuts[$name]);
    }

    /**
     * 根据短名称获取选项
     * @param string $shortcut 短名称
     * @return Option
     */
    public function getOptionForShortcut($shortcut)
    {
        return $this->getOption($this->shortcutToName($shortcut));
    }

    /**
     * 获取所有选项的默认值
     * @return array
     */
    public function getOptionDefaults()
    {
        $values = [];
        foreach ($this->options as $option) {
            $values[$option->getName()] = $option->getDefault();
        }

        return $values;
    }

    /**
     * 根据短名称获取选项名
     * @param string $shortcut 短名称
     * @return string
     * @throws \InvalidArgumentException
     */
    private function shortcutToName($shortcut)
    {
        if (!isset($this->shortcuts[$shortcut])) {
            throw new \InvalidArgumentException(sprintf('The "-%s" option does not exist.', $shortcut));
        }

        return $this->shortcuts[$shortcut];
    }

    /**
     * 获取该指令的介绍
     * @param bool $short 是否简洁介绍
     * @return string
     */
    public function getSynopsis($short = false)
    {
        $elements = [];

        if ($short && $this->getOptions()) {
            $elements[] = '[options]';
        } elseif (!$short) {
            foreach ($this->getOptions() as $option) {
                $value = '';
                if ($option->acceptValue()) {
                    $value = sprintf(' %s%s%s', $option->isValueOptional() ? '[' : '', strtoupper($option->getName()), $option->isValueOptional() ? ']' : '');
                }

                $shortcut   = $option->getShortcut() ? sprintf('-%s|', $option->getShortcut()) : '';
                $elements[] = sprintf('[%s--%s%s]', $shortcut, $option->getName(), $value);
            }
        }

        if (count($elements) && $this->getArguments()) {
            $elements[] = '[--]';
        }

        foreach ($this->getArguments() as $argument) {
            $element = '<' . $argument->getName() . '>';
            if (!$argument->isRequired()) {
                $element = '[' . $element . ']';
            } elseif ($argument->isArray()) {
                $element .= ' (' . $element . ')';
            }

            if ($argument->isArray()) {
                $element .= '...';
            }

            $elements[] = $element;
        }

        return implode(' ', $elements);
    }
}
