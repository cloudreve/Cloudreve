<?php
namespace app\index\controller;

use think\Controller;
use app\index\model\User;
use think\Cookie;
use think\Db;
use \app\index\model\Option;
use \app\index\model\Avatar;
use \PHPGangsta_GoogleAuthenticator;
use \app\index\model\TwoFactor;

class Member extends Controller{

	public $userObj;

	/**
	 * [index description]
	 * @return [type] [description]
	 */
	public function index(){
		echo "Pong";
	}

	/**
	 * [Register description]
	 */
	public function Register(){
		if(input('?post.username-reg') && input('?post.password-reg')){
			$regAction = User::register(input('post.username-reg'),input('post.password-reg'),input('post.captchaCode'));
			if ($regAction[0]){
				return json(['code' => '200','message' => $regAction[1]]);
			}else{
				return json(['code' => '1','message' => $regAction[1]]);
			}
		}else{
			return json(['code' => '1','message' => "信息不完整"]);
		}
	}

	public function ForgetPwd(){
		if(input('?post.regEmail')  && !empty(input('post.regEmail'))){
			$findAction = User::findPwd(input('post.regEmail'),input('post.captchaCode'));
			if ($findAction[0]){
				return json(['code' => '200','message' => $findAction[1]]);
			}else{
				return json(['code' => '1','message' => $findAction[1]]);
			}
		}else{
			return json(['code' => '1','message' => "信息不完整"]);
		}
	}

	/**
	 * [Login description]
	 */
	public function Login(){
		if(input('?post.userMail') && input('?post.userPass')){
			$logAction = User::login(input('post.userMail'),input('post.userPass'),input('post.captchaCode'));
			if ($logAction[0]){
				return json(['code' => '200','message' => '登陆成功']);
			}else{
				return json(['code' => '1','message' => $logAction[1]]);
			}
		}else{
			return json(['code' => '1','message' => "信息不完整"]);
		}
	}

	/**
	 * [LogOut description]
	 */
	public function LogOut(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$this->userObj->clear();
		$this->redirect("/Login",302);

	}

	public function Memory(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$this->userObj->getMemory();
	}

	public function LoginForm(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$this->isLoginStatusCheck();
		return view('login', [
			'options'  => Option::getValues(['basic']),
			'RegOptions'  => Option::getValues(['register','login']),
			'loginStatus' => $this->userObj->loginStatus,
		]);
	} 

	public function TwoStepCheck(){
		$checkCode = input("post.code");
		if(empty($checkCode)){
			return json(['code' => '1','message' => "验证码不能为空"]);
		}
		$userId = session("user_id_tmp");
		$userData = Db::name('users')->where('id',$userId)->find();
		$ga = new PHPGangsta_GoogleAuthenticator();
		$checkResult = $ga->verifyCode($userData["two_step"], $checkCode, 2);
		if($checkResult) {
			cookie('user_id',session("user_id_tmp"),604800);
			cookie('login_status',session("login_status_tmp"),604800);
			cookie('login_key',session("login_key_tmp"),604800);
			return json(['code' => '200','message' => '登陆成功']);
		}else{
			return json(['code' => '1','message' => "验证失败"]);
		}
	}

	public function TwoStep(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$this->isLoginStatusCheck();
		return view('two_step', [
			'options'  => Option::getValues(['basic']),
			'RegOptions'  => Option::getValues(['register','login']),
			'loginStatus' => $this->userObj->loginStatus,
		]);
	}

	public function setWebdavPwd(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$this->loginStatusCheck();
		Db::name("users")->where("id",$this->userObj->uid)
		->update([
			"webdav_key" => md5($this->userObj->userSQLData["user_email"].":CloudreveWebDav:".input("post.pwd")),
			]);
		return json(['error' => '200','msg' => '设置成功']);
	}

	public function emailActivate(){
		$activationKey = input('param.key');
		$basicOptions = Option::getValues(['basic']);
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$this->isLoginStatusCheck();
		$activeAction = User::activicateUser($activationKey);
		if($activeAction[0]){
			return view('active_user', [
			'options'  => $basicOptions,
			'loginStatus' => $this->userObj->loginStatus,
		]);
		}else{
			$this->error($activeAction[1],403,$basicOptions);
		}
	}

