$.material.init();
$(".navbar.navbar-inverse").animate({
	"backgroundColor": "#009688",

});
$(".header-panel").animate({
	"backgroundColor": "#009688",

});
$(".top-color").animate({
	"backgroundColor": "#4CAF50",

});
$("[data-change='true']").animate({
	"color": "#009688",

});
$(".btn.btn-primary").animate({
	"backgroundColor": "#009688",

});
$(".captcha_img").click(function() {
	$("[alt='captcha']:visible").attr('src', "/captcha");
})
$("#loginButton").click(function(){
	$("#loginButton").attr("disabled","true");
	$.post("/Member/TwoStepCheck", $("#loginForm").serialize(), function(data) {
		if(data.code != "200"){
				toastr["error"](data.message, "登录失败");
				$("#loginButton").removeAttr("disabled");
			
		}else{
			toastr["success"](data.message, "登录成功");
			window.location.href="/Home";
		}
	})
})