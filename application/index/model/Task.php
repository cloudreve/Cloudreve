<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \app\index\model\Option;
use \app\index\model\FileManage;

use think\console\Input;
use think\console\Output;

use \Krizalys\Onedrive\Client;
use Sabre\DAV\Mock\File;

class Task extends Model{

	public $taskModel; 
	public $taskName;
	public $taskType;
	public $taskContent;
	public $input;
	public $output;
	public $userId;

	public $status = "success";
	public $errorMsg;
	public $policyModel;


	public function __construct($id=null){
		if($id!==null){

		}
	}

	/**
	 * 保存任务至数据库
	 *
	 * @return void
	 */
	public function saveTask(){
		Db::name("task")->insert([
			"task_name" => $this->taskName,
			"attr" => $this->taskContent,
			"type" => $this->taskType,
			"status" => "todo",
			"uid" => $this->userId,
		]);
	}

	/**
	 * 开始执行任务
	 *
	 * @return void
	 */
	public function Doit(){
		switch ($this->taskModel["type"]){
			case "uploadSingleToOnedrive":
				$this->uploadSingleToOnedrive();
				break;
			case "UploadRegularRemoteDownloadFileToOnedrive":
				$this->uploadSingleToOnedrive();
				break;
			case "uploadChunksToOnedrive":
				$this->uploadChunksToOnedrive();
				break;
			case "UploadLargeRemoteDownloadFileToOnedrive":
				$this->uploadUnchunkedFile();
				break;
			default:
				$this->output->writeln("Unknown task type (".$this->taskModel["type"].")");
				break;
		}
	}

	/**
	 * 上传未分片的大文件至Onedrive
	 *
	 * @return void
	 */
	private function uploadUnchunkedFile(){
		$this->taskContent = json_decode($this->taskModel["attr"],true);
		$policyData = Db::name("policy")->where("id",$this->taskContent["policyId"])->find();
		$this->policyModel = $policyData;
		$onedrive = new Client([
			'stream_back_end' => \Krizalys\Onedrive\StreamBackEnd::TEMP,
			'client_id' => $policyData["bucketname"],
		
			// Restore the previous state while instantiating this client to proceed in
			// obtaining an access token.
			'state' => json_decode($policyData["sk"]),
		]);

		//创建分片上传Session,获取上传URL
		try{
			$uploadUrl = $onedrive->apiPost("/me/drive/root:/".rawurlencode($this->taskContent["savePath"] . "/" . $this->taskContent["objname"]).":/createUploadSession",[])->uploadUrl;
		}catch(\Exception $e){
			$this->status="error";
			$this->errorMsg = $e->getMessage();
			$this->cleanTmpChunk();
			return;
		}
		//创建分片上传Session,获取上传URL
		try{
			$uploadUrl = $onedrive->apiPost("/me/drive/root:/".rawurlencode($this->taskContent["savePath"] . "/" . $this->taskContent["objname"]).":/createUploadSession",[])->uploadUrl;
		}catch(\Exception $e){
			$this->status="error";
			$this->errorMsg = $e->getMessage();
			$this->cleanTmpChunk();
			return;
		}

		//每次4MB上传文件
		
		if(!$file = @fopen($this->taskContent["originPath"],"r")){
			$this->status="error";
			$this->errorMsg = "File not exist.";
			$this->cleanTmpChunk();
			return;
		}
		$offset = 0;
		$totalSize = filesize($this->taskContent["originPath"]);
		while (1) {
			//移动文件指针
			fseek($file, $offset);

			$chunksize = (($offset+4*1024*1024)>$totalSize)?($totalSize-$offset):1024*4*1024;
			$headers = [];
			$headers[] = "Content-Length: ".$chunksize;
			$headers[] = "Content-Range: bytes ".$offset."-".($offset+$chunksize-1)."/".$this->taskContent["fsize"];

			//发送单个分片数据
			try{
				$onedrive->sendFileChunk($uploadUrl,$headers,fread($file,$chunksize));
			}catch(\Exception $e){
				$this->status="error";
				$this->errorMsg = $e->getMessage();
				$this->cleanTmpChunk();
				return;
			}
			$this->output->writeln("[Info] Chunk Uploaded. Offset:".$offset);
			$offset+=$chunksize;
			if($offset+1 >=$totalSize){
				break;
			}
			
		}
		fclose($file);
		$jsonData = array(
			"path" => $this->taskContent["path"], 
			"fname" => $this->taskContent["fname"],
			"objname" => $this->taskContent["savePath"]."/".$this->taskContent["objname"],
			"fsize" => $this->taskContent["fsize"],
		);

		$addAction = FileManage::addFile($jsonData,$policyData,$this->taskModel["uid"],$this->taskContent["picInfo"]);
		if(!$addAction[0]){
			$this->setError($addAction[1],true,"/me/drive/root:/".rawurlencode($this->taskContent["savePath"] . "/" . $this->taskContent["objname"]),$onedrive);
			$this->cleanTmpChunk();
			return;
		}

		$this->cleanTmpChunk();
		
	}

