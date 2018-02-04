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

namespace think\process\pipes;

abstract class Pipes
{

    /** @var array */
    public $pipes = [];

    /** @var string */
    protected $inputBuffer = '';
    /** @var resource|null */
    protected $input;

    /** @var bool */
    private $blocked = true;

    const CHUNK_SIZE = 16384;

    /**
     * 返回用于 proc_open 描述符的数组
     * @return array
     */
    abstract public function getDescriptors();

    /**
     * 返回一个数组的索引由其相关的流，以防这些管道使用的临时文件的文件名。
     * @return string[]
     */
    abstract public function getFiles();

    /**
     * 文件句柄和管道中读取数据。
     * @param bool $blocking 是否使用阻塞调用
     * @param bool $close    是否要关闭管道，如果他们已经到达 EOF。
     * @return string[]
     */
    abstract public function readAndWrite($blocking, $close = false);

    /**
     * 返回当前状态如果有打开的文件句柄或管道。
     * @return bool
     */
    abstract public function areOpen();

    /**
     * {@inheritdoc}
     */
    public function close()
    {
        foreach ($this->pipes as $pipe) {
            fclose($pipe);
        }
        $this->pipes = [];
    }

    /**
     * 检查系统调用已被中断
     * @return bool
     */
    protected function hasSystemCallBeenInterrupted()
    {
        $lastError = error_get_last();

        return isset($lastError['message']) && false !== stripos($lastError['message'], 'interrupted system call');
    }

    protected function unblock()
    {
        if (!$this->blocked) {
            return;
        }

        foreach ($this->pipes as $pipe) {
            stream_set_blocking($pipe, 0);
        }
        if (null !== $this->input) {
            stream_set_blocking($this->input, 0);
        }

        $this->blocked = false;
    }
}
