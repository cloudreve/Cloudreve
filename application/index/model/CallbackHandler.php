<?php
namespace app\index\model;

require_once   'extend/Qiniu/functions.php';

use think\Model;
use think\Db;
use Qiniu\Auth;
use \app\index\model\Option;
use \app\index\model\FileManage;
use \app\index\model\UploadHandler;

class CallbackHandler extends Model{

	public $CallbackData;
	public $policyData;
	public $userData;
	
	public function __construct($data){
		$this->CallbackData = $data;
	}

	public function remoteHandler($header){
		$jsonData = json_decode(base64_decode($this->CallbackData),true);
		$CallbackSqlData = Db::name('callback')->where('callback_key',$jsonData['callbackkey'])->find();
		$this->policyData = Db::name('policy')->where('id',$CallbackSqlData['pid'])->find();
		if(!$this->IsRemoteCallback($header)){
			$this->setError("Undelegated Request");
		}
		if($this->policyData == null){
			$this->setError("CallbackKey Not Exist.");
		}
		if(!FileManage::sotrageCheck($CallbackSqlData["uid"],$jsonData["fsize"])){
			$this->setError("空间容量不足",true);
		}
		$picInfo = $jsonData["picinfo"];
		$addAction = FileManage::addFile($jsonData,$this->policyData,$CallbackSqlData["uid"],$picInfo);
		if(!$addAction[0]){
			$this->setError($addAction[1],true);
		}
		FileManage::storageCheckOut($CallbackSqlData["uid"],$jsonData["fsize"]);
		$this->setSuccess($jsonData['fname']);
	}

	public function qiniuHandler($header){
		$jsonData = json_decode($this->CallbackData,true);
		$CallbackSqlData = Db::name('callback')->where('callback_key',$jsonData['callbackkey'])->find();
		$this->policyData = Db::name('policy')->where('id',$CallbackSqlData['pid'])->find();

		if(!$this->IsQiniuCallback($header)){
			$this->setError("Undelegated Request");
		}
		if($this->policyData == null){
			$this->setError("CallbackKey Not Exist.");
		}
		if(!FileManage::sotrageCheck($CallbackSqlData["uid"],$jsonData["fsize"])){
			$this->setError("空间容量不足",true);
		}
		$picInfo = $jsonData["picinfo"];
		$addAction = FileManage::addFile($jsonData,$this->policyData,$CallbackSqlData["uid"],$picInfo);
		if(!$addAction[0]){
			$this->setError($addAction[1],true);
		}
		FileManage::storageCheckOut($CallbackSqlData["uid"],$jsonData["fsize"]);
		$this->setSuccess($jsonData['fname']);
	}

	public function ossHandler($auth,$pubKey){
		if(!$this->IsOssCallback($auth,$pubKey)){
			$this->setError("Undelegated Request");
		}
		$jsonData = json_decode($this->CallbackData,true);
		$jsonData["fname"] = urldecode($jsonData["fname"]);
		$jsonData["objname"] = urldecode($jsonData["objname"]);
		$jsonData["path"] = urldecode($jsonData["path"]);
		$CallbackSqlData = Db::name('callback')->where('callback_key',$jsonData['callbackkey'])->find();
		$this->policyData = Db::name('policy')->where('id',$CallbackSqlData['pid'])->find();
		if($this->policyData == null){
			$this->setError("CallbackKey Not Exist.");
		}
		if(!FileManage::sotrageCheck($CallbackSqlData["uid"],$jsonData["fsize"])){
			$this->setError("空间容量不足",true);
		}
		$picInfo = $jsonData["picinfo"];
		$addAction = FileManage::addFile($jsonData,$this->policyData,$CallbackSqlData["uid"],$picInfo);
		if(!$addAction[0]){
			$this->setError($addAction[1],true);
		}
		FileManage::storageCheckOut($CallbackSqlData["uid"],$jsonData["fsize"]);
		$this->setSuccess($jsonData['fname']);
	}

	public function upyunHandler($token,$date,$md5){
		$this->policyData = Db::name("policy")->where("id",$this->CallbackData["ext-param"]["pid"])->find();
		if(!$this->IsUpyunCallback($token,$date,$md5)){
			$this->setError("Undelegated Request",false,true);
		}
		if(!FileManage::sotrageCheck($this->CallbackData["ext-param"]["uid"],$this->CallbackData["file_size"])){
			FileManage::deleteFile($this->CallbackData["url"],$this->policyData);
			$this->setError("空间容量不足",false,true);
		}
		$picInfo = empty($this->CallbackData["image-width"]) ? "" :$this->CallbackData["image-width"].",".$this->CallbackData["image-height"];
		$fileNameExplode = explode("CLSUFF",$this->CallbackData["url"]);
		$jsonData["fname"] = end($fileNameExplode);
		$jsonData["objname"] = $this->CallbackData["url"];
		$jsonData["path"] = $this->CallbackData["ext-param"]["path"];
		$jsonData["fsize"] = $this->CallbackData["file_size"];
		$addAction = FileManage::addFile($jsonData,$this->policyData,$this->CallbackData["ext-param"]["uid"],$picInfo);
		if(!$addAction[0]){
			FileManage::deleteFile($this->CallbackData["url"],$this->policyData);
			$this->setError($addAction[1],false,true);
		}
		FileManage::storageCheckOut($this->CallbackData["ext-param"]["uid"],$jsonData["fsize"]);
		$this->setSuccess($jsonData['fname']);
	}

