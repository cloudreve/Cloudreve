<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use \app\index\model\Option;
use \app\index\model\User;
use think\Session;
use \app\index\model\FileManage;
use \app\index\model\ShareHandler;

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
		$url = "/File/Preview?action=preview&path=".$path;
		if(input("get.share")==true){
			$url = "/Share/Preview/".input("get.shareKey")."/?path=".$path;
		}else if(input("get.single")==true){
			$url = "/Share/Preview/".input("get.shareKey");
		}
		return view('video', [
			'options'  => Option::getValues(['basic'],$this->userObj->userSQLData),
			'userInfo' => $userInfo,
			'groupData' => $groupData,
			'url' => $url,
			'fileName' => end($pathSplit),
			'isSharePage' => input("?get.share")?"true":"false",
		]);
	}

	public function Markdown(){
		$path = input("get.path");
		$pathSplit = explode("/",urldecode($path));
		$userInfo = $this->userObj->getInfo();
		$groupData =  $this->userObj->getGroupData();
		$url = "/File/Content?action=preview&path=".$path;
		if(input("get.share")==true){
			$url = "/Share/Content/".input("get.shareKey")."/?path=".$path;
		}else if(input("get.single")==true){
			$url = "/Share/Content/".input("get.shareKey");
		}
		return view('markdown', [
			'options'  => Option::getValues(['basic'],$this->userObj->userSQLData),
			'userInfo' => $userInfo,
			'groupData' => $groupData,
			'url' => $url,
			'fileName' => end($pathSplit),
			'path' => urldecode($path),
			'isSharePage' => input("?get.share")?"true":"false",
		]);
	}
		
		
}
