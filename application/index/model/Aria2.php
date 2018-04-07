<?php
namespace app\index\model;

use think\Model;
use think\Db;

class Aria2 extends Model{

	private $authToken;
	private $apiUrl;
	private $savePath;
	private $saveOptions;
	public $reqStatus;
	public $reqMsg;
	public $pathId;
	public $pid;
	private $uid;
	private $policy;

	public function __construct($options){
		$this->authToken = $options["aria2_token"];
		$this->apiUrl = rtrim($options["aria2_rpcurl"],"/")."/";
		$this->saveOptions = json_decode($options["aria2_options"],true);
		$this->savePath = rtrim(rtrim($options["aria2_tmppath"],"/"),"\\").DS;
	}

	public function addUrl($url){
		$this->pathId = uniqid();
		$reqFileds = [
				"params" => ["token:".$this->authToken,
						[$url],["dir" => $this->savePath.$this->pathId],
					],
				"jsonrpc" => "2.0",
				"id" => $this->pathId,
				"method" => "aria2.addUri"
			];
		$reqFileds["params"][2] = array_merge($reqFileds["params"][2],$this->saveOptions);
		$reqFileds = json_encode($reqFileds,JSON_OBJECT_AS_ARRAY);
		$respondData = $this->sendReq($reqFileds);
		if(isset($respondData["result"])){
			$this->reqStatus = 1;
			$this->pid = $respondData["result"];
		}else{
			$this->reqStatus = 0;
			$this->reqMsg = $respondData["error"]["message"];
		}
	}

	public function flushStatus($id,$uid,$policy){
		$this->uid = $uid;
		$this->policy = $policy;
		$downloadInfo = Db::name("download")->where("id",$id)->find();
		if(empty($downloadInfo)){
			$this->reqStatus = 0;
			$this->reqMsg = "未找到下载记录";
			return false;
		}
		if(in_array($downloadInfo["status"], ["error","complete"])){
			$this->reqStatus = 1;
			return true;
		}
		if($uid != $downloadInfo["owner"]){
			$this->reqStatus = 0;
			$this->reqMsg = "无权操作";
			return false;
		}
		$reqFileds = [
				"params" => ["token:".$this->authToken,$downloadInfo["pid"]],
				"jsonrpc" => "2.0",
				"id" => uniqid(),
				"method" => "aria2.tellStatus"
			];
		$reqFileds = json_encode($reqFileds,JSON_OBJECT_AS_ARRAY);
		$respondData = $this->sendReq($reqFileds);
		if(isset($respondData["result"])){
			if($this->storageCheck($respondData["result"],$downloadInfo)){
				Db::name("download")->where("id",$id)
				->update([
					"status" => $respondData["result"]["status"],
					"last_update" => date("Y-m-d h:i:s"),
					"info" => json_encode([
							"completedLength" => $respondData["result"]["completedLength"],
							"totalLength" => $respondData["result"]["totalLength"],
							"dir" => $respondData["result"]["files"][0]["path"],
							"downloadSpeed" => $respondData["result"]["downloadSpeed"],
							"errorMessage" => isset($respondData["result"]["errorMessage"]) ? $respondData["result"]["errorMessage"] : "",
						]),
					"msg" => isset($respondData["result"]["errorMessage"]) ? $respondData["result"]["errorMessage"] : "",
					]);
				switch ($respondData["result"]["status"]) {
					case 'complete':
						$this->setComplete($respondData["result"],$downloadInfo);
						break;
					
					default:
						# code...
						break;
				}
			}else{
				$this->reqStatus = 0;
				$this->reqMsg = "空间容量不足";
				$this->setError($respondData["result"],$downloadInfo,"空间容量不足");
				return false;
			}
		}else{
			$this->reqStatus = 0;
			$this->reqMsg = $respondData["error"]["message"];
		}
		return true;
	}

