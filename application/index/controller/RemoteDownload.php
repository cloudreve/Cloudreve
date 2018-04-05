<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use \app\index\model\Option;
use \app\index\model\User;
use \app\index\model\Aria2;
use think\Session;


class RemoteDownload extends Controller{

	public $userObj;

	public function _initialize(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		if(!$this->userObj->loginStatus){
			echo "Bad request";
			exit();
		}
	}

	private function checkPerimission($permissionId){
		$permissionData = $this->userObj->groupData["aria2"];
		if(explode(",",$permissionData)[$permissionId] != "1"){
			return false;
		}
		return true;
	}

	public function addUrl(){
		if(!$this->checkPerimission(0)){
			return json(['error'=>1,'message'=>'您当前的无用户无法执行此操作']);
		}
		$aria2Options = Option::getValues(["aria2"]);
		$aria2 = new Aria2($aria2Options);
	}

}