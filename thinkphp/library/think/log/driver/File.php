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

/**
 * 本地化调试输出到文件
 */
class File
{
    protected $config = [
        'time_format' => ' c ',
        'file_size'   => 2097152,
        'path'        => LOG_PATH,
        'apart_level' => [],
    ];

    protected $writed = [];

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
     * @param array $log 日志信息
     * @return bool
     */
    public function save(array $log = [])
    {
        $cli         = IS_CLI ? '_cli' : '';
        $destination = $this->config['path'] . date('Ym') . DS . date('d') . $cli . '.log';

        $path = dirname($destination);
        !is_dir($path) && mkdir($path, 0755, true);

        $info = '';
        foreach ($log as $type => $val) {
            $level = '';
            foreach ($val as $msg) {
                if (!is_string($msg)) {
                    $msg = var_export($msg, true);
                }
                $level .= '[ ' . $type . ' ] ' . $msg . "\r\n";
            }
            if (in_array($type, $this->config['apart_level'])) {
                // 独立记录的日志级别
                $filename = $path . DS . date('d') . '_' . $type . $cli . '.log';
                $this->write($level, $filename, true);
            } else {
                $info .= $level;
            }
        }
        if ($info) {
            return $this->write($info, $destination);
        }
        return true;
    }

    protected function write($message, $destination, $apart = false)
    {
        //检测日志文件大小，超过配置大小则备份日志文件重新生成
        if (is_file($destination) && floor($this->config['file_size']) <= filesize($destination)) {
            rename($destination, dirname($destination) . DS . time() . '-' . basename($destination));
            $this->writed[$destination] = false;
        }

        if (empty($this->writed[$destination]) && !IS_CLI) {
            if (App::$debug && !$apart) {
                // 获取基本信息
                if (isset($_SERVER['HTTP_HOST'])) {
                    $current_uri = $_SERVER['HTTP_HOST'] . $_SERVER['REQUEST_URI'];
                } else {
                    $current_uri = "cmd:" . implode(' ', $_SERVER['argv']);
                }

                $runtime    = round(microtime(true) - THINK_START_TIME, 10);
                $reqs       = $runtime > 0 ? number_format(1 / $runtime, 2) : '∞';
                $time_str   = ' [运行时间：' . number_format($runtime, 6) . 's][吞吐率：' . $reqs . 'req/s]';
                $memory_use = number_format((memory_get_usage() - THINK_START_MEM) / 1024, 2);
                $memory_str = ' [内存消耗：' . $memory_use . 'kb]';
                $file_load  = ' [文件加载：' . count(get_included_files()) . ']';

                $message = '[ info ] ' . $current_uri . $time_str . $memory_str . $file_load . "\r\n" . $message;
            }
            $now     = date($this->config['time_format']);
            $server  = isset($_SERVER['SERVER_ADDR']) ? $_SERVER['SERVER_ADDR'] : '0.0.0.0';
            $remote  = isset($_SERVER['REMOTE_ADDR']) ? $_SERVER['REMOTE_ADDR'] : '0.0.0.0';
            $method  = isset($_SERVER['REQUEST_METHOD']) ? $_SERVER['REQUEST_METHOD'] : 'CLI';
            $uri     = isset($_SERVER['REQUEST_URI']) ? $_SERVER['REQUEST_URI'] : '';
            $message = "---------------------------------------------------------------\r\n[{$now}] {$server} {$remote} {$method} {$uri}\r\n" . $message;

            $this->writed[$destination] = true;
        }

        if (IS_CLI) {
            $now     = date($this->config['time_format']);
            $message = "[{$now}]" . $message;
        }

        return error_log($message, 3, $destination);
    }

}
