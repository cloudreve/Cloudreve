$("#saveTask").click(function() {
	$("#saveTask").attr("disabled", "true");
	$.post("/Admin/SaveTaskOption", 
		$("#taskOptions").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveTask").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("设置已保存");
			$("#saveTask").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveTask").removeAttr("disabled");
		}
	});
})
$("#generateToken").click(function(){
	len = 64;
　　var $chars = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678';
	var maxPos = $chars.length;
　　var pwd = '';
　　for (i = 0; i < len; i++) {
　　　　pwd += $chars.charAt(Math.floor(Math.random() * maxPos));
　　}
　　$("#task_queue_token").val(pwd);
})
$(document).ready(function(){
	if(document.location.href.indexOf("page")!=-1){
		$("[href='#list']").click();
	}
})