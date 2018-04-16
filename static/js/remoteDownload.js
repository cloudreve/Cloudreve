$.get("/RemoteDownload/FlushUser",  function(){
	$("#loadStatus").html("加载下载列表中");
	loadDownloadingList();
})
function loadDownloadingList(){
	$.getJSON("/RemoteDownload/ListDownloading",  function(data){
		console.log(data);
	})
}