	/**
	 * 上传已分片的大文件至Onedrive
	 *
	 * @return void
	 */
	private function uploadChunksToOnedrive(){
		$this->taskContent = json_decode($this->taskModel["attr"],true);
		$policyData = Db::name("policy")->where("id",$this->taskContent["policyId"])->find();
		$this->policyModel = $policyData;
		$onedrive = new Client([
			'stream_back_end' => \Krizalys\Onedrive\StreamBackEnd::TEMP,
			'client_id' => $policyData["bucketname"],
		
			// Restore the previous state while instantiating this client to proceed in
			// obtaining an access token.
			'state' => json_decode($policyData["sk"]),
		]);

		//创建分片上传Session,获取上传URL
		try{
			$uploadUrl = $onedrive->apiPost("/me/drive/root:/".rawurlencode($this->taskContent["savePath"] . "/" . $this->taskContent["objname"]).":/createUploadSession",[])->uploadUrl;
		}catch(\Exception $e){
			$this->status="error";
			$this->errorMsg = $e->getMessage();
			$this->cleanTmpChunk();
			return;
		}

		//逐个上传文件分片
		$offset = 0;
		foreach ($this->taskContent["chunks"] as $key => $value) {
			$chunkPath = ROOT_PATH . 'public/uploads/chunks/'.$value["obj_name"].".chunk";
			if(!$file = @fopen($chunkPath,"r")){
				$this->status="error";
				$this->errorMsg = "File chunk not exist.";
				$this->cleanTmpChunk();
				return;
			}
			$headers = [];
			$chunksize = filesize($chunkPath);
			$headers[] = "Content-Length: ".$chunksize;
			$headers[] = "Content-Range: bytes ".$offset."-".($offset+$chunksize-1)."/".$this->taskContent["fsize"];

			//发送单个分片数据
			try{
				$onedrive->sendFileChunk($uploadUrl,$headers,$file);
			}catch(\Exception $e){
				$this->status="error";
				$this->errorMsg = $e->getMessage();
				$this->cleanTmpChunk();
				return;
			}
			$this->output->writeln("[Info] Chunk Uploaded. Offset:".$offset);
			$offset += $chunksize;
			fclose($file);
			
		}

		$jsonData = array(
			"path" => $this->taskContent["path"], 
			"fname" => $this->taskContent["fname"],
			"objname" => $this->taskContent["savePath"]."/".$this->taskContent["objname"],
			"fsize" => $this->taskContent["fsize"],
		);

		$addAction = FileManage::addFile($jsonData,$policyData,$this->taskModel["uid"],$this->taskContent["picInfo"]);
		if(!$addAction[0]){
			$this->setError($addAction[1],true,"/me/drive/root:/".rawurlencode($this->taskContent["savePath"] . "/" . $this->taskContent["objname"]),$onedrive);
			$this->cleanTmpChunk();
			return;
		}

		$this->cleanTmpChunk();


	}

