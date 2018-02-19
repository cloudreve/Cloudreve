<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use \app\index\model\Option;
use \app\index\model\User;

class Index extends Controller{

	public $userObj;

    public function index(){

    	$this->userObj = new User(cookie('user_id'),cookie('login_key'));
    	if($this->userObj->loginStatus){
    		$this->redirect(url('/Home','',''));
    		exit();
    	}
    	$userInfo = $this->userObj->getInfo();
    	return view('index', [
    		'options'  => Option::getValues(['basic']),
    		'userInfo' => $userInfo,
		]);
    }

}
