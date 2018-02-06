<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \app\index\model\Option;

class B2 extends Model{

	public $policy;
	private $ak;

	public function __construct($policy){
		$this->policy = $policy;
		$akInfo = explode(":",$this->policy["ak"]);
		$this->ak = $akInfo[0];
		if(time()-$akInfo[1]>=86000){
			$this->updateAuth();
		}
	}

	private function updateAuth(){
		$credentials = base64_encode($this->ak . ":" . $this->policy["sk"]);
		$url = "https://api.backblazeb2.com/b2api/v1/b2_authorize_account";
		$session = curl_init($url);
		$headers = array();
		$headers[] = "Accept: application/json";
		$headers[] = "Authorization: Basic " . $credentials;
		curl_setopt($session, CURLOPT_HTTPHEADER, $headers); 
		curl_setopt($session, CURLOPT_HTTPGET, true); 
		curl_setopt($session, CURLOPT_RETURNTRANSFER, true);
		curl_setopt($session, CURLOPT_SSL_VERIFYPEER, false); //不验证证书
		curl_setopt($session, CURLOPT_SSL_VERIFYHOST, false); //不验证证书
		$server_output = curl_exec($session);
		curl_close ($session);
		$authInfo = json_decode($server_output,true);
		Db::name("policy")->where("id",$this->policy["id"])
		->update([
			"ak" => $this->ak . ":" . time(),
			"op_name" => $authInfo["apiUrl"],
			"url" => $authInfo["downloadUrl"],
			"op_pwd" => $authInfo["authorizationToken"],
			]);
		$this->policy = Db::name("policy")->where("id",$this->policy["id"])->find();
		$this->updateUploadAuth();
	}

	private function updateUploadAuth(){
		$session = curl_init($this->policy["op_name"] .  "/b2api/v1/b2_get_upload_url");
		$data = array("bucketId" => $this->policy["bucketname"]);
		$post_fields = json_encode($data);
		curl_setopt($session, CURLOPT_POSTFIELDS, $post_fields); 
		$headers = array();
		$headers[] = "Authorization: " . $this->policy["op_pwd"];
		curl_setopt($session, CURLOPT_HTTPHEADER, $headers); 
		curl_setopt($session, CURLOPT_SSL_VERIFYPEER, false); //不验证证书
		curl_setopt($session, CURLOPT_SSL_VERIFYHOST, false); //不验证证书
		curl_setopt($session, CURLOPT_POST, true);
		curl_setopt($session, CURLOPT_RETURNTRANSFER, true);
		$server_output = curl_exec($session);
		curl_close ($session);
		$authInfo = json_decode($server_output,true);
		Db::name("policy")->where("id",$this->policy["id"])
		->update([
			"server" => $authInfo["uploadUrl"] . "|" .  $authInfo["authorizationToken"],
			]);
		$this->policy = Db::name("policy")->where("id",$this->policy["id"])->find();
		die("");
	}

}
?>