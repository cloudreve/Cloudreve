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

	private function insertRecord($aria2,$url){
		Db::name("download")->insert([
				"pid" => $aria2->pid,
				"path_id" => $aria2->pathId,
				"owner" => $this->userObj->uid,
				"save_dir" => 1,
				"status" => "ready",
				"msg" => "",
				"info"=>"",
				"source" =>$url,
				"file_index" => 0,
				"is_single" => 1,
			]);
	}

	public function addUrl(){
		$policyData = Db::name("policy")->where("id",$this->userObj->groupData["policy_name"])->find();
		if(!$this->checkPerimission(0) || $policyData["policy_type"] != "local"){
			return json(['error'=>1,'message'=>'您当前的无用户无法执行此操作']);
		}
		$aria2Options = Option::getValues(["aria2"]);
		$aria2 = new Aria2($aria2Options);
		$downloadStart = $aria2->addUrl(input("post.url"));
		if($aria2->reqStatus){
			$this->insertRecord($aria2,input("post.url"));
		}else{
			return json(['error'=>1,'message'=>$aria2->reqMsg]);
		}
	}

	public function AddTorrent(){
		$policyData = Db::name("policy")->where("id",$this->userObj->groupData["policy_name"])->find();
		if(!$this->checkPerimission(0) || $policyData["policy_type"] != "local"){
			return json(['error'=>1,'message'=>'您当前的无用户无法执行此操作']);
		}
		$aria2Options = Option::getValues(["aria2"]);
		$aria2 = new Aria2($aria2Options);
		$torrentObj = new \app\index\model\FileManage(input("post.path"),$this->userObj->uid);
		$downloadStart = $aria2->addTorrent($torrentObj->signTmpUrl());
		if($aria2->reqStatus){
			$this->insertRecord($aria2,input("post.path"));
		}else{
			return json(['error'=>1,'message'=>$aria2->reqMsg]);
		}
	}

	public function FlushStatus(){
		$aria2Options = Option::getValues(["aria2"]);
		$aria2 = new Aria2($aria2Options);
		if(!input("?post.id")){
			return json(['error'=>1,'message'=>"信息不完整"]);
		}
		$policyData = Db::name("policy")->where("id",$this->userObj->groupData["policy_name"])->find();
		if(!$aria2->flushStatus(input("post.id"),$this->userObj->uid,$policyData)){
			return json(['error'=>1,'message'=>$aria2->reqMsg]);
		}
	}

}