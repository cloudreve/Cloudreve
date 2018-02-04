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

namespace think\process\exception;

use think\Process;

class Failed extends \RuntimeException
{

    private $process;

    public function __construct(Process $process)
    {
        if ($process->isSuccessful()) {
            throw new \InvalidArgumentException('Expected a failed process, but the given process was successful.');
        }

        $error = sprintf('The command "%s" failed.' . "\nExit Code: %s(%s)", $process->getCommandLine(), $process->getExitCode(), $process->getExitCodeText());

        if (!$process->isOutputDisabled()) {
            $error .= sprintf("\n\nOutput:\n================\n%s\n\nError Output:\n================\n%s", $process->getOutput(), $process->getErrorOutput());
        }

        parent::__construct($error);

        $this->process = $process;
    }

    public function getProcess()
    {
        return $this->process;
    }
}
