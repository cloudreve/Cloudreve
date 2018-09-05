<?php
namespace app\index\command;

use think\console\Command;
use think\console\Input;
use think\console\Output;

class Task extends Command
{
    protected function configure()
    {
        $this->setName('run')->setDescription('Start processing tasks for Cloudreve');
    }

    protected function execute(Input $input, Output $output)
    {
        $output->writeln("TestCommand:");
    }
}
?>