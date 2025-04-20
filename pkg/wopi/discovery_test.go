package wopi

import (
	"fmt"
	"testing"
)

func TestDiscoveryXmlToViewerGroup(t *testing.T) {
	xmlSrc := `<wopi-discovery>
<net-zone name="external-http">
<!--  Writer documents  -->
<app favIconUrl="https://127.0.0.1:9980/browser/80a6f97/images/x-office-document.svg" name="writer">
<action default="true" ext="sxw" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="odt" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="fodt" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  Text template documents  -->
<action default="true" ext="stw" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="ott" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  MS Word  -->
<action default="true" ext="doc" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="dot" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  OOXML wordprocessing  -->
<action default="true" ext="docx" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="docm" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="dotx" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="dotm" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  Others  -->
<action default="true" ext="wpd" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="pdb" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="hwp" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="wps" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="wri" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="lrf" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="mw" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="rtf" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="txt" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="fb2" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="cwk" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="pages" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="abw" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="602" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="writer-global">
<!--  Text master documents  -->
<action default="true" ext="sxg" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="odm" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  Writer master document templates  -->
<action default="true" ext="otm" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="writer-web">
<action default="true" ext="oth" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Calc documents  -->
<app favIconUrl="https://127.0.0.1:9980/browser/80a6f97/images/x-office-spreadsheet.svg" name="calc">
<action default="true" ext="sxc" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="ods" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="fods" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  Spreadsheet template documents  -->
<action default="true" ext="stc" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="ots" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  MS Excel  -->
<action default="true" ext="xls" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="xla" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  OOXML spreadsheet  -->
<action default="true" ext="xltx" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="xltm" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="xlsx" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="xlsb" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="xlsm" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  Others  -->
<action default="true" ext="dif" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="slk" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="csv" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="dbf" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="wk1" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="gnumeric" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="numbers" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Impress documents  -->
<app favIconUrl="https://127.0.0.1:9980/browser/80a6f97/images/x-office-presentation.svg" name="impress">
<action default="true" ext="sxi" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="odp" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="fodp" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  Presentation template documents  -->
<action default="true" ext="sti" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="otp" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  MS PowerPoint  -->
<action default="true" ext="ppt" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="pot" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  OOXML presentation  -->
<action default="true" ext="pptx" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="pptm" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="potx" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="potm" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="ppsx" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  Others  -->
<action default="true" ext="cgm" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="key" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Draw documents  -->
<app name="draw">
<action default="true" ext="sxd" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="odg" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="fodg" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  Drawing template documents  -->
<action default="true" ext="std" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="otg" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<!--  Others  -->
<action ext="svg" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="dxf" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="emf" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="wmf" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="cdr" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="vsd" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="vsdx" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="vss" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="pub" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="p65" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="wpg" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action default="true" ext="fh" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action ext="bmp" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action ext="png" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action ext="gif" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action ext="tiff" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action ext="jpg" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action ext="jpeg" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
<action ext="pdf" name="view_comment" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Math documents  -->
<!--  In fact Math documents are not supported at all.
             See: https://bugs.documentfoundation.org/show_bug.cgi?id=97006
        <app name="math">
            <action name="view" default="true" ext="sxm"/>
            <action name="edit" default="true" ext="odf"/>
        </app>
         -->
<!--  Legacy MIME-type actions (compatibility)  -->
<app name="image/svg+xml">
<action ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-powerpoint">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-excel">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Writer documents  -->
<app name="application/vnd.sun.xml.writer">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.text">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.text-flat-xml">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Calc documents  -->
<app name="application/vnd.sun.xml.calc">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.spreadsheet">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.spreadsheet-flat-xml">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Impress documents  -->
<app name="application/vnd.sun.xml.impress">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.presentation">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.presentation-flat-xml">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Draw documents  -->
<app name="application/vnd.sun.xml.draw">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.graphics">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.graphics-flat-xml">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Chart documents  -->
<app name="application/vnd.oasis.opendocument.chart">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Text master documents  -->
<app name="application/vnd.sun.xml.writer.global">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.text-master">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Math documents  -->
<!--  In fact Math documents are not supported at all.
             See: https://bugs.documentfoundation.org/show_bug.cgi?id=97006
        <app name="application/vnd.sun.xml.math">
            <action name="view" default="true" ext=""/>
        </app>
        <app name="application/vnd.oasis.opendocument.formula">
            <action name="edit" default="true" ext=""/>
        </app>
         -->
<!--  Text template documents  -->
<app name="application/vnd.sun.xml.writer.template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.text-template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Writer master document templates  -->
<app name="application/vnd.oasis.opendocument.text-master-template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Spreadsheet template documents  -->
<app name="application/vnd.sun.xml.calc.template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.spreadsheet-template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Presentation template documents  -->
<app name="application/vnd.sun.xml.impress.template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.presentation-template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Drawing template documents  -->
<app name="application/vnd.sun.xml.draw.template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.graphics-template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  MS Word  -->
<app name="application/msword">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/msword">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  MS Excel  -->
<app name="application/vnd.ms-excel">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  MS PowerPoint  -->
<app name="application/vnd.ms-powerpoint">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  OOXML wordprocessing  -->
<app name="application/vnd.openxmlformats-officedocument.wordprocessingml.document">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-word.document.macroEnabled.12">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.openxmlformats-officedocument.wordprocessingml.template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-word.template.macroEnabled.12">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  OOXML spreadsheet  -->
<app name="application/vnd.openxmlformats-officedocument.spreadsheetml.template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-excel.template.macroEnabled.12">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-excel.sheet.binary.macroEnabled.12">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-excel.sheet.macroEnabled.12">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  OOXML presentation  -->
<app name="application/vnd.openxmlformats-officedocument.presentationml.presentation">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-powerpoint.presentation.macroEnabled.12">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.openxmlformats-officedocument.presentationml.template">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-powerpoint.template.macroEnabled.12">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  Others  -->
<app name="application/vnd.wordperfect">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-aportisdoc">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-hwp">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-works">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-mswrite">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-dif-document">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="text/spreadsheet">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="text/csv">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-dbase">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.lotus-1-2-3">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/cgm">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/vnd.dxf">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/x-emf">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/x-wmf">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/coreldraw">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.visio2013">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.visio">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.ms-visio.drawing">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-mspublisher">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-sony-bbeb">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-gnumeric">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/macwriteii">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-iwork-numbers-sffnumbers">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.oasis.opendocument.text-web">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-pagemaker">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="text/rtf">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="text/plain">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-fictionbook+xml">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/clarisworks">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/x-wpg">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-iwork-pages-sffpages">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.openxmlformats-officedocument.presentationml.slideshow">
<action default="true" ext="" name="edit" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-iwork-keynote-sffkey">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-abiword">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/x-freehand">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/vnd.sun.xml.chart">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/x-t602">
<action default="true" ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/bmp">
<action ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/png">
<action ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/gif">
<action ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/tiff">
<action ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/jpg">
<action ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="image/jpeg">
<action ext="" name="view" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<app name="application/pdf">
<action ext="" name="view_comment" urlsrc="https://127.0.0.1:9980/browser/80a6f97/cool.html?"/>
</app>
<!--  End of legacy MIME-type actions  -->
<app name="Capabilities">
<action ext="" name="getinfo" urlsrc="https://127.0.0.1:9980/hosting/capabilities"/>
</app>
</net-zone>
</wopi-discovery>`
	group, _ := DiscoveryXmlToViewerGroup(xmlSrc)
	fmt.Print(group)
}
