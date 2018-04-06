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