// /*global Qiniu */
// /*global plupload */
// /*global FileProgress */
// /*global hljs */
// function getCookieByString(cookieName){
//  var start = document.cookie.indexOf(cookieName+'=');
//  if (start == -1) return false;
//  start = start+cookieName.length+1;
//  var end = document.cookie.indexOf(';', start);
//  if (end == -1) end=document.cookie.length;
//  return document.cookie.substring(start, end);
// }

// if(uploadConfig.saveType == "oss" || uploadConfig.saveType == "upyun" || uploadConfig.saveType == "s3"){
//     ChunkSize = "0";
// }else{
//     ChunkSize = "4mb";
// }

//      uploader = Qiniu.uploader({
//         runtimes: 'html5,flash,html4',
//         browse_button: 'pickfiles',
//         container: 'container',
//         drop_element: 'container',
//         max_file_size: uploadConfig.maxSize,
//         flash_swf_url: '/bower_components/plupload/js/Moxie.swf',
//         dragdrop: true,
//         chunk_size: ChunkSize,
//         filters: {
//          mime_types :uploadConfig.allowedType,
//      },
//         multi_selection: !(moxie.core.utils.Env.OS.toLowerCase() === "ios"),
//         uptoken_url: "/Upload/Token",
//         // uptoken_func: function(){
//         //     var ajax = new XMLHttpRequest();
//         //     ajax.open('GET', $('#uptoken_url').val(), false);
//         //     ajax.setRequestHeader("If-Modified-Since", "0");
//         //     ajax.send();
//         //     if (ajax.status === 200) {
//         //         var res = JSON.parse(ajax.responseText);
//         //         console.log('custom uptoken_func:' + res.uptoken);
//         //         return res.uptoken;
//         //     } else {
//         //         console.log('custom uptoken_func err');
//         //         return '';
//         //     }
//         // },
//         domain: $('#domain').val(),
//         get_new_uptoken: true,
//         // downtoken_url: '/downtoken',
//         // unique_names: true,
//         // save_key: true,
//         // x_vars: {
//         //     'id': '1234',
//         //     'time': function(up, file) {
//         //         var time = (new Date()).getTime();
//         //         // do something with 'time'
//         //         return time;
//         //     },
//         // },
//         auto_start: true,
//         log_level: 5,
//         init: {
//             'FilesAdded': function(up, files) {
//                 $('table').show();
//                 $('#upload_box').show();
//                 $('#success').hide();
//                 $('#info_box').hide();

//                   $.cookie('path', decodeURI(getCookieByString("path_tmp"))); 
//                 plupload.each(files, function(file) {
//                     var progress = new FileProgress(file, 'fsUploadProgress');
//                     progress.setStatus("等待...");
//                     progress.bindUploadCancel(up);
//                 });
            
//             },
//             'BeforeUpload': function(up, file) {
//                 var progress = new FileProgress(file, 'fsUploadProgress');
//                 var chunk_size = plupload.parseSize(this.getOption('chunk_size'));
//                 if (up.runtime === 'html5' && chunk_size) {
//                     progress.setChunkProgess(chunk_size);
//                 }
//             },
//             'UploadProgress': function(up, file) {
//                 var progress = new FileProgress(file, 'fsUploadProgress');
//                 var chunk_size = plupload.parseSize(this.getOption('chunk_size'));
//                 progress.setProgress(file.percent + "%", file.speed, chunk_size);
//             },
//             'UploadComplete': function(up, file) {
//                 $('#success').show();
//                 toastr["success"]("队列全部文件处理完毕");
//                 getMemory();
//             },
//             'FileUploaded': function(up, file, info) {
//                 var progress = new FileProgress(file, 'fsUploadProgress');
//                 progress.setComplete(up, info);
//             },
//             'Error': function(up, err, errTip) {
//                 $('#upload_box').show();
//                     $('table').show();
//                     $('#info_box').hide();
//                     var progress = new FileProgress(err.file, 'fsUploadProgress');
//                     progress.setError();
//                     progress.setStatus(errTip);
//                     toastr["error"]("上传时遇到错误");
//                 }
//                 // ,
//                 // 'Key': function(up, file) {
//                 //     var key = "";
//                 //     // do something with key
//                 //     return key
//                 // }
//         }
//     });

