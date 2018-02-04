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

namespace think\process;

use think\Process;

class Builder
{
    private $arguments;
    private $cwd;
    private $env = null;
    private $input;
    private $timeout        = 60;
    private $options        = [];
    private $inheritEnv     = true;
    private $prefix         = [];
    private $outputDisabled = false;

    /**
     * 构造方法
     * @param string[] $arguments 参数
     */
    public function __construct(array $arguments = [])
    {
        $this->arguments = $arguments;
    }

    /**
     * 创建一个实例
     * @param string[] $arguments 参数
     * @return self
     */
    public static function create(array $arguments = [])
    {
        return new static($arguments);
    }

    /**
     * 添加一个参数
     * @param string $argument 参数
     * @return self
     */
    public function add($argument)
    {
        $this->arguments[] = $argument;

        return $this;
    }

    /**
     * 添加一个前缀
     * @param string|array $prefix
     * @return self
     */
    public function setPrefix($prefix)
    {
        $this->prefix = is_array($prefix) ? $prefix : [$prefix];

        return $this;
    }

    /**
     * 设置参数
     * @param string[] $arguments
     * @return  self
     */
    public function setArguments(array $arguments)
    {
        $this->arguments = $arguments;

        return $this;
    }

    /**
     * 设置工作目录
     * @param null|string $cwd
     * @return  self
     */
    public function setWorkingDirectory($cwd)
    {
        $this->cwd = $cwd;

        return $this;
    }

    /**
     * 是否初始化环境变量
     * @param bool $inheritEnv
     * @return self
     */
    public function inheritEnvironmentVariables($inheritEnv = true)
    {
        $this->inheritEnv = $inheritEnv;

        return $this;
    }

    /**
     * 设置环境变量
     * @param string      $name
     * @param null|string $value
     * @return self
     */
    public function setEnv($name, $value)
    {
        $this->env[$name] = $value;

        return $this;
    }

    /**
     *  添加环境变量
     * @param array $variables
     * @return self
     */
    public function addEnvironmentVariables(array $variables)
    {
        $this->env = array_replace($this->env, $variables);

        return $this;
    }

    /**
     * 设置输入
     * @param mixed $input
     * @return self
     */
    public function setInput($input)
    {
        $this->input = Utils::validateInput(sprintf('%s::%s', __CLASS__, __FUNCTION__), $input);

        return $this;
    }

    /**
     * 设置超时时间
     * @param float|null $timeout
     * @return self
     */
    public function setTimeout($timeout)
    {
        if (null === $timeout) {
            $this->timeout = null;

            return $this;
        }

        $timeout = (float) $timeout;

        if ($timeout < 0) {
            throw new \InvalidArgumentException('The timeout value must be a valid positive integer or float number.');
        }

        $this->timeout = $timeout;

        return $this;
    }

    /**
     * 设置proc_open选项
     * @param string $name
     * @param string $value
     * @return self
     */
    public function setOption($name, $value)
    {
        $this->options[$name] = $value;

        return $this;
    }

    /**
     * 禁止输出
     * @return self
     */
    public function disableOutput()
    {
        $this->outputDisabled = true;

        return $this;
    }

    /**
     * 开启输出
     * @return self
     */
    public function enableOutput()
    {
        $this->outputDisabled = false;

        return $this;
    }

    /**
     * 创建一个Process实例
     * @return Process
     */
    public function getProcess()
    {
        if (0 === count($this->prefix) && 0 === count($this->arguments)) {
            throw new \LogicException('You must add() command arguments before calling getProcess().');
        }

        $options = $this->options;

        $arguments = array_merge($this->prefix, $this->arguments);
        $script    = implode(' ', array_map([__NAMESPACE__ . '\\Utils', 'escapeArgument'], $arguments));

        if ($this->inheritEnv) {
            // include $_ENV for BC purposes
            $env = array_replace($_ENV, $_SERVER, $this->env);
        } else {
            $env = $this->env;
        }

        $process = new Process($script, $this->cwd, $env, $this->input, $this->timeout, $options);

        if ($this->outputDisabled) {
            $process->disableOutput();
        }

        return $process;
    }
}
