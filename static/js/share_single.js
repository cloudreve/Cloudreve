            $.material.init();
            $(".btn-fab").mouseover(function() {
                $(this).addClass("animated jello");
            });
            $(".btn-fab").mouseout(function() {
                $(this).removeClass("animated jello");
            });
            $(window).load(function() {
                $('.file_title_inside').map(function() {
                    if (this.offsetWidth < this.scrollWidth) {
                        $('.file_title_inside').liMarquee();
                    }
                });

            });
            jQuery.ajaxSetup({
                cache: true
            });

            function getSize(size) {
                var filetype = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
                var i, bit;
                for (i = 0; size >= 1024; i++) {
                    size = size / 1024;
                }
                return parseInt(size * 100) / 100 + filetype[i];
            }
            function formateDate(dateString){
                var bigDate = dateString.split(" ");
                var smallDate = bigDate[0].split("-");
                return smallDate[0]+"年"+smallDate[1]+"月"+smallDate[2]+"日 "+bigDate[1];
            }

            function audioPause() {
                document.getElementById('preview-target').pause();
                dp.pause();
            }
            $("#size").html(getSize(shareInfo.fileSize));
            $("#down_num").html(shareInfo.downloadNum);
            $("#view_num").html(shareInfo.ViewNum);
            shareTime = formateDate(shareInfo.shareDate);
            $("#share_time").html(shareTime);
            $("#download").click(function() {
                $.post("/Share/getDownloadUrl", {
                    key: shareInfo.shareId
                }, function(result) {
                    if (result.error) {
                        toastr["error"](result.msg)
                    } else {
                        location.href = result.result;
                    }
                });
            });
            vplayderLoad = false;
        function includeCss(filename) {
            var head = document.getElementsByTagName('head')[0];
            var link = document.createElement('link');
            link.href = filename;
            link.rel = 'stylesheet';
            link.type = 'text/css';
            head.appendChild(link)
        }
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
                container: document.getElementById("preview-target"),
                screenshot: true,
                video: {
                    url: url
                },
                danmaku: {
                       id: 'Cloudreve_'+shareInfo.shareId,
                       api: 'https://api.prprpr.me/dplayer/'
                }
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
            var openPhotoSwipe = function(pic, ww, hh) {
                var pswpElement = document.querySelectorAll('.pswp')[0];
                items = [{
                    src: pic,
                    w: ww,
                    h: hh
                }];
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

            if (/\.(gif|jpg|jpeg|png|svg|SVG|GIF|JPG|PNG)$/.test(shareInfo.fileName)) {
                $.getScript("/static/js/photoswipe.min.js").done(function() {
                    $.getScript("/static/js/photoswipe-ui-default.js").done(function() {
                        $("#previewButton").click(function() {
                            if (shareInfo.allowPreview) {
                                x = shareInfo.picSize.split(",")[0];
                                y = shareInfo.picSize.split(",")[1];
                                openPhotoSwipe("/Share/Preview/" + shareInfo.shareId, x, y);
                            } else {
                                toastr["error"]("请先登录")
                            }
                        })

                    })

                })
            } else if (/\.(mp4|flv|avi|tff|MP4|FLV|AVI|TFF)$/.test(shareInfo.fileName)) {
                $(".file-sign").html('<i class="fa fa-file-movie-o" aria-hidden="true"></i>')
                $("#previewButton").click(function() {
                    if (shareInfo.allowPreview) {
                        $(".previewContent").html('<div id="preview-target" style="width: 100%;object-fit: fill;height:400px" class="preview"></div>');
                        loadDPlayer('/Share/Preview/' + shareInfo.shareId);
                        $('#previewModal').modal();
                    } else {
                        toastr["error"]("请先登录")
                    }
                })
            } else if (/\.(MP3|mp3|wav|WAV|ogg|OGG)$/.test(shareInfo.fileName)) {
                $(".file-sign").html('<i class="fa fa-file-audio-o" aria-hidden="true"></i>');
                $("#previewButton").click(function() {
                    if (shareInfo.allowPreview) {
                        $(".previewContent").html('<audio id="preview-target" style="width: 100%;object-fit: fill" controls="controls" class="preview"  src="/Share/Preview/' + shareInfo.shareId + '" ></audio>');
                        $('#previewModal').modal();
                    } else {
                        toastr["error"]("请先登录")
                    }
                })
            } else {
                $("#previewButton").click(function() {
                    toastr["warning"]("当前文件暂不支持预览")
                });
                if (/\.(doc|DOC|docx|DOCX|ogg)$/.test(shareInfo.fileName)) {
                    $(".file-sign").html('<i class="fa fa-file-word-o" aria-hidden="true"></i>');
                } else if (/\.(ppt|PPT|pptx|PPTX)$/.test(shareInfo.fileName)) {
                    $(".file-sign").html('<i class="fa fa-file-powerpoint-o" aria-hidden="true"></i>');
                } else if (/\.(pdf|PDF)$/.test(shareInfo.fileName)) {
                    $(".file-sign").html('<i class="fa fa-file-pdf-o" aria-hidden="true"></i>');
                } else if (/\.(zip|ZIP|RAR|rar|7z|7Z)$/.test(shareInfo.fileName)) {
                    $(".file-sign").html('<i class="fa fa-file-archive-o" aria-hidden="true"></i>');
                } else {
                    $(".file-sign").html('<i class="fa fa-file-text" aria-hidden="true"></i>');
                }
            }