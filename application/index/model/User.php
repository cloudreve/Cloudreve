<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \think\Session;
use \think\Cookie;
use \app\index\model\Option;
use \app\index\model\Mail;
use think\Validate;

class User extends Model{

	public $uid;
	public $varifyKey;
	public $groupId;
	public $regDate;
	public $loginStatus = false;
	public $userNick;
	public $userMail;
	public $groupData;
	public $userSQLData;
	
	/**
	 * [__construct description]
	 * @param [type] $userId  [description]
	 * @param [type] $userKey [description]
	 */
	public function __construct($userId,$userKey,$ignoreLogin = false){
		$userData = Db::name('users')->where('id',$userId)->where('user_status',0)->find();
		if(empty($userData)){
			$this->loginStatus = false;
			$this->setUser();
			return false;
		}
		if(!$ignoreLogin){
			if(md5($userData['user_email'].$userData['user_pass'].config('salt')) != $userKey){
				$this->loginStatus = false;
				$this->setUser();
				return false;
			}
		}
		$this->groupData = Db::name('groups')->where('id',$userData['user_group'])->find();
		$this->loginStatus = true;
		$this->uid = $userId;
		$this->groupId = $userData['user_group'];
		$this->regDate = $userData['user_date'];
		$this->userNick = $userData['user_nick'];
		$this->userMail = $userData['user_email'];
		$this->varifyKey = $userKey;
		$this->userSQLData = $userData;
	}


	public function setUser(){
		$this->groupData = Db::name('groups')->where('id',2)->find();
		$this->uid = -1;
	}

	static function resetPwd($key,$pwd){
		$key = explode("_",$key);
		$resetKey = $key[0]."_".$key[1];
		$userId = $key[2];
		$keyCheck = self::resetUser($resetKey,$userId);
		if(!$keyCheck[0]){
			return $keyCheck;
		}else{
			if ((mb_strlen($pwd,'UTF8')>64) || (mb_strlen($pwd,'UTF8')<4)){
				return [false,"密码不符合规范"];
			}
			self::Reset($userId,$pwd);
			return [true,"密码重设成功"];
		}
	}

	static function Reset($uid,$pwd){
		Db::name('users')->where('id',$uid)
		->update([
			'user_pass' => md5(config('salt').$pwd),
		]);
	}

	/**
	 * [register description]
	 * @param  [type] $userEmail [description]
	 * @param  [type] $userPass  [description]
	 * @return [type]            [description]
	 */
	static function register($userEmail,$userPass,$captchaCode){
		if(Option::getValue("reg_captcha")=="1"){
			if(!self::checkCaptcha($captchaCode)){
				return [false,"验证码错误"];
			}
		}
		if (\app\index\model\Option::getValue("regStatus") == '1'){
			return [false,"当前站点关闭注册"];			
		}
		$userName = str_replace(" ", "", $userEmail);
		$passWord = $userPass;
		if ( !filter_var($userName,FILTER_VALIDATE_EMAIL) || (mb_strlen($userName,'UTF8')>22) || (mb_strlen($userName,'UTF8')<4) || (mb_strlen($passWord,'UTF8')>64) || (mb_strlen($passWord,'UTF8')<4)){
			return [false,"邮箱或密码不符合规范"];
		}
		if(Db::name('users')->where('user_email',$userName)->find() !=null){
			return [false,"该邮箱已被注册"];
		}
		$defaultGroup = (int)\app\index\model\Option::getValue("defaultGroup");
		$regOptions = Option::getValues(["register"]);
		if($regOptions["email_active"] == "1"){
			$activationKey = md5(uniqid(rand(), TRUE));
			$userStatus = 1;
		}else{
			$activationKey = "n";
			$userStatus = 0;
		}
		$sqlData = [
			'user_email' => $userName,
			'user_pass' => md5(config('salt').$passWord),
			'user_status' => $userStatus,
			'user_group' => $defaultGroup,
			'group_primary' => 0,
			'user_date' => date("Y-m-d H:i:s"),
			'user_nick' => explode("@",$userName)[0],
			'user_activation_key' => $activationKey,
			'used_storage' => 0,
			'two_step'=>"0",
			'webdav_key' =>md5(config('salt').$passWord),
			'delay_time' =>0,
			'avatar' => "default",
			'profile' => true,
		];
		if(Db::name('users')->insert($sqlData)){
			$userId = Db::name('users')->getLastInsID();
			Db::name('folders')->insert( [
				'folder_name' => '根目录',
				'parent_folder' => 0,
				'position' => '.',
				'owner' => $userId,
				'date' => date("Y-m-d H:i:s"),
				'position_absolute' => '/',
			]);
			if($regOptions["email_active"] == "1"){
				$options = Option::getValues(["basic","mail_template"]);
				$replace = array(
					'{siteTitle}' =>$options["siteName"],
					'{userName}' =>explode("@",$userName)[0],
					'{siteUrl}' =>$options["siteURL"],
					'{siteSecTitle}' =>$options["siteTitle"],
					'{activationUrl}' =>$options["siteURL"]."Member/emailActivate/".$activationKey,
					);
				$mailContent = strtr($options["mail_activation_template"],$replace);
				$mailObj = new Mail();
				$mailObj->Send($userName,explode("@",$userName)[0],"【".$options["siteName"]."】"."注册激活",$mailContent);
				return [true,"ec"];
			}
			return [true,"注册成功"];
		}
	}

