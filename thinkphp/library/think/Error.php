<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2018 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: 麦当苗儿 <zuojiazi@vip.qq.com> <http://zjzit.cn>
// +----------------------------------------------------------------------

namespace think;

use think\console\Output as ConsoleOutput;
use think\exception\ErrorException;
use think\exception\Handle;
use think\exception\ThrowableError;

class Error
{
    /**
     * 注册异常处理
     * @access public
     * @return void
     */
    public static function register()
    {
        error_reporting(E_ALL);
        set_error_handler([__CLASS__, 'appError']);
        set_exception_handler([__CLASS__, 'appException']);
        register_shutdown_function([__CLASS__, 'appShutdown']);
    }

    /**
     * 异常处理
     * @access public
     * @param  \Exception|\Throwable $e 异常
     * @return void
     */
    public static function appException($e)
    {
        if (!$e instanceof \Exception) {
            $e = new ThrowableError($e);
        }

        $handler = self::getExceptionHandler();
        $handler->report($e);

        if (IS_CLI) {
            $handler->renderForConsole(new ConsoleOutput, $e);
        } else {
            $handler->render($e)->send();
        }
    }

    /**
     * 错误处理
     * @access public
     * @param  integer $errno      错误编号
     * @param  integer $errstr     详细错误信息
     * @param  string  $errfile    出错的文件
     * @param  integer $errline    出错行号
     * @return void
     * @throws ErrorException
     */
    public static function appError($errno, $errstr, $errfile = '', $errline = 0)
    {
        $exception = new ErrorException($errno, $errstr, $errfile, $errline);

        // 符合异常处理的则将错误信息托管至 think\exception\ErrorException
        if (error_reporting() & $errno) {
            throw $exception;
        }

        self::getExceptionHandler()->report($exception);
    }

    /**
     * 异常中止处理
     * @access public
     * @return void
     */
    public static function appShutdown()
    {
        // 将错误信息托管至 think\ErrorException
        if (!is_null($error = error_get_last()) && self::isFatal($error['type'])) {
            self::appException(new ErrorException(
                $error['type'], $error['message'], $error['file'], $error['line']
            ));
        }

        // 写入日志
        Log::save();
    }

    /**
     * 确定错误类型是否致命
     * @access protected
     * @param  int $type 错误类型
     * @return bool
     */
    protected static function isFatal($type)
    {
        return in_array($type, [E_ERROR, E_CORE_ERROR, E_COMPILE_ERROR, E_PARSE]);
    }

    /**
     * 获取异常处理的实例
     * @access public
     * @return Handle
     */
    public static function getExceptionHandler()
    {
        static $handle;

        if (!$handle) {
            // 异常处理 handle
            $class = Config::get('exception_handle');

            if ($class && is_string($class) && class_exists($class) &&
                is_subclass_of($class, "\\think\\exception\\Handle")
            ) {
                $handle = new $class;
            } else {
                $handle = new Handle;

                if ($class instanceof \Closure) {
                    $handle->setRender($class);
                }

            }
        }

        return $handle;
    }
}
