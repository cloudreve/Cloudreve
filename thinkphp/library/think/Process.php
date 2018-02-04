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

namespace think;

use think\process\exception\Failed as ProcessFailedException;
use think\process\exception\Timeout as ProcessTimeoutException;
use think\process\pipes\Pipes;
use think\process\pipes\Unix as UnixPipes;
use think\process\pipes\Windows as WindowsPipes;
use think\process\Utils;

class Process
{

    const ERR = 'err';
    const OUT = 'out';

    const STATUS_READY      = 'ready';
    const STATUS_STARTED    = 'started';
    const STATUS_TERMINATED = 'terminated';

    const STDIN  = 0;
    const STDOUT = 1;
    const STDERR = 2;

    const TIMEOUT_PRECISION = 0.2;

    private $callback;
    private $commandline;
    private $cwd;
    private $env;
    private $input;
    private $starttime;
    private $lastOutputTime;
    private $timeout;
    private $idleTimeout;
    private $options;
    private $exitcode;
    private $fallbackExitcode;
    private $processInformation;
    private $outputDisabled = false;
    private $stdout;
    private $stderr;
    private $enhanceWindowsCompatibility = true;
    private $enhanceSigchildCompatibility;
    private $process;
    private $status                       = self::STATUS_READY;
    private $incrementalOutputOffset      = 0;
    private $incrementalErrorOutputOffset = 0;
    private $tty;
    private $pty;

    private $useFileHandles = false;

    /** @var Pipes */
    private $processPipes;

    private $latestSignal;

    private static $sigchild;

    /**
     * @var array
     */
    public static $exitCodes = [
        0   => 'OK',
        1   => 'General error',
        2   => 'Misuse of shell builtins',
        126 => 'Invoked command cannot execute',
        127 => 'Command not found',
        128 => 'Invalid exit argument',
        // signals
        129 => 'Hangup',
        130 => 'Interrupt',
        131 => 'Quit and dump core',
        132 => 'Illegal instruction',
        133 => 'Trace/breakpoint trap',
        134 => 'Process aborted',
        135 => 'Bus error: "access to undefined portion of memory object"',
        136 => 'Floating point exception: "erroneous arithmetic operation"',
        137 => 'Kill (terminate immediately)',
        138 => 'User-defined 1',
        139 => 'Segmentation violation',
        140 => 'User-defined 2',
        141 => 'Write to pipe with no one reading',
        142 => 'Signal raised by alarm',
        143 => 'Termination (request to terminate)',
        // 144 - not defined
        145 => 'Child process terminated, stopped (or continued*)',
        146 => 'Continue if stopped',
        147 => 'Stop executing temporarily',
        148 => 'Terminal stop signal',
        149 => 'Background process attempting to read from tty ("in")',
        150 => 'Background process attempting to write to tty ("out")',
        151 => 'Urgent data available on socket',
        152 => 'CPU time limit exceeded',
        153 => 'File size limit exceeded',
        154 => 'Signal raised by timer counting virtual time: "virtual timer expired"',
        155 => 'Profiling timer expired',
        // 156 - not defined
        157 => 'Pollable event',
        // 158 - not defined
        159 => 'Bad syscall',
    ];

    /**
     * 构造方法
     * @param string         $commandline 指令
     * @param string|null    $cwd         工作目录
     * @param array|null     $env         环境变量
     * @param string|null    $input       输入
     * @param int|float|null $timeout     超时时间
     * @param array          $options     proc_open的选项
     * @throws \RuntimeException
     * @api
     */
    public function __construct($commandline, $cwd = null, array $env = null, $input = null, $timeout = 60, array $options = [])
    {
        if (!function_exists('proc_open')) {
            throw new \RuntimeException('The Process class relies on proc_open, which is not available on your PHP installation.');
        }

        $this->commandline = $commandline;
        $this->cwd         = $cwd;

        if (null === $this->cwd && (defined('ZEND_THREAD_SAFE') || '\\' === DS)) {
            $this->cwd = getcwd();
        }
        if (null !== $env) {
            $this->setEnv($env);
        }

        $this->input = $input;
        $this->setTimeout($timeout);
        $this->useFileHandles               = '\\' === DS;
        $this->pty                          = false;
        $this->enhanceWindowsCompatibility  = true;
        $this->enhanceSigchildCompatibility = '\\' !== DS && $this->isSigchildEnabled();
        $this->options                      = array_replace([
            'suppress_errors' => true,
            'binary_pipes'    => true,
        ], $options);
    }

