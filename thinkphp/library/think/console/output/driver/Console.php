<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK IT ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006-2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: yunwuxin <448901948@qq.com>
// +----------------------------------------------------------------------

namespace think\console\output\driver;

use think\console\Output;
use think\console\output\Formatter;

class Console
{

    /** @var  Resource */
    private $stdout;

    /** @var  Formatter */
    private $formatter;

    private $terminalDimensions;

    /** @var  Output */
    private $output;

    public function __construct(Output $output)
    {
        $this->output    = $output;
        $this->formatter = new Formatter();
        $this->stdout    = $this->openOutputStream();
        $decorated       = $this->hasColorSupport($this->stdout);
        $this->formatter->setDecorated($decorated);
    }

    public function getFormatter()
    {
        return $this->formatter;
    }

    public function setDecorated($decorated)
    {
        $this->formatter->setDecorated($decorated);
    }

    public function write($messages, $newline = false, $type = Output::OUTPUT_NORMAL, $stream = null)
    {
        if (Output::VERBOSITY_QUIET === $this->output->getVerbosity()) {
            return;
        }

        $messages = (array) $messages;

        foreach ($messages as $message) {
            switch ($type) {
                case Output::OUTPUT_NORMAL:
                    $message = $this->formatter->format($message);
                    break;
                case Output::OUTPUT_RAW:
                    break;
                case Output::OUTPUT_PLAIN:
                    $message = strip_tags($this->formatter->format($message));
                    break;
                default:
                    throw new \InvalidArgumentException(sprintf('Unknown output type given (%s)', $type));
            }

            $this->doWrite($message, $newline, $stream);
        }
    }

    public function renderException(\Exception $e)
    {
        $stderr    = $this->openErrorStream();
        $decorated = $this->hasColorSupport($stderr);
        $this->formatter->setDecorated($decorated);

        do {
            $title = sprintf('  [%s]  ', get_class($e));

            $len = $this->stringWidth($title);

            $width = $this->getTerminalWidth() ? $this->getTerminalWidth() - 1 : PHP_INT_MAX;

            if (defined('HHVM_VERSION') && $width > 1 << 31) {
                $width = 1 << 31;
            }
            $lines = [];
            foreach (preg_split('/\r?\n/', $e->getMessage()) as $line) {
                foreach ($this->splitStringByWidth($line, $width - 4) as $line) {

                    $lineLength = $this->stringWidth(preg_replace('/\[[^m]*m/', '', $line)) + 4;
                    $lines[]    = [$line, $lineLength];

                    $len = max($lineLength, $len);
                }
            }

            $messages   = ['', ''];
            $messages[] = $emptyLine = sprintf('<error>%s</error>', str_repeat(' ', $len));
            $messages[] = sprintf('<error>%s%s</error>', $title, str_repeat(' ', max(0, $len - $this->stringWidth($title))));
            foreach ($lines as $line) {
                $messages[] = sprintf('<error>  %s  %s</error>', $line[0], str_repeat(' ', $len - $line[1]));
            }
            $messages[] = $emptyLine;
            $messages[] = '';
            $messages[] = '';

            $this->write($messages, true, Output::OUTPUT_NORMAL, $stderr);

            if (Output::VERBOSITY_VERBOSE <= $this->output->getVerbosity()) {
                $this->write('<comment>Exception trace:</comment>', true, Output::OUTPUT_NORMAL, $stderr);

                // exception related properties
                $trace = $e->getTrace();
                array_unshift($trace, [
                    'function' => '',
                    'file'     => $e->getFile() !== null ? $e->getFile() : 'n/a',
                    'line'     => $e->getLine() !== null ? $e->getLine() : 'n/a',
                    'args'     => [],
                ]);

                for ($i = 0, $count = count($trace); $i < $count; ++$i) {
                    $class    = isset($trace[$i]['class']) ? $trace[$i]['class'] : '';
                    $type     = isset($trace[$i]['type']) ? $trace[$i]['type'] : '';
                    $function = $trace[$i]['function'];
                    $file     = isset($trace[$i]['file']) ? $trace[$i]['file'] : 'n/a';
                    $line     = isset($trace[$i]['line']) ? $trace[$i]['line'] : 'n/a';

                    $this->write(sprintf(' %s%s%s() at <info>%s:%s</info>', $class, $type, $function, $file, $line), true, Output::OUTPUT_NORMAL, $stderr);
                }

                $this->write('', true, Output::OUTPUT_NORMAL, $stderr);
                $this->write('', true, Output::OUTPUT_NORMAL, $stderr);
            }
        } while ($e = $e->getPrevious());

    }

    /**
     * 获取终端宽度
     * @return int|null
     */
    protected function getTerminalWidth()
    {
        $dimensions = $this->getTerminalDimensions();

        return $dimensions[0];
    }

    /**
     * 获取终端高度
     * @return int|null
     */
    protected function getTerminalHeight()
    {
        $dimensions = $this->getTerminalDimensions();

        return $dimensions[1];
    }

