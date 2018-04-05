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

	public function __construct($options){
		$this->authToken = $options["aria2_token"];
		$this->apiUrl = rtrim($options["aria2_rpcurl"],"/")."/";
		$this->saveOptions = json_decode($options["aria2_options"],true);
		$this->savePath = $options["aria2_tmppath"];
	}

	public function addUrl($url){
		//{"params": ["token:123123132",["https://www.baidu.com/img/baidu_jgylogo3.gif"],{"dir":"../"}], "jsonrpc": "2.0", "id": "qer", "method": "aria2.addUri"}
		$reqFileds = [
				"params" => ["token:".$this->authToken,[
						$url,["dir" => $this->savePath],
					]],
				"jsonrpc" => "2.0",
				"id" => uniqid(),
				"method" => "aria2.addUri"
			];
		$reqFileds["params"][1][1] = array_merge($reqFileds["params"][1][1],$this->saveOptions);
		$reqFileds = json_encode($reqFileds,JSON_OBJECT_AS_ARRAY);
		$respondData = $this->sendReq($reqFileds);
	}

	private function sendReq($data){
		$curl = curl_init();
	    curl_setopt($curl, CURLOPT_URL, $this->apiUrl."jsonrpc");
	    curl_setopt($curl, CURLOPT_POST, 1);
	    curl_setopt($curl, CURLOPT_POSTFIELDS, $data);
	    curl_setopt($curl, CURLOPT_TIMEOUT, 15); 
	    curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
	    $tmpInfo = curl_exec($curl); // 执行操作
	    if (curl_errno($curl)) {
	    	$this->reqStatus = 0;
	    	$this->reqMsg = "请求失败,".curl_error($curl);
	    }
	    curl_close($curl); // 关闭CURL会话
	    return json_decode($tmpInfo); // 返回数据，json格式	
	}

}
?>