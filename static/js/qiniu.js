/*!
 * qiniu-js-sdk v@VERSION
 *
 * Copyright 2015 by Qiniu
 * Released under GPL V2 License.
 *
 * GitHub: http://github.com/qiniu/js-sdk
 *
 * Date: @DATE
*/
/*!
 * 
 * Rebuild By Aaron@2018
 * 
*/
/*global plupload ,mOxie*/
/*global ActiveXObject */
/*exported Qiniu */
/*exported QiniuJsSDK */

function getCookieByString(cookieName){
 var start = document.cookie.indexOf(cookieName+'=');
 if (start == -1) return false;
 start = start+cookieName.length+1;
 var end = document.cookie.indexOf(';', start);
 if (end == -1) end=document.cookie.length;
 return document.cookie.substring(start, end);
}
;(function( global ){

/**
 * Creates new cookie or removes cookie with negative expiration
 * @param  key       The key or identifier for the store
 * @param  value     Contents of the store
 * @param  exp       Expiration - creation defaults to 30 days
 */
function createCookie(key, value, exp) {
	var date = new Date();
	date.setTime(date.getTime() + (exp * 24 * 60 * 60 * 1000));
	var expires = "; expires=" + date.toGMTString();
	document.cookie = key + "=" + value + expires + "; path=/";
}

/**
 * Returns contents of cookie
 * @param  key       The key or identifier for the store
 */
function readCookie(key) {
	var nameEQ = key + "=";
	var ca = document.cookie.split(';');
	for (var i = 0, max = ca.length; i < max; i++) {
		var c = ca[i];
		while (c.charAt(0) === ' ') {
			c = c.substring(1, c.length);
		}
		if (c.indexOf(nameEQ) === 0) {
			return c.substring(nameEQ.length, c.length);
		}
	}
	return null;
}

// if current browser is not support localStorage
// use cookie to make a polyfill
if ( !window.localStorage ) {
	window.localStorage = {
		setItem: function (key, value) {
			createCookie(key, value, 30);
		},
		getItem: function (key) {
			return readCookie(key);
		},
		removeItem: function (key) {
			createCookie(key, '', -1);
		}
	};
}

function QiniuJsSDK() {

	var that = this;

	/**
	 * detect IE version
	 * if current browser is not IE
	 *     it will return false
	 * else
	 *     it will return version of current IE browser
	 * @return {Number|Boolean} IE version or false
	 */
	this.detectIEVersion = function() {
		var v = 4,
			div = document.createElement('div'),
			all = div.getElementsByTagName('i');
		while (
			div.innerHTML = '<!--[if gt IE ' + v + ']><i></i><![endif]-->',
			all[0]
		) {
			v++;
		}
		return v > 4 ? v : false;
	};

	var logger = {
		MUTE: 0,
		FATA: 1,
		ERROR: 2,
		WARN: 3,
		INFO: 4,
		DEBUG: 5,
		TRACE: 6,
		level: 0
	};

	function log(type, args){
		var header = "[Cloudreve-uploader]["+type+"]";
		var msg = header;
		for (var i = 0; i < args.length; i++) {
			if (typeof args[i] === "string") {
				msg += " " + args[i];
			} else {
				msg += " " + that.stringifyJSON(args[i]);
			}
		}
		if (that.detectIEVersion()) {
			// http://stackoverflow.com/questions/5538972/console-log-apply-not-working-in-ie9
			//var log = Function.prototype.bind.call(console.log, console);
			//log.apply(console, args);
			console.log(msg);
		}else{
			args.unshift(header);
			console.log.apply(console, args);
		}
		if (document.getElementById('qiniu-js-sdk-log')) {
			document.getElementById('qiniu-js-sdk-log').innerHTML += '<p>'+msg+'</p>';
		}
	}

	function makeLogFunc(code){
		var func = code.toLowerCase();
		logger[func] = function(){
			// logger[func].history = logger[func].history || [];
			// logger[func].history.push(arguments);
			if(window.console && window.console.log && logger.level>=logger[code]){
				var args = Array.prototype.slice.call(arguments);
				log(func,args);
			}
		};
	}

	for (var property in logger){
		if (logger.hasOwnProperty(property) && (typeof logger[property]) === "number" && !logger.hasOwnProperty(property.toLowerCase())) {
			makeLogFunc(property);
		}
	}


	var qiniuUploadUrl;
  

	/**
	 * qiniu upload urls
	 * 'qiniuUploadUrls' is used to change target when current url is not avaliable
	 * @type {Array}
	 */
	if(uploadConfig.saveType == "qiniu"){
		  if (window.location.protocol === 'https:') {
		qiniuUploadUrl = 'https://up.qbox.me';
	} else {
		qiniuUploadUrl = 'http://upload.qiniu.com';
	}
	var qiniuUploadUrls = [
		"http://upload.qiniu.com",
		"http://up.qiniu.com"
	];

	var qiniuUpHosts = {
	   "http": [
		   "http://upload.qiniu.com",
		   "http://up.qiniu.com"
	   ],
	   "https": [
		   "https://up.qbox.me"
	   ]
	};
}else if(uploadConfig.saveType == "local" || uploadConfig.saveType == "oss" ||uploadConfig.saveType == "upyun"||uploadConfig.saveType == "s3"|| uploadConfig.saveType == "remote"||uploadConfig.saveType == "onedrive"){
 qiniuUploadUrl = uploadConfig.upUrl;
 var qiniuUploadUrls = [uploadConfig.upUrl,];
 var qiniuUpHosts = {
	"http": [uploadConfig.upUrl,],
	"https": [uploadConfig.upUrl,]
 }
}

	var changeUrlTimes = 0;

	/**
	 * reset upload url
	 * if current page protocal is https
	 *     it will always return 'https://up.qbox.me'
	 * else
	 *     it will set 'qiniuUploadUrl' value with 'qiniuUploadUrls' looply
	 */
	this.resetUploadUrl = function(){
	var hosts = window.location.protocol === 'https:' ? qiniuUpHosts.https : qiniuUpHosts.http;
	var i = changeUrlTimes % hosts.length;
	qiniuUploadUrl = hosts[i];
	changeUrlTimes++;
	logger.debug('resetUploadUrl: '+qiniuUploadUrl);
	};

	// this.resetUploadUrl();


	/**
	 * is image
	 * @param  {String}  url of a file
	 * @return {Boolean} file is a image or not
	 */
	this.isImage = function(url) {
		url = url.split(/[?#]/)[0];
		return (/\.(png|jpg|jpeg|gif|bmp)$/i).test(url);
	};

	/**
	 * get file extension
	 * @param  {String} filename
	 * @return {String} file extension
	 * @example
	 *     input: test.txt
	 *     output: txt
	 */
	this.getFileExtension = function(filename) {
		var tempArr = filename.split(".");
		var ext;
		if (tempArr.length === 1 || (tempArr[0] === "" && tempArr.length === 2)) {
			ext = "";
		} else {
			ext = tempArr.pop().toLowerCase(); //get the extension and make it lower-case
		}
		return ext;
	};

	/**
	 * encode string by utf8
	 * @param  {String} string to encode
	 * @return {String} encoded string
	 */
	this.utf8_encode = function(argString) {
		// http://kevin.vanzonneveld.net
		// +   original by: Webtoolkit.info (http://www.webtoolkit.info/)
		// +   improved by: Kevin van Zonneveld (http://kevin.vanzonneveld.net)
		// +   improved by: sowberry
		// +    tweaked by: Jack
		// +   bugfixed by: Onno Marsman
		// +   improved by: Yves Sucaet
		// +   bugfixed by: Onno Marsman
		// +   bugfixed by: Ulrich
		// +   bugfixed by: Rafal Kukawski
		// +   improved by: kirilloid
		// +   bugfixed by: kirilloid
		// *     example 1: this.utf8_encode('Kevin van Zonneveld');
		// *     returns 1: 'Kevin van Zonneveld'

		if (argString === null || typeof argString === 'undefined') {
			return '';
		}

		var string = (argString + ''); // .replace(/\r\n/g, '\n').replace(/\r/g, '\n');
		var utftext = '',
			start, end, stringl = 0;

		start = end = 0;
		stringl = string.length;
		for (var n = 0; n < stringl; n++) {
			var c1 = string.charCodeAt(n);
			var enc = null;

			if (c1 < 128) {
				end++;
			} else if (c1 > 127 && c1 < 2048) {
				enc = String.fromCharCode(
					(c1 >> 6) | 192, (c1 & 63) | 128
				);
			} else if (c1 & 0xF800 ^ 0xD800 > 0) {
				enc = String.fromCharCode(
					(c1 >> 12) | 224, ((c1 >> 6) & 63) | 128, (c1 & 63) | 128
				);
			} else { // surrogate pairs
				if (c1 & 0xFC00 ^ 0xD800 > 0) {
					throw new RangeError('Unmatched trail surrogate at ' + n);
				}
				var c2 = string.charCodeAt(++n);
				if (c2 & 0xFC00 ^ 0xDC00 > 0) {
					throw new RangeError('Unmatched lead surrogate at ' + (n - 1));
				}
				c1 = ((c1 & 0x3FF) << 10) + (c2 & 0x3FF) + 0x10000;
				enc = String.fromCharCode(
					(c1 >> 18) | 240, ((c1 >> 12) & 63) | 128, ((c1 >> 6) & 63) | 128, (c1 & 63) | 128
				);
			}
			if (enc !== null) {
				if (end > start) {
					utftext += string.slice(start, end);
				}
				utftext += enc;
				start = end = n + 1;
			}
		}

		if (end > start) {
			utftext += string.slice(start, stringl);
		}

		return utftext;
	};

	this.base64_decode = function (data) {
		// http://kevin.vanzonneveld.net
		// +   original by: Tyler Akins (http://rumkin.com)
		// +   improved by: Thunder.m
		// +      input by: Aman Gupta
		// +   improved by: Kevin van Zonneveld (http://kevin.vanzonneveld.net)
		// +   bugfixed by: Onno Marsman
		// +   bugfixed by: Pellentesque Malesuada
		// +   improved by: Kevin van Zonneveld (http://kevin.vanzonneveld.net)
		// +      input by: Brett Zamir (http://brett-zamir.me)
		// +   bugfixed by: Kevin van Zonneveld (http://kevin.vanzonneveld.net)
		// *     example 1: base64_decode('S2V2aW4gdmFuIFpvbm5ldmVsZA==');
		// *     returns 1: 'Kevin van Zonneveld'
		// mozilla has this native
		// - but breaks in 2.0.0.12!
		//if (typeof this.window['atob'] == 'function') {
		//    return atob(data);
		//}
		var b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";
		var o1, o2, o3, h1, h2, h3, h4, bits, i = 0,
		ac = 0,
		dec = "",
		tmp_arr = [];

		if (!data) {
			return data;
		}

		data += '';

		do { // unpack four hexets into three octets using index points in b64
			h1 = b64.indexOf(data.charAt(i++));
			h2 = b64.indexOf(data.charAt(i++));
			h3 = b64.indexOf(data.charAt(i++));
			h4 = b64.indexOf(data.charAt(i++));

			bits = h1 << 18 | h2 << 12 | h3 << 6 | h4;

			o1 = bits >> 16 & 0xff;
			o2 = bits >> 8 & 0xff;
			o3 = bits & 0xff;

			if (h3 === 64) {
				tmp_arr[ac++] = String.fromCharCode(o1);
			} else if (h4 === 64) {
				tmp_arr[ac++] = String.fromCharCode(o1, o2);
			} else {
				tmp_arr[ac++] = String.fromCharCode(o1, o2, o3);
			}
		} while (i < data.length);

		dec = tmp_arr.join('');

		return dec;
	};

	/**
	 * encode data by base64
	 * @param  {String} data to encode
	 * @return {String} encoded data
	 */
	this.base64_encode = function(data) {
		// http://kevin.vanzonneveld.net
		// +   original by: Tyler Akins (http://rumkin.com)
		// +   improved by: Bayron Guevara
		// +   improved by: Thunder.m
		// +   improved by: Kevin van Zonneveld (http://kevin.vanzonneveld.net)
		// +   bugfixed by: Pellentesque Malesuada
		// +   improved by: Kevin van Zonneveld (http://kevin.vanzonneveld.net)
		// -    depends on: this.utf8_encode
		// *     example 1: this.base64_encode('Kevin van Zonneveld');
		// *     returns 1: 'S2V2aW4gdmFuIFpvbm5ldmVsZA=='
		// mozilla has this native
		// - but breaks in 2.0.0.12!
		//if (typeof this.window['atob'] == 'function') {
		//    return atob(data);
		//}
		var b64 = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
		var o1, o2, o3, h1, h2, h3, h4, bits, i = 0,
			ac = 0,
			enc = '',
			tmp_arr = [];

		if (!data) {
			return data;
		}

		data = this.utf8_encode(data + '');

		do { // pack three octets into four hexets
			o1 = data.charCodeAt(i++);
			o2 = data.charCodeAt(i++);
			o3 = data.charCodeAt(i++);

			bits = o1 << 16 | o2 << 8 | o3;

			h1 = bits >> 18 & 0x3f;
			h2 = bits >> 12 & 0x3f;
			h3 = bits >> 6 & 0x3f;
			h4 = bits & 0x3f;

			// use hexets to index into b64, and append result to encoded string
			tmp_arr[ac++] = b64.charAt(h1) + b64.charAt(h2) + b64.charAt(h3) + b64.charAt(h4);
		} while (i < data.length);

		enc = tmp_arr.join('');

		switch (data.length % 3) {
			case 1:
				enc = enc.slice(0, -2) + '==';
				break;
			case 2:
				enc = enc.slice(0, -1) + '=';
				break;
		}

		return enc;
	};

	/**
	 * encode string in url by base64
	 * @param {String} string in url
	 * @return {String} encoded string
	 */
	this.URLSafeBase64Encode = function(v) {
		v = this.base64_encode(v);
		return v.replace(/\//g, '_').replace(/\+/g, '-');
	};

	this.URLSafeBase64Decode = function(v) {
		v = v.replace(/_/g, '/').replace(/-/g, '+');
		return this.base64_decode(v);
	};

	// TODO: use mOxie
	/**
	 * craete object used to AJAX
	 * @return {Object}
	 */
	this.createAjax = function(argument) {
		var xmlhttp = {};
		if (window.XMLHttpRequest) {
			xmlhttp = new XMLHttpRequest();
		} else {
			xmlhttp = new ActiveXObject("Microsoft.XMLHTTP");
		}
		return xmlhttp;
	};

	// TODO: enhance IE compatibility
	/**
	 * parse json string to javascript object
	 * @param  {String} json string
	 * @return {Object} object
	 */
	this.parseJSON = function(data) {
		// Attempt to parse using the native JSON parser first
		if (window.JSON && window.JSON.parse) {
			return window.JSON.parse(data);
		}

		//var rx_one = /^[\],:{}\s]*$/,
		//    rx_two = /\\(?:["\\\/bfnrt]|u[0-9a-fA-F]{4})/g,
		//    rx_three = /"[^"\\\n\r]*"|true|false|null|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?/g,
		//    rx_four = /(?:^|:|,)(?:\s*\[)+/g,
		var    rx_dangerous = /[\u0000\u00ad\u0600-\u0604\u070f\u17b4\u17b5\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufeff\ufff0-\uffff]/g;

		//var json;

		var text = String(data);
		rx_dangerous.lastIndex = 0;
		if(rx_dangerous.test(text)){
			text = text.replace(rx_dangerous, function(a){
			   return '\\u' + ('0000' + a.charCodeAt(0).toString(16)).slice(-4);
			});
		}

		// todo 使用一下判断,增加安全性
		//if (
		//    rx_one.test(
		//        text
		//            .replace(rx_two, '@')
		//            .replace(rx_three, ']')
		//            .replace(rx_four, '')
		//    )
		//) {
		//    return eval('(' + text + ')');
		//}

		return eval('('+text+')');
	};

	/**
	 * parse javascript object to json string
	 * @param  {Object} object
	 * @return {String} json string
	 */
	this.stringifyJSON = function(obj) {
		// Attempt to parse using the native JSON parser first
		if (window.JSON && window.JSON.stringify) {
			return window.JSON.stringify(obj);
		}
		switch (typeof (obj)) {
			case 'string':
				return '"' + obj.replace(/(["\\])/g, '\\$1') + '"';
			case 'array':
				return '[' + obj.map(that.stringifyJSON).join(',') + ']';
			case 'object':
				if (obj instanceof Array) {
					var strArr = [];
					var len = obj.length;
					for (var i = 0; i < len; i++) {
						strArr.push(that.stringifyJSON(obj[i]));
					}
					return '[' + strArr.join(',') + ']';
				} else if (obj === null) {
					return 'null';
				} else {
					var string = [];
					for (var property in obj) {
						if (obj.hasOwnProperty(property)) {
							string.push(that.stringifyJSON(property) + ':' + that.stringifyJSON(obj[property]));
						}
					}
					return '{' + string.join(',') + '}';
				}
				break;
			case 'number':
				return obj;
			case false:
				return obj;
			case 'boolean':
				return obj;
		}
	};

	/**
	 * trim space beside text
	 * @param  {String} untrimed string
	 * @return {String} trimed string
	 */
	this.trim = function(text) {
		return text === null ? "" : text.replace(/^\s+|\s+$/g, '');
	};

	/**
	 * create a uploader by QiniuJsSDK
	 * @param  {object} options to create a new uploader
	 * @return {object} uploader
	 */
	this.uploader = function(op) {

		/********** inner function define start **********/

		// according the different condition to reset chunk size
		// and the upload strategy according with the chunk size
		// when chunk size is zero will cause to direct upload
		// see the statement binded on 'BeforeUpload' event
		var reset_chunk_size = function() {
			var ie = that.detectIEVersion();
			var BLOCK_BITS, MAX_CHUNK_SIZE, chunk_size;
			// case Safari 5、Windows 7、iOS 7 set isSpecialSafari to true
			var isSpecialSafari = (moxie.core.utils.Env.browser  === "Safari" && moxie.core.utils.Env.version  <= 5 && moxie.core.utils.Env.os === "Windows" && moxie.core.utils.Env.osVersion === "7") || (moxie.core.utils.Env.browser === "Safari" && moxie.core.utils.Env.os === "iOS" && moxie.core.utils.Env.osVersion === "7");
			// case IE 9-，chunk_size is not empty and flash is included in runtimes
			// set op.chunk_size to zero
			//if (ie && ie < 9 && op.chunk_size && op.runtimes.indexOf('flash') >= 0) {
			if (ie && ie < 9 && op.chunk_size && op.runtimes.indexOf('flash') >= 0) {
				//  link: http://www.plupload.com/docs/Frequently-Asked-Questions#when-to-use-chunking-and-when-not
				//  when plupload chunk_size setting is't null ,it cause bug in ie8/9  which runs  flash runtimes (not support html5) .
				op.chunk_size = 0;
			} else if (isSpecialSafari) {
				// win7 safari / iOS7 safari have bug when in chunk upload mode
				// reset chunk_size to 0
				// disable chunk in special version safari
				op.chunk_size = 0;
			} else {
				BLOCK_BITS = 20;
				MAX_CHUNK_SIZE = 4 << BLOCK_BITS; //4M

				chunk_size = plupload.parseSize(op.chunk_size);
				if (chunk_size > MAX_CHUNK_SIZE) {
					op.chunk_size = MAX_CHUNK_SIZE;
				}
				// qiniu service  max_chunk_size is 4m
				// reset chunk_size to max_chunk_size(4m) when chunk_size > 4m
			}
			// if op.chunk_size set 0 will be cause to direct upload
		};

		var getHosts = function(hosts) {
			var result = [];
			for (var i = 0; i < hosts.length; i++) {
				var host = hosts[i];
				if (host.indexOf('-H') === 0) {
					result.push(host.split(' ')[2]);
				} else {
					result.push(host);
				}
			}
			return result;
		};

		var getPutPolicy = function (uptoken) {
			if(uploadConfig.saveType =="oss" || uploadConfig.saveType =="upyun"||uploadConfig.saveType =="s3"||uploadConfig.saveType =="remote"){
				return "oss";
			}else{
				var segments = uptoken.split(":");
				var ak = segments[0];
				var putPolicy = that.parseJSON(that.URLSafeBase64Decode(segments[2]));
				putPolicy.ak = ak;
				if (putPolicy.scope.indexOf(":") >= 0) {
					putPolicy.bucket = putPolicy.scope.split(":")[0];
					putPolicy.key = putPolicy.scope.split(":")[1];
				} else {
					putPolicy.bucket = putPolicy.scope;
				}
				return putPolicy;
			}
		};

		var getUpHosts = function(uptoken) {
			var putPolicy = getPutPolicy(uptoken);
			// var uphosts_url = "//uc.qbox.me/v1/query?ak="+ak+"&bucket="+putPolicy.scope;
			// IE 9- is not support protocal relative url
			var uphosts_url = window.location.protocol + "//uc.qbox.me/v1/query?ak=" + putPolicy.ak + "&bucket=" + putPolicy.bucket;
			logger.debug("putPolicy: ", putPolicy);
			logger.debug("get uphosts from: ", uphosts_url);
			var ie = that.detectIEVersion();
			var ajax;
			if (ie && ie <= 9) {
				ajax = new moxie.XMLHttpRequest();
			   moxie.core.utils.Env.swf_url = op.flash_swf_url;
			}else{
				ajax = that.createAjax();
			}
			if(uploadConfig.saveType != "qiniu"){
				qiniuUpHosts.http = [uploadConfig.upUrl];
				qiniuUpHosts.http = [uploadConfig.upUrl];
				that.resetUploadUrl();
			}else{
				ajax.open('GET', uphosts_url, false);
				var onreadystatechange = function(){
					logger.debug("ajax.readyState: ", ajax.readyState);
					if (ajax.readyState === 4) {
						logger.debug("ajax.status: ", ajax.status);
						if (ajax.status < 400) {
							var res = that.parseJSON(ajax.responseText);
							qiniuUpHosts.http = getHosts(res.http.up);
							qiniuUpHosts.https = getHosts(res.https.up);
							logger.debug("get new uphosts: ", qiniuUpHosts);
							that.resetUploadUrl();
						} else {
							logger.error("get uphosts error: ", ajax.responseText);
						}
					}
				};
				if (ie && ie <= 9) {
					ajax.bind('readystatechange', onreadystatechange);
				}else{
					ajax.onreadystatechange = onreadystatechange;
				}
				ajax.send();
				// ajax.send();
				// if (ajax.status < 400) {
				//     var res = that.parseJSON(ajax.responseText);
				//     qiniuUpHosts.http = getHosts(res.http.up);
				//     qiniuUpHosts.https = getHosts(res.https.up);
				//     logger.debug("get new uphosts: ", qiniuUpHosts);
				//     that.resetUploadUrl();
				// } else {
				//     logger.error("get uphosts error: ", ajax.responseText);
				// }
				}
			return;
		};

		var getUptoken = function(file) {
			if(uploadConfig.saveType == "remote"){
				return that.token;
			}
			if (!that.token || (op.uptoken_url && that.tokenInfo.isExpired())) {
				return getNewUpToken(file);
			} else {
				return that.token;
			}
		};

		// getNewUptoken maybe called at Init Event or BeforeUpload Event
		// case Init Event, the file param of getUptken will be set a null value
		// if op.uptoken has value, set uptoken with op.uptoken
		// else if op.uptoken_url has value, set uptoken from op.uptoken_url
		// else if op.uptoken_func has value, set uptoken by result of op.uptoken_func
		var getNewUpToken = function(file) {
			if (op.uptoken) {
				that.token = op.uptoken;
			} else if (op.uptoken_url) {
				logger.debug("get uptoken from: ", that.uptoken_url);
				// TODO: use mOxie
				var ajax = that.createAjax();
				ajax.open('GET', that.uptoken_url, false);
				ajax.setRequestHeader("If-Modified-Since", "0");
				// ajax.onreadystatechange = function() {
				//     if (ajax.readyState === 4 && ajax.status === 200) {
				//         var res = that.parseJSON(ajax.responseText);
				//         that.token = res.uptoken;
				//     }
				// };
				ajax.send();
				if (ajax.status === 200) {
					var res = that.parseJSON(ajax.responseText);
					that.token = res.uptoken;
					if (uploadConfig.saveType == "oss"){
						var putPolicy = that.token;
						that.sign = res.sign;
						that.access = res.id;
						that.file_name = res.key;
						that.callback = res.callback;
					}else if(uploadConfig.saveType == "s3"){
						var putPolicy = that.token;
						that.sign = res.sign;
						that.policy = res.policy;
						that.file_name = res.key;
						that.credential = res.credential;
						that.x_amz_date = res.x_amz_date;
						that.surl = res.siteUrl;
						that.callbackKey = res.callBackKey;
					}else if(uploadConfig.saveType == "upyun"){
						var putPolicy = that.token;
						that.token = res.token;
						that.policy = res.policy;
					}else if(uploadConfig.saveType == "remote"){
						var putPolicy = that.token;
						that.policy = res.uptoken;
					}else{
						var segments = that.token.split(":");
						var putPolicy = that.parseJSON(that.URLSafeBase64Decode(segments[2]));
						if (!that.tokenMap) {
							that.tokenMap = {};
						}
						var getTimestamp = function(time) {
							return Math.ceil(time.getTime()/1000);
						};
						var serverTime = getTimestamp(new Date(ajax.getResponseHeader("date")));
						var clientTime = getTimestamp(new Date());
						that.tokenInfo = {
							serverDelay: clientTime - serverTime,
							deadline: putPolicy.deadline,
							isExpired: function() {
								var leftTime = this.deadline - getTimestamp(new Date()) + this.serverDelay;
								return leftTime < 600;
							}
						}; 
						logger.debug("get token info: ", that.tokenInfo);
					}
					
					logger.debug("get new uptoken: ", that.token);
				   
				} else {
					logger.error("get uptoken error: ", ajax.responseText);
				}
			} else if (op.uptoken_func) {
				logger.debug("get uptoken from uptoken_func");
				that.token = op.uptoken_func(file);
				logger.debug("get new uptoken: ", that.token);
			} else {
				logger.error("one of [uptoken, uptoken_url, uptoken_func] settings in options is required!");
			}
			if (that.token) {
				getUpHosts(that.token);
			}
			return that.token;
		};

		// get file key according with the user passed options
		var getFileKey = function(up, file, func) {
			// WARNING
			// When you set the key in putPolicy by "scope": "bucket:key"
			// You should understand the risk of override a file in the bucket
			// So the code below that automatically get key from uptoken has been commented
			// var putPolicy = getPutPolicy(that.token)
			// if (putPolicy.key) {
			//     logger.debug("key is defined in putPolicy.scope: ", putPolicy.key)
			//     return putPolicy.key
			// }
			var key = '',
				unique_names = false;
			if (!op.save_key) {
				unique_names = up.getOption && up.getOption('unique_names');
				unique_names = unique_names || (up.settings && up.settings.unique_names);
				if (unique_names) {
					var ext = that.getFileExtension(file.name);
					key = ext ? file.id + '.' + ext : file.id;
				} else if (typeof func === 'function') {
					key = func(up, file);
				} else {
					key = file.name;
				}
			}
			if(uploadConfig.saveType == "qiniu"){
				return "";
			}
			return key;
		};

		/********** inner function define end **********/

		if (op.log_level) {
			logger.level = op.log_level;
		}

		if (!op.domain) {
			throw 'domain setting in options is required!';
		}

		if (!op.browse_button) {
			throw 'browse_button setting in options is required!';
		}

		if (!op.uptoken && !op.uptoken_url && !op.uptoken_func) {
			throw 'one of [uptoken, uptoken_url, uptoken_func] settings in options is required!';
		}

		logger.debug("init uploader start");

		logger.debug("environment: ", moxie.core.utils.Env);

		logger.debug("userAgent: ", navigator.userAgent);

		var option = {};

		// hold the handler from user passed options
		var _Error_Handler = op.init && op.init.Error;
		var _FileUploaded_Handler = op.init && op.init.FileUploaded;

		// replace the handler for intercept
		op.init.Error = function() {};
		op.init.FileUploaded = function() {};

		that.uptoken_url = op.uptoken_url;
		that.token = '';
		that.key_handler = typeof op.init.Key === 'function' ? op.init.Key : '';
		this.domain = op.domain;
		// TODO: ctx is global in scope of a uploader instance
		// this maybe cause error
		var ctx = '';
		var speedCalInfo = {
			isResumeUpload: false,
			resumeFilesize: 0,
			startTime: '',
			currentTime: ''
		};

		reset_chunk_size();
		logger.debug("invoke reset_chunk_size()");
		logger.debug("op.chunk_size: ", op.chunk_size);

		var defaultSetting = {
			url: qiniuUploadUrl,
			multipart_params: {
				token: ''
			}
		};
		var ie = that.detectIEVersion();
		// case IE 9-
		// add accept in multipart params
		if (ie && ie <= 9) {
			defaultSetting.multipart_params.accept = 'text/plain; charset=utf-8';
			logger.debug("add accept text/plain in multipart params");
		}

		// compose options with user passed options and default setting
		plupload.extend(option, op, defaultSetting);

		logger.debug("option: ", option);

		// create a new uploader with composed options
		var uploader = new plupload.Uploader(option);

		logger.debug("new plupload.Uploader(option)");

		// bind getNewUpToken to 'Init' event
		uploader.bind('Init', function(up, params) {
			logger.debug("Init event activated");
			// if op.get_new_uptoken is not true
			//      invoke getNewUptoken when uploader init
			// else
			//      getNewUptoken everytime before a new file upload
			if(!op.get_new_uptoken){
				getNewUpToken(null);
			}
			//getNewUpToken(null);
		});

		logger.debug("bind Init event");

		// bind 'FilesAdded' event
		// when file be added and auto_start has set value
		// uploader will auto start upload the file
		uploader.bind('FilesAdded', function(up, files) {
			logger.debug("FilesAdded event activated");
			var auto_start = up.getOption && up.getOption('auto_start');
			auto_start = auto_start || (up.settings && up.settings.auto_start);
			logger.debug("auto_start: ", auto_start);
			logger.debug("files: ", files);

			
			 for (var i = 0; i < files.length; i++) {
					var file = files[i];
					file.path = decodeURIComponent(getCookieByString("path_tmp"));
				   
		 }
			
			// detect is iOS
			var is_ios = function (){
				if(moxie.core.utils.Env.OS.toLowerCase()==="ios") {
					return true;
				} else {
					return false;
				}
			};

			// if current env os is iOS change file name to [time].[ext]
			if (is_ios()) {
				for (var i = 0; i < files.length; i++) {
					var file = files[i];
					var ext = that.getFileExtension(file.name);
					file.name = file.id + "." + ext;
				}
			}

			if (auto_start) {
				setTimeout(function(){
					up.start();
					logger.debug("invoke up.start()");
				}, 0);
				// up.start();
				// plupload.each(files, function(i, file) {
				//     up.start();
				//     logger.debug("invoke up.start()")
				//     logger.debug("file: ", file);
				// });
			}
			up.refresh(); // Reposition Flash/Silverlight
		});

		logger.debug("bind FilesAdded event");

		// bind 'BeforeUpload' event
		// intercept the process of upload
		// - prepare uptoken
		// - according the chunk size to make differnt upload strategy
		// - resume upload with the last breakpoint of file
		uploader.bind('BeforeUpload', function(up, file) {
			logger.debug("BeforeUpload event activated");
			getNewUpToken(file);
			// add a key named speed for file object
			file.speed = file.speed || 0;
			ctx = '';

			var directUpload = function(up, file, func) {
				speedCalInfo.startTime = new Date().getTime();
		  
				
		  
				var multipart_params_obj;
				if (op.save_key) {
					multipart_params_obj = {
						'token': that.token
					};
				} else {
					multipart_params_obj = {
						'key': getFileKey(up, file, func),
						'token': that.token
					};
					if(uploadConfig.saveType == "qiniu"){
						multipart_params_obj = {
							'token': that.token,
							'x:path': file.path
						}; 
					}else if(uploadConfig.saveType == "local"||uploadConfig.saveType == "onedrive"){
						multipart_params_obj = {
						'path': file.path
						}; 
					}else if(uploadConfig.saveType == "remote"){
						multipart_params_obj = {
						'path': file.path,
						"token" :that.policy,
						"MAX_FILE_SIZE":4194304,
						}; 
					}else if(uploadConfig.saveType == "oss"){
						multipart_params_obj = {
						'policy': that.token,
						'x:path': file.path,
						"signature":that.sign,
						"OSSAccessKeyId":that.access,
						'x:fname':file.name,
						"key" :that.file_name.replace("${filename}", file.name),
						"callback":that.callback,
						}; 
					}else if(uploadConfig.saveType == "s3"){
						multipart_params_obj = {
							'policy': that.policy,
							"key" :that.file_name+"/"+(file.path.replace(",","/") == "" ? "" : file.path.replace(",","/")+"/")+file.name,
							"success_action_redirect" :that.surl+"Callback/S3/key/"+that.callbackKey,
							"x-amz-algorithm":"AWS4-HMAC-SHA256",
							"x-amz-credential":that.credential,
							"x-amz-date":that.x_amz_date,
							"x-amz-signature":that.sign,
							"Content-Type":file.type,
						}; 
					}else if(uploadConfig.saveType == "upyun"){
						multipart_params_obj = {
						'authorization': that.token,
						'policy': that.policy,
						}; 
					}
				}
				var ie = that.detectIEVersion();
				// case IE 9-
				// add accept in multipart params
				if (ie && ie <= 9) {
					multipart_params_obj.accept = 'text/plain; charset=utf-8';
					logger.debug("add accept text/plain in multipart params");
				}

				logger.debug("directUpload multipart_params_obj: ", multipart_params_obj);

				var x_vars = op.x_vars;
				if (x_vars !== undefined && typeof x_vars === 'object') {
					for (var x_key in x_vars) {
						if (x_vars.hasOwnProperty(x_key)) {
							if (typeof x_vars[x_key] === 'function') {
								multipart_params_obj['x:' + x_key] = x_vars[x_key](up, file);
							} else if (typeof x_vars[x_key] !== 'object') {
								multipart_params_obj['x:' + x_key] = x_vars[x_key];
							}
						}
					}
				}

				up.setOption({
					'url': qiniuUploadUrl,
					'multipart': true,
					'chunk_size': is_android_weixin_or_qq() ? op.max_file_size : undefined,
					'multipart_params': multipart_params_obj
				});
			};

			// detect is weixin or qq inner browser
			var is_android_weixin_or_qq = function (){
				var ua = navigator.userAgent.toLowerCase();
				if((ua.match(/MicroMessenger/i) || moxie.core.utils.Env.browser === "QQBrowser" || ua.match(/V1_AND_SQ/i)) && moxie.core.utils.Env.OS.toLowerCase()==="android") {
					return true;
				} else {
					return false;
				}
			};

			var chunk_size = up.getOption && up.getOption('chunk_size');
			chunk_size = chunk_size || (up.settings && up.settings.chunk_size);

			logger.debug("uploader.runtime: ",uploader.runtime);
			logger.debug("chunk_size: ",chunk_size);

			// TODO: flash support chunk upload
			if ((uploader.runtime === 'html5' || uploader.runtime === 'flash') && chunk_size) {
				if (file.size < chunk_size || is_android_weixin_or_qq()) {
					logger.debug("directUpload because file.size < chunk_size || is_android_weixin_or_qq()");
					// direct upload if file size is less then the chunk size
					directUpload(up, file, that.key_handler);
				} else {
					// TODO: need a polifill to make it work in IE 9-
					// ISSUE: if file.name is existed in localStorage
					// but not the same file maybe cause error
					var localFileInfo = localStorage.getItem(file.name);
					var blockSize = chunk_size;
					if (localFileInfo) {
						// TODO: although only the html5 runtime will enter this statement
						// but need uniform way to make convertion between string and json
						localFileInfo = that.parseJSON(localFileInfo);
						var now = (new Date()).getTime();
						var before = localFileInfo.time || 0;
						var aDay = 24 * 60 * 60 * 1000; //  milliseconds of one day
						// if the last upload time is within one day
						//      will upload continuously follow the last breakpoint
						// else
						//      will reupload entire file
						if (now - before < aDay) {

							if (localFileInfo.percent !== 100) {
								if (file.size === localFileInfo.total) {
									// TODO: if file.name and file.size is the same
									// but not the same file will cause error
									file.percent = localFileInfo.percent;
									file.loaded = localFileInfo.offset;
									ctx = localFileInfo.ctx;

									// set speed info
									speedCalInfo.isResumeUpload = true;
									speedCalInfo.resumeFilesize = localFileInfo.offset;

									// set block size
									if (localFileInfo.offset + blockSize > file.size) {
										blockSize = file.size - localFileInfo.offset;
									}
								} else {
									// remove file info when file.size is conflict with file info
									localStorage.removeItem(file.name);
								}

							} else {
								// remove file info when upload percent is 100%
								// avoid 499 bug
								localStorage.removeItem(file.name);
							}
						} else {
							// remove file info when last upload time is over one day
							localStorage.removeItem(file.name);
						}
					}
					speedCalInfo.startTime = new Date().getTime();
					var multipart_params_obj = {};
					var ie = that.detectIEVersion();
					// case IE 9-
					// add accept in multipart params
					if (ie && ie <= 9) {
						multipart_params_obj.accept = 'text/plain; charset=utf-8';
						logger.debug("add accept text/plain in multipart params");
					}
					// TODO: to support bput
					// http://developer.qiniu.com/docs/v6/api/reference/up/bput.html
					if(uploadConfig.saveType == "remote"){
						up.setOption({
						    'url': qiniuUploadUrl + 'chunk.php',
						    'multipart': false,
						    'chunk_size': chunk_size,
						    'required_features': "chunks",
						    'headers': {
						        'Authorization': getUptoken(file)
						    },
						    'multipart_params': multipart_params_obj
						});
					}else{
						up.setOption({
							'url': qiniuUploadUrl + '/mkblk/' + blockSize,
							'multipart': false,
							'chunk_size': chunk_size,
							'required_features': "chunks",
							'headers': {
								'Authorization': 'UpToken ' + getUptoken(file)
							},
							'multipart_params': multipart_params_obj
						});
					}
				}
			} else {
				logger.debug("directUpload because uploader.runtime !== 'html5' || uploader.runtime !== 'flash' || !chunk_size");
				// direct upload if runtime is not html5
				directUpload(up, file, that.key_handler);
			}
		});

		logger.debug("bind BeforeUpload event");

		// bind 'UploadProgress' event
		// calculate upload speed
		uploader.bind('UploadProgress', function(up, file) {
			logger.trace("UploadProgress event activated");
			speedCalInfo.currentTime = new Date().getTime();
			var timeUsed = speedCalInfo.currentTime - speedCalInfo.startTime; // ms
			var fileUploaded = file.loaded || 0;
			if (speedCalInfo.isResumeUpload) {
				fileUploaded = file.loaded - speedCalInfo.resumeFilesize;
			}
			file.speed = (fileUploaded / timeUsed * 1000).toFixed(0) || 0; // unit: byte/s
		});

		logger.debug("bind UploadProgress event");

		// bind 'ChunkUploaded' event
		// store the chunk upload info and set next chunk upload url
		uploader.bind('ChunkUploaded', function(up, file, info) {
			logger.debug("ChunkUploaded event activated");
			logger.debug("file: ", file);
			logger.debug("info: ", info);
			var res = that.parseJSON(info.response);
			logger.debug("res: ", res);
			// ctx should look like '[chunk01_ctx],[chunk02_ctx],[chunk03_ctx],...'
			ctx = ctx ? ctx + ',' + res.ctx : res.ctx;
			var leftSize = info.total - info.offset;
			var chunk_size = up.getOption && up.getOption('chunk_size');
			chunk_size = chunk_size || (up.settings && up.settings.chunk_size);
			if (leftSize < chunk_size) {
				up.setOption({
					'url': qiniuUploadUrl + '/mkblk/' + leftSize
				});
				if(uploadConfig.saveType == "remote"){
					up.setOption({
					'url': qiniuUploadUrl + 'chunk.php'
				});
				}
				logger.debug("up.setOption url: ", qiniuUploadUrl + '/mkblk/' + leftSize);
			}
			if(uploadConfig.saveType == "remote"){
				up.setOption({
					'headers': {
						'Authorization': getUptoken(file)
					}
				});
			}else{
				up.setOption({
					'headers': {
						'Authorization': 'UpToken ' + getUptoken(file)
					}
				});
			}
			localStorage.setItem(file.name, that.stringifyJSON({
				ctx: ctx,
				percent: file.percent,
				total: info.total,
				offset: info.offset,
				time: (new Date()).getTime()
			}));
		});

		logger.debug("bind ChunkUploaded event");

		var retries = qiniuUploadUrls.length;

		// if error is unkown switch upload url and retry
		var unknow_error_retry = function(file){
			if (retries-- > 0) {
				setTimeout(function(){
					that.resetUploadUrl();
					file.status = plupload.QUEUED;
					uploader.stop();
					uploader.start();
				}, 0);
				return true;
			}else{
				retries = qiniuUploadUrls.length;
				return false;
			}
		};

		// bind 'Error' event
		// check the err.code and return the errTip
		uploader.bind('Error', (function(_Error_Handler) {
			return function(up, err) {
				logger.error("Error event activated");
				logger.error("err: ", err);
				var errTip = '';
				var file = err.file;
				if (file) {
					switch (err.code) {
						case plupload.FAILED:
							errTip = '上传失败。请稍后再试。';
							break;
						case plupload.FILE_SIZE_ERROR:
							var max_file_size = up.getOption && up.getOption('max_file_size');
							max_file_size = max_file_size || (up.settings && up.settings.max_file_size);
							errTip = '文件过大，您当前用户组最多可上传' + max_file_size + '的文件';
							break;
						case plupload.FILE_EXTENSION_ERROR:
							errTip = '您当前的用户组不可上传此文件';
							break;
						case plupload.HTTP_ERROR:
							if (err.response === '') {
								// Fix parseJSON error ,when http error is like net::ERR_ADDRESS_UNREACHABLE
								errTip = err.message || '未知网络错误。';
								if (!unknow_error_retry(file)) {
									return;
								}
								break;
							}
							if(uploadConfig.saveType == "oss"){
								var str = err.response
									  var a = $.parseXML(str);
									   $(a).find('Error').each(function () {
										  errTip = $(this).children('Message').text();
										  var errorText = "Error";
									  });
								if(err.status == 203){
									errTip = "上传失败，请检查空间容量、是否重名";
								}
							}else if(uploadConfig.saveType == "s3" && err.status!=401){
								var str = err.response
									  var a = $.parseXML(str);
									   $(a).find('Error').each(function () {
										  errTip = $(this).children('Message').text();
										  var errorText = "Error";
									  });
							}else{
								var errorObj = that.parseJSON(err.response);
								var errorText = errorObj.error;
								if (err.status==579){
									var errorObj2 = that.parseJSON(errorText);
									errorText=errorObj2.error;
								}
								switch (err.status) {
									case 400:
										errTip = "请求报文格式错误。";
										break;
									case 401:
										errTip = "客户端认证授权失败。请重试或提交反馈。";
										break;
									case 405:
										errTip = "客户端请求错误。请重试或提交反馈。";
										break;
									case 579:
										errTip = "资源上传成功，但回调失败。";
										break;
									case 599:
										errTip = "网络连接异常。请重试或提交反馈。";
										if (!unknow_error_retry(file)) {
											return;
										}
										break;
									case 614:
										errTip = "文件已存在。";
										try {
											errorObj = that.parseJSON(errorObj.error);
											errorText = errorObj.error || 'file exists';
										} catch (e) {
											errorText = errorObj.error || 'file exists';
										}
										break;
									case 631:
										errTip = "指定空间不存在。";
										break;
									case 701:
										errTip = "上传数据块校验出错。请重试或提交反馈。";
										break;
									default:
										errTip = "未知错误。";
										if (!unknow_error_retry(file)) {
											return;
										}
										break;
							}
							
							}
							errTip = errTip + '(' + err.status + '：' + errorText + ')';
							break;
						case plupload.SECURITY_ERROR:
							errTip = '安全配置错误。请联系网站管理员。';
							break;
						case plupload.GENERIC_ERROR:
							errTip = '上传失败。请稍后再试。';
							break;
						case plupload.IO_ERROR:
							errTip = '上传失败。请稍后再试。';
							break;
						case plupload.INIT_ERROR:
							errTip = '网站配置错误。请联系网站管理员。';
							uploader.destroy();
							break;
						default:
							errTip = err.message + err.details;
							if (!unknow_error_retry(file)) {
								return;
							}
							break;
					}
					if (_Error_Handler) {
						_Error_Handler(up, err, errTip);
					}
				}
				up.refresh(); // Reposition Flash/Silverlight
			};
		})(_Error_Handler));

		logger.debug("bind Error event");

		// bind 'FileUploaded' event
		// intercept the complete of upload
		// - get downtoken from downtoken_url if bucket is private
		// - invoke mkfile api to compose chunks if upload strategy is chunk upload
		uploader.bind('FileUploaded', (function(_FileUploaded_Handler) {
			return function(up, file, info) {
				logger.debug("FileUploaded event activated");
				logger.debug("file: ", file);
				logger.debug("info: ", info);
				if(uploadConfig.saveType == "s3"){
				}
				var last_step = function(up, file, info) {
					if (op.downtoken_url) {
						// if op.dowontoken_url is not empty
						// need get downtoken before invoke the _FileUploaded_Handler
						var ajax_downtoken = that.createAjax();
						ajax_downtoken.open('POST', op.downtoken_url, true);
						ajax_downtoken.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
						ajax_downtoken.onreadystatechange = function() {
							if (ajax_downtoken.readyState === 4) {

								if (ajax_downtoken.status === 200 || ajax_downtoken.status === 204||ajax_downtoken.status === 303) {

										var res_downtoken;
										try {
											res_downtoken = that.parseJSON(ajax_downtoken.responseText);
										} catch (e) {
											throw ('invalid json format');
										}
										var info_extended = {};
										plupload.extend(info_extended, that.parseJSON(info), res_downtoken);
										if (_FileUploaded_Handler) {
											_FileUploaded_Handler(up, file, that.stringifyJSON(info_extended));
										}
							  
								} else {
									uploader.trigger('Error', {
										status: ajax_downtoken.status,
										response: ajax_downtoken.responseText,
										file: file,
										code: plupload.HTTP_ERROR
									});
								}
							}
						};
						ajax_downtoken.send('key=' + that.parseJSON(info).key + '&domain=' + op.domain);
					} else if (_FileUploaded_Handler) {
						_FileUploaded_Handler(up, file, info);
					}
				};        
				if(uploadConfig.saveType == "oss" || uploadConfig.saveType == "upyun"){
					ctx = 0;
				}else{
					var res = that.parseJSON(info.response);
					ctx = ctx ? ctx : res.ctx;
				}
				// if ctx is not empty
				//      that means the upload strategy is chunk upload
				//      befroe the invoke the last_step
				//      we need request the mkfile to compose all uploaded chunks
				// else
				//      invalke the last_step
				logger.debug("ctx: ", ctx);
				if (ctx) {
					var key = '';
					logger.debug("save_key: ", op.save_key);
					if (!op.save_key) {
						key = getFileKey(up, file, that.key_handler);
						key = key ? '/key/' + that.URLSafeBase64Encode(key) : '';
					}

					var fname = '/fname/' + that.URLSafeBase64Encode(file.name);
					if(uploadConfig.saveType=="remote"){
						if (!op.save_key) {
							key = getFileKey(up, file, that.key_handler);
							key = key ? that.URLSafeBase64Encode(key) : '';
						}
						fname = '' + that.URLSafeBase64Encode(file.name);
						op.x_vars= {
							'path': file.path,
						};
					}
					logger.debug("op.x_vars: ", op.x_vars);
					if(uploadConfig.saveType == "qiniu"){
						op.x_vars= {
							'path': file.path,
						};
				}
					var x_vars = op.x_vars,
						x_val = '',
						x_vars_url = '';
					if (x_vars !== undefined && typeof x_vars === 'object') {
						for (var x_key in x_vars) {
							if (x_vars.hasOwnProperty(x_key)) {
								if (typeof x_vars[x_key] === 'function') {
									x_val = that.URLSafeBase64Encode(x_vars[x_key](up, file));
								} else if (typeof x_vars[x_key] !== 'object') {
									x_val = that.URLSafeBase64Encode(x_vars[x_key]);
								}
								x_vars_url += '/x:' + x_key + '/' + x_val;
							}
						}
					}
					local_path = "";
					if(uploadConfig.saveType == "local"||uploadConfig.saveType == "onedrive"){
						pathTmp = file.path;
						if(file.path == ""){
							pathTmp = "ROOTDIR";
						}
						local_path = '/path/'+that.URLSafeBase64Encode(pathTmp);
					}
					if(uploadConfig.saveType == "remote"){
						pathTmp = file.path;
						local_path = that.URLSafeBase64Encode(pathTmp);
						var url = qiniuUploadUrl + 'mkfile.php?size=' + file.size +"&key="+ key+"&fname="+ fname +"&path="+local_path;
					}else{
						var url = qiniuUploadUrl + '/mkfile/' + file.size + key + fname + x_vars_url+local_path;
					}
					var ie = that.detectIEVersion();
					var ajax;
					if (ie && ie <= 9) {
						ajax = new moxie.xhr.XMLHttpRequest();
						moxie.core.utils.Env.swf_url = op.flash_swf_url;
					}else{
						ajax = that.createAjax();
					}
					ajax.open('POST', url, true);
					ajax.setRequestHeader('Content-Type', 'text/plain;charset=UTF-8');
					if(uploadConfig.saveType == "remote"){
						ajax.setRequestHeader('Authorization',that.token);
					}else{
						ajax.setRequestHeader('Authorization', 'UpToken ' + that.token);
					}
					var onreadystatechange = function(){
						logger.debug("ajax.readyState: ", ajax.readyState);
						if (ajax.readyState === 4) {
							localStorage.removeItem(file.name);
							var info;
							if (ajax.status === 200) {
								info = ajax.responseText;
								logger.debug("mkfile is success: ", info);
								last_step(up, file, info);
							} else {
								info = {
									status: ajax.status,
									response: ajax.responseText,
									file: file,
									code: -200,
									responseHeaders: ajax.getAllResponseHeaders()
								};
								logger.debug("mkfile is error: ", info);
								uploader.trigger('Error', info);
							}
						}
					};
					if (ie && ie <= 9) {
						ajax.bind('readystatechange', onreadystatechange);
					}else{
						ajax.onreadystatechange = onreadystatechange;
					}
					ajax.send(ctx);
					logger.debug("mkfile: ", url);
				} else {
					last_step(up, file, info.response);
				}

			};
		})(_FileUploaded_Handler));

		logger.debug("bind FileUploaded event");

		// init uploader
		uploader.init();

		logger.debug("invoke uploader.init()");

		logger.debug("init uploader end");

		return uploader;
	};

	/**
	 * get url by key
	 * @param  {String} key of file
	 * @return {String} url of file
	 */
	this.getUrl = function(key) {
		if (!key) {
			return false;
		}
		key = encodeURI(key);
		var domain = this.domain;
		if (domain.slice(domain.length - 1) !== '/') {
			domain = domain + '/';
		}
		return domain + key;
	};

	/**
	 * invoke the imageView2 api of Qiniu
	 * @param  {Object} api params
	 * @param  {String} key of file
	 * @return {String} url of processed image
	 */
	this.imageView2 = function(op, key) {

		if (!/^\d$/.test(op.mode)) {
			return false;
		}

		var mode = op.mode,
			w = op.w || '',
			h = op.h || '',
			q = op.q || '',
			format = op.format || '';

		if (!w && !h) {
			return false;
		}

		var imageUrl = 'imageView2/' + mode;
		imageUrl += w ? '/w/' + w : '';
		imageUrl += h ? '/h/' + h : '';
		imageUrl += q ? '/q/' + q : '';
		imageUrl += format ? '/format/' + format : '';
		if (key) {
			imageUrl = this.getUrl(key) + '?' + imageUrl;
		}
		return imageUrl;
	};

	/**
	 * invoke the imageMogr2 api of Qiniu
	 * @param  {Object} api params
	 * @param  {String} key of file
	 * @return {String} url of processed image
	 */
	this.imageMogr2 = function(op, key) {
		var auto_orient = op['auto-orient'] || '',
			thumbnail = op.thumbnail || '',
			strip = op.strip || '',
			gravity = op.gravity || '',
			crop = op.crop || '',
			quality = op.quality || '',
			rotate = op.rotate || '',
			format = op.format || '',
			blur = op.blur || '';
		//Todo check option

		var imageUrl = 'imageMogr2';

		imageUrl += auto_orient ? '/auto-orient' : '';
		imageUrl += thumbnail ? '/thumbnail/' + thumbnail : '';
		imageUrl += strip ? '/strip' : '';
		imageUrl += gravity ? '/gravity/' + gravity : '';
		imageUrl += quality ? '/quality/' + quality : '';
		imageUrl += crop ? '/crop/' + crop : '';
		imageUrl += rotate ? '/rotate/' + rotate : '';
		imageUrl += format ? '/format/' + format : '';
		imageUrl += blur ? '/blur/' + blur : '';

		if (key) {
			imageUrl = this.getUrl(key) + '?' + imageUrl;
		}
		return imageUrl;
	};

	/**
	 * invoke the watermark api of Qiniu
	 * @param  {Object} api params
	 * @param  {String} key of file
	 * @return {String} url of processed image
	 */
	this.watermark = function(op, key) {
		var mode = op.mode;
		if (!mode) {
			return false;
		}

		var imageUrl = 'watermark/' + mode;

		if (mode === 1) {
			var image = op.image || '';
			if (!image) {
				return false;
			}
			imageUrl += image ? '/image/' + this.URLSafeBase64Encode(image) : '';
		} else if (mode === 2) {
			var text = op.text ? op.text : '',
				font = op.font ? op.font : '',
				fontsize = op.fontsize ? op.fontsize : '',
				fill = op.fill ? op.fill : '';
			if (!text) {
				return false;
			}
			imageUrl += text ? '/text/' + this.URLSafeBase64Encode(text) : '';
			imageUrl += font ? '/font/' + this.URLSafeBase64Encode(font) : '';
			imageUrl += fontsize ? '/fontsize/' + fontsize : '';
			imageUrl += fill ? '/fill/' + this.URLSafeBase64Encode(fill) : '';
		} else {
			// Todo mode3
			return false;
		}

		var dissolve = op.dissolve || '',
			gravity = op.gravity || '',
			dx = op.dx || '',
			dy = op.dy || '';

		imageUrl += dissolve ? '/dissolve/' + dissolve : '';
		imageUrl += gravity ? '/gravity/' + gravity : '';
		imageUrl += dx ? '/dx/' + dx : '';
		imageUrl += dy ? '/dy/' + dy : '';

		if (key) {
			imageUrl = this.getUrl(key) + '?' + imageUrl;
		}
		return imageUrl;
	};

	/**
	 * invoke the imageInfo api of Qiniu
	 * @param  {String} key of file
	 * @return {Object} image info
	 */
	this.imageInfo = function(key) {
		if (!key) {
			return false;
		}
		var url = this.getUrl(key) + '?imageInfo';
		var xhr = this.createAjax();
		var info;
		var that = this;
		xhr.open('GET', url, false);
		xhr.onreadystatechange = function() {
			if (xhr.readyState === 4 && xhr.status === 200) {
				info = that.parseJSON(xhr.responseText);
			}
		};
		xhr.send();
		return info;
	};

	/**
	 * invoke the exif api of Qiniu
	 * @param  {String} key of file
	 * @return {Object} image exif
	 */
	this.exif = function(key) {
		if (!key) {
			return false;
		}
		var url = this.getUrl(key) + '?exif';
		var xhr = this.createAjax();
		var info;
		var that = this;
		xhr.open('GET', url, false);
		xhr.onreadystatechange = function() {
			if (xhr.readyState === 4 && xhr.status === 200) {
				info = that.parseJSON(xhr.responseText);
			}
		};
		xhr.send();
		return info;
	};

	/**
	 * invoke the exif or imageInfo api of Qiniu
	 * according with type param
	 * @param  {String} ['exif'|'imageInfo']type of info
	 * @param  {String} key of file
	 * @return {Object} image exif or info
	 */
	this.get = function(type, key) {
		if (!key || !type) {
			return false;
		}
		if (type === 'exif') {
			return this.exif(key);
		} else if (type === 'imageInfo') {
			return this.imageInfo(key);
		}
		return false;
	};

	/**
	 * invoke api of Qiniu like a pipeline
	 * @param  {Array of Object} params of a series api call
	 * each object in array is options of api which name is set as 'fop' property
	 * each api's output will be next api's input
	 * @param  {String} key of file
	 * @return {String|Boolean} url of processed image
	 */
	this.pipeline = function(arr, key) {
		var isArray = Object.prototype.toString.call(arr) === '[object Array]';
		var option, errOp, imageUrl = '';
		if (isArray) {
			for (var i = 0, len = arr.length; i < len; i++) {
				option = arr[i];
				if (!option.fop) {
					return false;
				}
				switch (option.fop) {
					case 'watermark':
						imageUrl += this.watermark(option) + '|';
						break;
					case 'imageView2':
						imageUrl += this.imageView2(option) + '|';
						break;
					case 'imageMogr2':
						imageUrl += this.imageMogr2(option) + '|';
						break;
					default:
						errOp = true;
						break;
				}
				if (errOp) {
					return false;
				}
			}
			if (key) {
				imageUrl = this.getUrl(key) + '?' + imageUrl;
				var length = imageUrl.length;
				if (imageUrl.slice(length - 1) === '|') {
					imageUrl = imageUrl.slice(0, length - 1);
				}
			}
			return imageUrl;
		}
		return false;
	};
}

var Qiniu = new QiniuJsSDK();

global.Qiniu = Qiniu;

global.QiniuJsSDK = QiniuJsSDK;

})( window );