//     uploader.bind('FileUploaded', function(up,file) {
//         console.log('a file is uploaded');
//     });
//     $('#container').on(
//         'dragenter',
//         function(e) {
//             e.preventDefault();
//             $('#container').addClass('draging');
//             e.stopPropagation();
//         }
//     ).on('drop', function(e) {
//         e.preventDefault();
//         $('#container').removeClass('draging');
//         e.stopPropagation();
//     }).on('dragleave', function(e) {
//         e.preventDefault();
//         $('#container').removeClass('draging');
//         e.stopPropagation();
//     }).on('dragover', function(e) {
//         e.preventDefault();
//         $('#container').addClass('draging');
//         e.stopPropagation();
//     });



//     $('#show_code').on('click', function() {
//         $('#myModal-code').modal();
//         $('pre code').each(function(i, e) {
//             hljs.highlightBlock(e);
//         });
//     });


//     $('body').on('click', 'table button.btn', function() {
//         $(this).parents('tr').next().toggle();
//     });


//     var getRotate = function(url) {
//         if (!url) {
//             return 0;
//         }
//         var arr = url.split('/');
//         for (var i = 0, len = arr.length; i < len; i++) {
//             if (arr[i] === 'rotate') {
//                 return parseInt(arr[i + 1], 10);
//             }
//         }
//         return 0;
//     };

//     $('#myModal-img .modal-body-footer').find('a').on('click', function() {
//         var img = $('#myModal-img').find('.modal-body img');
//         var key = img.data('key');
//         var oldUrl = img.attr('src');
//         var originHeight = parseInt(img.data('h'), 10);
//         var fopArr = [];
//         var rotate = getRotate(oldUrl);
//         if (!$(this).hasClass('no-disable-click')) {
//             $(this).addClass('disabled').siblings().removeClass('disabled');
//             if ($(this).data('imagemogr') !== 'no-rotate') {
//                 fopArr.push({
//                     'fop': 'imageMogr2',
//                     'auto-orient': true,
//                     'strip': true,
//                     'rotate': rotate,
//                     'format': 'png'
//                 });
//             }
//         } else {
//             $(this).siblings().removeClass('disabled');
//             var imageMogr = $(this).data('imagemogr');
//             if (imageMogr === 'left') {
//                 rotate = rotate - 90 < 0 ? rotate + 270 : rotate - 90;
//             } else if (imageMogr === 'right') {
//                 rotate = rotate + 90 > 360 ? rotate - 270 : rotate + 90;
//             }
//             fopArr.push({
//                 'fop': 'imageMogr2',
//                 'auto-orient': true,
//                 'strip': true,
//                 'rotate': rotate,
//                 'format': 'png'
//             });
//         }

//         $('#myModal-img .modal-body-footer').find('a.disabled').each(function() {

//             var watermark = $(this).data('watermark');
//             var imageView = $(this).data('imageview');
//             var imageMogr = $(this).data('imagemogr');

//             if (watermark) {
//                 fopArr.push({
//                     fop: 'watermark',
//                     mode: 1,
//                     image: 'http://www.b1.qiniudn.com/images/logo-2.png',
//                     dissolve: 100,
//                     gravity: watermark,
//                     dx: 100,
//                     dy: 100
//                 });
//             }

//             if (imageView) {
//                 var height;
//                 switch (imageView) {
//                     case 'large':
//                         height = originHeight;
//                         break;
//                     case 'middle':
//                         height = originHeight * 0.5;
//                         break;
//                     case 'small':
//                         height = originHeight * 0.1;
//                         break;
//                     default:
//                         height = originHeight;
//                         break;
//                 }
//                 fopArr.push({
//                     fop: 'imageView2',
//                     mode: 3,
//                     h: parseInt(height, 10),
//                     q: 100,
//                     format: 'png'
//                 });
//             }

//             if (imageMogr === 'no-rotate') {
//                 fopArr.push({
//                     'fop': 'imageMogr2',
//                     'auto-orient': true,
//                     'strip': true,
//                     'rotate': 0,
//                     'format': 'png'
//                 });
//             }
//         });



//         var newUrl = Qiniu.pipeline(fopArr, key);

//         var newImg = new Image();
//         img.attr('src', 'images/loading.gif');
//         newImg.onload = function() {
//             img.attr('src', newUrl);
//             img.parent('a').attr('href', newUrl);
//         };
//         newImg.src = newUrl;
//         return false;
//     });

// function t(){
//     uploader.getNewUpToken();
// }