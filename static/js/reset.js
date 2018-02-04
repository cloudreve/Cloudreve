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
$("#regButton").click(function(){
	$("#regButton").attr("disabled","true");
	var pw=$("input[name='password-reg']").val();
	var pw1=$("input[name='password-check']").val();
	var resetKey=$("#resetKey").val();
	if(pw!=pw1){
		toastr["error"]("两次密码输入不一致", "注册失败");
		$("#regButton").removeAttr("disabled");
	}else{
		$.post("/Member/Reset", {pwd:pw,key:resetKey}, function(data) {
			if(data.code != "200"){
				toastr["error"](data.message);
				$("#regButton").removeAttr("disabled");
			}else{
				toastr["success"](data.message, "重设成功");
				$("#regButton").removeAttr("disabled");

			}
		})
	}
});