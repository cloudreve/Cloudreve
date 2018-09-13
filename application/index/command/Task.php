<?php
namespace app\index\command;

use think\Db;
use think\console\Command;
use think\console\Input;
use think\console\Output;

class Task extends Command
{

    const SLEEP_TIME = 1;

    protected function configure()
    {
        $this->setName('run')->setDescription('Start processing tasks for Cloudreve');
    }

    protected function Init(Output $output){
        $output->writeln("Cloudreve tasks processor started.");
    }

    protected function setComplete($taskId,Output $output){
        $output->writeln("Cloudreve tasks processor started.");
    }

    protected function execute(Input $input, Output $output)
    {
        self::Init($output);
        while (1){
            $newTaskInfo = Db::name("task")->where("status","todo")->find();
            if(empty($newTaskInfo)){
                sleep(self::SLEEP_TIME);
                continue;
            }
            Db::name("task")->where("id",$newTaskInfo["id"])->update(["status"=>"processing"]);
            $output->writeln("[New task] Name:".$newTaskInfo["task_name"]." Type:".$newTaskInfo["type"]);
            $task = new \app\index\model\Task();
            $task->taskModel = $newTaskInfo;
            $task->input = $input;
            $task->output = $output;
            $task->Doit();
            if($task->status=="error"){
                $output->writeln("[Error] ".$task->errorMsg);
                Db::name("task")->where("id",$newTaskInfo["id"])->update(["status"=>"error|".$task->errorMsg]);
            }else{
                $output->writeln("[Complete]");
                Db::name("task")->where("id",$newTaskInfo["id"])->update(["status"=>"complete"]);
            }
        }
    }
}
?>