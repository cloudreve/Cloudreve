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

	/**
	 * 新建普通URL下载任务
	 *
	 * @param string $url
	 * @return void
	 */
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
			$this->reqMsg = isset($respondData["error"]["message"]) ? $respondData["error"]["message"] : $this->reqMsg;
		}
	}

	/**
	 * 新建种子下载任务
	 *
	 * @param string $torrentUrl 种子URL
	 * @return void
	 */
	public function addTorrent($torrentUrl){
		$this->pathId = uniqid();
		$reqFileds = [
				"params" => ["token:".$this->authToken,
						[$torrentUrl],["dir" => $this->savePath.$this->pathId],
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
			$this->reqMsg = isset($respondData["error"]["message"]) ? $respondData["error"]["message"] : $this->reqMsg;
		}
	}

	/**
	 * 刷新下载状态
	 *
	 * @param int $id       任务ID
	 * @param int $uid      用户ID
	 * @param array $policy 上传策略
	 * @return void
	 */
	public function flushStatus($id,$uid,$policy){
		$this->uid = $uid;
		if(empty($policy)){
			$user = Db::name("users")->where("id",$uid)->find();
			$group = Db::name("groups")->where("id",$user["user_group"])->find();
			$policy = Db::name("policy")->where("id",$group["policy_name"])->find();
		}
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
				if($downloadInfo["is_single"] && count($respondData["result"]["files"]) >1){
					$this->updateToMuiltpe($respondData["result"],$downloadInfo);
					return false;
				}
				if(isset($respondData["result"]["followedBy"])){
					Db::name("download")->where("id",$id)
						->update([
								"pid" => $respondData["result"]["followedBy"][0],
							]);
					return false;
				}
				Db::name("download")->where("id",$id)
				->update([
					"status" => $respondData["result"]["status"],
					"last_update" => date("Y-m-d h:i:s"),
					"info" => json_encode([
							"completedLength" => $respondData["result"]["files"][$downloadInfo["file_index"]]["completedLength"],
							"totalLength" => $respondData["result"]["files"][$downloadInfo["file_index"]]["length"],
							"dir" => $respondData["result"]["files"][$downloadInfo["file_index"]]["path"],
							"downloadSpeed" => $respondData["result"]["downloadSpeed"],
							"errorMessage" => isset($respondData["result"]["errorMessage"]) ? $respondData["result"]["errorMessage"] : "",
						]),
					"msg" => isset($respondData["result"]["errorMessage"]) ? $respondData["result"]["errorMessage"] : "",
					"total_size" =>  $respondData["result"]["files"][$downloadInfo["file_index"]]["length"],
					]);
				switch ($respondData["result"]["status"]) {
					case 'complete':
						$this->setComplete($respondData["result"],$downloadInfo);
						break;
					case 'removed':
						$this->setCanceled($respondData["result"],$downloadInfo);
						break;
					default:
						# code...
						break;
				}
				if(($respondData["result"]["files"][$downloadInfo["file_index"]]["completedLength"] == $respondData["result"]["files"][$downloadInfo["file_index"]]["length"] && ($respondData["result"]["files"][$downloadInfo["file_index"]]["length"] !=0 )) && $respondData["result"]["status"]=="active"){
					$this->setComplete($respondData["result"],$downloadInfo,$downloadInfo["file_index"]);
					Db::name("download")->where("id",$id)
				    ->update([
				    		"status" => "complete",
				    	]);
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
			$this->setError($respondData,$downloadInfo,$respondData["error"]["message"],"error",true);
				return false;
		}
		return true;
	}

	/**
	 * 取消任务
	 *
	 * @param array $quenInfo 任务信息（aria2）
	 * @param array $sqlData  任务信息（数据库）
	 * @return void
	 */
	private function setCanceled($quenInfo,$sqlData){
		@self::remove_directory($this->savePath.$sqlData["path_id"]);
		if(!is_dir($this->savePath.$sqlData["path_id"])){
			Db::name("download")->where("id",$sqlData["id"])->update([
				"status" => "canceled",
			]);
		}
	}

	/**
	 * 移除整个目录
	 *
	 * @param string $dir
	 * @return void
	 */
	static function remove_directory($dir){
		if($handle=opendir("$dir")){
			while(false!==($item=readdir($handle))){
				if($item!="."&&$item!=".."){
					if(is_dir("$dir/$item")){
						self::remove_directory("$dir/$item");
					}else{
						unlink("$dir/$item");
					}
				}
			}
			closedir($handle);
			rmdir($dir);
		}
	}

	/**
	 * 将单文件任务升级至多文件任务
	 *
	 * @param array $quenInfo
	 * @param array $sqlData
	 * @return void
	 */
	private function updateToMuiltpe($quenInfo,$sqlData){
		foreach ($quenInfo["files"] as $key => $value) {
			Db::name("download")->insert([
				"pid" => $sqlData["pid"],
				"path_id" => $sqlData["path_id"],
				"owner" => $sqlData["owner"],
				"save_dir" => $sqlData["save_dir"],
				"status" => "ready",
				"msg" => "",
				"info"=>"",
				"source" =>$sqlData["source"],
				"file_index" => $key,
				"is_single" => 0,
				"total_size" => 0,
			]);
		}
		Db::name("download")->where("id",$sqlData["id"])->delete();
	}

	/**
	 * 下载完成后续处理
	 *
	 * @param array $quenInfo
	 * @param array $sqlData
	 * @param int $fileIndex
	 * @return void
	 */
	private function setComplete($quenInfo,$sqlData,$fileIndex=null){
		if($this->policy["policy_type"] != "local" && $this->policy["policy_type"] != "onedrive"){
			$this->setError($quenInfo,$sqlData,"您当前的上传策略无法使用离线下载");
			return false;
		}
		if($fileIndex==null){
			$this->forceRemove($sqlData["pid"]);
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
			$fileName = $uploadHandller->getObjName($this->policy['namerule'],"local",basename($quenInfo["files"][$sqlData["file_index"]]["path"]));
		}else{
			$fileName = basename($quenInfo["files"][$sqlData["file_index"]]["path"]);
		}
		$generatePath = $uploadHandller->getDirName($this->policy['dirrule']);

		if($this->policy["policy_type"] == "onedrive"){

			$savePath = ROOT_PATH . 'public/uploads/'.$generatePath.DS.$fileName;
			$task = new Task();
			$task->taskName = "Upload RemoteDownload File " .  $quenInfo["files"][$sqlData["file_index"]]["path"] . " to Onedrive";
			$task->taskType = $quenInfo["files"][$sqlData["file_index"]]["length"]<=4*1024*1024 ? "UploadRegularRemoteDownloadFileToOnedrive" :"UploadLargeRemoteDownloadFileToOnedrive";
			@list($width, $height, $type, $attr) = getimagesize($quenInfo["files"][$sqlData["file_index"]]["path"]);
			$picInfo = empty($width)?"":$width.",".$height;
			$task->taskContent = json_encode([
				"path" => ltrim(str_replace("/", ",", $sqlData["save_dir"]),","),
				"fname" => basename($quenInfo["files"][$sqlData["file_index"]]["path"]),
				"originPath" => $quenInfo["files"][$sqlData["file_index"]]["path"],
				"objname" => $fileName,
				"savePath" => $generatePath,
				"fsize" => $quenInfo["files"][$sqlData["file_index"]]["length"],
				"picInfo" => $picInfo,
				"policyId" => $this->policy["id"],
			]);
			$task->userId = $this->uid;
			$task->saveTask();

		}else{
			
			$savePath = ROOT_PATH . 'public/uploads/'.$generatePath.DS.$fileName;
			is_dir(dirname($savePath))? :mkdir(dirname($savePath),0777,true);
			rename($quenInfo["files"][$sqlData["file_index"]]["path"],$savePath);
			@unlink(dirname($quenInfo["files"][$sqlData["file_index"]]["path"]));
			$jsonData = array(
				"path" => ltrim(str_replace("/", ",", $sqlData["save_dir"]),","),
				"fname" => basename($quenInfo["files"][$sqlData["file_index"]]["path"]),
				"objname" => $generatePath.DS.$fileName,
				"fsize" => $quenInfo["files"][$sqlData["file_index"]]["length"],
			);
			@list($width, $height, $type, $attr) = getimagesize($savePath);
			$picInfo = empty($width)?" ":$width.",".$height;
			$addAction = FileManage::addFile($jsonData,$this->policy,$this->uid,$picInfo);
			if(!$addAction[0]){
				//取消任务
				$this->setError($quenInfo,$sqlData,$addAction[1]);
				return false;
			}

		}
		
		FileManage::storageCheckOut($this->uid,$quenInfo["files"][$sqlData["file_index"]]["length"]);
	}

	/**
	 * 设置任务为失败状态
	 *
	 * @param array $quenInfo
	 * @param array $sqlData
	 * @param string $msg     失败消息
	 * @param string $status  状态
	 * @param boolean $delete 是否删除下载文件
	 * @return void
	 */
	private function setError($quenInfo,$sqlData,$msg,$status="error",$delete=true){
		$this->Remove($sqlData["pid"],$sqlData);
		$this->removeDownloadResult($sqlData["pid"],$sqlData);
		if($delete){
			if(isset($quenInfo["files"][$sqlData["file_index"]]["path"]) && file_exists($quenInfo["files"][$sqlData["file_index"]]["path"])){
				@unlink($quenInfo["files"][$sqlData["file_index"]]["path"]);
				@self::remove_directory(dirname($quenInfo["files"][$sqlData["file_index"]]["path"]));
			}
		}
		Db::name("download")->where("id",$sqlData["id"])->update([
			"msg" => $msg,
			"status" => $status,
			]);
	}

	/**
	 * 移除任务
	 *
	 * @param int $gid
	 * @param array $sqlData
	 * @return void
	 */
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

	/**
	 * 删除下载结果
	 *
	 * @param int $gid
	 * @param array $sqlData
	 * @return void
	 */
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

	/**
	 * 强制移除任务
	 *
	 * @param int $gid
	 * @return void
	 */
	public function forceRemove($gid){
		$reqFileds = [
				"params" => ["token:".$this->authToken,$gid],
				"jsonrpc" => "2.0",
				"id" => uniqid(),
				"method" => "aria2.forceRemove"
			];
		$reqFileds = json_encode($reqFileds,JSON_OBJECT_AS_ARRAY);
		$respondData = $this->sendReq($reqFileds);
		if(isset($respondData["result"])){
			return true;
		}
		return false;
	}

	/**
	 * 检查容量
	 *
	 * @param array $quenInfo
	 * @param array $sqlData
	 * @return void
	 */
	private function storageCheck($quenInfo,$sqlData){
		if(!FileManage::sotrageCheck($this->uid,$quenInfo["totalLength"])){
			return false;
		}
		if(!FileManage::sotrageCheck($this->uid,$quenInfo["completedLength"])){
			return false;
		}
		return true;
	}

	/**
	 * 发送请求
	 *
	 * @param string $data
	 * @return array
	 */
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