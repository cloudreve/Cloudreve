<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use think\Request;

use \app\index\model\Option;

class Queue extends Controller{

	public function __construct(\think\Request $request = null){
		$token = Option::getValue("task_queue_token");
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
		if(empty($tasks)){
			return "none";
		}else{
			return json_encode($tasks);
		}
	}

}
