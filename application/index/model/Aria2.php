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
							"dir" => $respondData["result"]["dir"],
							"downloadSpeed" => $respondData["result"]["downloadSpeed"],
							"errorMessage" => $respondData["result"]["errorMessage"],
						]),
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
				//取消离线下载
				return false;
			}
		}else{
			$this->reqStatus = 0;
			$this->reqMsg = $respondData["error"]["message"];
		}
		return true;
	}

	private function setComplete($quenInfo,$sqlData){
		FileManage::storageCheckOut($this->uid,(int)$quenInfo["totalLength"]);
		if($this->policy["policy_type"] != "local"){
			return false;
		}
		$suffixTmp = explode('.', $quenInfo["dir"]);
		$fileSuffix = array_pop($suffixTmp);
		$allowedSuffix = explode(',', UploadHandler::getAllowedExt(json_decode($this->policy["filetype"],true)));
		$sufficCheck = !in_array($fileSuffix,$allowedSuffix);
		if(empty(UploadHandler::getAllowedExt(json_decode($this->policy["filetype"],true)))){
			$sufficCheck = false;
		}
		var_dump($sufficCheck);
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