	public function resetPwd(){
		$resetKey = input('param.key');
		$userId = input('get.uid');
		$basicOptions = Option::getValues(['basic']);
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$this->isLoginStatusCheck();
		$resetAction = User::resetUser($resetKey,$userId);
		if($resetAction[0]){
			return view('reset_user', [
			'options'  => $basicOptions,
			'loginStatus' => $this->userObj->loginStatus,
			'key' => $resetKey."_".$userId,
		]);
		}else{
			$this->error($resetAction[1],403,$basicOptions);
		}
	}

	public function Reset(){
		$newPwd = input('post.pwd');
		$resetKey = input('post.key');
		$resetAction = User::resetPwd($resetKey,$newPwd);
		if($resetAction[0]){
			return json(['code' => '200','message' => '重设成功，请前往登录页登录']);
		}else{
			return json(['code' => '1','message' => $resetAction[1]]);
		}
	}

	public function Setting(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$userInfo = $this->userObj->getInfo();
		$this->loginStatusCheck();
		$policyList=[];
		foreach (explode(",",$this->userObj->groupData["policy_list"]) as $key => $value) {
			$policyList[$key] = $value;
		}
		$avaliablePolicy = Db::name("policy")->where("id","in",$policyList)->select();
		$basicOptions = Option::getValues(['basic']);
		return view('setting', [
			'options'  => $basicOptions,
			'userInfo' => $userInfo,
			'userSQL' => $this->userObj->userSQLData,
			'groupData' => $this->userObj->groupData,
			'loginStatus' => $this->userObj->loginStatus,
			'avaliablePolicy' => $avaliablePolicy,
		]);
	}

	public function SaveAvatar(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$file = request()->file("avatar");
		$avatarObj = new Avatar(true,$file);
		if(!$avatarObj->SaveAvatar()){
			return json_encode($avatarObj->errorMsg);
		}else{
			$avatarObj->bindUser($this->userObj->uid);
			return json_encode(["result" => "success"]);
		}
	}

	public function Avatar(){
		if(!input("get.cache")=="no"){
			header("Cache-Control: max-age=10800");
		}
		$userId = input("param.uid");
		$avatarObj = new Avatar(false,$userId);
		$avatarImg = $avatarObj->Out(input("param.size"));
		$this->redirect($avatarImg,302);
	}

	public function SetGravatar(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$avatarObj = new Avatar(false,$this->userObj->uid);
		$avatarObj->setGravatar();
	}

	public function Nick(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$userInfo = $this->userObj->getInfo();
		$this->loginStatusCheck();
		$saveAction = $this->userObj->changeNick(input("post.nick"));
		if($saveAction[0]){
			return json(['error' => '200','msg' => '设置成功']);
		}else{
			return json(['error' => '1','msg' => $saveAction[1]]);
		}
	}

	public function HomePage(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$userInfo = $this->userObj->getInfo();
		$this->loginStatusCheck();
		$saveAction = $this->userObj->homePageToggle(input("post.status"));
		if($saveAction[0]){
			return json(['error' => '200','msg' => '设置成功']);
		}else{
			return json(['error' => '1','msg' => $saveAction[1]]);
		}
	}

	public function EnableTwoFactor(){
		$twoFactor = new TwoFactor();
		$twoFactor->qrcodeRender();
	}

	public function TwoFactorConfirm(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$userInfo = $this->userObj->getInfo();
		$this->loginStatusCheck();
		$twoFactor = new TwoFactor();
		$confirmResult = $twoFactor->confirmCode(session("two_factor_enable"),input("post.code"));
		if($confirmResult[0]){
			$twoFactor->bindUser($this->userObj->uid);
			return json(['error' => '200','msg' => '设置成功']);
		}else{
			return json(['error' => '1','msg' => $confirmResult[1]]);
		}
	}

	public function ChangePwd(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		$userInfo = $this->userObj->getInfo();
		$this->loginStatusCheck();
		$changeAction = $this->userObj->changePwd(input("post.origin"),input("post.new"));
		if($changeAction[0]){
			return json(['error' => '200','msg' => '设置成功']);
		}else{
			return json(['error' => '1','msg' => $changeAction[1]]);
		}
	}

	private function loginStatusCheck($login=true){
		if(!$this->userObj->loginStatus){
			if($login){
				$this->redirect(url('/Login','',''));
			}else{
				$this->redirect(url('/Home','',''));
			}
			exit();
		}
	}

	private function isLoginStatusCheck(){
		if($this->userObj->loginStatus){
			$this->redirect(url('/Home','',''));
			exit();
		}
	}

}
