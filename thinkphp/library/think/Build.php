<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2017 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------

namespace think;

class Build
{
    /**
     * 根据传入的build资料创建目录和文件
     * @access protected
     * @param  array  $build build列表
     * @param  string $namespace 应用类库命名空间
     * @param  bool   $suffix 类库后缀
     * @return void
     */
    public static function run(array $build = [], $namespace = 'app', $suffix = false)
    {
        // 锁定
        $lockfile = APP_PATH . 'build.lock';
        if (is_writable($lockfile)) {
            return;
        } elseif (!touch($lockfile)) {
            throw new Exception('应用目录[' . APP_PATH . ']不可写，目录无法自动生成！<BR>请手动生成项目目录~', 10006);
        }
        foreach ($build as $module => $list) {
            if ('__dir__' == $module) {
                // 创建目录列表
                self::buildDir($list);
            } elseif ('__file__' == $module) {
                // 创建文件列表
                self::buildFile($list);
            } else {
                // 创建模块
                self::module($module, $list, $namespace, $suffix);
            }
        }
        // 解除锁定
        unlink($lockfile);
    }

    /**
     * 创建目录
     * @access protected
     * @param  array $list 目录列表
     * @return void
     */
    protected static function buildDir($list)
    {
        foreach ($list as $dir) {
            if (!is_dir(APP_PATH . $dir)) {
                // 创建目录
                mkdir(APP_PATH . $dir, 0755, true);
            }
        }
    }

    /**
     * 创建文件
     * @access protected
     * @param  array $list 文件列表
     * @return void
     */
    protected static function buildFile($list)
    {
        foreach ($list as $file) {
            if (!is_dir(APP_PATH . dirname($file))) {
                // 创建目录
                mkdir(APP_PATH . dirname($file), 0755, true);
            }
            if (!is_file(APP_PATH . $file)) {
                file_put_contents(APP_PATH . $file, 'php' == pathinfo($file, PATHINFO_EXTENSION) ? "<?php\n" : '');
            }
        }
    }

    /**
     * 创建模块
     * @access public
     * @param  string $module 模块名
     * @param  array  $list build列表
     * @param  string $namespace 应用类库命名空间
     * @param  bool   $suffix 类库后缀
     * @return void
     */
    public static function module($module = '', $list = [], $namespace = 'app', $suffix = false)
    {
        $module = $module ? $module : '';
        if (!is_dir(APP_PATH . $module)) {
            // 创建模块目录
            mkdir(APP_PATH . $module);
        }
        if (basename(RUNTIME_PATH) != $module) {
            // 创建配置文件和公共文件
            self::buildCommon($module);
            // 创建模块的默认页面
            self::buildHello($module, $namespace, $suffix);
        }
        if (empty($list)) {
            // 创建默认的模块目录和文件
            $list = [
                '__file__' => ['config.php', 'common.php'],
                '__dir__'  => ['controller', 'model', 'view'],
            ];
        }
        // 创建子目录和文件
        foreach ($list as $path => $file) {
            $modulePath = APP_PATH . $module . DS;
            if ('__dir__' == $path) {
                // 生成子目录
                foreach ($file as $dir) {
                    if (!is_dir($modulePath . $dir)) {
                        // 创建目录
                        mkdir($modulePath . $dir, 0755, true);
                    }
                }
            } elseif ('__file__' == $path) {
                // 生成（空白）文件
                foreach ($file as $name) {
                    if (!is_file($modulePath . $name)) {
                        file_put_contents($modulePath . $name, 'php' == pathinfo($name, PATHINFO_EXTENSION) ? "<?php\n" : '');
                    }
                }
            } else {
                // 生成相关MVC文件
                foreach ($file as $val) {
                    $val      = trim($val);
                    $filename = $modulePath . $path . DS . $val . ($suffix ? ucfirst($path) : '') . EXT;
                    $space    = $namespace . '\\' . ($module ? $module . '\\' : '') . $path;
                    $class    = $val . ($suffix ? ucfirst($path) : '');
                    switch ($path) {
                        case 'controller': // 控制器
                            $content = "<?php\nnamespace {$space};\n\nclass {$class}\n{\n\n}";
                            break;
                        case 'model': // 模型
                            $content = "<?php\nnamespace {$space};\n\nuse think\Model;\n\nclass {$class} extends Model\n{\n\n}";
                            break;
                        case 'view': // 视图
                            $filename = $modulePath . $path . DS . $val . '.html';
                            if (!is_dir(dirname($filename))) {
                                // 创建目录
                                mkdir(dirname($filename), 0755, true);
                            }
                            $content = '';
                            break;
                        default:
                            // 其他文件
                            $content = "<?php\nnamespace {$space};\n\nclass {$class}\n{\n\n}";
                    }

                    if (!is_file($filename)) {
                        file_put_contents($filename, $content);
                    }
                }
            }
        }
    }

    /**
     * 创建模块的欢迎页面
     * @access public
     * @param  string $module 模块名
     * @param  string $namespace 应用类库命名空间
     * @param  bool   $suffix 类库后缀
     * @return void
     */
    protected static function buildHello($module, $namespace, $suffix = false)
    {
        $filename = APP_PATH . ($module ? $module . DS : '') . 'controller' . DS . 'Index' . ($suffix ? 'Controller' : '') . EXT;
        if (!is_file($filename)) {
            $content = file_get_contents(THINK_PATH . 'tpl' . DS . 'default_index.tpl');
            $content = str_replace(['{$app}', '{$module}', '{layer}', '{$suffix}'], [$namespace, $module ? $module . '\\' : '', 'controller', $suffix ? 'Controller' : ''], $content);
            if (!is_dir(dirname($filename))) {
                mkdir(dirname($filename), 0755, true);
            }
            file_put_contents($filename, $content);
        }
    }

    /**
     * 创建模块的公共文件
     * @access public
     * @param  string $module 模块名
     * @return void
     */
    protected static function buildCommon($module)
    {
        $filename = CONF_PATH . ($module ? $module . DS : '') . 'config.php';

        if (!is_dir(dirname($filename))) {
            mkdir(dirname($filename, 0755, true));
        }
        if (!is_file($filename)) {
            file_put_contents($filename, "<?php\n//配置文件\nreturn [\n\n];");
        }
        $filename = APP_PATH . ($module ? $module . DS : '') . 'common.php';
        if (!is_file($filename)) {
            file_put_contents($filename, "<?php\n");
        }
    }
}