    public function __destruct()
    {
        $this->stop();
    }

    public function __clone()
    {
        $this->resetProcessData();
    }

    /**
     * 运行指令
     * @param callback|null $callback
     * @return int
     */
    public function run($callback = null)
    {
        $this->start($callback);

        return $this->wait();
    }

    /**
     * 运行指令
     * @param callable|null $callback
     * @return self
     * @throws \RuntimeException
     * @throws ProcessFailedException
     */
    public function mustRun($callback = null)
    {
        if ($this->isSigchildEnabled() && !$this->enhanceSigchildCompatibility) {
            throw new \RuntimeException('This PHP has been compiled with --enable-sigchild. You must use setEnhanceSigchildCompatibility() to use this method.');
        }

        if (0 !== $this->run($callback)) {
            throw new ProcessFailedException($this);
        }

        return $this;
    }

    /**
     * 启动进程并写到 STDIN 输入后返回。
     * @param callable|null $callback
     * @throws \RuntimeException
     * @throws \RuntimeException
     * @throws \LogicException
     */
    public function start($callback = null)
    {
        if ($this->isRunning()) {
            throw new \RuntimeException('Process is already running');
        }
        if ($this->outputDisabled && null !== $callback) {
            throw new \LogicException('Output has been disabled, enable it to allow the use of a callback.');
        }

        $this->resetProcessData();
        $this->starttime = $this->lastOutputTime = microtime(true);
        $this->callback  = $this->buildCallback($callback);
        $descriptors     = $this->getDescriptors();

        $commandline = $this->commandline;

        if ('\\' === DS && $this->enhanceWindowsCompatibility) {
            $commandline = 'cmd /V:ON /E:ON /C "(' . $commandline . ')';
            foreach ($this->processPipes->getFiles() as $offset => $filename) {
                $commandline .= ' ' . $offset . '>' . Utils::escapeArgument($filename);
            }
            $commandline .= '"';

            if (!isset($this->options['bypass_shell'])) {
                $this->options['bypass_shell'] = true;
            }
        }

        $this->process = proc_open($commandline, $descriptors, $this->processPipes->pipes, $this->cwd, $this->env, $this->options);

        if (!is_resource($this->process)) {
            throw new \RuntimeException('Unable to launch a new process.');
        }
        $this->status = self::STATUS_STARTED;

        if ($this->tty) {
            return;
        }

        $this->updateStatus(false);
        $this->checkTimeout();
    }

    /**
     * 重启进程
     * @param callable|null $callback
     * @return Process
     * @throws \RuntimeException
     * @throws \RuntimeException
     */
    public function restart($callback = null)
    {
        if ($this->isRunning()) {
            throw new \RuntimeException('Process is already running');
        }

        $process = clone $this;
        $process->start($callback);

        return $process;
    }

    /**
     * 等待要终止的进程
     * @param callable|null $callback
     * @return int
     */
    public function wait($callback = null)
    {
        $this->requireProcessIsStarted(__FUNCTION__);

        $this->updateStatus(false);
        if (null !== $callback) {
            $this->callback = $this->buildCallback($callback);
        }

        do {
            $this->checkTimeout();
            $running = '\\' === DS ? $this->isRunning() : $this->processPipes->areOpen();
            $close   = '\\' !== DS || !$running;
            $this->readPipes(true, $close);
        } while ($running);

        while ($this->isRunning()) {
            usleep(1000);
        }

        if ($this->processInformation['signaled'] && $this->processInformation['termsig'] !== $this->latestSignal) {
            throw new \RuntimeException(sprintf('The process has been signaled with signal "%s".', $this->processInformation['termsig']));
        }

        return $this->exitcode;
    }

