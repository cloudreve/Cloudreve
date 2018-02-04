<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use think\Request;
use \app\index\model\CronHandler;
use think\Session;

class Cron extends Controller{

	public function index(){
		$cornObj = new CronHandler();
		$cornObj->Doit();
	}

}