    /**
     * 获取当前终端的尺寸
     * @return array
     */
    public function getTerminalDimensions()
    {
        if ($this->terminalDimensions) {
            return $this->terminalDimensions;
        }

        if ('\\' === DS) {
            if (preg_match('/^(\d+)x\d+ \(\d+x(\d+)\)$/', trim(getenv('ANSICON')), $matches)) {
                return [(int) $matches[1], (int) $matches[2]];
            }
            if (preg_match('/^(\d+)x(\d+)$/', $this->getMode(), $matches)) {
                return [(int) $matches[1], (int) $matches[2]];
            }
        }

        if ($sttyString = $this->getSttyColumns()) {
            if (preg_match('/rows.(\d+);.columns.(\d+);/i', $sttyString, $matches)) {
                return [(int) $matches[2], (int) $matches[1]];
            }
            if (preg_match('/;.(\d+).rows;.(\d+).columns/i', $sttyString, $matches)) {
                return [(int) $matches[2], (int) $matches[1]];
            }
        }

        return [null, null];
    }

    /**
     * 获取stty列数
     * @return string
     */
    private function getSttyColumns()
    {
        if (!function_exists('proc_open')) {
            return;
        }

        $descriptorspec = [1 => ['pipe', 'w'], 2 => ['pipe', 'w']];
        $process        = proc_open('stty -a | grep columns', $descriptorspec, $pipes, null, null, ['suppress_errors' => true]);
        if (is_resource($process)) {
            $info = stream_get_contents($pipes[1]);
            fclose($pipes[1]);
            fclose($pipes[2]);
            proc_close($process);

            return $info;
        }
        return;
    }

    /**
     * 获取终端模式
     * @return string <width>x<height> 或 null
     */
    private function getMode()
    {
        if (!function_exists('proc_open')) {
            return;
        }

        $descriptorspec = [1 => ['pipe', 'w'], 2 => ['pipe', 'w']];
        $process        = proc_open('mode CON', $descriptorspec, $pipes, null, null, ['suppress_errors' => true]);
        if (is_resource($process)) {
            $info = stream_get_contents($pipes[1]);
            fclose($pipes[1]);
            fclose($pipes[2]);
            proc_close($process);

            if (preg_match('/--------+\r?\n.+?(\d+)\r?\n.+?(\d+)\r?\n/', $info, $matches)) {
                return $matches[2] . 'x' . $matches[1];
            }
        }
        return;
    }

    private function stringWidth($string)
    {
        if (!function_exists('mb_strwidth')) {
            return strlen($string);
        }

        if (false === $encoding = mb_detect_encoding($string)) {
            return strlen($string);
        }

        return mb_strwidth($string, $encoding);
    }

    private function splitStringByWidth($string, $width)
    {
        if (!function_exists('mb_strwidth')) {
            return str_split($string, $width);
        }

        if (false === $encoding = mb_detect_encoding($string)) {
            return str_split($string, $width);
        }

        $utf8String = mb_convert_encoding($string, 'utf8', $encoding);
        $lines      = [];
        $line       = '';
        foreach (preg_split('//u', $utf8String) as $char) {
            if (mb_strwidth($line . $char, 'utf8') <= $width) {
                $line .= $char;
                continue;
            }
            $lines[] = str_pad($line, $width);
            $line    = $char;
        }
        if (strlen($line)) {
            $lines[] = count($lines) ? str_pad($line, $width) : $line;
        }

        mb_convert_variables($encoding, 'utf8', $lines);

        return $lines;
    }

    private function isRunningOS400()
    {
        $checks = [
            function_exists('php_uname') ? php_uname('s') : '',
            getenv('OSTYPE'),
            PHP_OS,
        ];
        return false !== stripos(implode(';', $checks), 'OS400');
    }

    /**
     * 当前环境是否支持写入控制台输出到stdout.
     *
     * @return bool
     */
    protected function hasStdoutSupport()
    {
        return false === $this->isRunningOS400();
    }

    /**
     * 当前环境是否支持写入控制台输出到stderr.
     *
     * @return bool
     */
    protected function hasStderrSupport()
    {
        return false === $this->isRunningOS400();
    }

    /**
     * @return resource
     */
    private function openOutputStream()
    {
        if (!$this->hasStdoutSupport()) {
            return fopen('php://output', 'w');
        }
        return @fopen('php://stdout', 'w') ?: fopen('php://output', 'w');
    }

    /**
     * @return resource
     */
    private function openErrorStream()
    {
        return fopen($this->hasStderrSupport() ? 'php://stderr' : 'php://output', 'w');
    }

    /**
     * 将消息写入到输出。
     * @param string $message 消息
     * @param bool   $newline 是否另起一行
     * @param null   $stream
     */
    protected function doWrite($message, $newline, $stream = null)
    {
        if (null === $stream) {
            $stream = $this->stdout;
        }
        if (false === @fwrite($stream, $message . ($newline ? PHP_EOL : ''))) {
            throw new \RuntimeException('Unable to write output.');
        }

        fflush($stream);
    }

    /**
     * 是否支持着色
     * @param $stream
     * @return bool
     */
    protected function hasColorSupport($stream)
    {
        if (DIRECTORY_SEPARATOR === '\\') {
            return
            '10.0.10586' === PHP_WINDOWS_VERSION_MAJOR . '.' . PHP_WINDOWS_VERSION_MINOR . '.' . PHP_WINDOWS_VERSION_BUILD
            || false !== getenv('ANSICON')
            || 'ON' === getenv('ConEmuANSI')
            || 'xterm' === getenv('TERM');
        }

        return function_exists('posix_isatty') && @posix_isatty($stream);
    }

}
