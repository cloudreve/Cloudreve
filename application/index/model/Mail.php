<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \app\index\model\Option;
use \PHPMailer\PHPMailer\PHPMailer;
use \PHPMailer\PHPMailer\Exception;

class Mail extends Model{

	public $fromName;
	public $fromAdress;
	public $smtpHost;
	public $smtpPort;
	public $replyTo;
	public $smtpUser;
	public $smtpPass;
	public $encriptionType;
	public $errorMsg;

	public function __construct(){
		$mailOptions = Option::getValues(["mail"]);
		$this->fromName = $mailOptions["fromName"];
		$this->fromAdress = $mailOptions["fromAdress"];
		$this->smtpHost = $mailOptions["smtpHost"];
		$this->smtpPort = $mailOptions["smtpPort"];
		$this->replyTo = $mailOptions["replyTo"];
		$this->smtpUser = $mailOptions["smtpUser"];
		$this->smtpPass = $mailOptions["smtpPass"];
		$this->encriptionType = $mailOptions["encriptionType"];
	}

	public function Send($to,$name,$title,$content){
		$mail = new PHPMailer();
		$mail->isSMTP();
		$mail->SMTPAuth=true;
		$mail->Host = $this->smtpHost;
		if(!empty($this->encriptionType) && $this->encriptionType != "no"){
			$mail->SMTPSecure = $this->encriptionType;
		}
		$mail->Port = $this->smtpPort;
		$mail->CharSet = 'UTF-8';
		$mail->FromName = $this->fromName;
		$mail->Username =$this->smtpUser;
		$mail->Password = $this->smtpPass;
		$mail->From = $this->fromAdress;
		$mail->SMTPDebug = 1;
		$mail->Debugoutput = function($str, $level) {
			$this->errorMsg .= $str;
		}; 
		$mail->isHTML(true); 
		$mail->addAddress($to,$name);
		$mail->Subject = $title;
		$mail->Body = $content;
		$status = $mail->send();
		if(!$status){
			return false;
		}
		return true;
	}
}
?>