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
namespace think\console\command\optimize;

use think\App;
use think\console\Command;
use think\console\Input;
use think\console\input\Option;
use think\console\Output;
use think\Db;

class Schema extends Command
{
    /** @var  Output */
    protected $output;

    protected function configure()
    {
        $this->setName('optimize:schema')
            ->addOption('config', null, Option::VALUE_REQUIRED, 'db config .')
            ->addOption('db', null, Option::VALUE_REQUIRED, 'db name .')
            ->addOption('table', null, Option::VALUE_REQUIRED, 'table name .')
            ->addOption('module', null, Option::VALUE_REQUIRED, 'module name .')
            ->setDescription('Build database schema cache.');
    }

    protected function execute(Input $input, Output $output)
    {
        if (!is_dir(RUNTIME_PATH . 'schema')) {
            @mkdir(RUNTIME_PATH . 'schema', 0755, true);
        }
        $config = [];
        if ($input->hasOption('config')) {
            $config = $input->getOption('config');
        }
        if ($input->hasOption('module')) {
            $module = $input->getOption('module');
            // 读取模型
            $path = APP_PATH . $module . DS . 'model';
            $list = is_dir($path) ? scandir($path) : [];
            $app  = App::$namespace;
            foreach ($list as $file) {
                if (0 === strpos($file, '.')) {
                    continue;
                }
                $class = '\\' . $app . '\\' . $module . '\\model\\' . pathinfo($file, PATHINFO_FILENAME);
                $this->buildModelSchema($class);
            }
            $output->writeln('<info>Succeed!</info>');
            return;
        } elseif ($input->hasOption('table')) {
            $table = $input->getOption('table');
            if (!strpos($table, '.')) {
                $dbName = Db::connect($config)->getConfig('database');
            }
            $tables[] = $table;
        } elseif ($input->hasOption('db')) {
            $dbName = $input->getOption('db');
            $tables = Db::connect($config)->getTables($dbName);
        } elseif (!\think\Config::get('app_multi_module')) {
            $app  = App::$namespace;
            $path = APP_PATH . 'model';
            $list = is_dir($path) ? scandir($path) : [];
            foreach ($list as $file) {
                if (0 === strpos($file, '.')) {
                    continue;
                }
                $class = '\\' . $app . '\\model\\' . pathinfo($file, PATHINFO_FILENAME);
                $this->buildModelSchema($class);
            }
            $output->writeln('<info>Succeed!</info>');
            return;
        } else {
            $tables = Db::connect($config)->getTables();
        }

        $db = isset($dbName) ? $dbName . '.' : '';
        $this->buildDataBaseSchema($tables, $db, $config);

        $output->writeln('<info>Succeed!</info>');
    }

    protected function buildModelSchema($class)
    {
        $reflect = new \ReflectionClass($class);
        if (!$reflect->isAbstract() && $reflect->isSubclassOf('\think\Model')) {
            $table   = $class::getTable();
            $dbName  = $class::getConfig('database');
            $content = '<?php ' . PHP_EOL . 'return ';
            $info    = $class::getConnection()->getFields($table);
            $content .= var_export($info, true) . ';';
            file_put_contents(RUNTIME_PATH . 'schema' . DS . $dbName . '.' . $table . EXT, $content);
        }
    }

    protected function buildDataBaseSchema($tables, $db, $config)
    {
        if ('' == $db) {
            $dbName = Db::connect($config)->getConfig('database') . '.';
        } else {
            $dbName = $db;
        }
        foreach ($tables as $table) {
            $content = '<?php ' . PHP_EOL . 'return ';
            $info    = Db::connect($config)->getFields($db . $table);
            $content .= var_export($info, true) . ';';
            file_put_contents(RUNTIME_PATH . 'schema' . DS . $dbName . $table . EXT, $content);
        }
    }
}
