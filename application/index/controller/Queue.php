<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use think\Request;

use \app\index\model\Task;
use \app\index\model\Option;

class Queue extends Controller{

	public function __construct(\think\Request $request = null){
		$token = Option::getValue("task_queue_token");
		if($token==""){
			abort(403);
		}
		if(Request::instance()->header("Authorization") !="Bearer ".$token){
			abort(403);
		}
	}

	public function index(){
		
	}

	public function basicInfo(){
		return json_encode([
			"basePath" => ROOT_PATH,
		]);
	}

	public function getList(){
		$size = input("get.num");
		$tasks = Db::name("task")->where("status","todo")->limit($size)->select();
		$taskID = array_column($tasks,"id");
		Db::name("task")->where("id","in",$taskID)->update(["status"=>"processing"]);
		if(empty($tasks)){
			return "none";
		}else{
			return json_encode($tasks);
		}
	}

	public function getPolicy(){
		$id = input("get.id");
		$policy  = Db::name("policy")->where("id",$id)->find();
		if(empty(($policy))){
			abort(404);
		}else{
			return json($policy);
		}
	}

	public function setSuccess(){
		$id = input("get.id");
		$task = new Task($id);
		$task->taskModel = Db::name("task")
		->where("id",$id)
		//->where("status","processing")
		->find();
		if(empty($task->taskModel)){
			return json(["error"=>true,"msg"=>"未找到任务"]);
		}
		return json($task->setSuccess());
	}

	public function setError(){
		$id = input("get.id");
		$task = new Task($id);
		$task->taskModel = Db::name("task")
		->where("id",$id)
		//->where("status","processing")
		->find();
		if(empty($task->taskModel)){
			return json(["error"=>true,"msg"=>"未找到任务"]);
		}
		return json($task->Error());
	}

}
