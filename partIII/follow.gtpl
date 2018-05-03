<!DOCTYPE html>
<html>
<head>
    <title></title>
</head>
<div class="row">
    <div class="container">
        <h1>Discover Page</h1>
        <form name="follow_form" method="post" action="/follow" onSubmit="return InputCheck(this)" autocomplete="off">
            <label for="username">Username </label>
            <input class="form-control" style="width:20%" type="text" id="username" name="username" value=""><br />
            <button type="submit" class="btn btn-default">Search</button>    
        </form>
        {{if .}}
        <div style="width:100%; height:100px">
            <div style="float:left; width:50%">
                <h1>{{.Username}}'s Post</h1>
            </div>
        </div>
        <div id="messages" style="float:left; width:30%">
            {{range .Messages}}
            <p><font size="3">{{.Text}}</font></p>
            <p><font size="2">{{.Username}} posted at </font><font size="1">{{.DisplayTime}}</font></p><br />
            {{end}}
        </div>
        {{if .Status}}
        <form name="add_form" method="post" action="/add" autocomplete="off">
            <button type="submit" class="btn btn-default">Add</button>    
        </form>
        {{end}}
        {{end}}
    </div>
</div>

<script language=JavaScript>
<!--
function InputCheck(follow_form) {
    if (follow_form.username.value == "") {
        alert("Please enter username!");  
        login_form.username.focus();  
        return (false);
    }
}
//-->
</script>
<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">
