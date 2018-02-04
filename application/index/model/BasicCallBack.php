<?php
namespace app\index\model;

use think\Model;
use Sabre\DAV;
use Sabre\DAV\Auth\Backend;

class BasicCallBack extends \Sabre\DAV\Auth\Backend\AbstractDigest {

	/**
	 * Callback
	 *
	 * @var callable
	 */
	protected $callBack;

	/**
	 * Creates the backend.
	 *
	 * A callback must be provided to handle checking the username and
	 * password.
	 *
	 * @param callable $callBack
	 * @return void
	 */
	function __construct(callable $callBack) {

	    $this->callBack = $callBack;
	    $this->realm = 'CloudreveWebDav';

	}


	public function getDigestHash($realm, $username){
		$cb = $this->callBack;
		return $cb($realm,$username);
	}

}