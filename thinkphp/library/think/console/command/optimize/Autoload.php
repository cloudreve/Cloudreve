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
use think\Config;
use think\console\Command;
use think\console\Input;
use think\console\Output;

class Autoload extends Command
{

    protected function configure()
    {
        $this->setName('optimize:autoload')
            ->setDescription('Optimizes PSR0 and PSR4 packages to be loaded with classmaps too, good for production.');
    }

    protected function execute(Input $input, Output $output)
    {

        $classmapFile = <<<EOF
<?php
/**
 * 类库映射
 */

return [

EOF;

        $namespacesToScan = [
            App::$namespace . '\\' => realpath(rtrim(APP_PATH)),
            'think\\'              => LIB_PATH . 'think',
            'behavior\\'           => LIB_PATH . 'behavior',
            'traits\\'             => LIB_PATH . 'traits',
            ''                     => realpath(rtrim(EXTEND_PATH)),
        ];

        $root_namespace = Config::get('root_namespace');
        foreach ($root_namespace as $namespace => $dir) {
            $namespacesToScan[$namespace . '\\'] = realpath($dir);
        }

        krsort($namespacesToScan);
        $classMap = [];
        foreach ($namespacesToScan as $namespace => $dir) {

            if (!is_dir($dir)) {
                continue;
            }

            $namespaceFilter = $namespace === '' ? null : $namespace;
            $classMap        = $this->addClassMapCode($dir, $namespaceFilter, $classMap);
        }

        ksort($classMap);
        foreach ($classMap as $class => $code) {
            $classmapFile .= '    ' . var_export($class, true) . ' => ' . $code;
        }
        $classmapFile .= "];\n";

        if (!is_dir(RUNTIME_PATH)) {
            @mkdir(RUNTIME_PATH, 0755, true);
        }

        file_put_contents(RUNTIME_PATH . 'classmap' . EXT, $classmapFile);

        $output->writeln('<info>Succeed!</info>');
    }

    protected function addClassMapCode($dir, $namespace, $classMap)
    {
        foreach ($this->createMap($dir, $namespace) as $class => $path) {

            $pathCode = $this->getPathCode($path) . ",\n";

            if (!isset($classMap[$class])) {
                $classMap[$class] = $pathCode;
            } elseif ($classMap[$class] !== $pathCode && !preg_match('{/(test|fixture|example|stub)s?/}i', strtr($classMap[$class] . ' ' . $path, '\\', '/'))) {
                $this->output->writeln(
                    '<warning>Warning: Ambiguous class resolution, "' . $class . '"' .
                    ' was found in both "' . str_replace(["',\n"], [
                        '',
                    ], $classMap[$class]) . '" and "' . $path . '", the first will be used.</warning>'
                );
            }
        }
        return $classMap;
    }

    protected function getPathCode($path)
    {

        $baseDir    = '';
        $libPath    = $this->normalizePath(realpath(LIB_PATH));
        $appPath    = $this->normalizePath(realpath(APP_PATH));
        $extendPath = $this->normalizePath(realpath(EXTEND_PATH));
        $rootPath   = $this->normalizePath(realpath(ROOT_PATH));
        $path       = $this->normalizePath($path);

        if ($libPath !== null && strpos($path, $libPath . '/') === 0) {
            $path    = substr($path, strlen(LIB_PATH));
            $baseDir = 'LIB_PATH';
        } elseif ($appPath !== null && strpos($path, $appPath . '/') === 0) {
            $path    = substr($path, strlen($appPath) + 1);
            $baseDir = 'APP_PATH';
        } elseif ($extendPath !== null && strpos($path, $extendPath . '/') === 0) {
            $path    = substr($path, strlen($extendPath) + 1);
            $baseDir = 'EXTEND_PATH';
        } elseif ($rootPath !== null && strpos($path, $rootPath . '/') === 0) {
            $path    = substr($path, strlen($rootPath) + 1);
            $baseDir = 'ROOT_PATH';
        }

        if ($path !== false) {
            $baseDir .= " . ";
        }

        return $baseDir . (($path !== false) ? var_export($path, true) : "");
    }

    protected function normalizePath($path)
    {
        if ($path === false) {
            return;
        }
        $parts    = [];
        $path     = strtr($path, '\\', '/');
        $prefix   = '';
        $absolute = false;

        if (preg_match('{^([0-9a-z]+:(?://(?:[a-z]:)?)?)}i', $path, $match)) {
            $prefix = $match[1];
            $path   = substr($path, strlen($prefix));
        }

        if (substr($path, 0, 1) === '/') {
            $absolute = true;
            $path     = substr($path, 1);
        }

        $up = false;
        foreach (explode('/', $path) as $chunk) {
            if ('..' === $chunk && ($absolute || $up)) {
                array_pop($parts);
                $up = !(empty($parts) || '..' === end($parts));
            } elseif ('.' !== $chunk && '' !== $chunk) {
                $parts[] = $chunk;
                $up      = '..' !== $chunk;
            }
        }

        return $prefix . ($absolute ? '/' : '') . implode('/', $parts);
    }

    protected function createMap($path, $namespace = null)
    {
        if (is_string($path)) {
            if (is_file($path)) {
                $path = [new \SplFileInfo($path)];
            } elseif (is_dir($path)) {

                $objects = new \RecursiveIteratorIterator(new \RecursiveDirectoryIterator($path), \RecursiveIteratorIterator::SELF_FIRST);

                $path = [];

                /** @var \SplFileInfo $object */
                foreach ($objects as $object) {
                    if ($object->isFile() && $object->getExtension() == 'php') {
                        $path[] = $object;
                    }
                }
            } else {
                throw new \RuntimeException(
                    'Could not scan for classes inside "' . $path .
                    '" which does not appear to be a file nor a folder'
                );
            }
        }

        $map = [];

        /** @var \SplFileInfo $file */
        foreach ($path as $file) {
            $filePath = $file->getRealPath();

            if (pathinfo($filePath, PATHINFO_EXTENSION) != 'php') {
                continue;
            }

            $classes = $this->findClasses($filePath);

            foreach ($classes as $class) {
                if (null !== $namespace && 0 !== strpos($class, $namespace)) {
                    continue;
                }

                if (!isset($map[$class])) {
                    $map[$class] = $filePath;
                } elseif ($map[$class] !== $filePath && !preg_match('{/(test|fixture|example|stub)s?/}i', strtr($map[$class] . ' ' . $filePath, '\\', '/'))) {
                    $this->output->writeln(
                        '<warning>Warning: Ambiguous class resolution, "' . $class . '"' .
                        ' was found in both "' . $map[$class] . '" and "' . $filePath . '", the first will be used.</warning>'
                    );
                }
            }
        }

        return $map;
    }

    protected function findClasses($path)
    {
        $extraTypes = '|trait';

        $contents = @php_strip_whitespace($path);
        if (!$contents) {
            if (!file_exists($path)) {
                $message = 'File at "%s" does not exist, check your classmap definitions';
            } elseif (!is_readable($path)) {
                $message = 'File at "%s" is not readable, check its permissions';
            } elseif ('' === trim(file_get_contents($path))) {
                return [];
            } else {
                $message = 'File at "%s" could not be parsed as PHP, it may be binary or corrupted';
            }
            $error = error_get_last();
            if (isset($error['message'])) {
                $message .= PHP_EOL . 'The following message may be helpful:' . PHP_EOL . $error['message'];
            }
            throw new \RuntimeException(sprintf($message, $path));
        }

        if (!preg_match('{\b(?:class|interface' . $extraTypes . ')\s}i', $contents)) {
            return [];
        }

        // strip heredocs/nowdocs
        $contents = preg_replace('{<<<\s*(\'?)(\w+)\\1(?:\r\n|\n|\r)(?:.*?)(?:\r\n|\n|\r)\\2(?=\r\n|\n|\r|;)}s', 'null', $contents);
        // strip strings
        $contents = preg_replace('{"[^"\\\\]*+(\\\\.[^"\\\\]*+)*+"|\'[^\'\\\\]*+(\\\\.[^\'\\\\]*+)*+\'}s', 'null', $contents);
        // strip leading non-php code if needed
        if (substr($contents, 0, 2) !== '<?') {
            $contents = preg_replace('{^.+?<\?}s', '<?', $contents, 1, $replacements);
            if ($replacements === 0) {
                return [];
            }
        }
        // strip non-php blocks in the file
        $contents = preg_replace('{\?>.+<\?}s', '?><?', $contents);
        // strip trailing non-php code if needed
        $pos = strrpos($contents, '?>');
        if (false !== $pos && false === strpos(substr($contents, $pos), '<?')) {
            $contents = substr($contents, 0, $pos);
        }

        preg_match_all('{
            (?:
                 \b(?<![\$:>])(?P<type>class|interface' . $extraTypes . ') \s++ (?P<name>[a-zA-Z_\x7f-\xff:][a-zA-Z0-9_\x7f-\xff:\-]*+)
               | \b(?<![\$:>])(?P<ns>namespace) (?P<nsname>\s++[a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*+(?:\s*+\\\\\s*+[a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*+)*+)? \s*+ [\{;]
            )
        }ix', $contents, $matches);

        $classes   = [];
        $namespace = '';

        for ($i = 0, $len = count($matches['type']); $i < $len; $i++) {
            if (!empty($matches['ns'][$i])) {
                $namespace = str_replace([' ', "\t", "\r", "\n"], '', $matches['nsname'][$i]) . '\\';
            } else {
                $name = $matches['name'][$i];
                if ($name[0] === ':') {
                    $name = 'xhp' . substr(str_replace(['-', ':'], ['_', '__'], $name), 1);
                } elseif ($matches['type'][$i] === 'enum') {
                    $name = rtrim($name, ':');
                }
                $classes[] = ltrim($namespace . $name, '\\');
            }
        }

        return $classes;
    }

}
