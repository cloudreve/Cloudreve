$("#saveAria2").click(function() {
	$("#saveAria2").attr("disabled", "true");
	$.post("/Admin/SaveAria2Setting", 
		$("#aria2Options").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveAria2").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("设置已保存");
			$("#saveAria2").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveAria2").removeAttr("disabled");
		}
	});
})
function cancel(id){
	$.post("/Admin/CancelDownload", {id:id}, function(data){
		if(data.error){
			toastr["warning"](data.message);
		}else{
			var pid = $("#i-"+id).attr("data-pid");
			$("[data-pid='"+pid+"'").remove();
			toastr["success"](data.message);

		}
	})
}
$(document).ready(function(){
	if(document.location.href.indexOf("page")!=-1){
		$("[href='#list']").click();
	}
})