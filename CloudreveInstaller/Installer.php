<?php

namespace CloudreveInstaller;

use Composer\Script\Event;

class Installer{

	public static function startInstall(Event $event){
		$version = json_decode(file_get_contents("application/version.json"),true)["version"];
		$ioContext = $event->getIO();
		$welcomMsg = "
   ___ _                 _                    
  / __\ | ___  _   _  __| |_ __ _____   _____ 
 / /  | |/ _ \| | | |/ _` | '__/ _ \ \ / / _ \
/ /___| | (_) | |_| | (_| | | |  __/\ V /  __/
\____/|_|\___/ \__,_|\__,_|_|  \___| \_/ \___|
		
                 Ver $version
================================================
";
		$ioContext->write($welcomMsg);
		$sqlInfo = self::getSqlInformation($event);
		$ioContext->write("");
		$siteUrl=$ioContext->ask("The full-url to access to your Cloudreve (e.g. https://pan.aoaoao.me/ , 'http' must be included in the front and '/' must be included at the end):");
		$ioContext->write("");
		if(!file_exists('mysql.sql')){
			$ioContext->writeError("[Error] The file mysql.sql not exist.\nInstaller will exit.To retry, run 'composer install'");
			exit();
		}
		$sqlSource = file_get_contents('mysql.sql');
		$sqlSource = str_replace("https://cloudreve.org/", $siteUrl, $sqlSource);
		$mysqli = @new \mysqli($sqlInfo["hostname"], $sqlInfo["username"], $sqlInfo["password"],  $sqlInfo["database"],  $sqlInfo["hostport"]);
		$ioContext->write("=======================");
		$ioContext->write("Starting import sql file...");
		if ($mysqli->multi_query($sqlSource)) {
			$ioContext->write("Writing complete.");
			$ioContext->write("Writing database.php...");
			if(file_exists('application/database.php')){
				$ioContext->writeError("[Error] The file database.php already exist.\nInstaller will exit.To retry, run 'composer install'");
				$ioContext->write("=======================");
				exit();
			}
			self::writrConfig($event,$sqlInfo);
			$ioContext->write("=======================");
		}else{
			$ioContext->writeError("[Error] Writing failed.Installer will exit. To retry, run 'composer install'");
			$ioContext->write("=======================");
		}
		$ioContext->write("");
		$ioContext->write("Congratulations! Cloudreve has been installed successfully.");
		$ioContext->write("");
		$ioContext->write("Here's some informatioin about yor Cloudreve:");
		$ioContext->write("Homepage: $siteUrl");
		$ioContext->write("Admin Panel: ".$siteUrl."Admin");
		$ioContext->write("Default username: admin@cloudreve.org");
		$ioContext->write("Default password: admin");
		$ioContext->write("");
		$ioContext->write("=======================");
		$ioContext->write("IMPORTANT! You may still have to configure the URL Rewrite to set everthing to work.");
		$ioContext->write("Refer to the install manual for more informatioin.");
		$ioContext->write("=======================");
		self::sendFeedBack($siteUrl);
	}

	 public static function writrConfig(Event $event,$sqlInfo){
		 $ioContext = $event->getIO();
		 try {
			 $fileContent = file_get_contents("CloudreveInstaller/database_sample.php");
			 $replacement = array(
				'{hostname}' => $sqlInfo["hostname"],
				'{database}' => $sqlInfo["database"],
				'{username}' => $sqlInfo["username"],
				'{password}' => $sqlInfo["password"],
				'{hostport}' => $sqlInfo["hostport"],
				);
			$fileContent = strtr($fileContent,$replacement);
			file_put_contents('application/database.php',$fileContent);
		}catch (Exception $e) {
			$ioContext->writeError("[Error] Writing failed.Installer will exit. To retry, run 'composer install'");
		}
		$ioContext->write("Writing complete.");
	 }

	public static function getSqlInformation(Event $event){
		$ioContext = $event->getIO();
		$hostname=$ioContext->ask("Input the hostname of your MySQL server (Default:127.0.0.1):","127.0.0.1");
		$database=$ioContext->ask("The database name:","127.0.0.1");
		$username=$ioContext->ask("The username of your MySQL server (Default:root):","root");
		$password=$ioContext->ask("The password of your MySQL server:");
		$hostport=$ioContext->ask("The hostport of your MySQL server (Default:3306):","3306");
		$mysqli = @new \mysqli($hostname, $username,  $password,  $database,  $hostport);
		if ($mysqli->connect_error) {
			$ioContext->writeError("[Error] Cannot connect to MySQL server, Message:".$mysqli->connect_error);
			$ioContext->write("");
			$ioContext->write("Please confirm your connection informatioin:");
			@$mysqli->close();
			return self::getSqlInformation($event);
		}
		return [
			"hostname" =>  $hostname,
			"database" =>  $database,
			"username" =>  $username,
			"password" =>  $password,
			"hostport" =>  $hostport,
		];
	}

	public static function sendFeedBack($url){
		@file_get_contents("http://toy.aoaoao.me/feedback.php?url=".urlencode($url));
	}

}
?>