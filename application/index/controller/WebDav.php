<?php
namespace app\index\controller;

use think\Controller;
use think\Db;

use \app\index\model\Option;
use \app\index\model\User;
use \app\index\model\Directory;
use \app\index\model\Objects;
use \app\index\model\DavAuth;
use \app\index\model\BasicCallBack;

use Sabre\DAV;
use Sabre\DAV\Auth;

class WebDav extends Controller{

	public $userObj;
	public $uid;

	public function index(){

	}

	public function Api(){
		$this->uid = input("param.uid");
		$publicDir = new Directory($this->uid."/");
		
		$server = new DAV\Server($publicDir);
		$server->setBaseUri('/WebDav/Api/uid/'.$this->uid."/");
		$lockBackend = new DAV\Locks\Backend\File(ROOT_PATH.'public/locks');
		$lockPlugin = new DAV\Locks\Plugin($lockBackend);
		$server->addPlugin($lockPlugin);
		$check = new DavAuth($this->uid);
		$callBack = new BasicCallBack($check);
		$authPlugin = new Auth\Plugin($callBack);
		$server->addPlugin($authPlugin);
		$server->addPlugin(new DAV\Browser\Plugin());
		$server->addPlugin(new \Sabre\DAV\Browser\GuessContentType());
		ob_end_clean(); 
		$server->exec();
	}

}
