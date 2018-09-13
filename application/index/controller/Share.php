<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use \app\index\model\Option;
use \app\index\model\User;
use \app\index\model\ShareHandler;
use think\Session;
use think\Request;

class Share extends Controller{

	public $userObj;
	public $siteOptions;

	public function _initialize(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$this->siteOptions = Option::getValues(["basic"]);
	}

	public function index(){
		$shareKey = input('param.key');
		$shareObj = new ShareHandler($shareKey);
		if(!$shareObj->querryStatus){
			 header('HTTP/1.1 404 Not Found');
			 $this->error('当前分享不存在或者已经失效',404,$this->siteOptions);
		}
		if(!$shareObj->lockStatus){
			$shareObj->numIncrease("view_num");
			if($shareObj->shareData["source_type"] == "dir"){
				return view('share_dir', [
					'options'  => Option::getValues(['basic','share']),
					'userInfo' => $shareObj->shareOwner->userSQLData,
					'dirData' => $shareObj->dirData,
					'shareData' => $shareObj->shareData,
					'loginStatus' => $this->userObj->loginStatus,
					'userData' => $this->userObj->userSQLData,
					'groupData' =>  $shareObj->shareOwner->groupData,
					'allowPreview' => Option::getValue("allowdVisitorDownload"),
				]);
			}else{
				return view('share_single', [
					'options'  => Option::getValues(['basic','share']),
					'userInfo' => $shareObj->shareOwner->userSQLData,
					'fileData' => $shareObj->fileData,
					'shareData' => $shareObj->shareData,
					'loginStatus' => $this->userObj->loginStatus,
					'userData' => $this->userObj->userSQLData,
					'allowPreview' => Option::getValue("allowdVisitorDownload"),
				]);
			}
		}else{
			return view('share_lock', [
				'options'  => Option::getValues(['basic','share']),
				'userInfo' => $shareObj->shareOwner->userSQLData,
				'fileData' => $shareObj->fileData,
				'shareData' => $shareObj->shareData,
				'loginStatus' => $this->userObj->loginStatus,
				'userData' => $this->userObj->userSQLData,
				'pwd' => input("?get.pwd") ? input("get.pwd") : "",
			]);
		}
	}

	public function getDownloadUrl(){
		$shareId = input('key');
		$shareObj = new ShareHandler($shareId,false);
		return $shareObj->getDownloadUrl($this->userObj);
	}

	public function Download(){
		$shareId = input('param.key');
		$filePath = input('get.path');
		$shareObj = new ShareHandler($shareId,false);
		if(empty($filePath)){
			$DownloadHandler = $shareObj->Download($this->userObj);
		}else{
			$DownloadHandler = $shareObj->DownloadFolder($this->userObj,$filePath);
		}
		if($DownloadHandler[0]){
			$this->redirect($DownloadHandler[1],302);
		}else{
			$this->error($DownloadHandler[1],404,$this->siteOptions);
		}
	}

	public function chekPwd(){
		$shareId = input('key');
		$inputPwd = input('password');
		$shareObj = new ShareHandler($shareId,false);
		if(!$shareObj->querryStatus){
			 return array(
				"error" => 1,
				"msg" => "分享不存在"
				);
		}
		return $shareObj->checkPwd($inputPwd);
	}

	public function Preview(){
		$shareId = input('param.key');
		$filePath = input('get.path');
		$shareObj = new ShareHandler($shareId,false);
		if(empty($filePath)){
			$previewHandler = $shareObj->Preview($this->userObj);
		}else{
			if(!empty(input('get.folder'))){
				 $previewHandler = $shareObj->PreviewFolder($this->userObj,$filePath,true);
			}else{
				$previewHandler = $shareObj->PreviewFolder($this->userObj,$filePath);
			}
		}
		if($previewHandler[0]){
			$this->redirect($previewHandler[1],302);
		}else{
			$this->error($previewHandler[1],404,$this->siteOptions);
		}
	}
	
	public function ListFile(){
		$shareId = input('param.key');
		$reqPathTo = stripslashes(json_decode(file_get_contents("php://input"),true)['path']);
		$shareObj = new ShareHandler($shareId,false);
		return $shareObj->ListFile($reqPathTo);
	}

	public function ListPic(){
		$filePath = input('get.path');
		$shareId = input('get.id');
		$shareObj = new ShareHandler($shareId,false);
		return $shareObj->listPic($shareId,$filePath);
	}

	public function Thumb(){
		$shareId = input('get.shareKey');
		$filePath = input('get.path');
		if(input("get.isImg") != "true"){
			return "";
		}
		$shareObj = new ShareHandler($shareId,false);
		$Redirect = $shareObj->getThumb($this->userObj,$filePath);
		if($Redirect[0]){
			$this->redirect($Redirect[1],302);
		}else{
			$this->error($Redirect[1],403,$this->siteOptions);
		}
	}

	public function Delete(){
		$shareId = input('post.id');
		$shareObj = new ShareHandler($shareId,false);
		if(!$shareObj->querryStatus){
			 return array(
				"error" => 1,
				"msg" => "分享不存在"
				);
		}
		return $shareObj->deleteShare($this->userObj->uid);
	}

	public function ChangePromission(){
		$shareId = input('post.id');
		$shareObj = new ShareHandler($shareId,false);
		if(!$shareObj->querryStatus){
			 return array(
				"error" => 1,
				"msg" => "分享不存在"
				);
		}
		return $shareObj->changePromission($this->userObj->uid);
	}

	public function My(){
		if(!$this->userObj->loginStatus){
			$this->redirect(url('/Login','',''));
			exit();
		}
		$userInfo = $this->userObj->getInfo();
		$groupData =  $this->userObj->getGroupData();
		$list = Db::name('shares')->where('owner',$this->userObj->uid)->order('share_time DESC')->paginate(30);
		$listData = $list->all();
		foreach ($listData as $key => $value) {
			if($value["source_type"]=="file"){
				$listData[$key]["fileData"] = Db::name('files')->where('id',$value["source_name"])->find()["orign_name"];

			}else{
				$listData[$key]["fileData"] = $value["source_name"];
			}
		}
		return view('share_home', [
			'options'  => Option::getValues(['basic','share']),
			'userInfo' => $userInfo,
			'groupData' => $groupData,
			'list' => $listData,
			'listOrigin' => $list
		]);
	}

}
