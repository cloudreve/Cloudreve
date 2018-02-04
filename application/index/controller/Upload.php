<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use think\Request;
use \app\index\model\Option;
use \app\index\model\User;
use \app\index\model\UploadHandler;
use think\Session;

class Upload extends Controller{

	public $userObj;

	public function _initialize(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		if(!$this->userObj->loginStatus){
			echo "Bad request";
			exit();
		}
	}

	public function index(){
		ob_end_clean();
		$file = request()->file('file');
		$fileInfo = Request::instance()->request();
		$UploadHandler = new UploadHandler($this->userObj->groupData['policy_name'],$this->userObj->uid);
		return $UploadHandler->fileReceive($file,$fileInfo);
	}
	
	public function Token(){
		$uploadObj = new UploadHandler($this->userObj->groupData['policy_name'],$this->userObj->uid);
		$upToken = $uploadObj->getToken();
		if(!empty($uploadObj->upyunPolicy)){
			return json([
				"token" => $upToken,
				"policy" => $uploadObj->upyunPolicy,
				]);
		}
		if(!empty($uploadObj->s3Policy)){
			return json([
				"policy" => $uploadObj->s3Policy,
				"sign" =>  $uploadObj->s3Sign,
				"key" => $uploadObj->dirName,
				"credential" => $uploadObj->s3Credential,
				"x_amz_date" => $uploadObj->x_amz_date,
				"siteUrl"=>$uploadObj->siteUrl,
				"callBackKey" => $uploadObj->callBackKey,
				]);
		}
		if(!$uploadObj->getToken()){
			return json([
				"uptoken" => $uploadObj->ossToken,
				"sign" => $uploadObj->ossSign,
				"id" => $uploadObj->ossAccessId,
				"key" => $uploadObj->ossFileName,
				"callback" => $uploadObj->ossCallBack,
				]);
		}
		return json(["uptoken" => $uploadObj->getToken()]);
	}

	public function chunk(){
		$file = file_get_contents('php://input');
		$uploadObj = new UploadHandler($this->userObj->groupData['policy_name'],$this->userObj->uid);
		$uploadObj->setChunk(input('param.chunk'),input('param.chunks'),$file);
	}

	public function mkFile(){
		$ctx = file_get_contents('php://input');
		$originName = UploadHandler::b64Decode(input('param.fname'));
		$filePath = UploadHandler::b64Decode(input('param.path'));
		$uploadObj = new UploadHandler($this->userObj->groupData['policy_name'],$this->userObj->uid);
		$uploadObj->generateFile($ctx,$originName,$filePath);
	}

}