	static function activicateUser($key){
		$userData = Db::name('users')
		->where('user_activation_key','neq','n')
		->where("user_activation_key",$key)->find();
		if(empty($userData)){
			return [0,"激活失败，用户在不存在"];
		}else{
			Db::name('users')->where("id",$userData["id"])->update([
				"user_activation_key" => "n",
				"user_status" => 0,
			]);
			return [1,1];
		}
	}

	static function resetUser($key,$uid){
		$timeNow = time();
		if(empty($key)||empty($uid)){
			return [0,"URL参数错误"];
		}
		$key = explode("_",$key);
		$userData = Db::name('users')
		->where('user_status',0)
		->where("id",$uid)->find();
		if(empty($userData)){
			return [0,"用户不存在"];
		}
		if(md5($userData["user_pass"].$key[1]) != $key[0]){
			return [0,"参数无效，请检查邮件链接"];
		}
		if(($timeNow - $key[1])>7200){
			return [0,"重设链接过期，请重新提交"];
		}
		return [1,1];
	}

	static function findPwd($email,$captchaCode){
		if(Option::getValue("forget_captcha")=="1"){
			if(!self::checkCaptcha($captchaCode)){
				return [false,"验证码错误"];
			}
		}
		$userData = Db::name('users')->where('user_email',$email)->find();
		if(empty($userData)){
			return [1,1];
		}
		$timeNow = time();
		$resetHash = md5($userData["user_pass"].$timeNow);
		$resetKey = $resetHash."_".$timeNow;
		$options = Option::getValues(["basic","mail_template"]);
		$replace = array(
			'{siteTitle}' =>$options["siteName"],
			'{userName}' =>$userData["user_nick"],
			'{siteUrl}' =>$options["siteURL"],
			'{siteSecTitle}' =>$options["siteTitle"],
			'{resetUrl}' =>$options["siteURL"]."Member/resetPwd/".$resetKey."?uid=".$userData["id"],
			);
		$mailContent = strtr($options["mail_reset_pwd_template"],$replace);
		$mailObj = new Mail();
		$mailObj->Send($email,$userData["user_nick"],"【".$options["siteName"]."】"."密码重置",$mailContent);
		return [true,"ec"];
	}

	/**
	 * [login description]
	 * @param  [type] $userEmail [description]
	 * @param  [type] $userPass  [description]
	 * @return [type]            [description]
	 */
	static function login($userEmail,$userPass,$captchaCode){
		$userEmail = str_replace(" ", "", $userEmail);
		$userData =Db::name('users')->where('user_email',$userEmail)->find();
		if(empty($userEmail) || empty($userPass)){
			return [false,"表单不完整"];
		}
		if(Option::getValue("login_captcha")=="1"){
			if(!self::checkCaptcha($captchaCode)){
				return [false,"验证码错误"];
			}
		}
		if(Db::name('users')->where('user_email',$userEmail)->value('user_pass') != md5(config('salt').$userPass)){
			return [false,"用户名或密码错误"];
		}
		if(Db::name('users')->where('user_email',$userEmail)->value('user_status') != 0){
			return [false,"账号被禁用或未激活"];
		}
		if($userData["two_step"] != "0"){
			session("user_id_tmp",Db::name('users')->where('user_email',$userEmail)->value('id'));
			session("login_status_tmp","ok");
			session("login_key_tmp",md5($userEmail.md5(config('salt').$userPass).config('salt')));
			return [false,"tsp"];
		}
		$loginKey = md5($userEmail.md5(config('salt').$userPass).config('salt'));
		cookie('user_id',Db::name('users')->where('user_email',$userEmail)->value('id'),604800);
		cookie('login_status','ok',604800);
		cookie('login_key',$loginKey,604800);
		return [true,"登录成功",$loginKey];
	}

