<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \app\index\model\Option;
use \app\index\model\FileManage;

use think\console\Input;
use think\console\Output;

use \Krizalys\Onedrive\Client;

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


	public function __construct($id=null){
		if($id!==null){

		}
	}

	public function saveTask(){
		Db::name("task")->insert([
			"task_name" => $this->taskName,
			"attr" => $this->taskContent,
			"type" => $this->taskType,
			"status" => "todo",
			"uid" => $this->userId,
		]);
	}

	public function Do(){
		switch ($this->taskModel["type"]){
			case "uploadSingleToOnedrive":
				$this->uploadSingleToOnedrive();
				break;
			default:
				$this->output->writeln("Unknown task type");
				break;
		}
	}

	private function uploadSingleToOnedrive(){
		$this->taskContent = json_decode($this->taskModel["attr"],true);
		$policyData = Db::name("policy")->where("id",$this->taskContent["policyId"])->find();
		$onedrive = new Client([
			'stream_back_end' => \Krizalys\Onedrive\StreamBackEnd::TEMP,
			'client_id' => $policyData["bucketname"],
		
			// Restore the previous state while instantiating this client to proceed in
			// obtaining an access token.
			'state' => json_decode($policyData["sk"]),
		]);
		
		$filePath = ROOT_PATH . 'public/uploads/'.$this->taskContent["savePath"] . "/" . $this->taskContent["objname"];
		if($file = @fopen($filePath,"r")){
			try{
				$onedrive->createFile(urlencode($this->taskContent["objname"]),"/me/drive/root:/".$this->taskContent["savePath"],$file);
			}catch(\Exception $e){
				$this->status="error";
				$this->errorMsg = $e->getMessage();
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
				// $tmpFileName = $Uploadinfo->getSaveName();
				// unset($Uploadinfo);
				// $this->setError($addAction[1],true,$tmpFileName,$savePath);
			}

			//TO-DO删除本地文件
			
			fclose($file);
		}else{
			$this->status = "error";
			$this->errorMsg = "Failed to open file [".$filePath."]";
		}
		
	}

}

?>