    /**
     * 获取PID
     * @return int|null
     * @throws \RuntimeException
     */
    public function getPid()
    {
        if ($this->isSigchildEnabled()) {
            throw new \RuntimeException('This PHP has been compiled with --enable-sigchild. The process identifier can not be retrieved.');
        }

        $this->updateStatus(false);

        return $this->isRunning() ? $this->processInformation['pid'] : null;
    }

    /**
     * 将一个 POSIX 信号发送到进程中
     * @param int $signal
     * @return Process
     */
    public function signal($signal)
    {
        $this->doSignal($signal, true);

        return $this;
    }

    /**
     * 禁用从底层过程获取输出和错误输出。
     * @return Process
     */
    public function disableOutput()
    {
        if ($this->isRunning()) {
            throw new \RuntimeException('Disabling output while the process is running is not possible.');
        }
        if (null !== $this->idleTimeout) {
            throw new \LogicException('Output can not be disabled while an idle timeout is set.');
        }

        $this->outputDisabled = true;

        return $this;
    }

    /**
     * 开启从底层过程获取输出和错误输出。
     * @return Process
     * @throws \RuntimeException
     */
    public function enableOutput()
    {
        if ($this->isRunning()) {
            throw new \RuntimeException('Enabling output while the process is running is not possible.');
        }

        $this->outputDisabled = false;

        return $this;
    }

    /**
     * 输出是否禁用
     * @return bool
     */
    public function isOutputDisabled()
    {
        return $this->outputDisabled;
    }

    /**
     * 获取当前的输出管道
     * @return string
     * @throws \LogicException
     * @throws \LogicException
     * @api
     */
    public function getOutput()
    {
        if ($this->outputDisabled) {
            throw new \LogicException('Output has been disabled.');
        }

        $this->requireProcessIsStarted(__FUNCTION__);

        $this->readPipes(false, '\\' === DS ? !$this->processInformation['running'] : true);

        return $this->stdout;
    }

    /**
     * 以增量方式返回的输出结果。
     * @return string
     */
    public function getIncrementalOutput()
    {
        $this->requireProcessIsStarted(__FUNCTION__);

        $data = $this->getOutput();

        $latest = substr($data, $this->incrementalOutputOffset);

        if (false === $latest) {
            return '';
        }

        $this->incrementalOutputOffset = strlen($data);

        return $latest;
    }

    /**
     * 清空输出
     * @return Process
     */
    public function clearOutput()
    {
        $this->stdout                  = '';
        $this->incrementalOutputOffset = 0;

        return $this;
    }

    /**
     * 返回当前的错误输出的过程 (STDERR)。
     * @return string
     */
    public function getErrorOutput()
    {
        if ($this->outputDisabled) {
            throw new \LogicException('Output has been disabled.');
        }

        $this->requireProcessIsStarted(__FUNCTION__);

        $this->readPipes(false, '\\' === DS ? !$this->processInformation['running'] : true);

        return $this->stderr;
    }

    /**
     * 以增量方式返回 errorOutput
     * @return string
     */
    public function getIncrementalErrorOutput()
    {
        $this->requireProcessIsStarted(__FUNCTION__);

        $data = $this->getErrorOutput();

        $latest = substr($data, $this->incrementalErrorOutputOffset);

        if (false === $latest) {
            return '';
        }

        $this->incrementalErrorOutputOffset = strlen($data);

        return $latest;
    }

    /**
     * 清空 errorOutput
     * @return Process
     */
    public function clearErrorOutput()
    {
        $this->stderr                       = '';
        $this->incrementalErrorOutputOffset = 0;

        return $this;
    }

