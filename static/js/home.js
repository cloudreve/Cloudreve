		upload_load = 0;
		cliLoad = 0;
		previewLoad = 0;
		mobileMode = true;
		blankClick = 0;
		authC = true;
		mdLoad = 0;
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
		if (uploadConfig.allowSource == "1") {
			Source = true;
		} else {
			Source = false;
		}
		if (uploadConfig.allowShare == "1") {
			allowShare = true;
		} else {
			allowShare = false;
		}
		angular.module('FileManagerApp').config(['fileManagerConfigProvider', function(config) {

			var defaults = config.$get();
			config.set({
				appName: 'angular-filemanager',
				defaultLang: 'zh_cn',
				sidebar: false,
				pickCallback: function(item) {
					var msg = 'Picked %s "%s" for external use'
						.replace('%s', item.type)
						.replace('%s', item.fullPath());
					window.alert(msg);
				},
				allowedActions: angular.extend(defaults.allowedActions, {
					pickFiles: false,
					pickFolders: false,
					changePermissions: false,
					upload: false,
					shareFile: uploadConfig.allowShare == "1" ? true : false,
				}),
			});
		}]);
		window.onload = function() {
			$("[href='/Home']").addClass("active");
			$.material.init();
			getMemory();

		}
		jQuery.ajaxSetup({
			cache: true
		});
		openUpload = function() {
			$("#pickfiles").css("z-index", "9999");
			$('#upload_modal').modal();
			if (!upload_load) {
				$("#up_text").html("正在加载上传组件...");
				$.getScript("/static/js/moxie.js").done(function() {
					$.getScript("/static/js/plupload.dev.js").done(function() {
						$.getScript("/static/js/i18n/zh_CN.js").done(function() {
							$.getScript("/static/js/ui.js").done(function() {
								$.getScript("/static/js/qiniu.js").done(function() {
									$.getScript("/static/js/main.js");
									$("#up_text").html("选择文件");
									toastr.clear();
									upload_load = 1;
								});
							});
						});
					});
				});
			} else {
				$("[class='moxie-shim moxie-shim-html5']").show();
			}
		 }
		function closeUpload() {
			$('#upload_modal').modal('hide');
			$("[class='moxie-shim moxie-shim-html5']").hide();
		}
		$(function() {
			$('#upload_modal').on('hide.bs.modal', function() {
				$("[class='moxie-shim moxie-shim-html5']").hide();

			})
		});
		function includeCss(filename) {
			var head = document.getElementsByTagName('head')[0];
			var link = document.createElement('link');
			link.href = filename;
			link.rel = 'stylesheet';
			link.type = 'text/css';
			head.appendChild(link)
		}
		function loadMdEditor(result){
			if (mdLoad == 0) {
				toastr["info"]("加载编辑器...");
				includeCss("/static/css/mdeditor/codemirror.css");
				$.getScript("/static/js/mdeditor/markdown-it.min.js").done(function() {
					$.getScript("/static/js/mdeditor/toMark.min.js").done(function() {
						$.getScript("/static/js/mdeditor/tui-code-snippet.min.js").done(function() {
							$.getScript("/static/js/mdeditor/codemirror.js").done(function() {
								$.getScript("/static/js/mdeditor/highlight.pack.min.js").done(function() {
									$.getScript("/static/js/mdeditor/squire-raw.js").done(function() {
										$.when(
											$.ajax({
												async: false,
												url: "/static/js/mdeditor/tui-editor-Editor-all.min.js",
												dataType: "script"
											})).done(function(){
												editor = new tui.Editor({
														el: document.querySelector('#md'),
														initialEditType: 'markdown',
														previewStyle: 'vertical',
														height: 'auto',
														initialValue: result,
														language:"zh",
												});

												toastr.clear();
												mdLoad = 1;
										});
											
									
									});
								});
							});
						});
					});
				});
			}else{
				editor.setMarkdown(result);
			}
		}
		function addCSS(url) {
			var link = document.createElement('link');
			link.type = 'text/css';
			link.rel = 'stylesheet';
			link.href = url;
			document.getElementsByTagName("head")[0].appendChild(link);
		}



		function audioPause() {
			document.getElementById('audiopreview-target').pause();
			dp.pause()
		}

		
		var openPhotoSwipe = function(items) {
			var pswpElement = document.querySelectorAll('.pswp')[0];

			var options = {
				history: false,
				focus: false,
				showAnimationDuration: 5,
				hideAnimationDuration: 0,
				bgOpacity: 0.8,
				closeOnScroll: 0,

			};

			var gallery = new PhotoSwipe(pswpElement, PhotoSwipeUI_Default, items, options);
			gallery.init();
		};
		var loadPreview = function(t) {

			if (!previewLoad) {
				toastr["info"]("加载预览组件...");
				$.getScript("/static/js/photoswipe.min.js").done(function() {
					$.getScript("/static/js/photoswipe-ui-default.js").done(function() {
						openPhotoSwipe(t);
						toastr.clear();
						previewLoad = 1;
					})

				})
			} else {
				openPhotoSwipe(t);
			}
		}
		vplayderLoad = false;
		var loadDPlayer = function(url){
			if (!vplayderLoad) {
				toastr["info"]("加载预览组件...");
				includeCss("/static/css/DPlayer.min.css");
				$.getScript("/static/js/DPlayer.min.js").done(function() {
					toastr.clear();
					vplayderLoad = 1;
					playVideo(url);
				});
			}else{
				playVideo(url);
			}
		}
		var playVideo = function(url){
			 dp = new DPlayer({
				container: document.getElementById("videopreview-target"),
				screenshot: true,
				video: {
					url: url
				},
			});
			dp.on("fullscreen", function(){
				$(".modal-backdrop").hide();
				$("#side").hide();
				return false;
			});
			dp.on("fullscreen_cancel", function(){
				$(".modal-backdrop").show();
				$("#side").show();
				return false;
			})
		}
		function suffix($fname){
			alert($fname);
		}	
		function mobileBind(){
			if($(window).width()<768){
			$('a[ng-click|="selectOrUnselect(item, $event)"]').click(function(event){ 
				var menu = $("#context-menu");
				if (event.pageX >= window.innerWidth - menu.width()) {
							event.pageX -= menu.width();
						}
						if (event.pageY >= window.innerHeight - menu.height()) {
							event.pageY -= menu.height();
						}
				  $("#context-menu").css({
						"left": event.pageX,
						"top": event.pageY
					});
				$(this).contextmenu();
				return false;
			})
		}
		}

		function loadThumb(obj){
			alert("ds");
		}