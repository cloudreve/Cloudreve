<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \think\Session;
use \app\index\model\FileManage;
use \app\index\model\Option;
use \app\index\model\User;

class ShareHandler extends Model{

	public $shareData;
	public $querryStatus = true;
	public $shareOwner;
	public $fileData;
	public $dirData;
	public $lockStatus = false;

	public function __construct($key,$notExist=true){
		$this->shareData = Db::name('shares')->where('share_key',$key)->find();
		if(empty($this->shareData)){
			$this->querryStatus = false;
		}else{
			if($notExist){
				if($this->shareData["source_type"] == "file"){
					$this->fileShareHandler();
				}else{
					$this->dirShareHandler();
				}
			}
		}
		if($this->shareData["type"] == "private"){
			$this->lockStatus = Session::has("share".$this->shareData["id"])?false:true;
		}

	}

	public function deleteShare($uid){
		if($this->shareData["owner"] != $uid){
			return array(
                "error" => 1,
                "msg" => "无权操作"
                );
		}
		Db::name('shares')->where('share_key',$this->shareData["share_key"])->delete();
		return array(
                "error" => 0,
                "msg" => "分享取消成功"
                );
	}

	public function changePromission($uid,$ignore=false){
		if($this->shareData["owner"] != $uid && $ignore == false){
			return array(
                "error" => 1,
                "msg" => "无权操作"
                );
		}
		Db::name('shares')->where('share_key',$this->shareData["share_key"])->update([
				'type' => $this->shareData["type"] == "public"?"private":"public",
				'share_pwd' => self::getRandomKey(6)
			]);
		return array(
					"error" =>0,
					"msg" => "更改成功"
				);
	}

	public function checkPwd($pwd){
		if($pwd == $this->shareData["share_pwd"]){
			Session::set("share".$this->shareData["id"],"1");
			return array(
					"error" =>0,
				);
		}else{
			return array(
					"error" =>1,
					"msg" => "密码错误"
				);
		}
	}

	public function getThumb($user,$path,$folder=false){
		$checkStatus = $this->checkSession($user);
		if(!$checkStatus[0]){
			return [$checkStatus[0],$checkStatus[1]];
		}
		$reqPath = Db::name('folders')->where('position_absolute',$this->shareData["source_name"])->find();
		if($folder){
			$fileObj = new FileManage($path,$this->shareData["owner"]);
		}else{
			$fileObj = new FileManage($reqPath["position_absolute"].$path,$this->shareData["owner"]);
		}
		return $fileObj->getThumb();
	}

	public function checkSession($user){
		if($this->lockStatus){
			return [false,"会话过期，请刷新页面"];
		}
		if(Option::getValue("allowdVisitorDownload") == "false" && !$user->loginStatus){
			return [false,"未登录用户禁止下载，请先登录"];
		}
		if(!$this->querryStatus){
			return [false,"分享不存在，请检查链接是否正确"];
		}
		return[true,null];
	}

	public function Download($user){
		$checkStatus = $this->checkSession($user);
		if(!$checkStatus[0]){
			return [$checkStatus[0],$checkStatus[1]];
		}
		$reqPath = Db::name('files')->where('id',$this->shareData["source_name"])->find();
		if($reqPath["dir"] == "/"){
			$reqPath["dir"] = $reqPath["dir"].$reqPath["orign_name"];
		}else{
			$reqPath["dir"] = $reqPath["dir"]."/".$reqPath["orign_name"];
		}
		$fileObj = new FileManage($reqPath["dir"],$this->shareData["owner"]);
		$FileHandler = $fileObj->Download();
		return $FileHandler;
	}

	public function DownloadFolder($user,$path){
		$checkStatus = $this->checkSession($user);
		if(!$checkStatus[0]){
			return [$checkStatus[0],$checkStatus[1]];
		}
		$reqPath = Db::name('folders')->where('position_absolute',$this->shareData["source_name"])->find();
		$path = $path == "/"?"":$path;
		$fileObj = new FileManage($reqPath["position_absolute"].$path,$this->shareData["owner"]);
		$this->numIncrease("download_num");
		$FileHandler = $fileObj->Download();
		return $FileHandler;
	}

	public function ListFile($path){
		if($this->lockStatus){
			die('{ "result": { "success": false, "error": "会话过期，请重新登陆" } }');
		}
		if(!$this->querryStatus){
			return [false,"分享不存在，请检查链接是否正确"];
		}
		$reqPath = Db::name('folders')->where('position_absolute',$this->shareData["source_name"])->find();
		$path = $path == "/"?"":$path;
		return FileManage::ListFile($this->shareData["source_name"].$path,$this->shareData["owner"]);
	}

	public function Preview($user){
		$checkStatus = $this->checkSession($user);
		if(!$checkStatus[0]){
			return [$checkStatus[0],$checkStatus[1]];
		}
		$reqPath = Db::name('files')->where('id',$this->shareData["source_name"])->find();
		if($reqPath["dir"] == "/"){
			$reqPath["dir"] = $reqPath["dir"].$reqPath["orign_name"];
		}else{
			$reqPath["dir"] = $reqPath["dir"]."/".$reqPath["orign_name"];
		}
		$fileObj = new FileManage($reqPath["dir"],$this->shareData["owner"]);
		return $fileObj->PreviewHandler();
	}

