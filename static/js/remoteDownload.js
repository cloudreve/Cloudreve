$.get("/RemoteDownload/FlushUser", function() {
	$("#loadStatus").html("加载下载列表中...");
	loadDownloadingList();
})

function loadDownloadingList() {
	$.getJSON("/RemoteDownload/ListDownloading", function(data) {
		$("#itemContent").html();
			if(data.length == 0){
				$("#loadStatus").html("下载列表为空");
			}
			data.forEach(function(e) {
				$("#itemContent").append(function() {
					var row = '<tr id="i-' + e["id"] + '"><th scope="row" class="centerTable">' + e["id"] + '</th><td>' + e["fileName"] + '</td>';
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
		var k = 1000, // or 1024
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
			console.log(data);
			if(data.error){
				toastr["warning"](data.message);
			}else{
				toastr["success"](data.message);
			}
		})
	}