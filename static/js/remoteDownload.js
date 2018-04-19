function getMemory() {
	$.get("/Member/Memory", function(data) {
		var dataObj = eval("(" + data + ")");
		if (dataObj.rate >= 100) {
			$("#memory_bar").css("width", "100%");
			$("#memory_bar").addClass("progress-bar-warning");
			toastr["error"]("您的已用容量已超过容量配额，请尽快删除多余文件或购买容量");

		} else {
			$("#memory_bar").css("width", dataObj.rate + "%");
		}
		$("#used").html(dataObj.used);
		$("#total").html(dataObj.total);
	});
}

page = 1;
window.onload = function() {
	$.material.init();

	getMemory();
}
$(function() {
	$("#loadFinished").click(function(){
		$.getJSON("/RemoteDownload/ListFinished?page="+page, function(data) {
			if(data.length == 0){
				$("#loadFinished").html("已加载全部");
				$("#loadFinished").attr("disabled","true");
			}else{
				$("#loadFinished").html("继续加载");
			}
			data.forEach(function(e) {
				$("#completeItemContent").append(function() {
					var row = '<tr id="i-' + e["id"] + '" data-pid="'+e["pid"]+'"><th scope="row" class="centerTable">' + e["id"] + '</th><td>' + e["fileName"] + '</td>';
					row = row + '<td class="centerTable">' + bytesToSize(e["totalLength"]) + '</td>';
					row = row + '<td class="centerTable">' + e["save_dir"] + '</td>';
					switch(e["status"]){
						case "error":
							row = row +'<td class="centerTable"><span class="download-error" data-toggle="tooltip" data-placement="top" title="'+e["msg"]+'">失败</span></td>'
							break;
						case "canceled":
							row = row +'<td class="centerTable"><span >取消</span></td>'
							break;
						case "canceled":
							row = row +'<td class="centerTable"><span >取消</span></td>'
							break;
						case "complete":
							row = row +'<td class="centerTable"><span class="download-success">完成</span></td>'
							break;
					}
					return row + "</tr>";
				});
				switch(e["status"]){
					case "error":
						$("#i-" + e["id"]).addClass("td-error");
						$('[data-toggle="tooltip"]').tooltip()
						break;
					case "complete":
						$("#i-" + e["id"]).addClass("td-success");
						break;
				}
			});
			page++;
		})
	})
})
$.get("/RemoteDownload/FlushUser", function() {
	$("#loadStatus").html("加载下载列表中...");
	loadDownloadingList();
})
function loadDownloadingList() {
	$("#itemContent").html("");
	$.getJSON("/RemoteDownload/ListDownloading", function(data) {
			if(data.length == 0){
				$("#loadStatus").html("下载列表为空");
			}
			data.forEach(function(e) {
				$("#itemContent").append(function() {
					var row = '<tr id="i-' + e["id"] + '" data-pid="'+e["pid"]+'"><th scope="row" class="centerTable">' + e["id"] + '</th><td>' + e["fileName"] + '</td>';
					row = row + '<td class="centerTable">' + bytesToSize(e["totalLength"]) + '</td>';
					row = row + '<td class="centerTable">' + e["save_dir"] + '</td>';
					if (e["downloadSpeed"] == "0") {
						row = row + '<td class="centerTable">-</td>';
					} else {
						row = row + '<td class="centerTable">' + bytesToSize(e["downloadSpeed"]) + '/s</td>';
					}
					row = row + '<td class="centerTable">' + GetPercent(e["completedLength"], e["totalLength"]) + '</td>'
					row = row + '<td class="centerTable"><a href="javascript:" onclick="cancel('+e["id"]+')" >取消</a></td>'
					return row + "</tr>";
				});
				$("#i-" + e["id"]).css({
					"background-image": "-webkit-gradient(linear, left top, right top, from(#ecefff), to(white), color-stop("+e["completedLength"]/e["totalLength"]+", #ecefff), color-stop("+e["completedLength"]/e["totalLength"]+", white))",

				});
				$(".table-responsive").slideDown();
				$("#loadStatus").slideUp();
			});
		})
	}

	function bytesToSize(bytes) {
		if (bytes === 0) return '0 B';
		var k = 1024, // or 1024
			sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
			i = Math.floor(Math.log(bytes) / Math.log(k));

		return (bytes / Math.pow(k, i)).toPrecision(3) + ' ' + sizes[i];
	}

	function GetPercent(num, total) {
		num = parseFloat(num);
		total = parseFloat(total);
		if (isNaN(num) || isNaN(total)) {
			return "-";
		}
		return total <= 0 ? "0%" : (Math.round(num / total * 10000) / 100.00 + "%");
	}

	function cancel(id){
		$.post("/RemoteDownload/Cancel", {id:id}, function(data){
			if(data.error){
				toastr["warning"](data.message);
			}else{
				var pid = $("#i-"+id).attr("data-pid");
				$("[data-pid='"+pid+"'").remove();
				toastr["success"](data.message);

			}
		})
	}

	function refresh(){
		$.get("/RemoteDownload/FlushUser?i="+Math.random(), function() {
			$("#loadStatus").html("加载下载列表中...");
			loadDownloadingList();
		})
	}