	public function PreviewFolder($user,$path,$folder=false){
		$checkStatus = $this->checkSession($user);
		if(!$checkStatus[0]){
			return [$checkStatus[0],$checkStatus[1]];
		}
		$reqPath = Db::name('folders')->where('position_absolute',$this->shareData["source_name"])->find();
		if($folder){
			$fileObj = new FileManage($path,$this->shareData["owner"]);
		}else{
			$fileObj = new FileManage($reqPath["position_absolute"].$path,$this->shareData["owner"]);
		}
		return $fileObj->PreviewHandler();
	}

	public function listPic($id,$path){
		if($this->lockStatus){
			die('{ "result": { "success": false, "error": "会话过期，请重新登陆" } }');
		}
		if(!$this->querryStatus){
			return [false,"分享不存在，请检查链接是否正确"];
		}
		$reqPath = Db::name('folders')->where('position_absolute',$this->shareData["source_name"])->find();
		$path = $path == "/"?"":$path;
		return FileManage::listPic($this->shareData["source_name"].$path,$this->shareData["owner"],"/Share/Preview/".$this->shareData["share_key"]."?folder=true");
	}

	public function numIncrease($name){
		Db::name('shares')->where('share_key',$this->shareData["share_key"])->setInc($name);
	}

	public function getDownloadUrl($user){
		if(Option::getValue("allowdVisitorDownload") == "false" && !$user->loginStatus){
			return array(
				"error" => 1,
				"msg" => "未登录用户禁止下载，请先登录",
				);
		}else{
			$this->numIncrease("download_num");
			return array(
				"error" => 0,
				"result" => "/Share/Download/".$this->shareData["share_key"],
				);
		}
	}

	public function fileShareHandler(){
		$this->shareOwner = new User($this->shareData["owner"],null,true);
		$this->fileData = Db::name('files')
			->where('upload_user',$this->shareData["owner"])
			->where('id',(int)$this->shareData["source_name"])
			->find();
		if(!$this->shareOwner->loginStatus || empty($this->fileData)){
			$this->querryStatus = false;
		}else{
			$this->querryStatus = true;
		}
	}

	public function dirShareHandler(){
		$this->shareOwner = new User($this->shareData["owner"],null,true);
		$this->dirData = Db::name('folders')
			->where('owner',$this->shareData["owner"])
			->where('position_absolute',$this->shareData["source_name"])
			->find();
		if(!$this->shareOwner->loginStatus || empty($this->dirData)){
			$this->querryStatus = false;
		}else{
			$this->querryStatus = true;
		}
	}

	static function createShare($fname,$type,$user,$group){
		if(!$group["allow_share"]){
			self::setError("您当前的用户组无权分享文件");
		}
		$path = FileManage::getFileName($fname)[1];
		$fnameTmp = FileManage::getFileName($fname)[0];
		$fileRecord = Db::name('files')->where('upload_user',$user["id"])->where('orign_name',$fnameTmp)->where('dir',$path)->find();
		if(empty($fileRecord)){
			self::createDirShare($fname,$type,$user,$group);
		}else{
			self::createFileShare($fileRecord,$type,$user,$group);
		}
	}

	static function setError($text){
		die('{ "result": { "success": false, "error": "'.$text.'" } }');
	}

	static function setSuccess($text){
		die('{ "result": "'.$text.'" }');
	}

	static function createDirShare($fname,$type,$user,$group){
		$dirRecord = Db::name('folders')->where('owner',$user["id"])->where('position_absolute',$fname)->find();
		if(empty($dirRecord)){
			self::setError("目录不存在");
		}
		$shareKey = self::getRandomKey(8);
		$sharePwd = $type=="public" ? "0" : self::getRandomKey(6);
		$SQLData = [
			'type' => $type=="public" ? "public" : "private",
			'share_time' => date("Y-m-d H:i:s"),
			'owner' => $user["id"],
			'source_name' => $fname,
			'origin_name' => $fname,
			'download_num' => 0,
			'view_num' => 0,
			'source_type' => "dir",
			'share_key' => $shareKey,
			'share_pwd' => $sharePwd,
		];
		if(Db::name('shares')->insert($SQLData)){
			if($sharePwd == "0"){
				self::setSuccess(Option::getValue("siteURL")."s/".$shareKey);
			}else{
				self::setSuccess("链接：".Option::getValue("siteURL")."s/".$shareKey."   密码：".$sharePwd);
			}
		}
	}

	static function createFileShare($file,$type,$user,$group){
		$shareKey = self::getRandomKey(8);
		$sharePwd = $type=="public" ? "0" : self::getRandomKey(6);
		$SQLData = [
			'type' => $type=="public" ? "public" : "private",
			'share_time' => date("Y-m-d H:i:s"),
			'owner' => $user["id"],
			'source_name' => $file["id"],
			'origin_name' => $file["orign_name"],
			'download_num' => 0,
			'view_num' => 0,
			'source_type' => "file",
			'share_key' => $shareKey,
			'share_pwd' => $sharePwd,
		];
		if(Db::name('shares')->insert($SQLData)){
			if($sharePwd == "0"){
				self::setSuccess(Option::getValue("siteURL")."s/".$shareKey);
			}else{
				self::setSuccess("链接：".Option::getValue("siteURL")."s/".$shareKey."   密码：".$sharePwd);
			}
		}
	}

	static function getRandomKey($length = 16){
		$charTable = 'abcdefghijklmnopqrstuvwxyz0123456789';
		$result = ""; 
		for ( $i = 0; $i < $length; $i++ ){ 
			$result .= $charTable[ mt_rand(0, strlen($charTable) - 1) ]; 
		} 
		return $result; 
	}

}
?>