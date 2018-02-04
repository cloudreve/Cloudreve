<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------

/**
 * build测试
 * @author    刘志淳 <chun@engineer.com>
 */

namespace tests\thinkphp\library\think;

use think\Build;

class buildTest extends \PHPUnit_Framework_TestCase
{
    public function testRun()
    {
        $build = [
            // Test run directory
            '__dir__'  => ['runtime/cache', 'runtime/log', 'runtime/temp', 'runtime/template'],
            '__file__' => ['common.php'],

            // Test generation module
            'demo'     => [
                '__file__'   => ['common.php'],
                '__dir__'    => ['behavior', 'controller', 'model', 'view', 'service'],
                'controller' => ['Index', 'Test', 'UserType'],
                'model'      => ['User', 'UserType'],
                'service'    => ['User', 'UserType'],
                'view'       => ['index/index'],
            ],
        ];
        Build::run($build);

        $this->buildFileExists($build);
    }

    protected function buildFileExists($build)
    {
        foreach ($build as $module => $list) {
            if ('__dir__' == $module || '__file__' == $module) {
                foreach ($list as $file) {
                    $this->assertFileExists(APP_PATH . $file);
                }
            } else {
                foreach ($list as $path => $moduleList) {
                    if ('__file__' == $path || '__dir__' == $path) {
                        foreach ($moduleList as $file) {
                            $this->assertFileExists(APP_PATH . $module . '/' . $file);
                        }
                    } else {
                        foreach ($moduleList as $file) {
                            if ('view' == $path) {
                                $file_name = APP_PATH . $module . '/' . $path . '/' . $file . '.html';
                            } else {
                                $file_name = APP_PATH . $module . '/' . $path . '/' . $file . EXT;
                            }
                            $this->assertFileExists($file_name);
                        }
                    }
                }
                $this->assertFileExists(APP_PATH . ($module ? $module . DS : '') . 'config.php');
            }
        }
    }
}