	public function s3Handler($key){
		$CallbackSqlData = Db::name('callback')->where('callback_key',$key)->find();
		//删除callback记录
		if(empty($CallbackSqlData)){
			$this->setError("Undelegated Request",false,true);
		}
		$this->policyData = Db::name('policy')->where('id',$CallbackSqlData['pid'])->find();
		$this->userData =  Db::name('users')->where('id',$CallbackSqlData['uid'])->find();
		$paths = explode("/",$this->CallbackData["key"]);
		$jsonData["fname"] = end($paths);
		$jsonData["objname"] = $this->CallbackData["key"];
		$jsonData["path"] ="";
		foreach ($paths as $key => $value) {
			if($key == 0 || $key == count($paths)-1) continue;
			$jsonData["path"].=$value.",";
		}
		$jsonData["path"] = rtrim($jsonData["path"],",");
		$jsonData["fsize"] = $this->getS3FileInfo();
		if(!$jsonData["fsize"]){
			$this->setError("File not exist",false,true);
		}
		$jsonData["fsize"] = $jsonData["fsize"]["size"];
		$picInfo = "";
		$addAction = FileManage::addFile($jsonData,$this->policyData,$this->userData["id"],"");
		if(!$addAction[0]){
			FileManage::deleteFile($this->CallbackData["key"],$this->policyData);
			$this->setError($addAction[1],false,true);
		}
		FileManage::storageCheckOut($this->userData["id"],$jsonData["fsize"]);
		$this->setSuccess($jsonData['fname']);
	}

	private function getS3FileInfo(){
		$s3 = new \S3\S3($this->policyData["ak"], $this->policyData["sk"],false,$this->policyData["op_pwd"]);
		$s3->setSignatureVersion('v4');
		try {
			$returnVal = $s3->getObjectInfo($this->policyData["bucketname"],$this->CallbackData["key"]);
		} catch (Exception $e) {
			return false;
		}
		return $returnVal;
	}

	public function setSuccess($fname){
		die(json_encode(["key"=> $fname]));
	}

	public function setError($text,$delete = false,$ignore=false){
		header("HTTP/1.1 401 Unauthorized");
		if(!$ignore){
			$deletedFile = json_decode($this->CallbackData,true);
			$fileNmae = $deletedFile['objname'];
			if($delete){
				FileManage::deleteFile($fileNmae,$this->policyData);
			}
		}
		die(json_encode(["error"=> $text]));
	}

	private function isUpyunCallback($token,$date,$md5){
		if(UploadHandler::upyunSign($this->policyData["op_name"],md5($this->policyData["op_pwd"]),"POST","/Callback/Upyun",$date,$md5) != $token){
			return false;
		}
		return true;
	}

	public function IsQiniuCallback($httpHeader){
		$auth = new Auth($this->policyData['ak'], $this->policyData['sk']);
		$callbackBody = $this->CallbackData;
		$contentType = 'application/json';
		$authorization = $httpHeader;
		$url = Option::getValue("siteUrl")."Callback/Qiniu";
		$isQiniuCallback = $auth->verifyCallback($contentType, $authorization, $url,$callbackBody);
		if ($isQiniuCallback) {
			return true;
		} else {
			return false;
		}
	}

	private function IsRemoteCallback($header){
		$signKey = hash_hmac("sha256",$this->CallbackData,$this->policyData["sk"]);
		return ($signKey == $header);
	}

	public function IsOssCallback($auth,$pubKey){
		if (empty($auth) || empty($pubKey)){
			header("http/1.1 403 Forbidden");
			exit();
		}
		$authorization = base64_decode($auth);
		$pubKeyUrl = base64_decode($pubKey);
		$pubOssKey = file_get_contents($pubKeyUrl);
		if ($pubOssKey == ""){
			return false;
		}
		$body = file_get_contents('php://input');
		$authStr = '';
		$path = $_SERVER['REQUEST_URI'];
		$pos = strpos($path, '?');
		if ($pos === false){
			$authStr = urldecode($path)."\n".$body;
		}else{
			$authStr = urldecode(substr($path, 0, $pos)).substr($path, $pos, strlen($path) - $pos)."\n".$body;
		}
		$ok = openssl_verify($authStr, $authorization, $pubOssKey, OPENSSL_ALGO_MD5);
		if ($ok == 1){
			return true;
		}else{
			return false;
		}
	}

}


?>