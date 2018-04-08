<!DOCTYPE html>
<html>
<head>
    <title></title>
</head>
<div class="row">
    <div class="container">
        <div style="width:100%; height:100px">
            <div style="float:left; width:50%">
                <h1>Home Page</h1>
            </div>
            <div style="float:right; width:50%; height:100%; text-align:center; margin-top:20px">
                <a href="/cancel">[Cancel Account]</a>
            </div>
        </div>
        <div id="following" style="float:left; width:30%">
            <div style="float:left; width:80px">Following</div>
            <div style="margin-left: 80px">
                <a href="/addfollow">[Add Follow]</a>
            </div><br />
            {{range $key, $value := .Following}}
            <p>{{$key}}</p>
            {{end}}
        </div>
        <div id="follower" style="margin-left: 30%">
            <div style="float:left; width:80px">Follower</div><br /><br />
            {{range $key, $value := .Follower}}
            <p>{{$key}}</p>
            {{end}}
        </div>
        <div style="width:100%; height:100px">
            <div style="float:left; width:50%">
                <h1>Message Box</h1>
            </div>
        </div>
        <div id="messages" style="float:left; width:30%">
            {{range .Messages}}
            <p><font size="3">{{.Text}}</font></p>
            <p><font size="2">{{.Username}} posted at </font><font size="1">{{.DisplayTime}}</font></p><br />
            {{end}}
        </div>
    </div>
</div>

<body>
    <div style="color:green; text-align:center; position:absolute; bottom:0; width:100%;" class="footer">
        <form action="post.php" method="post">
            <p>Share with your friends:</p>
                <input type="text" name="postcontent" style="width:80%; height:100px"><br>
            <p></p><p></p><input type="submit" value="post"></p>
        </form>
    </div>
</body>


<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">
