<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use \app\index\model\Option;
use \app\index\model\User;
use think\Session;

class Home extends Controller{

	public $userObj;

	public function _initialize(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		if(!$this->userObj->loginStatus){
			$this->redirect(url('/Login','',''));
			exit();
		}
	}

	public function index(){
		$userInfo = $this->userObj->getInfo();
		$policyData = $this->userObj->getPolicy();
		$groupData =  $this->userObj->getGroupData();
		$extJson = json_decode($policyData["filetype"],true);
		$extLimit="";
		foreach ($extJson as $key => $value) {
			$extLimit.='{ title : "'.$value["title"].'", extensions : "'.$value["ext"].'" },';
		}
		$policyData["max_size"] = $policyData["max_size"]/(1024*1024);
		return view('home', [
			'options'  => Option::getValues(['basic','upload'],$this->userObj->userSQLData),
			'userInfo' => $userInfo,
			'extLimit' => $extLimit,
			'policyData' => $policyData,
			'groupData' => $groupData,
			'chunkSize' => config('upload.chunk_size'),
			'path' => empty(input("get.path"))?"/":input("get.path"),
		]);
	}

	public function Download(){
		$userInfo = $this->userObj->getInfo();
		$groupData =  $this->userObj->getGroupData();
		return view('download', [
			'options'  => Option::getValues(['basic','group_sell'],$this->userObj->userSQLData),
			'userInfo' => $userInfo,
			'groupData' => $groupData,
		]);
	}
		
}