    /**
     * 获取退出码
     * @return null|int
     */
    public function getExitCode()
    {
        if ($this->isSigchildEnabled() && !$this->enhanceSigchildCompatibility) {
            throw new \RuntimeException('This PHP has been compiled with --enable-sigchild. You must use setEnhanceSigchildCompatibility() to use this method.');
        }

        $this->updateStatus(false);

        return $this->exitcode;
    }

    /**
     * 获取退出文本
     * @return null|string
     */
    public function getExitCodeText()
    {
        if (null === $exitcode = $this->getExitCode()) {
            return;
        }

        return isset(self::$exitCodes[$exitcode]) ? self::$exitCodes[$exitcode] : 'Unknown error';
    }

    /**
     * 检查是否成功
     * @return bool
     */
    public function isSuccessful()
    {
        return 0 === $this->getExitCode();
    }

    /**
     * 是否未捕获的信号已被终止子进程
     * @return bool
     */
    public function hasBeenSignaled()
    {
        $this->requireProcessIsTerminated(__FUNCTION__);

        if ($this->isSigchildEnabled()) {
            throw new \RuntimeException('This PHP has been compiled with --enable-sigchild. Term signal can not be retrieved.');
        }

        $this->updateStatus(false);

        return $this->processInformation['signaled'];
    }

    /**
     * 返回导致子进程终止其执行的数。
     * @return int
     */
    public function getTermSignal()
    {
        $this->requireProcessIsTerminated(__FUNCTION__);

        if ($this->isSigchildEnabled()) {
            throw new \RuntimeException('This PHP has been compiled with --enable-sigchild. Term signal can not be retrieved.');
        }

        $this->updateStatus(false);

        return $this->processInformation['termsig'];
    }

    /**
     * 检查子进程信号是否已停止
     * @return bool
     */
    public function hasBeenStopped()
    {
        $this->requireProcessIsTerminated(__FUNCTION__);

        $this->updateStatus(false);

        return $this->processInformation['stopped'];
    }

    /**
     * 返回导致子进程停止其执行的数。
     * @return int
     */
    public function getStopSignal()
    {
        $this->requireProcessIsTerminated(__FUNCTION__);

        $this->updateStatus(false);

        return $this->processInformation['stopsig'];
    }

    /**
     * 检查是否正在运行
     * @return bool
     */
    public function isRunning()
    {
        if (self::STATUS_STARTED !== $this->status) {
            return false;
        }

        $this->updateStatus(false);

        return $this->processInformation['running'];
    }

    /**
     * 检查是否已开始
     * @return bool
     */
    public function isStarted()
    {
        return self::STATUS_READY != $this->status;
    }

    /**
     * 检查是否已终止
     * @return bool
     */
    public function isTerminated()
    {
        $this->updateStatus(false);

        return self::STATUS_TERMINATED == $this->status;
    }

    /**
     * 获取当前的状态
     * @return string
     */
    public function getStatus()
    {
        $this->updateStatus(false);

        return $this->status;
    }

    /**
     * 终止进程
     */
    public function stop()
    {
        if ($this->isRunning()) {
            if ('\\' === DS && !$this->isSigchildEnabled()) {
                exec(sprintf('taskkill /F /T /PID %d 2>&1', $this->getPid()), $output, $exitCode);
                if ($exitCode > 0) {
                    throw new \RuntimeException('Unable to kill the process');
                }
            } else {
                $pids = preg_split('/\s+/', `ps -o pid --no-heading --ppid {$this->getPid()}`);
                foreach ($pids as $pid) {
                    if (is_numeric($pid)) {
                        posix_kill($pid, 9);
                    }
                }
            }
        }

        $this->updateStatus(false);
        if ($this->processInformation['running']) {
            $this->close();
        }

        return $this->exitcode;
    }

    /**
     * 添加一行输出
     * @param string $line
     */
    public function addOutput($line)
{
        $this->lastOutputTime = microtime(true);
        $this->stdout .= $line;
    }

