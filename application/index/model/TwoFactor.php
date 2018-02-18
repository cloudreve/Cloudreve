<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \think\Session;
use \PHPGangsta_GoogleAuthenticator;
use \Endroid\QrCode\QrCode;
use \app\index\model\Option;

class TwoFactor extends Model{
	
	public $ga;
	public $secretKey;
	public $checkResult;

	public function __construct(){
		$this->ga = new PHPGangsta_GoogleAuthenticator();
	}

	public function qrcodeRender(){
		ob_end_clean();
		$this->secretKey = $this->ga->createSecret();
		session("two_factor_enable",$this->secretKey);
		$qrCode = new QrCode(urldecode(str_replace("https://chart.googleapis.com/chart?chs=200x200&chld=M|0&cht=qr&chl=","",$this->ga->getQRCodeGoogleUrl(Option::getValue("siteName"), $this->secretKey))));
		$qrCode->setSize(165);
		$qrCode->setMargin(0);
		header('Content-Type: '.$qrCode->getContentType());
		echo $qrCode->writeString();
	}


	public function confirmCode($key,$code){
		$this->secretKey = $key;
		if(empty($code)){
			return [0,"验证码不能为空"];
		}
		if(empty($key)){
			return [0,"二维码过期，请刷新页面后重新扫描"];
		}
		$this->checkResult = $this->ga->verifyCode($key, $code, 2);
		if($this->checkResult){
			return [1,"验证成功"];
		}else{
			return [0,"验证码错误"];
		}
	}

	public function bindUser($uid){
		Db::name("users")->where("id",$uid)->update(["two_step" => $this->secretKey]);
	}

}
?>
