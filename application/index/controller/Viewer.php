<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use \app\index\model\Option;
use \app\index\model\User;
use think\Session;
use \app\index\model\FileManage;

class Viewer extends Controller{

	public $userObj;

	public function _initialize(){
		// $this->userObj = new User(cookie('user_id'),cookie('login_key'));
		// if(!$this->userObj->loginStatus){
		// 	$this->redirect(url('/Login','',''));
		// 	exit();
		// }
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
	}

	public function Video(){
		$path = input("get.path");
		$pathSplit = explode("/",urldecode($path));
		$userInfo = $this->userObj->getInfo();
		$groupData =  $this->userObj->getGroupData();
		return view('video', [
			'options'  => Option::getValues(['basic']),
			'userInfo' => $userInfo,
			'groupData' => $groupData,
			'url' => "/File/Preview?action=preview&path=".$path,
			'fileName' => end($pathSplit),
		]);
	}
		
		
}
