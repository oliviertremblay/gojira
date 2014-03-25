package main

var defaultTemplate string = `<html>
<head>
</head>
<body>
{{#Issues}}
<div style="width:3.75in;height:3.75in; border:2px solid black; position:relative; float:left; margin-left:0.5in; margin-top:0.5in; page-break-inside:avoid">
    <h1 style="padding:5px; text-align:center;background-color:darkgray;color:white;margin:0;">{{Key}}</h1>
    <div style="margin-top:15px; padding:5px 15px;font-size:1.5em">{{Summary}}
    </div>
    <div style="position:absolute; bottom:1px; left:1px;">
    <h2 style="margin:0; padding:10px">{{Type}}</h2>
    </div>
    <img src="data:image/png;base64,{{QRCodeBase64}}" style="position:absolute;bottom:1px;right:1px;"/>
</div>
{{/Issues}}
</body>
</html>`
