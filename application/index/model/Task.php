<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \app\index\model\Option;

class Task extends Model{

	public $taskModel;
	public $taskName;
	public $taskType;
	public $taskContent;

	public function __construct($id=null){
		if($id!==null){

		}
	}

	public function saveTask(){
		Db::name("task")->insert([
			"task_name" => $this->taskName,
			"attr" => $this->taskContent,
			"type" => $this->taskType,
			"status" => "todo",
		]);
	}

}
?>