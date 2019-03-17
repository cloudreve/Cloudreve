$("#addRow").click(function(){
	$("#colorTable tbody").append('<tr><td></td><td><textarea class="form-control" rows="4" name="color[]"></textarea></td></tr>');
})
$("[data-action='removeRow']").on("click",function(e){
	$(this).parent().parent().remove();
});
$("#saveColor").click(function() {
	$("#saveColor").attr("disabled", "true");
	$.post("/Admin/SaveColorSetting", 
		$("#colorForm").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveColor").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("设置已保存");
			$("#saveColor").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveColor").removeAttr("disabled");
		}
	});
})