	/**
	 * 上传单文件(<=4mb)至Onedrive
	 *
	 * @return void
	 */
	private function uploadSingleToOnedrive(){
		$this->taskContent = json_decode($this->taskModel["attr"],true);
		$policyData = Db::name("policy")->where("id",$this->taskContent["policyId"])->find();
		$this->policyModel = $policyData;
		$onedrive = new Client([
			'stream_back_end' => \Krizalys\Onedrive\StreamBackEnd::TEMP,
			'client_id' => $policyData["bucketname"],
		
			// Restore the previous state while instantiating this client to proceed in
			// obtaining an access token.
			'state' => json_decode($policyData["sk"]),
		]);
		
		$filePath = $this->taskModel["type"] == "UploadRegularRemoteDownloadFileToOnedrive"?$this->taskContent["originPath"]:ROOT_PATH . 'public/uploads/'.$this->taskContent["savePath"] . "/" . $this->taskContent["objname"];
		if($file = @fopen($filePath,"r")){
			try{
				$onedrive->createFile(rawurlencode($this->taskContent["objname"]),"/me/drive/root:/".$this->taskContent["savePath"],$file);
			}catch(\Exception $e){
				$this->status="error";
				$this->errorMsg = $e->getMessage();
				$this->cleanTmpFile();
				return;
			}

			$jsonData = array(
				"path" => $this->taskContent["path"], 
				"fname" => $this->taskContent["fname"],
				"objname" => $this->taskContent["savePath"]."/".$this->taskContent["objname"],
				"fsize" => $this->taskContent["fsize"],
			);

			$addAction = FileManage::addFile($jsonData,$policyData,$this->taskModel["uid"],$this->taskContent["picInfo"]);
			if(!$addAction[0]){
				$this->setError($addAction[1],true,"/me/drive/root:/".$this->taskContent["savePath"]."/".rawurlencode($this->taskContent["objname"]),$onedrive);
				$this->cleanTmpFile();
				return;
			}
			
			fclose($file);
			$this->cleanTmpFile();
		}else{
			$this->status = "error";
			$this->errorMsg = "Failed to open file [".$filePath."]";
		}
		
	}
	
	/**
	 * 删除本地临时文件
	 *
	 * @return bool 是否成功
	 */
	private function cleanTmpFile(){
		if($this->taskModel["type"] == "UploadRegularRemoteDownloadFileToOnedrive"){
			return @unlink($this->taskContent["originPath"]);
		}else{
			return @unlink(ROOT_PATH . 'public/uploads/'.$this->taskContent["savePath"] . "/" . $this->taskContent["objname"]);
		}
		
	}

	/**
	 * 删除本地临时分片
	 *
	 * @return void
	 */
	private function cleanTmpChunk(){
		if($this->taskModel["type"] == "UploadLargeRemoteDownloadFileToOnedrive"){
			@unlink($this->taskContent["originPath"]);
		}else{
			foreach ($this->taskContent["chunks"] as $key => $value) {
				@unlink( ROOT_PATH . 'public/uploads/chunks/'.$value["obj_name"].".chunk");
			}
		}
	}

	/**
	 * 设置为出错状态并清理远程文件
	 *
	 * @param string $msg    错误消息
	 * @param bool $delete   是否删除文件
	 * @param string $path   文件路径
	 * @param mixed $adapter 远程操作适配器
	 * @return void
	 */
	private function setError($msg,$delete,$path,$adapter){
		$this->status="error";
		$this->errorMsg = $msg;
		if($delete){
			switch($this->taskModel["type"]){
			case "uploadSingleToOnedrive":
				$adapter->deleteObject($path);
				break;
			case "uploadChunksToOnedrive":
				$adapter->deleteObject($path);
				break;
			default:
				
				break;
			}
		}
		FileManage::storageGiveBack($this->taskModel["uid"],$this->taskContent["fsize"]);
	}

}

?>