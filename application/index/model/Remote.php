<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \app\index\model\Option;

class Remote extends Model{

	public $sk;
	private $policy;
	private $serverOutput;
	private $httpCode;

	public function __construct($policy){
		$this->policy = $policy;
	}

	public function remove($fileList){
		$signKey = $this->sign($fileList,"DELETE");
		$this->send("manager.php",$signKey,"DELETE",base64_encode(json_encode($fileList)));
	}

	public function preview($fname){
		return $this->signUrl($this->policy["url"]."object.php?action=preview&name=".urlencode($fname)."&expires=".(time()+(int)Option::getValue("timeout")));
	}

	public function clean(){
		return $this->signUrl($this->policy["url"]."object.php?action=clean&expires=".(time()+(int)Option::getValue("timeout")));
	}

	public function download($fname,$attnanme){
		return $this->signUrl($this->policy["url"]."object.php?action=download&name=".urlencode($fname)."&attaname=".urlencode($attnanme)."&expires=".(time()+(int)Option::getValue("timeout")));
	}
	
	public function thumb($fname,$picInfo){
		return $this->signUrl($this->policy["url"]."object.php?action=thumb&name=".urlencode($fname)."&expires=".(time()+(int)Option::getValue("timeout"))."&w=".$picInfo[0]."&h=".$picInfo[1]);
	}

	public function signUrl($url){
		$signKey = hash_hmac("sha256",$url,"GET".$this->policy["sk"]);
		return $url."&auth=".$signKey;
	}

	public function updateContent($fname,$content){
		$object = ["fname"=>$fname,"content"=>$content];
		$signKey = $this->sign($object,"UPDATE");
		$this->send("manager.php",$signKey,"UPDATE",base64_encode(json_encode($object)));
	}

	public function send($target,$auth,$action,$object){
		$session = curl_init($this->policy["server"].$target);
		$postData = array(
			"action" => $action,
			"auth" => $auth,
			"object" => $object,
		);
		curl_setopt($session, CURLOPT_POST, 1);
		curl_setopt($session, CURLOPT_POSTFIELDS, $postData);
		curl_setopt($session, CURLOPT_RETURNTRANSFER, 1);
		curl_setopt($session, CURLOPT_SSL_VERIFYPEER, false);
		curl_setopt($session, CURLOPT_SSL_VERIFYHOST, false);
		$this->serverOutput = curl_exec($session);
		$this->httpCode = curl_getinfo($session,CURLINFO_HTTP_CODE); 
		echo $this->serverOutput;
	}

	public function sign($content,$method = null){
		return hash_hmac("sha256",base64_encode(json_encode($content)),$method.$this->policy["sk"]);
	}

}
?>