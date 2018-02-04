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
namespace think\console\command\optimize;

use think\console\Command;
use think\console\Input;
use think\console\Output;

class Route extends Command
{
    /** @var  Output */
    protected $output;

    protected function configure()
    {
        $this->setName('optimize:route')
            ->setDescription('Build route cache.');
    }

    protected function execute(Input $input, Output $output)
    {
        file_put_contents(RUNTIME_PATH . 'route.php', $this->buildRouteCache());
        $output->writeln('<info>Succeed!</info>');
    }

    protected function buildRouteCache()
    {
        $files = \think\Config::get('route_config_file');
        foreach ($files as $file) {
            if (is_file(CONF_PATH . $file . CONF_EXT)) {
                $config = include CONF_PATH . $file . CONF_EXT;
                if (is_array($config)) {
                    \think\Route::import($config);
                }
            }
        }
        $rules = \think\Route::rules(true);
        array_walk_recursive($rules, [$this, 'buildClosure']);
        $content = '<?php ' . PHP_EOL . 'return ';
        $content .= var_export($rules, true) . ';';
        $content = str_replace(['\'[__start__', '__end__]\''], '', stripcslashes($content));
        return $content;
    }

    protected function buildClosure(&$value)
    {
        if ($value instanceof \Closure) {
            $reflection = new \ReflectionFunction($value);
            $startLine  = $reflection->getStartLine();
            $endLine    = $reflection->getEndLine();
            $file       = $reflection->getFileName();
            $item       = file($file);
            $content    = '';
            for ($i = $startLine - 1; $i <= $endLine - 1; $i++) {
                $content .= $item[$i];
            }
            $start = strpos($content, 'function');
            $end   = strrpos($content, '}');
            $value = '[__start__' . substr($content, $start, $end - $start + 1) . '__end__]';
        }
    }
}