    /**
     * 添加一行错误输出
     * @param string $line
     */
    public function addErrorOutput($line)
{
        $this->lastOutputTime = microtime(true);
        $this->stderr .= $line;
    }

    /**
     * 获取被执行的指令
     * @return string
     */
    public function getCommandLine()
{
        return $this->commandline;
    }

    /**
     * 设置指令
     * @param string $commandline
     * @return self
     */
    public function setCommandLine($commandline)
{
        $this->commandline = $commandline;

        return $this;
    }

    /**
     * 获取超时时间
     * @return float|null
     */
    public function getTimeout()
{
        return $this->timeout;
    }

    /**
     * 获取idle超时时间
     * @return float|null
     */
    public function getIdleTimeout()
{
        return $this->idleTimeout;
    }

    /**
     * 设置超时时间
     * @param int|float|null $timeout
     * @return self
     */
    public function setTimeout($timeout)
{
        $this->timeout = $this->validateTimeout($timeout);

        return $this;
    }

    /**
     * 设置idle超时时间
     * @param int|float|null $timeout
     * @return self
     */
    public function setIdleTimeout($timeout)
{
        if (null !== $timeout && $this->outputDisabled) {
            throw new \LogicException('Idle timeout can not be set while the output is disabled.');
        }

        $this->idleTimeout = $this->validateTimeout($timeout);

        return $this;
    }

    /**
     * 设置TTY
     * @param bool $tty
     * @return self
     */
    public function setTty($tty)
{
        if ('\\' === DS && $tty) {
            throw new \RuntimeException('TTY mode is not supported on Windows platform.');
        }
        if ($tty && (!file_exists('/dev/tty') || !is_readable('/dev/tty'))) {
            throw new \RuntimeException('TTY mode requires /dev/tty to be readable.');
        }

        $this->tty = (bool) $tty;

        return $this;
    }

    /**
     * 检查是否是tty模式
     * @return bool
     */
    public function isTty()
{
        return $this->tty;
    }

    /**
     * 设置pty模式
     * @param bool $bool
     * @return self
     */
    public function setPty($bool)
{
        $this->pty = (bool) $bool;

        return $this;
    }

    /**
     * 是否是pty模式
     * @return bool
     */
    public function isPty()
{
        return $this->pty;
    }

    /**
     * 获取工作目录
     * @return string|null
     */
    public function getWorkingDirectory()
{
        if (null === $this->cwd) {
            return getcwd() ?: null;
        }

        return $this->cwd;
    }

    /**
     * 设置工作目录
     * @param string $cwd
     * @return self
     */
    public function setWorkingDirectory($cwd)
{
        $this->cwd = $cwd;

        return $this;
    }

    /**
     * 获取环境变量
     * @return array
     */
    public function getEnv()
{
        return $this->env;
    }

    /**
     * 设置环境变量
     * @param array $env
     * @return self
     */
    public function setEnv(array $env)
{
        $env = array_filter($env, function ($value) {
            return !is_array($value);
        });

        $this->env = [];
        foreach ($env as $key => $value) {
            $this->env[(binary) $key] = (binary) $value;
        }

        return $this;
    }

    /**
     * 获取输入
     * @return null|string
     */
    public function getInput()
{
        return $this->input;
    }

    /**
     * 设置输入
     * @param mixed $input
     * @return self
     */
    public function setInput($input)
{
        if ($this->isRunning()) {
            throw new \LogicException('Input can not be set while the process is running.');
        }

        $this->input = Utils::validateInput(sprintf('%s::%s', __CLASS__, __FUNCTION__), $input);

        return $this;
    }

    /**
     * 获取proc_open的选项
     * @return array
     */
    public function getOptions()
{
        return $this->options;
    }

    /**
     * 设置proc_open的选项
     * @param array $options
     * @return self
     */
    public function setOptions(array $options)
{
        $this->options = $options;

        return $this;
    }

