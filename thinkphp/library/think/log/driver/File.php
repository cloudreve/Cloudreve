<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK IT ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006-2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------

namespace think\log\driver;

use think\App;
use think\Request;

/**
 * 本地化调试输出到文件
 */
class File
{
    protected $config = [
        'time_format' => ' c ',
        'single'      => false,
        'file_size'   => 2097152,
        'path'        => LOG_PATH,
        'apart_level' => [],
        'max_files'   => 0,
        'json'        => false,
    ];

    // 实例化并传入参数
    public function __construct($config = [])
    {
        if (is_array($config)) {
            $this->config = array_merge($this->config, $config);
        }
    }

    /**
     * 日志写入接口
     * @access public
     * @param  array    $log 日志信息
     * @param  bool     $append 是否追加请求信息
     * @return bool
     */
    public function save(array $log = [], $append = false)
    {
        $destination = $this->getMasterLogFile();

        $path = dirname($destination);
        !is_dir($path) && mkdir($path, 0755, true);

        $info = [];
        foreach ($log as $type => $val) {

            foreach ($val as $msg) {
                if (!is_string($msg)) {
                    $msg = var_export($msg, true);
                }

                $info[$type][] = $this->config['json'] ? $msg : '[ ' . $type . ' ] ' . $msg;
            }

            if (!$this->config['json'] && (true === $this->config['apart_level'] || in_array($type, $this->config['apart_level']))) {
                // 独立记录的日志级别
                $filename = $this->getApartLevelFile($path, $type);

                $this->write($info[$type], $filename, true, $append);
                unset($info[$type]);
            }
        }

        if ($info) {
            return $this->write($info, $destination, false, $append);
        }

        return true;
    }

    /**
     * 获取主日志文件名
     * @access public
     * @return string
     */
    protected function getMasterLogFile()
    {
        if ($this->config['single']) {
            $name = is_string($this->config['single']) ? $this->config['single'] : 'single';

            $destination = $this->config['path'] . $name . '.log';
        } else {
            $cli = PHP_SAPI == 'cli' ? '_cli' : '';

            if ($this->config['max_files']) {
                $filename = date('Ymd') . $cli . '.log';
                $files    = glob($this->config['path'] . '*.log');

                try {
                    if (count($files) > $this->config['max_files']) {
                        unlink($files[0]);
                    }
                } catch (\Exception $e) {
                }
            } else {
                $filename = date('Ym') . DIRECTORY_SEPARATOR . date('d') . $cli . '.log';
            }

            $destination = $this->config['path'] . $filename;
        }

        return $destination;
    }

    /**
     * 获取独立日志文件名
     * @access public
     * @param  string $path 日志目录
     * @param  string $type 日志类型
     * @return string
     */
    protected function getApartLevelFile($path, $type)
    {
        $cli = PHP_SAPI == 'cli' ? '_cli' : '';

        if ($this->config['single']) {
            $name = is_string($this->config['single']) ? $this->config['single'] : 'single';

            $name .= '_' . $type;
        } elseif ($this->config['max_files']) {
            $name = date('Ymd') . '_' . $type . $cli;
        } else {
            $name = date('d') . '_' . $type . $cli;
        }

        return $path . DIRECTORY_SEPARATOR . $name . '.log';
    }

    /**
     * 日志写入
     * @access protected
     * @param  array     $message 日志信息
     * @param  string    $destination 日志文件
     * @param  bool      $apart 是否独立文件写入
     * @param  bool      $append 是否追加请求信息
     * @return bool
     */
    protected function write($message, $destination, $apart = false, $append = false)
    {
        // 检测日志文件大小，超过配置大小则备份日志文件重新生成
        $this->checkLogSize($destination);

        // 日志信息封装
        $info['timestamp'] = date($this->config['time_format']);

        foreach ($message as $type => $msg) {
            $info[$type] = is_array($msg) ? implode("\r\n", $msg) : $msg;
        }

        if (PHP_SAPI == 'cli') {
            $message = $this->parseCliLog($info);
        } else {
            // 添加调试日志
            $this->getDebugLog($info, $append, $apart);

            $message = $this->parseLog($info);
        }

        return error_log($message, 3, $destination);
    }

    /**
     * 检查日志文件大小并自动生成备份文件
     * @access protected
     * @param  string    $destination 日志文件
     * @return void
     */
    protected function checkLogSize($destination)
    {
        if (is_file($destination) && floor($this->config['file_size']) <= filesize($destination)) {
            try {
                rename($destination, dirname($destination) . DIRECTORY_SEPARATOR . time() . '-' . basename($destination));
            } catch (\Exception $e) {
            }
        }
    }

    /**
     * CLI日志解析
     * @access protected
     * @param  array     $info 日志信息
     * @return string
     */
    protected function parseCliLog($info)
    {
        if ($this->config['json']) {
            $message = json_encode($info, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) . "\r\n";
        } else {
            $now = $info['timestamp'];
            unset($info['timestamp']);

            $message = implode("\r\n", $info);

            $message = "[{$now}]" . $message . "\r\n";
        }

        return $message;
    }

    /**
     * 解析日志
     * @access protected
     * @param  array     $info 日志信息
     * @return string
     */
    protected function parseLog($info)
    {
        $request     = Request::instance();
        $requestInfo = [
            'ip'     => $request->ip(),
            'method' => $request->method(),
            'host'   => $request->host(),
            'uri'    => $request->url(),
        ];

        if ($this->config['json']) {
            $info = $requestInfo + $info;
            return json_encode($info, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) . "\r\n";
        }

        array_unshift($info, "---------------------------------------------------------------\r\n[{$info['timestamp']}] {$requestInfo['ip']} {$requestInfo['method']} {$requestInfo['host']}{$requestInfo['uri']}");
        unset($info['timestamp']);

        return implode("\r\n", $info) . "\r\n";
    }

    protected function getDebugLog(&$info, $append, $apart)
    {
        if (App::$debug && $append) {

            if ($this->config['json']) {
                // 获取基本信息
                $runtime = round(microtime(true) - THINK_START_TIME, 10);
                $reqs    = $runtime > 0 ? number_format(1 / $runtime, 2) : '∞';

                $memory_use = number_format((memory_get_usage() - THINK_START_MEM) / 1024, 2);

                $info = [
                    'runtime' => number_format($runtime, 6) . 's',
                    'reqs'    => $reqs . 'req/s',
                    'memory'  => $memory_use . 'kb',
                    'file'    => count(get_included_files()),
                ] + $info;

            } elseif (!$apart) {
                // 增加额外的调试信息
                $runtime = round(microtime(true) - THINK_START_TIME, 10);
                $reqs    = $runtime > 0 ? number_format(1 / $runtime, 2) : '∞';

                $memory_use = number_format((memory_get_usage() - THINK_START_MEM) / 1024, 2);

                $time_str   = '[运行时间：' . number_format($runtime, 6) . 's] [吞吐率：' . $reqs . 'req/s]';
                $memory_str = ' [内存消耗：' . $memory_use . 'kb]';
                $file_load  = ' [文件加载：' . count(get_included_files()) . ']';

                array_unshift($info, $time_str . $memory_str . $file_load);
            }
        }
    }
}
