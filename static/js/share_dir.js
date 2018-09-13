            $.material.init();
            angular.module('FileManagerApp').config(['fileManagerConfigProvider', function(config) {

                var defaults = config.$get();
                config.set({
                    appName: 'angular-filemanager',
                    defaultLang: 'zh_cn',
                    sidebar: true,
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
                        shareFile: false,
                        shareFile: false,
                    }),
                });
            }]);
            function includeCss(filename) {
                var head = document.getElementsByTagName('head')[0];
                var link = document.createElement('link');
                link.href = filename;
                link.rel = 'stylesheet';
                link.type = 'text/css';
                head.appendChild(link)
            }
            previewLoad = 0;
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
            jQuery.ajaxSetup({
            cache: true
        });
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