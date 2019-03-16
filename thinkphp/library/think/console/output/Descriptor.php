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

namespace think\console\output;

use think\Console;
use think\console\Command;
use think\console\input\Argument as InputArgument;
use think\console\input\Definition as InputDefinition;
use think\console\input\Option as InputOption;
use think\console\Output;
use think\console\output\descriptor\Console as ConsoleDescription;

class Descriptor
{

    /**
     * @var Output
     */
    protected $output;

    /**
     * {@inheritdoc}
     */
    public function describe(Output $output, $object, array $options = [])
    {
        $this->output = $output;

        switch (true) {
            case $object instanceof InputArgument:
                $this->describeInputArgument($object, $options);
                break;
            case $object instanceof InputOption:
                $this->describeInputOption($object, $options);
                break;
            case $object instanceof InputDefinition:
                $this->describeInputDefinition($object, $options);
                break;
            case $object instanceof Command:
                $this->describeCommand($object, $options);
                break;
            case $object instanceof Console:
                $this->describeConsole($object, $options);
                break;
            default:
                throw new \InvalidArgumentException(sprintf('Object of type "%s" is not describable.', get_class($object)));
        }
    }

    /**
     * 输出内容
     * @param string $content
     * @param bool   $decorated
     */
    protected function write($content, $decorated = false)
    {
        $this->output->write($content, false, $decorated ? Output::OUTPUT_NORMAL : Output::OUTPUT_RAW);
    }

    /**
     * 描述参数
     * @param InputArgument $argument
     * @param array         $options
     * @return void
     */
    protected function describeInputArgument(InputArgument $argument, array $options = [])
    {
        if (null !== $argument->getDefault()
            && (!is_array($argument->getDefault())
                || count($argument->getDefault()))
        ) {
            $default = sprintf('<comment> [default: %s]</comment>', $this->formatDefaultValue($argument->getDefault()));
        } else {
            $default = '';
        }

        $totalWidth   = isset($options['total_width']) ? $options['total_width'] : strlen($argument->getName());
        $spacingWidth = $totalWidth - strlen($argument->getName()) + 2;

        $this->writeText(sprintf("  <info>%s</info>%s%s%s", $argument->getName(), str_repeat(' ', $spacingWidth), // + 17 = 2 spaces + <info> + </info> + 2 spaces
            preg_replace('/\s*\R\s*/', PHP_EOL . str_repeat(' ', $totalWidth + 17), $argument->getDescription()), $default), $options);
    }

    /**
     * 描述选项
     * @param InputOption $option
     * @param array       $options
     * @return void
     */
    protected function describeInputOption(InputOption $option, array $options = [])
    {
        if ($option->acceptValue() && null !== $option->getDefault()
            && (!is_array($option->getDefault())
                || count($option->getDefault()))
        ) {
            $default = sprintf('<comment> [default: %s]</comment>', $this->formatDefaultValue($option->getDefault()));
        } else {
            $default = '';
        }

        $value = '';
        if ($option->acceptValue()) {
            $value = '=' . strtoupper($option->getName());

            if ($option->isValueOptional()) {
                $value = '[' . $value . ']';
            }
        }

        $totalWidth = isset($options['total_width']) ? $options['total_width'] : $this->calculateTotalWidthForOptions([$option]);
        $synopsis   = sprintf('%s%s', $option->getShortcut() ? sprintf('-%s, ', $option->getShortcut()) : '    ', sprintf('--%s%s', $option->getName(), $value));

        $spacingWidth = $totalWidth - strlen($synopsis) + 2;

        $this->writeText(sprintf("  <info>%s</info>%s%s%s%s", $synopsis, str_repeat(' ', $spacingWidth), // + 17 = 2 spaces + <info> + </info> + 2 spaces
            preg_replace('/\s*\R\s*/', "\n" . str_repeat(' ', $totalWidth + 17), $option->getDescription()), $default, $option->isArray() ? '<comment> (multiple values allowed)</comment>' : ''), $options);
    }

    /**
     * 描述输入
     * @param InputDefinition $definition
     * @param array           $options
     * @return void
     */
    protected function describeInputDefinition(InputDefinition $definition, array $options = [])
    {
        $totalWidth = $this->calculateTotalWidthForOptions($definition->getOptions());
        foreach ($definition->getArguments() as $argument) {
            $totalWidth = max($totalWidth, strlen($argument->getName()));
        }

        if ($definition->getArguments()) {
            $this->writeText('<comment>Arguments:</comment>', $options);
            $this->writeText("\n");
            foreach ($definition->getArguments() as $argument) {
                $this->describeInputArgument($argument, array_merge($options, ['total_width' => $totalWidth]));
                $this->writeText("\n");
            }
        }

        if ($definition->getArguments() && $definition->getOptions()) {
            $this->writeText("\n");
        }

        if ($definition->getOptions()) {
            $laterOptions = [];

            $this->writeText('<comment>Options:</comment>', $options);
            foreach ($definition->getOptions() as $option) {
                if (strlen($option->getShortcut()) > 1) {
                    $laterOptions[] = $option;
                    continue;
                }
                $this->writeText("\n");
                $this->describeInputOption($option, array_merge($options, ['total_width' => $totalWidth]));
            }
            foreach ($laterOptions as $option) {
                $this->writeText("\n");
                $this->describeInputOption($option, array_merge($options, ['total_width' => $totalWidth]));
            }
        }
    }