	private function setComplete($quenInfo,$sqlData){
		if($this->policy["policy_type"] != "local"){
			$this->setError($quenInfo,$sqlData,"您当前的上传策略无法使用离线下载");
			return false;
		}
		$suffixTmp = explode('.', $quenInfo["dir"]);
		$fileSuffix = array_pop($suffixTmp);
		$uploadHandller = new UploadHandler($this->policy["id"],$this->uid);
		$allowedSuffix = explode(',', $uploadHandller->getAllowedExt(json_decode($this->policy["filetype"],true)));
		$sufficCheck = !in_array($fileSuffix,$allowedSuffix);
		if(empty($uploadHandller->getAllowedExt(json_decode($this->policy["filetype"],true)))){
			$sufficCheck = false;
		}
		if($sufficCheck){
			//取消任务
			$this->setError($quenInfo,$sqlData,"文件类型不被允许");
			return false;
		}
		if($this->policy['autoname']){
			$fileName = $uploadHandller->getObjName($this->policy['namerule'],"local",basename($quenInfo["files"][0]["path"]));
		}else{
			$fileName = basename($quenInfo["files"][0]["path"]);
		}
		$generatePath = $uploadHandller->getDirName($this->policy['dirrule']);
		$savePath = ROOT_PATH . 'public/uploads/'.$generatePath.DS.$fileName;
		is_dir(dirname($savePath))? :mkdir(dirname($savePath),0777,true);
		rename($quenInfo["files"][0]["path"],$savePath);
		@unlink(dirname($quenInfo["files"][0]["path"]));
		$jsonData = array(
			"path" => "", 
			"fname" => basename($quenInfo["files"][0]["path"]),
			"objname" => $generatePath.DS.$fileName,
			"fsize" => $quenInfo["totalLength"],
		);
		@list($width, $height, $type, $attr) = getimagesize($savePath);
		$picInfo = empty($width)?" ":$width.",".$height;
		$addAction = FileManage::addFile($jsonData,$this->policy,$this->uid,$picInfo);
		if(!$addAction[0]){
			//取消任务
			$this->setError($quenInfo,$sqlData,$addAction[1]);
			return false;
		}
		FileManage::storageCheckOut($this->uid,(int)$quenInfo["totalLength"]);
	}

	private function setError($quenInfo,$sqlData,$msg,$status="error"){
		$this->Remove($sqlData["pid"],$sqlData);
		$this->removeDownloadResult($sqlData["pid"],$sqlData);
		if(file_exists($quenInfo["files"][0]["path"])){
			@unlink($quenInfo["files"][0]["path"]);
			@unlink(dirname($quenInfo["files"][0]["path"]));
		}
		Db::name("download")->where("id",$sqlData["id"])->update([
			"msg" => $msg,
			"status" => $status,
			]);
	}

	public function Remove($gid,$sqlData){
		$reqFileds = [
				"params" => ["token:".$this->authToken,$gid],
				"jsonrpc" => "2.0",
				"id" => uniqid(),
				"method" => "aria2.remove"
			];
		$reqFileds = json_encode($reqFileds,JSON_OBJECT_AS_ARRAY);
		$respondData = $this->sendReq($reqFileds);
		if(isset($respondData["result"])){
			return true;
		}
		return false;
	}

	public function removeDownloadResult($gid,$sqlData){
		$reqFileds = [
				"params" => ["token:".$this->authToken,$gid],
				"jsonrpc" => "2.0",
				"id" => uniqid(),
				"method" => "aria2.removeDownloadResult"
			];
		$reqFileds = json_encode($reqFileds,JSON_OBJECT_AS_ARRAY);
		$respondData = $this->sendReq($reqFileds);
		if(isset($respondData["result"])){
			return true;
		}
		return false;
	}

	private function storageCheck($quenInfo,$sqlData){
		if(!FileManage::sotrageCheck($this->uid,(int)$quenInfo["totalLength"])){
			return false;
		}
		if(!FileManage::sotrageCheck($this->uid,(int)$quenInfo["completedLength"])){
			return false;
		}
		return true;
	}

	private function sendReq($data){
		$curl = curl_init();
	    curl_setopt($curl, CURLOPT_URL, $this->apiUrl."jsonrpc");
	    curl_setopt($curl, CURLOPT_POST, 1);
	    curl_setopt($curl, CURLOPT_POSTFIELDS, $data);
	    curl_setopt($curl, CURLOPT_TIMEOUT, 15); 
	    curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
	    $tmpInfo = curl_exec($curl);
	    if (curl_errno($curl)) {
	    	$this->reqStatus = 0;
	    	$this->reqMsg = "请求失败,".curl_error($curl);
	    }
	    curl_close($curl);
	    return json_decode($tmpInfo,true);
	}

}
?>