    /**
     * 是否兼容windows
     * @return bool
     */
    public function getEnhanceWindowsCompatibility()
{
        return $this->enhanceWindowsCompatibility;
    }

    /**
     * 设置是否兼容windows
     * @param bool $enhance
     * @return self
     */
    public function setEnhanceWindowsCompatibility($enhance)
{
        $this->enhanceWindowsCompatibility = (bool) $enhance;

        return $this;
    }

    /**
     * 返回是否 sigchild 兼容模式激活
     * @return bool
     */
    public function getEnhanceSigchildCompatibility()
{
        return $this->enhanceSigchildCompatibility;
    }

    /**
     * 激活 sigchild 兼容性模式。
     * @param bool $enhance
     * @return self
     */
    public function setEnhanceSigchildCompatibility($enhance)
{
        $this->enhanceSigchildCompatibility = (bool) $enhance;

        return $this;
    }

    /**
     * 是否超时
     */
    public function checkTimeout()
{
        if (self::STATUS_STARTED !== $this->status) {
            return;
        }

        if (null !== $this->timeout && $this->timeout < microtime(true) - $this->starttime) {
            $this->stop();

            throw new ProcessTimeoutException($this, ProcessTimeoutException::TYPE_GENERAL);
        }

        if (null !== $this->idleTimeout && $this->idleTimeout < microtime(true) - $this->lastOutputTime) {
            $this->stop();

            throw new ProcessTimeoutException($this, ProcessTimeoutException::TYPE_IDLE);
        }
    }

    /**
     * 是否支持pty
     * @return bool
     */
    public static function isPtySupported()
{
        static $result;

        if (null !== $result) {
            return $result;
        }

        if ('\\' === DS) {
            return $result = false;
        }

        $proc = @proc_open('echo 1', [['pty'], ['pty'], ['pty']], $pipes);
        if (is_resource($proc)) {
            proc_close($proc);

            return $result = true;
        }

        return $result = false;
    }

    /**
     * 创建所需的 proc_open 的描述符
     * @return array
     */
    private function getDescriptors()
{
        if ('\\' === DS) {
            $this->processPipes = WindowsPipes::create($this, $this->input);
        } else {
            $this->processPipes = UnixPipes::create($this, $this->input);
        }
        $descriptors = $this->processPipes->getDescriptors($this->outputDisabled);

        if (!$this->useFileHandles && $this->enhanceSigchildCompatibility && $this->isSigchildEnabled()) {

            $descriptors = array_merge($descriptors, [['pipe', 'w']]);

            $this->commandline = '(' . $this->commandline . ') 3>/dev/null; code=$?; echo $code >&3; exit $code';
        }

        return $descriptors;
    }

    /**
     * 建立 wait () 使用的回调。
     * @param callable|null $callback
     * @return callable
     */
    protected function buildCallback($callback)
{
        $out      = self::OUT;
        $callback = function ($type, $data) use ($callback, $out) {
            if ($out == $type) {
                $this->addOutput($data);
            } else {
                $this->addErrorOutput($data);
            }

            if (null !== $callback) {
                call_user_func($callback, $type, $data);
            }
        };

        return $callback;
    }

    /**
     * 更新状态
     * @param bool $blocking
     */
    protected function updateStatus($blocking)
{
        if (self::STATUS_STARTED !== $this->status) {
            return;
        }

        $this->processInformation = proc_get_status($this->process);
        $this->captureExitCode();

        $this->readPipes($blocking, '\\' === DS ? !$this->processInformation['running'] : true);

        if (!$this->processInformation['running']) {
            $this->close();
        }
    }

    /**
     * 是否开启 '--enable-sigchild'
     * @return bool
     */
    protected function isSigchildEnabled()
{
        if (null !== self::$sigchild) {
            return self::$sigchild;
        }

        if (!function_exists('phpinfo')) {
            return self::$sigchild = false;
        }

        ob_start();
        phpinfo(INFO_GENERAL);

        return self::$sigchild = false !== strpos(ob_get_clean(), '--enable-sigchild');
    }

