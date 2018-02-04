$("#pack").submit(function() {
	$("#addPacButton").attr("disabled", "true");
	$.post("/Admin/AddPack", 
		$("#pack").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#addPacButton").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("商品已添加");
			document.getElementById("pack").reset();
			$("#addPacButton").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#addPacButton").removeAttr("disabled");
		}
	});
	return false;
})
$("#groupP").submit(function() {
	$("#addPacButton").attr("disabled", "true");
	$.post("/Admin/AddGroupPurchase", 
		$("#groupP").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#addPacButton").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("商品已添加");
			document.getElementById("groupP").reset();
			$("#addPacButton").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#addPacButton").removeAttr("disabled");
		}
	});
	return false;
})
$("[data-action='delete_pack'").click(function() {
	var packId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/DeletePack", {
		id: packId
	}, function(data) {
		if (data.error == true) {
			toastr["warning"]("删除失败");
			$(this).removeAttr("disabled");
		} else {
			toastr["success"]("商品已删除");
			$(this).removeAttr("disabled");
			$('[data-toggle="tooltip"]').tooltip("hide")
			thisObj.parent().parent().remove();
		}
	});
})
$("[data-action='delete_group'").click(function() {
	var packId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/DeleteGroupPurchase", {
		id: packId
	}, function(data) {
		if (data.error == true) {
			toastr["warning"]("删除失败");
			$(this).removeAttr("disabled");
		} else {
			toastr["success"]("商品已删除");
			$(this).removeAttr("disabled");
			$('[data-toggle="tooltip"]').tooltip("hide")
			thisObj.parent().parent().remove();
		}
	});
})