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
                <a href="/cancel">[Cancel Account] {{.Username}}</a><br />
                <a href="/login">[Sign In]</a><br />
                <a href="/register">[Sign Up]</a>
            </div>
        </div>
        <div id="following" style="float:left; width:30%">
            <div style="float:left; width:80px">Following</div>
            <div style="margin-left: 80px">
                <a href="/follow">[Add Follow]</a>
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
            <div style="float:right; color:green; text-align:center; width:50%;">
                <form name="post_form" method="post" action="/post" onSubmit="return InputCheck(this)" autocomplete="off">
                    <p>Share with your friends:</p>
                    <input type="text" name="postcontent" style="width:80%; height:100px"><br>
                    <p><p></p><button type="submit" class="btn btn-default">post</button></p>
                </form>
            </div>
            <div id="messages" style="float:left; width:30%">
                {{range .Messages}}
                <p><font size="3">{{.Text}}</font></p>
                <p><font size="2">{{.Username}} posted at </font><font size="1">{{.DisplayTime}}</font></p><br />
                {{end}}
            </div>        
        </div> 
        <div style="width:100%; height:100px">
            <div style="float:left; width:50%">
                <h1>My Posts</h1>
            </div>
        </div>
        <div id="usermessages" style="float:left; width:30%">
            {{range .UserMessages}}
            <p><font size="3">{{.Text}}</font></p>
            <p><font size="2">You posted at </font><font size="1">{{.DisplayTime}}</font></p><br />
            {{end}}
        </div>
    </div>
</div>

<script language=JavaScript>
<!--
function InputCheck(post_form) { 
    if (post_form.postcontent.value == "") {
        alert("Write something to post!");  
        post_form.postcontent.focus();  
        return (false);
    }
} 
//-->
</script>

<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">