    /**
     * 验证是否超时
     * @param int|float|null $timeout
     * @return float|null
     */
    private function validateTimeout($timeout)
{
        $timeout = (float) $timeout;

        if (0.0 === $timeout) {
            $timeout = null;
        } elseif ($timeout < 0) {
            throw new \InvalidArgumentException('The timeout value must be a valid positive integer or float number.');
        }

        return $timeout;
    }

    /**
     * 读取pipes
     * @param bool $blocking
     * @param bool $close
     */
    private function readPipes($blocking, $close)
{
        $result = $this->processPipes->readAndWrite($blocking, $close);

        $callback = $this->callback;
        foreach ($result as $type => $data) {
            if (3 == $type) {
                $this->fallbackExitcode = (int) $data;
            } else {
                $callback(self::STDOUT === $type ? self::OUT : self::ERR, $data);
            }
        }
    }

    /**
     * 捕获退出码
     */
    private function captureExitCode()
{
        if (isset($this->processInformation['exitcode']) && -1 != $this->processInformation['exitcode']) {
            $this->exitcode = $this->processInformation['exitcode'];
        }
    }

    /**
     * 关闭资源
     * @return int 退出码
     */
    private function close()
{
        $this->processPipes->close();
        if (is_resource($this->process)) {
            $exitcode = proc_close($this->process);
        } else {
            $exitcode = -1;
        }

        $this->exitcode = -1 !== $exitcode ? $exitcode : (null !== $this->exitcode ? $this->exitcode : -1);
        $this->status   = self::STATUS_TERMINATED;

        if (-1 === $this->exitcode && null !== $this->fallbackExitcode) {
            $this->exitcode = $this->fallbackExitcode;
        } elseif (-1 === $this->exitcode && $this->processInformation['signaled']
            && 0 < $this->processInformation['termsig']
        ) {
            $this->exitcode = 128 + $this->processInformation['termsig'];
        }

        return $this->exitcode;
    }

    /**
     * 重置数据
     */
    private function resetProcessData()
{
        $this->starttime                    = null;
        $this->callback                     = null;
        $this->exitcode                     = null;
        $this->fallbackExitcode             = null;
        $this->processInformation           = null;
        $this->stdout                       = null;
        $this->stderr                       = null;
        $this->process                      = null;
        $this->latestSignal                 = null;
        $this->status                       = self::STATUS_READY;
        $this->incrementalOutputOffset      = 0;
        $this->incrementalErrorOutputOffset = 0;
    }

    /**
     * 将一个 POSIX 信号发送到进程中。
     * @param int  $signal
     * @param bool $throwException
     * @return bool
     */
    private function doSignal($signal, $throwException)
{
        if (!$this->isRunning()) {
            if ($throwException) {
                throw new \LogicException('Can not send signal on a non running process.');
            }

            return false;
        }

        if ($this->isSigchildEnabled()) {
            if ($throwException) {
                throw new \RuntimeException('This PHP has been compiled with --enable-sigchild. The process can not be signaled.');
            }

            return false;
        }

        if (true !== @proc_terminate($this->process, $signal)) {
            if ($throwException) {
                throw new \RuntimeException(sprintf('Error while sending signal `%s`.', $signal));
            }

            return false;
        }

        $this->latestSignal = $signal;

        return true;
    }

    /**
     * 确保进程已经开启
     * @param string $functionName
     */
    private function requireProcessIsStarted($functionName)
{
        if (!$this->isStarted()) {
            throw new \LogicException(sprintf('Process must be started before calling %s.', $functionName));
        }
    }

    /**
     * 确保进程已经终止
     * @param string $functionName
     */
    private function requireProcessIsTerminated($functionName)
{
        if (!$this->isTerminated()) {
            throw new \LogicException(sprintf('Process must be terminated before calling %s.', $functionName));
        }
    }
}