	static function checkCaptcha($code){
		if(!captcha_check($code)){
		 	return false;
		}
		return true;
	}

	/**
	 * [clear description]
	 * @return [type] [description]
	 */
	public function clear(){
		$this->loginStatus = false;
		$this->uid = null;
		$this->groupId = null;
		$this->regDate = null;
		$this->varifyKey = null;
		cookie('user_id', null);
		cookie('login_status', null);
		cookie('login_key', null);
	}

	/**
	 * [getInfo description]
	 * @return [type] [description]
	 */
	public function getInfo(){
		return [
		'uid' => $this->uid,
		'groupId' => $this->groupId,
		'regDate' => $this->regDate,
		'loginStatus' => $this->loginStatus,
		'userNick' => $this->userNick,
		'userMail' => $this->userMail,
		'groupData' => $this->groupData,
		'sqlData' => $this->userSQLData,
		];
	}

	public function getSQLData(){
		return $this->userSQLData;
	}

	public function getPolicy(){
		return Db::name('policy')->where('id',$this->groupData["policy_name"])->find();
	}

	public function getGroupData(){
		return $this->groupData;
	}

	public function getMemory($notEcho = false){
		$usedMemory = $this->userSQLData["used_storage"];
		$groupStorage = $this->groupData["max_storage"];
		$packetStorage = Db::name('storage_pack')
		->where('uid',$this->uid)
		->where('dlay_time',">",time())
		->sum('pack_size');
		$returnData["used"] = self::countSize($usedMemory);
		$returnData["total"] = self::countSize($groupStorage+$packetStorage);
		$returnData["rate"] = floor($usedMemory/($groupStorage+$packetStorage)*100);
		$returnData["basic"] = self::countSize($groupStorage);
		$returnData["pack"] = self::countSize($packetStorage);
		if($usedMemory > $groupStorage){
			$returnData["r1"] = floor($usedMemory/($groupStorage+$packetStorage)*100);
			$returnData["r2"] = 0;
			$returnData["r3"] = 100-$returnData["r1"];
		}else{
			$returnData["r1"] = floor($usedMemory/($groupStorage+$packetStorage)*100);
			$returnData["r2"] = floor(($groupStorage-$usedMemory)/($groupStorage+$packetStorage)*100);;
			$returnData["r3"] = 100-$returnData["r1"]-$returnData["r2"];
		}
		if($notEcho){
			return json_encode($returnData);
		}
		echo json_encode($returnData);
	}

	static function countSize($bit)  {  
		$type = array('Bytes','KB','MB','GB','TB');  
		for($i = 0; $bit >= 1024; $i++) {  
			$bit/=1024;  
		}  
		return (floor($bit*100)/100).$type[$i];  
	}

	public function changeNick($nick){
		$nick=["nick" => $nick];
		$rules = [
		    'nick'  => ['require','max'=>'25','chsDash'],
		];
		$validate = new Validate($rules);
		if (!$validate->check($nick)) {
		    return [0,"昵称必须是1-25位字符，只能包含中英文等常见字符"];
		}else{
			Db::name("users")->where("id",$this->uid)->update(["user_nick" => $nick["nick"]]);
			return [1,1];
		}
	}

	public function changePwd($origin,$new){
		if(md5(config('salt').$origin) != $this->userSQLData["user_pass"]){
			return [0,"原密码错误"];
		}
		if ((mb_strlen($new,'UTF8')>64) || (mb_strlen($new,'UTF8')<4)){
			return [false,"密码不符合规范"];
		}
		self::Reset($this->uid,$new);
		return [true,"密码重设成功"];
	}

	public function homePageToggle($status){
		Db::name("users")->where("id",$this->uid)->update(["profile" => $status=="true"?1:0]);
			return [1,1];
	}

}
?>