    /**
     * 描述指令
     * @param Command $command
     * @param array   $options
     * @return void
     */
    protected function describeCommand(Command $command, array $options = [])
    {
        $command->getSynopsis(true);
        $command->getSynopsis(false);
        $command->mergeConsoleDefinition(false);

        $this->writeText('<comment>Usage:</comment>', $options);
        foreach (array_merge([$command->getSynopsis(true)], $command->getAliases(), $command->getUsages()) as $usage) {
            $this->writeText("\n");
            $this->writeText('  ' . $usage, $options);
        }
        $this->writeText("\n");

        $definition = $command->getNativeDefinition();
        if ($definition->getOptions() || $definition->getArguments()) {
            $this->writeText("\n");
            $this->describeInputDefinition($definition, $options);
            $this->writeText("\n");
        }

        if ($help = $command->getProcessedHelp()) {
            $this->writeText("\n");
            $this->writeText('<comment>Help:</comment>', $options);
            $this->writeText("\n");
            $this->writeText(' ' . str_replace("\n", "\n ", $help), $options);
            $this->writeText("\n");
        }
    }

    /**
     * 描述控制台
     * @param Console $console
     * @param array   $options
     * @return void
     */
    protected function describeConsole(Console $console, array $options = [])
    {
        $describedNamespace = isset($options['namespace']) ? $options['namespace'] : null;
        $description        = new ConsoleDescription($console, $describedNamespace);

        if (isset($options['raw_text']) && $options['raw_text']) {
            $width = $this->getColumnWidth($description->getCommands());

            foreach ($description->getCommands() as $command) {
                $this->writeText(sprintf("%-${width}s %s", $command->getName(), $command->getDescription()), $options);
                $this->writeText("\n");
            }
        } else {
            if ('' != $help = $console->getHelp()) {
                $this->writeText("$help\n\n", $options);
            }

            $this->writeText("<comment>Usage:</comment>\n", $options);
            $this->writeText("  command [options] [arguments]\n\n", $options);

            $this->describeInputDefinition(new InputDefinition($console->getDefinition()->getOptions()), $options);

            $this->writeText("\n");
            $this->writeText("\n");

            $width = $this->getColumnWidth($description->getCommands());

            if ($describedNamespace) {
                $this->writeText(sprintf('<comment>Available commands for the "%s" namespace:</comment>', $describedNamespace), $options);
            } else {
                $this->writeText('<comment>Available commands:</comment>', $options);
            }

            // add commands by namespace
            foreach ($description->getNamespaces() as $namespace) {
                if (!$describedNamespace && ConsoleDescription::GLOBAL_NAMESPACE !== $namespace['id']) {
                    $this->writeText("\n");
                    $this->writeText(' <comment>' . $namespace['id'] . '</comment>', $options);
                }

                foreach ($namespace['commands'] as $name) {
                    $this->writeText("\n");
                    $spacingWidth = $width - strlen($name);
                    $this->writeText(sprintf("  <info>%s</info>%s%s", $name, str_repeat(' ', $spacingWidth), $description->getCommand($name)
                            ->getDescription()), $options);
                }
            }

            $this->writeText("\n");
        }
    }

    /**
     * {@inheritdoc}
     */
    private function writeText($content, array $options = [])
    {
        $this->write(isset($options['raw_text'])
            && $options['raw_text'] ? strip_tags($content) : $content, isset($options['raw_output']) ? !$options['raw_output'] : true);
    }

    /**
     * 格式化
     * @param mixed $default
     * @return string
     */
    private function formatDefaultValue($default)
    {
        return json_encode($default, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
    }

    /**
     * @param Command[] $commands
     * @return int
     */
    private function getColumnWidth(array $commands)
    {
        $width = 0;
        foreach ($commands as $command) {
            $width = strlen($command->getName()) > $width ? strlen($command->getName()) : $width;
        }

        return $width + 2;
    }

    /**
     * @param InputOption[] $options
     * @return int
     */
    private function calculateTotalWidthForOptions($options)
    {
        $totalWidth = 0;
        foreach ($options as $option) {
            $nameLength = 4 + strlen($option->getName()) + 2; // - + shortcut + , + whitespace + name + --

            if ($option->acceptValue()) {
                $valueLength = 1 + strlen($option->getName()); // = + value
                $valueLength += $option->isValueOptional() ? 2 : 0; // [ + ]

                $nameLength += $valueLength;
            }
            $totalWidth = max($totalWidth, $nameLength);
        }

        return $totalWidth;
    }
}
