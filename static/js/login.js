$.material.init();
function changeColor(c1,c2){
	$(".navbar.navbar-inverse").animate({
		"backgroundColor": c1,

	});
	$(".header-panel").animate({
		"backgroundColor": c1,

	});

	$(".btn.btn-primary").animate({
		"backgroundColor": c1,

	});
	$(".top-color").animate({
		"backgroundColor": c2,

	});
	$("[data-change='true']").animate({
		"color": c2,

	});
}
changeColor("#4e64d9","#3f51b5");
$(".captcha_img").click(function() {
	$("[alt='captcha']:visible").attr('src', "/captcha");
})
function updateMetaThemeColor(themeColor) {
    $('meta[name=theme-color]').remove();
    $('head').append('<meta name="theme-color" content="'+themeColor+'">');
}
function switchToReg(){
	changeColor("#46adff","#2196F3");
	updateMetaThemeColor("#2196F3");
	$("#logForm").hide();
	$("#regForm").show();
	$("[alt='captcha']:visible").attr('src', "/captcha");
	$("#regForm").removeClass("animated zoomIn");
	$("#regForm").addClass("animated zoomIn");
}
function switchToLog(){
	changeColor("#4e64d9","#3f51b5");
	updateMetaThemeColor("#3f51b5");
	$("#regForm").hide();
	$("#forgetForm").hide();
	$("#logForm").show();
	$("[alt='captcha']:visible").attr('src', "/captcha");
	$("#logForm").removeClass("animated zoomIn");
	$("#logForm").addClass("animated zoomIn");
}
function switchToEmail(){
	changeColor("#009688","#4CAF50");
	updateMetaThemeColor("#4CAF50");
	$("#regForm").hide();
	$("#emailCheck").show();
	$("#emailCheck").removeClass("animated zoomIn");
	$("#emailCheck").addClass("animated zoomIn");
}
function switchToForget(){
	changeColor("#FF9800","#F44336");
	updateMetaThemeColor("#F44336");	
	$("#regForm").hide();
	$("#logForm").hide();
	$("#forgetForm").show();
	$("[alt='captcha']:visible").attr('src', "/captcha");
	$("#forgetForm").removeClass("animated zoomIn");
	$("#forgetForm").addClass("animated zoomIn");
}
$("#loginButton").click(function(){
	$("#loginButton").attr("disabled","true");
	$.post("/Member/Login", $("#loginForm").serialize(), function(data) {
		if(data.code != "200"){
			if(data.message=="tsp"){
				window.location.href="/Member/TwoStep";
			}else{
				toastr["error"](data.message, "登录失败");
				$("#loginButton").removeAttr("disabled");
				$("[alt='captcha']:visible").attr('src', "/captcha");
			}
		}else{
			toastr["success"](data.message, "登录成功");
			window.location.href="/Home";
		}
	})
});
$("#regButton").click(function(){
	$("#regButton").attr("disabled","true");
	var re =  /^([a-zA-Z0-9]+[_|_|.]?)*[a-zA-Z0-9]+@([a-zA-Z0-9]+[_|_|.]?)*[a-zA-Z0-9]+.[a-zA-Z]{2,3}$/;
	var um=$("input[name='username-reg']").val();
	var pw=$("input[name='password-reg']").val();
	var pw1=$("input[name='password-check']").val();
	if(um.match(re) == null){
		toastr["error"]("电子邮箱格式不正确", "注册失败");
		$("#regButton").removeAttr("disabled");
	}else if(pw!=pw1){
		toastr["error"]("两次密码输入不一致", "注册失败");
		$("#regButton").removeAttr("disabled");
	}else{
		$.post("/Member/Register", $("#registerForm").serialize(), function(data) {
			if(data.code != "200"){
				toastr["error"](data.message, "注册失败");
				$("#regButton").removeAttr("disabled");
				$("[alt='captcha']:visible").attr('src', "/captcha");
			}else{
				if(data.message=="ec"){
					switchToEmail();
				}else{
					toastr["success"](data.message, "注册成功");
					$("#regButton").removeAttr("disabled");
					switchToLog();
				}
			}
		})
	}
});
$("#findMyFuckingPwd").click(function(){
	$("#findMyFuckingPwd").attr("disabled","true");
	var re =  /^([a-zA-Z0-9]+[_|_|.]?)*[a-zA-Z0-9]+@([a-zA-Z0-9]+[_|_|.]?)*[a-zA-Z0-9]+.[a-zA-Z]{2,3}$/;
	var em=$("#regEmail").val();
	if(em.match(re) == null){
		toastr["error"]("电子邮箱格式不正确");
		$("#findMyFuckingPwd").removeAttr("disabled");
	}else{
		$.post("/Member/ForgetPwd", $("#forgetPwdForm").serialize(), function(data) {
			if(data.code != "200"){
				toastr["error"](data.message);
				$("#findMyFuckingPwd").removeAttr("disabled");
				$("[alt='captcha']:visible").attr('src', "/captcha");
			}else{
				switchToLog();
				toastr["success"]("如果此邮箱在本站注册过，你将会收到一封密码重置邮件","成功");
				$("#findMyFuckingPwd").removeAttr("disabled");
			}
		});
	}
})
$("#create").click(function(){switchToReg()});
$("#forgetSwitch,#forgetSwitch2").click(function(){switchToForget()});
$("#loginSwitch2,#loginSwitch,#loginSwitch3").click(function(){switchToLog()});
$("#qqLogin").click(function(){
	window.location.href="/Member/QQLogin";
})