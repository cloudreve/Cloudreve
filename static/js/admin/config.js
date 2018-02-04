$(document).ready(function(){
	configType = $("#configType").val();
	$("a[data-type='"+configType+"']").addClass("active");
	fileName = $("#fileName").val();
	$("a[data-name='"+fileName+"']").addClass("active");
})
$("#save").click(function(){
	var content = $("textarea").val();
	var type= $("#configType").val();
	$.post("/Admin/SaveConfigFile", {
		content:content,
		type:type,
	}, function(data) {
		if (data.error == false) {
			toastr["success"]("配置文件已保存");
		}else{
			toastr["warning"](data.msg);
		}
	});
})
$("#saveTheme").click(function(){
	var content = $("textarea").val();
	var name= $("#fileName").val();
	$.post("/Admin/SaveThemeFile", {
		content:content,
		name:name,
	}, function(data) {
		if (data.error == false) {
			toastr["success"]("配置文件已保存");
		}else{
			toastr["warning"](data.msg);
		}
	});
})